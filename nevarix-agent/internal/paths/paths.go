package paths

import (
	"os"
	"path/filepath"
	"strings"
)

// Name is the Windows service name and the folder name used under Program Files and LocalAppData.
const Name = "NevarixAgent"

// InstallDir is where the binary is installed: %ProgramFiles%\NevarixAgent.
func InstallDir() string {
	return filepath.Join(os.Getenv("ProgramFiles"), Name)
}

// DataDir is where state lives: NEVARIX_AGENT_DATA_DIR, else %LOCALAPPDATA%\NevarixAgent.
func DataDir() string {
	if p := strings.TrimSpace(os.Getenv("NEVARIX_AGENT_DATA_DIR")); p != "" {
		return p
	}
	// ponytail: UserCacheDir on Windows is %LOCALAPPDATA%; the error case falls back to a relative path.
	base, _ := os.UserCacheDir()
	return filepath.Join(base, Name)
}
