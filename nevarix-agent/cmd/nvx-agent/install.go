package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/ArminDashti/nevarix-agent/internal/paths"
	"github.com/ArminDashti/nevarix-agent/internal/service"
)

// cmdInstall copies nvx-agent and nevarix-agent into the install dir, creates the data dir,
// persists install env vars, and registers the OS service (Windows service or systemd).
func cmdInstall(_ []string) error {
	installDir := promptPath("Install directory", paths.InstallDir())
	dataDir := promptPath("Data directory", paths.DataDir())

	cliSrc, err := os.Executable()
	if err != nil {
		return err
	}
	daemonSrc := filepath.Join(filepath.Dir(cliSrc), binName("nevarix-agent"))
	if _, err := os.Stat(daemonSrc); err != nil {
		return fmt.Errorf("daemon binary %s not found next to nvx-agent; run rebuild first", daemonSrc)
	}

	if err := os.MkdirAll(installDir, 0o755); err != nil {
		return fmt.Errorf("install dir (%s): %w", elevateHint(), err)
	}
	cliDst := filepath.Join(installDir, binName("nvx-agent"))
	daemonDst := filepath.Join(installDir, binName("nevarix-agent"))
	if err := copyFile(cliSrc, cliDst); err != nil {
		return fmt.Errorf("copy CLI (%s): %w", elevateHint(), err)
	}
	if err := copyFile(daemonSrc, daemonDst); err != nil {
		return fmt.Errorf("copy daemon (%s): %w", elevateHint(), err)
	}
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		return err
	}
	env := map[string]string{
		"NEVARIX_AGENT_DATA_DIR": dataDir,
		"NEVARIX_OUTBOX_PATH":    filepath.Join(dataDir, "agent-outbox.db"),
	}
	if err := service.PersistEnv(env); err != nil {
		return err
	}
	if err := service.Create(daemonDst, dataDir); err != nil {
		return err
	}
	fmt.Printf("installed %s to %s, data in %s\n", paths.Name, installDir, dataDir)
	fmt.Println("set NEVARIX_AGENT_TOKEN in the service environment and restart the agent")
	return nil
}

// promptPath asks for a path and returns def when the answer is empty.
func promptPath(label, def string) string {
	fmt.Printf("%s [%s]: ", label, def)
	line, _ := stdin.ReadString('\n')
	if p := strings.TrimSpace(line); p != "" {
		return p
	}
	return def
}

func copyFile(src, dst string) error {
	raw, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, raw, 0o755)
}

func binName(base string) string {
	if runtime.GOOS == "windows" {
		return base + ".exe"
	}
	return base
}

func elevateHint() string {
	if runtime.GOOS == "windows" {
		return "run as administrator"
	}
	if runtime.GOOS == "linux" {
		return "run as root"
	}
	return "elevated permissions required"
}
