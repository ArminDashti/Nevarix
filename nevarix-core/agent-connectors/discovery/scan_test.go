package discovery

import "testing"

func TestHostsSkipsNetworkAndBroadcast(t *testing.T) {
	hosts, err := Hosts("10.0.0.0/30")
	if err != nil {
		t.Fatal(err)
	}
	if len(hosts) != 2 || hosts[0] != "10.0.0.1" || hosts[1] != "10.0.0.2" {
		t.Fatalf("unexpected hosts: %v", hosts)
	}
}

func TestHostsRejectsLargeRange(t *testing.T) {
	if _, err := Hosts("10.0.0.0/8"); err == nil {
		t.Fatal("expected error for /8")
	}
}
