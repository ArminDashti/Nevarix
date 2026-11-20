package store

import (
	"testing"
	"time"

	"github.com/ArminDashti/nevarix-core/core/model"
)

func TestFillAgentGrid(t *testing.T) {
	now := time.Date(2026, 11, 19, 12, 0, 0, 0, time.UTC)
	a := model.AgentSummary{LastSeenAt: now.Add(-2 * time.Minute)}
	disks := []byte(`[{"mountPoint":"C:","usedPercent":40},{"mountPoint":"D:"}]`)
	system := []byte(`{"os":"windows","platform":"Microsoft Windows 11 Pro","platformVersion":"10.0.26340","network":{"nics":[{"addrs":["127.0.0.1/8","10.0.0.5/24"]}]}}`)
	fillAgentGrid(&a, disks, system, now)
	if a.Status != "unstable" {
		t.Fatalf("status %q", a.Status)
	}
	if a.Disks != "C: 40%, D: 0%" || a.OS != "windows" || a.IP != "10.0.0.5" {
		t.Fatalf("grid %+v", a)
	}
	if a.Description != "Microsoft Windows 11 Pro 10.0.26340" {
		t.Fatalf("desc %q", a.Description)
	}
	fresh := model.AgentSummary{LastSeenAt: now.Add(-10 * time.Second)}
	fillAgentGrid(&fresh, []byte(`[]`), []byte(`{}`), now)
	if fresh.Status != "connected" {
		t.Fatalf("fresh %q", fresh.Status)
	}
	old := model.AgentSummary{LastSeenAt: now.Add(-time.Hour)}
	fillAgentGrid(&old, nil, nil, now)
	if old.Status != "disconnected" {
		t.Fatalf("old %q", old.Status)
	}
}
