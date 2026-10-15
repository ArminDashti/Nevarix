package core

import "testing"

func TestParseEntry(t *testing.T) {
	alias, a, err := ParseEntry("lab:10.0.0.5:8090")
	if err != nil || alias != "lab" || a.Address != "10.0.0.5" || a.Port != 8090 {
		t.Fatalf("valid entry rejected or misparsed: %v %v %v", alias, a, err)
	}
	for _, bad := range []string{"lab", ":10.0.0.5:8090", "lab:10.0.0.5", "lab:10.0.0.5:0", "lab:10.0.0.5:70000"} {
		if _, _, err := ParseEntry(bad); err == nil {
			t.Fatalf("expected error for %q", bad)
		}
	}
}
