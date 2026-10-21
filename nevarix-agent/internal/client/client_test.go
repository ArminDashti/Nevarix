package client

import (
	"encoding/json"
	"testing"
	"time"
)

func TestSnapshotToProto(t *testing.T) {
	body, _ := json.Marshal(wireSnapshot{
		AgentID: "a1", Hostname: "h1", CPUPercent: 12.5,
		Cores: []wireCore{{CoreIndex: 0, UsagePercent: 10}},
		Disks: []wireDisk{{MountPoint: "/", TotalBytes: 100}},
		System: json.RawMessage(`{"schemaVersion":2}`),
		CollectedAt: time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC),
	})
	pb, err := snapshotToProto(body)
	if err != nil {
		t.Fatal(err)
	}
	if pb.GetAgentId() != "a1" || pb.GetSystemJson() != `{"schemaVersion":2}` || pb.GetCpuPercent() != 12.5 {
		t.Fatalf("proto=%v", pb)
	}
	if pb.GetCollectedAt() == nil {
		t.Fatal("missing timestamp")
	}
}
