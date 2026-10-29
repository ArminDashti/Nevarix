//go:build linux

package paths

import (
	"path/filepath"
	"testing"
)

func TestInstallDir(t *testing.T) {
	if got := InstallDir(); got != filepath.Join("/opt", Name) {
		t.Fatalf("install dir = %q", got)
	}
}
