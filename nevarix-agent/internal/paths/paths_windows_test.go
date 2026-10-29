//go:build windows

package paths

import (
	"path/filepath"
	"testing"
)

func TestInstallDir(t *testing.T) {
	t.Setenv("ProgramFiles", `C:\Program Files`)
	if got := InstallDir(); got != filepath.Join(`C:\Program Files`, Name) {
		t.Fatalf("install dir = %q", got)
	}
}
