package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/ArminDashti/nevarix-core/internal/paths"
	"github.com/ArminDashti/nevarix-core/internal/service"
	"github.com/ArminDashti/nevarix-core/internal/state"
)

const scanTimeout = 5 * time.Second

var stdin = bufio.NewReader(os.Stdin)

const usage = `Usage: nvx-core <command>

  status
  service status | start | stop | restart
  help
  version
  install
  uninstall
  rebuild
  rebuild remove-data
  agent list
  agent add --alias=<> --address=<host:port> --protocol=<restful|grpc> --secret-key=<>
  agent remove <alias>
  agent enable <alias>
  agent disable <alias>
  agent scan
  secret-key show
  management list
  stats
  stats reset
  app status
  app version
  app run [--port=N]
  app stop
  app port [--port=N]
  app update
  app rebuild [--mode=docker|native] [--remove-data=no|yes] [-y]
  app uninstall [-y]`

func cmdHelp(_ []string) error {
	fmt.Println(usage)
	return nil
}

func cmdVersion(_ []string) error {
	fmt.Println(version)
	return nil
}

func cmdStatus(_ []string) error {
	if out, err := service.Status(); err != nil {
		fmt.Println("service: not available:", err)
	} else {
		fmt.Println(out)
	}
	st, err := state.Load(state.Path())
	if err != nil {
		return err
	}
	fmt.Printf("agents: %d total, %d enabled\n", len(st.Agents), st.EnabledCount())
	return nil
}

func cmdServiceStatus(_ []string) error {
	out, err := service.Status()
	fmt.Println(out)
	return err
}

func cmdServiceStart(_ []string) error   { return service.Start() }
func cmdServiceStop(_ []string) error    { return service.Stop() }
func cmdServiceRestart(_ []string) error { return service.Restart() }

func cmdRebuild(_ []string) error {
	if !confirm("Rebuild core binary into bin/? [y/N]: ") {
		return nil
	}
	return build()
}

func cmdRebuildRemoveData(_ []string) error {
	if !confirm("Delete all core data and rebuild? [y/N]: ") {
		return nil
	}
	if err := os.RemoveAll(paths.DataDir()); err != nil {
		return err
	}
	return build()
}

func build() error {
	ldflags := "-X main.version=" + version
	return runTool("go", "build", "-ldflags", ldflags, "-o", "bin/nvx-core.exe", "./cmd/nvx-core")
}

func cmdUninstall(_ []string) error {
	if !confirm("Stop and delete the service and local data? [y/N]: ") {
		return nil
	}
	_ = service.Stop() // ponytail: ignores stop error so uninstall still removes a stopped service
	if err := service.Delete(); err != nil {
		return err
	}
	// Removes the core data directory (agents, stats). The binary is left in place: it is the running file.
	return os.RemoveAll(paths.DataDir())
}

func cmdAgentList(_ []string) error {
	st, err := state.Load(state.Path())
	if err != nil {
		return err
	}
	for _, alias := range st.Aliases() {
		a := st.Agents[alias]
		fmt.Printf("%s\t%s\t%s\tenabled=%t\n", a.Alias, a.Address, a.Protocol, a.Enabled)
	}
	return nil
}

func cmdAgentAdd(args []string) error {
	fs := flag.NewFlagSet("agent add", flag.ContinueOnError)
	alias := fs.String("alias", "", "agent alias")
	address := fs.String("address", "", "agent host:port")
	protocol := fs.String("protocol", "", "restful or grpc")
	secretKey := fs.String("secret-key", "", "agent secret key")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *alias == "" {
		return errors.New("--alias is required")
	}
	addr, err := state.ParseAddress(*address)
	if err != nil {
		return fmt.Errorf("--address: %w", err)
	}
	if *protocol != "restful" && *protocol != "grpc" {
		return errors.New("--protocol must be restful or grpc")
	}
	if *secretKey == "" {
		return errors.New("--secret-key is required")
	}
	st, err := state.Load(state.Path())
	if err != nil {
		return err
	}
	if _, exists := st.Agents[*alias]; exists {
		return fmt.Errorf("alias %q already exists", *alias)
	}
	st.Agents[*alias] = state.Agent{
		Alias:     *alias,
		Address:   addr,
		Protocol:  *protocol,
		SecretKey: *secretKey,
		Enabled:   true,
	}
	return state.Save(state.Path(), st)
}

func cmdAgentRemove(args []string) error {
	st, err := state.Load(state.Path())
	if err != nil {
		return err
	}
	if _, ok := st.Agents[args[0]]; !ok {
		return fmt.Errorf("alias %q not found", args[0])
	}
	delete(st.Agents, args[0])
	return state.Save(state.Path(), st)
}

func cmdAgentEnable(args []string) error  { return setAgentEnabled(args[0], true) }
func cmdAgentDisable(args []string) error { return setAgentEnabled(args[0], false) }

func setAgentEnabled(alias string, enabled bool) error {
	st, err := state.Load(state.Path())
	if err != nil {
		return err
	}
	a, ok := st.Agents[alias]
	if !ok {
		return fmt.Errorf("alias %q not found", alias)
	}
	a.Enabled = enabled
	st.Agents[alias] = a
	return state.Save(state.Path(), st)
}

func cmdAgentScan(_ []string) error {
	st, err := state.Load(state.Path())
	if err != nil {
		return err
	}
	now := time.Now()
	scanned, failed := 0, 0
	for _, alias := range st.Aliases() {
		a := st.Agents[alias]
		if !a.Enabled {
			continue
		}
		scanned++
		start := time.Now()
		probeErr := probe(a)
		a.LastScanAt, a.LastScanOK = now, probeErr == nil
		st.Agents[alias] = a
		if probeErr != nil {
			failed++
			fmt.Printf("%s\t%s\tFAIL\t%v\n", a.Alias, a.Address, probeErr)
			continue
		}
		fmt.Printf("%s\t%s\tOK\t%s\n", a.Alias, a.Address, time.Since(start).Round(time.Millisecond))
	}
	if scanned == 0 {
		fmt.Println("no enabled agents")
		return nil
	}
	st.Stats.ScanRuns++
	st.Stats.ScanFailures += failed
	st.Stats.LastScanAt = now
	return state.Save(state.Path(), st)
}

// probe checks reachability: TCP dial for grpc, any HTTP response for restful.
func probe(a state.Agent) error {
	if a.Protocol == "grpc" {
		conn, err := net.DialTimeout("tcp", a.Address, scanTimeout)
		if err != nil {
			return err
		}
		return conn.Close()
	}
	client := http.Client{Timeout: scanTimeout}
	resp, err := client.Get("http://" + a.Address + "/")
	if err != nil {
		return err
	}
	return resp.Body.Close()
}

func cmdSecretKeyShow(_ []string) error {
	key := strings.TrimSpace(os.Getenv("NEVARIX_CORE_SECRET_KEY"))
	if key == "" {
		return errors.New("NEVARIX_CORE_SECRET_KEY is not set")
	}
	fmt.Println(key)
	return nil
}

func cmdManagementList(_ []string) error {
	st, err := state.Load(state.Path())
	if err != nil {
		return err
	}
	for _, alias := range st.Aliases() {
		a := st.Agents[alias]
		last := "never"
		if !a.LastScanAt.IsZero() {
			last = fmt.Sprintf("%s ok=%t", a.LastScanAt.Format(time.RFC3339), a.LastScanOK)
		}
		fmt.Printf("%s\t%s\tenabled=%t\tlast-scan=%s\n", a.Alias, a.Protocol, a.Enabled, last)
	}
	return nil
}

func cmdStats(_ []string) error {
	st, err := state.Load(state.Path())
	if err != nil {
		return err
	}
	last := "never"
	if !st.Stats.LastScanAt.IsZero() {
		last = st.Stats.LastScanAt.Format(time.RFC3339)
	}
	fmt.Printf("agents:        %d total, %d enabled\n", len(st.Agents), st.EnabledCount())
	fmt.Printf("scan runs:     %d\n", st.Stats.ScanRuns)
	fmt.Printf("scan failures: %d\n", st.Stats.ScanFailures)
	fmt.Printf("last scan:     %s\n", last)
	return nil
}

func cmdStatsReset(_ []string) error {
	st, err := state.Load(state.Path())
	if err != nil {
		return err
	}
	st.Stats = state.Stats{}
	return state.Save(state.Path(), st)
}

func confirm(prompt string) bool {
	fmt.Print(prompt)
	line, _ := stdin.ReadString('\n')
	return strings.EqualFold(strings.TrimSpace(line), "y")
}

// runTool runs an external tool in the current directory, streaming its output.
func runTool(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	return cmd.Run()
}
