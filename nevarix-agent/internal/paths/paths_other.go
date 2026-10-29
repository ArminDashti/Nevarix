//go:build !windows && !linux

package paths

import (
	"os"
	"path/filepath"
)

func InstallDir() string {
	return filepath.Join("/usr/local", Name)
}

func defaultDataDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".local", "share", Name)
}
