package paths

import (
	"path/filepath"
	"testing"
)

func TestDataDir(t *testing.T) {
	t.Setenv("NEVARIX_AGENT_DATA_DIR", "")
	if got := filepath.Base(DataDir()); got != Name {
		t.Fatalf("default data dir base = %q, want %q", got, Name)
	}
	t.Setenv("NEVARIX_AGENT_DATA_DIR", `D:\custom\agent`)
	if got := DataDir(); got != `D:\custom\agent` {
		t.Fatalf("override ignored: got %q", got)
	}
}

