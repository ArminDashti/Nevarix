package sysreport

import (
	"strings"

	"github.com/ArminDashti/nevarix-agent/internal/inventory"
)

type TempBuckets struct {
	CPU         *float64
	Memory      *float64
	Motherboard *float64
	NIC         *float64
	Disk        map[string]float64
}

func BucketTemps(temps []inventory.Temperature) TempBuckets {
	b := TempBuckets{Disk: map[string]float64{}}
	for _, t := range temps {
		k := strings.ToLower(t.Sensor)
		c := t.Celsius
		switch {
		case containsAny(k, "dimm", "sodimm", "memory"):
			if b.Memory == nil {
				b.Memory = ptr(c)
			}
		case containsAny(k, "nvme", "ssd", "hdd", "drive", "disk", "smart"):
			b.Disk[k] = c
		case containsAny(k, "nic", "ethernet", "lan"):
			if b.NIC == nil {
				b.NIC = ptr(c)
			}
		case containsAny(k, "cpu", "coretemp", "k10temp", "zenpower", "package id", "tdie", "tctl"):
			if b.CPU == nil {
				b.CPU = ptr(c)
			}
		case containsAny(k, "motherboard", "acpitz", "pch", "chipset", "system", "ambient", "board"):
			if b.Motherboard == nil {
				b.Motherboard = ptr(c)
			}
		}
	}
	return b
}

func DiskTemp(buckets TempBuckets, id, model string) *float64 {
	id = strings.ToLower(id)
	model = strings.ToLower(model)
	for k, v := range buckets.Disk {
		if (id != "" && strings.Contains(k, id)) || (model != "" && strings.Contains(k, model)) {
			return ptr(v)
		}
	}
	if strings.Contains(id, "nvme") || strings.Contains(model, "nvme") {
		for k, v := range buckets.Disk {
			if strings.Contains(k, "nvme") {
				return ptr(v)
			}
		}
	}
	return nil
}

func containsAny(s string, parts ...string) bool {
	for _, p := range parts {
		if strings.Contains(s, p) {
			return true
		}
	}
	return false
}

func ptr(v float64) *float64 { return &v }
