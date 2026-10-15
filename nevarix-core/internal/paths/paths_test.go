package paths

import (
	"path/filepath"
	"testing"
)

func TestDataDir(t *testing.T) {
	t.Setenv("NEVARIX_CORE_DATA_DIR", "")
	if got := filepath.Base(DataDir()); got != Name {
		t.Fatalf("default data dir base = %q, want %q", got, Name)
	}
	t.Setenv("NEVARIX_CORE_DATA_DIR", `D:\custom\core`)
	if got := DataDir(); got != `D:\custom\core` {
		t.Fatalf("override ignored: got %q", got)
	}
}

func TestInstallDir(t *testing.T) {
	t.Setenv("ProgramFiles", `C:\Program Files`)
	if got := InstallDir(); got != filepath.Join(`C:\Program Files`, Name) {
		t.Fatalf("install dir = %q", got)
	}
}
