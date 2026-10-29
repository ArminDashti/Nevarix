package paths

import (
	"os"
	"strings"
)

// Name is the service name and the folder name under the install and data roots.
const Name = "NevarixAgent"

// DataDir is where state lives: NEVARIX_AGENT_DATA_DIR, else the OS default.
func DataDir() string {
	if p := strings.TrimSpace(os.Getenv("NEVARIX_AGENT_DATA_DIR")); p != "" {
		return p
	}
	return defaultDataDir()
}
