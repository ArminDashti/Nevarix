//go:build linux

package service

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Name is the systemd unit name (without .service).
const Name = "nevarix-agent"

const (
	envDir   = "/etc/nevarix-agent"
	envFile  = envDir + "/environment"
	unitPath = "/etc/systemd/system/nevarix-agent.service"
	envScope = "run as root"
)

func Start() error   { return systemctl("start") }
func Stop() error    { return systemctl("stop") }
func Restart() error { return systemctl("restart") }

func Delete() error {
	_ = systemctl("disable", "--now")
	_ = os.Remove(unitPath)
	return reload()
}

func Status() (string, error) {
	out, err := exec.Command("systemctl", "status", Name, "--no-pager").CombinedOutput()
	return strings.TrimSpace(string(out)), err
}

// Create installs a systemd unit that runs daemonBin with env from envFile.
func Create(daemonBin, dataDir string) error {
	unit := fmt.Sprintf(`[Unit]
Description=Nevarix Agent
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
ExecStart=%s
EnvironmentFile=-%s
WorkingDirectory=%s
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
`, daemonBin, envFile, dataDir)
	if err := os.WriteFile(unitPath, []byte(unit), 0o644); err != nil {
		return fmt.Errorf("write %s (%s): %w", unitPath, envScope, err)
	}
	if err := reload(); err != nil {
		return err
	}
	if err := systemctl("enable", "--now"); err != nil {
		return err
	}
	fmt.Printf("enabled and started %s.service\n", Name)
	return nil
}

// PersistEnv writes /etc/nevarix-agent/environment for the systemd unit.
func PersistEnv(vars map[string]string) error {
	if err := os.MkdirAll(envDir, 0o755); err != nil {
		return fmt.Errorf("mkdir %s (%s): %w", envDir, envScope, err)
	}
	var b strings.Builder
	for key, val := range vars {
		fmt.Fprintf(&b, "%s=%s\n", key, val)
	}
	if err := os.WriteFile(envFile, []byte(b.String()), 0o644); err != nil {
		return fmt.Errorf("write %s (%s): %w", envFile, envScope, err)
	}
	return nil
}

func systemctl(args ...string) error {
	out, err := exec.Command("systemctl", append(args, Name)...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("systemctl %s %s (%s): %w: %s", strings.Join(args, " "), Name, envScope, err, strings.TrimSpace(string(out)))
	}
	if s := strings.TrimSpace(string(out)); s != "" {
		fmt.Println(s)
	}
	return nil
}

func reload() error {
	out, err := exec.Command("systemctl", "daemon-reload").CombinedOutput()
	if err != nil {
		return fmt.Errorf("systemctl daemon-reload (%s): %w: %s", envScope, err, strings.TrimSpace(string(out)))
	}
	return nil
}
