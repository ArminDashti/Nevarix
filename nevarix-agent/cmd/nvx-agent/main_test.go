package main

import "testing"

func TestRouteWhitelist(t *testing.T) {
	ok := [][]string{
		{"status"},
		{"core", "connect", "lab"},
		{"core", "address", "set", "lab:10.0.0.5:8090"},
		{"core", "secret-key", "show"},
		{"install"},
	}
	for _, args := range ok {
		if _, _, found := route(args); !found {
			t.Fatalf("whitelisted command rejected: %v", args)
		}
	}
	bad := [][]string{
		{"core"},
		{"core", "connect"},
		{"core", "address", "set"},
		{"core", "address", "set", "lab", "extra"},
		{"status", "--verbose"},
	}
	for _, args := range bad {
		if _, _, found := route(args); found {
			t.Fatalf("non-whitelisted command accepted: %v", args)
		}
	}
	if len(commands) != 22 {
		t.Fatalf("expected 22 commands, got %d", len(commands))
	}
}
