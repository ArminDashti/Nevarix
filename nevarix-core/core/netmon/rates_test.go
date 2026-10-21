package netmon

import (
	"testing"
	"time"
)

func TestCounterBps(t *testing.T) {
	bps := CounterBps(1000, 2000, time.Second)
	if bps != 8000 {
		t.Fatalf("expected 8000 bps, got %v", bps)
	}
}
