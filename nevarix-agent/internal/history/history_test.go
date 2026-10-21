package history

import (
	"testing"
	"time"

	"github.com/ArminDashti/nevarix-agent/internal/sysreport"
)

func TestRingKeepsLastN(t *testing.T) {
	s := New(10)
	now := time.Unix(0, 0).UTC()
	for i := 0; i < 12; i++ {
		s.Push(sysreport.KeyCPU, sysreport.Sample{Timestamp: now.Add(time.Duration(i) * time.Second), CPUPercent: float64(i)})
	}
	got := s.Get(sysreport.KeyCPU)
	if len(got) != 10 {
		t.Fatalf("len=%d want 10", len(got))
	}
	if got[0].CPUPercent != 2 || got[9].CPUPercent != 11 {
		t.Fatalf("got first=%v last=%v", got[0].CPUPercent, got[9].CPUPercent)
	}
}
