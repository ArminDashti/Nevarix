package state

import (
	"path/filepath"
	"testing"
)

func TestSaveLoadRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "core-state.json")
	st, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	st.Agents["lab"] = Agent{Alias: "lab", Address: "10.0.0.5:8090", Protocol: "grpc", SecretKey: "k", Enabled: true}
	st.Stats.ScanRuns = 3
	if err := Save(path, st); err != nil {
		t.Fatal(err)
	}
	got, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if a := got.Agents["lab"]; a.Address != "10.0.0.5:8090" || !a.Enabled || got.Stats.ScanRuns != 3 {
		t.Fatalf("round trip lost data: %+v %+v", a, got.Stats)
	}
}

func TestParseAddress(t *testing.T) {
	if got, err := ParseAddress("10.0.0.5:8090"); err != nil || got != "10.0.0.5:8090" {
		t.Fatalf("valid address rejected: %q %v", got, err)
	}
	for _, bad := range []string{"10.0.0.5", "10.0.0.5:0", "10.0.0.5:70000", ":8090"} {
		if _, err := ParseAddress(bad); err == nil {
			t.Fatalf("invalid address accepted: %s", bad)
		}
	}
}
