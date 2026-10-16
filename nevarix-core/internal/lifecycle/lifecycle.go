// Package lifecycle runs one app as a Docker Compose project or as a native detached process.
// The same files live in nevarix-core and nevarix-agent; only the App value differs.
package lifecycle

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

const logLimit = 10 << 20

// App describes one app. Each repo defines its own value.
type App struct {
	Name        string                  // short name: state folder and compose project suffix
	CLI         string                  // command users run, used in hints
	Binary      string                  // built binary name without .exe
	Pkg         string                  // Go package to build
	Image       string                  // docker image tag
	Dockerfile  string                  // dockerfile in the repo root
	Compose     string                  // compose file in the repo root
	DefaultPort int                     // 0 = no HTTP port
	PortEnv     string                  // compose variable that sets the published port
	URLPath     string                  // path probed after start; empty = no probe
	NativeEnv   func(port int) []string // extra env for native runs; nil = none
}

// State is persisted per app in the app's state folder.
type State struct {
	Installed bool   `json:"installed"`
	Mode      string `json:"mode"`
	Port      int    `json:"port,omitempty"`
	Version   string `json:"version,omitempty"`
}

type appPaths struct {
	base, install, data, logs, pid, staging string
}

// Staging lives outside the app base so wiping the app never kills an in-progress build.
func (a App) paths() (appPaths, error) {
	root, err := lifecycleRoot()
	if err != nil {
		return appPaths{}, err
	}
	base := filepath.Join(root, a.Name)
	return appPaths{
		base:    base,
		install: filepath.Join(base, "app"),
		data:    filepath.Join(base, "data"),
		logs:    filepath.Join(base, "logs"),
		pid:     filepath.Join(base, "run.pid"),
		staging: filepath.Join(root, ".staging", a.Name),
	}, nil
}

func lifecycleRoot() (string, error) {
	if runtime.GOOS == "windows" {
		if dir := os.Getenv("LOCALAPPDATA"); dir != "" {
			return filepath.Join(dir, "nevarix"), nil
		}
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".local", "share", "nevarix"), nil
}

// ponytail: root is the current directory. Upgrade path: walk up to go.mod.
func (a App) checkRoot() error {
	if _, err := os.Stat(a.Compose); err != nil {
		return fmt.Errorf("run %s from the %s repo root: %w", a.CLI, a.Name, err)
	}
	return nil
}

func (a App) loadState() (State, error) {
	p, err := a.paths()
	if err != nil {
		return State{}, err
	}
	s := State{Mode: "docker", Port: a.DefaultPort}
	b, err := os.ReadFile(filepath.Join(p.base, "config.json"))
	if errors.Is(err, fs.ErrNotExist) {
		return s, nil
	}
	if err != nil {
		return s, err
	}
	if err := json.Unmarshal(b, &s); err != nil {
		return s, fmt.Errorf("%s state: %w", a.Name, err)
	}
	if s.Mode == "" {
		s.Mode = "docker"
	}
	if s.Port == 0 {
		s.Port = a.DefaultPort
	}
	return s, nil
}

func (a App) saveState(s State) error {
	p, err := a.paths()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(p.base, 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(p.base, "config.json"), b, 0o600)
}

func (a App) requireInstalled() (State, error) {
	s, err := a.loadState()
	if err != nil {
		return s, err
	}
	if !s.Installed {
		return s, fmt.Errorf("%s is not installed; run: %s app rebuild", a.Name, a.CLI)
	}
	return s, nil
}

func (a App) url(port int) string {
	if port == 0 || a.URLPath == "" {
		return ""
	}
	return "http://127.0.0.1:" + strconv.Itoa(port) + a.URLPath
}

func (a App) sourceVersion() string {
	if out, ok := capture(nil, "git", "describe", "--tags", "--always", "--dirty"); ok {
		return out
	}
	return "unknown"
}

// Status prints mode, version, port, running state, and URL.
func (a App) Status(_ []string) error {
	if err := a.checkRoot(); err != nil {
		return err
	}
	s, err := a.loadState()
	if err != nil {
		return err
	}
	if !s.Installed {
		fmt.Printf("%s: not installed\n", a.Name)
		return nil
	}
	running := a.isRunning(s)
	url := a.url(s.Port)
	reach := ""
	if url != "" && running {
		reach = " (NOT reachable)"
		if probe(url) {
			reach = " (reachable)"
		}
	}
	path := a.Name
	if s.Mode == "native" {
		p, _ := a.paths()
		path = p.install
	} else if wd, err := os.Getwd(); err == nil {
		path = wd
	}
	fmt.Println(a.Name)
	fmt.Printf("  mode:     %s\n", s.Mode)
	fmt.Printf("  version:  %s\n", dashIf(s.Version))
	fmt.Printf("  port:     %s\n", portText(s.Port))
	fmt.Printf("  running:  %s\n", yesNo(running))
	fmt.Printf("  url:      %s%s\n", dashIf(url), reach)
	fmt.Printf("  paths:    %s\n", path)
	return nil
}

// Version prints the installed and source versions; errors when they differ.
func (a App) Version(_ []string) error {
	s, err := a.loadState()
	if err != nil {
		return err
	}
	src := a.sourceVersion()
	installed := "not installed"
	if s.Version != "" {
		installed = s.Version
	}
	fmt.Printf("%s installed: %s\n", a.Name, installed)
	fmt.Printf("%s source:    %s\n", a.Name, src)
	if s.Version != "" && s.Version != src {
		return errors.New(`MISMATCH: source changed since install; run "rebuild" or "update"`)
	}
	return nil
}

// Run starts the app. --port overrides and saves the port.
func (a App) Run(args []string) error {
	if err := a.checkRoot(); err != nil {
		return err
	}
	port, err := portFlag(args)
	if err != nil {
		return err
	}
	s, err := a.requireInstalled()
	if err != nil {
		return err
	}
	p := s.Port
	if port != 0 {
		p = port
	}
	if a.isRunning(s) {
		fmt.Printf("%s already running\n", a.Name)
		return a.Status(nil)
	}
	if err := a.start(s, p); err != nil {
		return err
	}
	if p != s.Port {
		s.Port = p
		if err := a.saveState(s); err != nil {
			return err
		}
	}
	return a.Status(nil)
}

// Stop stops the app. Idempotent.
func (a App) Stop(_ []string) error {
	if err := a.checkRoot(); err != nil {
		return err
	}
	s, err := a.loadState()
	if err != nil {
		return err
	}
	if !s.Installed {
		fmt.Printf("%s is not installed; nothing to stop\n", a.Name)
		return nil
	}
	if err := a.stop(s); err != nil {
		return err
	}
	fmt.Printf("%s stopped\n", a.Name)
	return nil
}

// Port shows the port, or sets it with --port and restarts the app if it runs.
func (a App) Port(args []string) error {
	if err := a.checkRoot(); err != nil {
		return err
	}
	port, err := portFlag(args)
	if err != nil {
		return err
	}
	s, err := a.requireInstalled()
	if err != nil {
		return err
	}
	if port == 0 {
		fmt.Printf("%s port: %s\n", a.Name, portText(s.Port))
		return nil
	}
	running := a.isRunning(s)
	if running {
		if err := a.stop(s); err != nil {
			return err
		}
	}
	s.Port = port
	if err := a.saveState(s); err != nil {
		return err
	}
	if running {
		if err := a.start(s, port); err != nil {
			return err
		}
	}
	return a.Status(nil)
}

// Update pulls the repo (fast-forward only), then rebuilds and keeps data.
func (a App) Update(_ []string) error {
	if err := a.checkRoot(); err != nil {
		return err
	}
	s, err := a.requireInstalled()
	if err != nil {
		return err
	}
	if err := run("git pull --ff-only ("+a.Name+")", nil, "git", "pull", "--ff-only"); err != nil {
		return err
	}
	return a.rebuild(s, "no", true, "")
}

// Rebuild rebuilds the app in docker or native mode. --remove-data=yes wipes data and config.
func (a App) Rebuild(args []string) error {
	if err := a.checkRoot(); err != nil {
		return err
	}
	flags := flag.NewFlagSet("rebuild", flag.ContinueOnError)
	mode := flags.String("mode", "", "docker or native (default: keep current)")
	removeData := flags.String("remove-data", "no", "yes wipes data and config")
	yes := flags.Bool("y", false, "skip confirmation")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if *removeData != "no" && *removeData != "yes" {
		return fmt.Errorf("--remove-data must be no or yes")
	}
	if *mode != "" && *mode != "docker" && *mode != "native" {
		return fmt.Errorf("--mode must be docker or native")
	}
	s, err := a.loadState()
	if err != nil {
		return err
	}
	return a.rebuild(s, *removeData, *yes, *mode)
}

func (a App) rebuild(s State, removeData string, yes bool, mode string) error {
	next := s.Mode
	if mode != "" {
		next = mode
	}
	wasRunning := s.Installed && a.isRunning(s)
	if removeData == "yes" && !confirm(fmt.Sprintf("Wipe ALL data and config for %s: %s? [y/N]: ", a.Name, a.wipeTargets(s)), yes) {
		fmt.Println("aborted")
		return nil
	}
	fmt.Printf("rebuild %s (mode %s, remove-data %s)\n", a.Name, next, removeData)

	if s.Installed {
		if err := a.stop(s); err != nil {
			return err
		}
	}
	p, err := a.paths()
	if err != nil {
		return err
	}
	if next == "docker" {
		if s.Installed && s.Mode == "docker" {
			if err := a.dockerDown(removeData == "yes"); err != nil {
				return err
			}
		}
		if err := run("docker build "+a.Name, nil, "docker", "build", "--no-cache", "--pull", "-f", a.Dockerfile, "-t", a.Image, "."); err != nil {
			return err
		}
	} else {
		if err := a.nativeBuild(p.staging); err != nil {
			return err
		}
		if removeData == "yes" {
			if err := os.RemoveAll(p.base); err != nil {
				return err
			}
		}
		if err := swap(p); err != nil {
			return err
		}
	}

	base := s
	if removeData == "yes" {
		base = State{}
	}
	ns := State{Installed: true, Mode: next, Port: base.Port, Version: a.sourceVersion()}
	if ns.Port == 0 {
		ns.Port = a.DefaultPort
	}
	if err := a.saveState(ns); err != nil {
		return err
	}
	if wasRunning {
		if err := a.start(ns, ns.Port); err != nil {
			return err
		}
	}
	return a.Status(nil)
}

// Uninstall removes the app, its data, and its state. Asks unless -y.
func (a App) Uninstall(args []string) error {
	if err := a.checkRoot(); err != nil {
		return err
	}
	flags := flag.NewFlagSet("uninstall", flag.ContinueOnError)
	yes := flags.Bool("y", false, "skip confirmation")
	if err := flags.Parse(args); err != nil {
		return err
	}
	s, err := a.loadState()
	if err != nil {
		return err
	}
	if !s.Installed {
		fmt.Printf("%s is not installed\n", a.Name)
		return nil
	}
	if !confirm(fmt.Sprintf("Remove %s: %s? [y/N]: ", a.Name, a.wipeTargets(s)), *yes) {
		fmt.Println("aborted")
		return nil
	}
	if s.Mode == "docker" {
		if err := a.dockerDown(true); err != nil {
			return err
		}
	} else if err := a.nativeStop(); err != nil {
		return err
	}
	p, err := a.paths()
	if err != nil {
		return err
	}
	if err := os.RemoveAll(p.base); err != nil {
		return err
	}
	fmt.Printf("%s removed\n", a.Name)
	return nil
}

func (a App) wipeTargets(s State) string {
	if s.Mode == "docker" {
		return fmt.Sprintf("containers, networks, volumes of project nevarix-%s, image %s, and config entry", a.Name, a.Image)
	}
	return fmt.Sprintf("app, data, logs of %s, and config entry", a.Name)
}

func (a App) isRunning(s State) bool {
	if s.Mode == "docker" {
		return a.dockerRunning(s.Port)
	}
	return a.nativeRunning()
}

func (a App) start(s State, port int) error {
	if port != 0 && !portFree(port) {
		return fmt.Errorf("port %d is already in use", port)
	}
	if s.Mode == "docker" {
		if err := a.ensureNetwork(); err != nil {
			return err
		}
		if err := a.compose(port, "up", "-d"); err != nil {
			return err
		}
	} else if err := a.nativeStart(port); err != nil {
		return err
	}
	if url := a.url(port); url != "" && !waitUntil(func() bool { return probe(url) }, 30) {
		return fmt.Errorf("%s not reachable after 30s", url)
	}
	return nil
}

func (a App) stop(s State) error {
	if s.Mode == "docker" {
		if err := a.compose(s.Port, "stop"); err != nil {
			return err
		}
	} else if err := a.nativeStop(); err != nil {
		return err
	}
	if a.isRunning(s) {
		return fmt.Errorf("%s still running after stop", a.Name)
	}
	if s.Port != 0 && !waitUntil(func() bool { return portFree(s.Port) }, 30) {
		return fmt.Errorf("port %d still busy after stop", s.Port)
	}
	return nil
}

func (a App) readPid() (int, error) {
	p, err := a.paths()
	if err != nil {
		return 0, err
	}
	b, err := os.ReadFile(p.pid)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(strings.TrimSpace(string(b)))
}

func (a App) nativeRunning() bool {
	pid, err := a.readPid()
	return err == nil && isAlive(pid)
}

func (a App) nativeStop() error {
	if pid, err := a.readPid(); err == nil && isAlive(pid) {
		killTree(pid)
	}
	p, err := a.paths()
	if err != nil {
		return err
	}
	if err := os.Remove(p.pid); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	return nil
}

func (a App) nativeStart(port int) error {
	p, err := a.paths()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(p.logs, 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll(p.data, 0o755); err != nil {
		return err
	}
	logPath := filepath.Join(p.logs, a.Name+".log")
	if info, err := os.Stat(logPath); err == nil && info.Size() > logLimit {
		_ = os.Rename(logPath, logPath+".1")
	}
	out, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer out.Close()

	cmd := exec.Command(filepath.Join(p.install, "bin", exeName(a.Binary)))
	cmd.Dir = p.install
	cmd.Stdout, cmd.Stderr = out, out
	cmd.Env = append(os.Environ(), "APP_DATA_DIR="+p.data)
	if a.NativeEnv != nil {
		cmd.Env = append(cmd.Env, a.NativeEnv(port)...)
	}
	detach(cmd)
	if err := cmd.Start(); err != nil {
		return err
	}
	if err := os.WriteFile(p.pid, []byte(strconv.Itoa(cmd.Process.Pid)), 0o644); err != nil {
		return err
	}
	return cmd.Process.Release()
}

func (a App) nativeBuild(staging string) error {
	if err := os.RemoveAll(staging); err != nil {
		return err
	}
	bin := filepath.Join(staging, "bin")
	if err := os.MkdirAll(bin, 0o755); err != nil {
		return err
	}
	return run("go build "+a.Name, []string{"CGO_ENABLED=0"}, "go", "build", "-o", filepath.Join(bin, exeName(a.Binary)), a.Pkg)
}

// swap moves the staged build into place only after the build succeeded, so a failed build keeps the old app.
func swap(p appPaths) error {
	if err := os.MkdirAll(filepath.Dir(p.install), 0o755); err != nil {
		return err
	}
	if err := os.RemoveAll(p.install); err != nil {
		return err
	}
	return os.Rename(p.staging, p.install)
}

func (a App) composeBase() []string {
	return []string{"compose", "-p", "nevarix-" + a.Name, "-f", a.Compose}
}

func (a App) composeEnv(port int) []string {
	env := []string{"IMAGE_TAG=" + a.Image}
	if a.PortEnv != "" && port != 0 {
		env = append(env, a.PortEnv+"="+strconv.Itoa(port))
	}
	return env
}

func (a App) compose(port int, args ...string) error {
	return run("docker compose "+strings.Join(args, " ")+" ("+a.Name+")", a.composeEnv(port), "docker", append(a.composeBase(), args...)...)
}

func (a App) dockerRunning(port int) bool {
	out, ok := capture(a.composeEnv(port), "docker", append(a.composeBase(), "ps", "-q", "--status", "running")...)
	return ok && out != ""
}

func (a App) dockerDown(volumes bool) error {
	args := []string{"down", "--remove-orphans"}
	if volumes {
		args = append(args, "-v", "--rmi", "local")
	}
	return a.compose(0, args...)
}

func (a App) ensureNetwork() error {
	net := os.Getenv("DOCKER_NETWORK")
	if net == "" {
		net = "nevarix-net"
	}
	if _, ok := capture(nil, "docker", "network", "inspect", net); ok {
		return nil
	}
	return run("create network "+net, nil, "docker", "network", "create", net)
}

// portFlag parses --port. 0 means not set.
func portFlag(args []string) (int, error) {
	flags := flag.NewFlagSet("port", flag.ContinueOnError)
	port := flags.Int("port", 0, "published port (1-65535)")
	if err := flags.Parse(args); err != nil {
		return 0, err
	}
	if flags.NArg() != 0 {
		return 0, fmt.Errorf("unexpected arguments: %s", strings.Join(flags.Args(), " "))
	}
	if *port != 0 {
		if err := validPort(*port); err != nil {
			return 0, err
		}
	}
	return *port, nil
}

func validPort(p int) error {
	if p < 1 || p > 65535 {
		return fmt.Errorf("invalid port: %d", p)
	}
	return nil
}

func exeName(binary string) string {
	if runtime.GOOS == "windows" {
		return binary + ".exe"
	}
	return binary
}

func confirm(prompt string, yes bool) bool {
	if yes {
		return true
	}
	fmt.Print(prompt)
	line, _ := bufio.NewReader(os.Stdin).ReadString('\n')
	ans := strings.ToLower(strings.TrimSpace(line))
	return ans == "y" || ans == "yes"
}

func dashIf(v string) string {
	if v == "" {
		return "-"
	}
	return v
}

func portText(p int) string {
	if p == 0 {
		return "-"
	}
	return strconv.Itoa(p)
}

func yesNo(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}
