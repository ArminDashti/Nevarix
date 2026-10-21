package store

import "testing"

func TestMetricDetailsJSON(t *testing.T) {
	raw := []byte(`{"storage":{"ssds":[{"id":"nvme0n1","kind":"ssd","readBps":1,"writeBps":2}],"hdds":[{"id":"sda","kind":"hdd"}]},"network":{"nics":[{"name":"eth0","rxBps":3,"txBps":4}]},"docker":{"containers":[{"id":"c1","name":"web","cpuPercent":1.5}]}}`)
	got := MetricDetailsJSON(raw)
	if string(got) == "{}" {
		t.Fatal(string(got))
	}
	if !containsAll(string(got), "nvme0n1", "eth0", "c1", "sda") {
		t.Fatalf("details=%s", got)
	}
}

func containsAll(s string, parts ...string) bool {
	for _, p := range parts {
		if !contains(s, p) {
			return false
		}
	}
	return true
}

func contains(s, p string) bool {
	return len(s) >= len(p) && (s == p || len(p) == 0 || (func() bool {
		for i := 0; i+len(p) <= len(s); i++ {
			if s[i:i+len(p)] == p {
				return true
			}
		}
		return false
	})())
}
