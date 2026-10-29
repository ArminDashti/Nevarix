//go:build linux

package paths

import (
	"path/filepath"
)

// InstallDir is where binaries are installed: /opt/NevarixAgent.
func InstallDir() string {
	return filepath.Join("/opt", Name)
}

func defaultDataDir() string {
	return filepath.Join("/var/lib", Name)
}
