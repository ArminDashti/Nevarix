package inventory

import "testing"

func TestRound1(t *testing.T) {
	if got := round1(42.46); got != 42.5 {
		t.Fatalf("round1(42.46) = %v, want 42.5", got)
	}
}

func TestCollectFillsSlicesNotNil(t *testing.T) {
	s := New().Collect()
	if s.Temperatures == nil || s.Firewall == nil {
		t.Fatal("expected non-nil slices so JSON encodes [] not null")
	}
}
