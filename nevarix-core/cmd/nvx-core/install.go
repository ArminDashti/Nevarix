package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ArminDashti/nevarix-core/internal/paths"
	"github.com/ArminDashti/nevarix-core/internal/service"
)

// cmdInstall copies the binary to the install dir, creates the data dir,
// stores the data dir as a machine env var, and registers the Windows service.
// Needs an elevated shell for Program Files, setx /M and sc create.
func cmdInstall(_ []string) error {
	installDir := promptPath("Install directory", paths.InstallDir())
	dataDir := promptPath("Data directory", paths.DataDir())

	exe, err := os.Executable()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(installDir, 0o755); err != nil {
		return fmt.Errorf("install dir (run as administrator?): %w", err)
	}
	bin := filepath.Join(installDir, "nvx-core.exe")
	raw, err := os.ReadFile(exe)
	if err != nil {
		return err
	}
	if err := os.WriteFile(bin, raw, 0o755); err != nil {
		return fmt.Errorf("copy binary (run as administrator?): %w", err)
	}
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		return err
	}
	if out, err := exec.Command("setx", "/M", "NEVARIX_CORE_DATA_DIR", dataDir).CombinedOutput(); err != nil {
		return fmt.Errorf("setx NEVARIX_CORE_DATA_DIR (run as administrator?): %w: %s", err, strings.TrimSpace(string(out)))
	}
	if err := service.Create(bin); err != nil {
		return err
	}
	fmt.Printf("installed %s to %s, data in %s\n", paths.Name, installDir, dataDir)
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
