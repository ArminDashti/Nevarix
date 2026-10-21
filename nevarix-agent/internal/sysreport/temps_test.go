package sysreport

import (
	"testing"

	"github.com/ArminDashti/nevarix-agent/internal/inventory"
)

func TestBucketTemps(t *testing.T) {
	b := BucketTemps([]inventory.Temperature{
		{Sensor: "coretemp_package_id_0", Celsius: 55},
		{Sensor: "DIMM0", Celsius: 41},
		{Sensor: "acpitz", Celsius: 38},
		{Sensor: "nvme_composite", Celsius: 47},
	})
	if b.CPU == nil || *b.CPU != 55 {
		t.Fatalf("cpu temp = %v", b.CPU)
	}
	if b.Memory == nil || *b.Memory != 41 {
		t.Fatalf("memory temp = %v", b.Memory)
	}
	if b.Motherboard == nil || *b.Motherboard != 38 {
		t.Fatalf("board temp = %v", b.Motherboard)
	}
	if DiskTemp(b, "nvme0n1", "") == nil {
		t.Fatal("expected nvme disk temp")
	}
}
