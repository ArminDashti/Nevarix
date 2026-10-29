//go:build windows

package paths

import (
	"os"
	"path/filepath"
)

// InstallDir is where binaries are installed: %ProgramFiles%\NevarixAgent.
func InstallDir() string {
	return filepath.Join(os.Getenv("ProgramFiles"), Name)
}

func defaultDataDir() string {
	base, _ := os.UserCacheDir()
	return filepath.Join(base, Name)
}
