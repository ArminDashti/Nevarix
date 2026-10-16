package main

import "testing"

func TestRouteWhitelist(t *testing.T) {
	ok := [][]string{
		{"status"},
		{"rebuild"},
		{"rebuild", "remove-data"},
		{"agent", "list"},
		{"agent", "add", "--alias=lab", "--address=10.0.0.5:8090", "--protocol=grpc", "--secret-key=x"},
		{"agent", "remove", "lab"},
		{"agent", "scan"},
		{"secret-key", "show"},
		{"management", "list"},
		{"stats", "reset"},
		{"install"},
		{"app", "status"},
		{"app", "run", "--port=8181"},
		{"app", "rebuild", "--remove-data=yes", "-y"},
	}
	for _, args := range ok {
		if _, _, found := route(args); !found {
			t.Fatalf("whitelisted command rejected: %v", args)
		}
	}
	bad := [][]string{
		{"agent"},
		{"agent", "remove"},
		{"agent", "remove", "lab", "extra"},
		{"status", "--verbose"},
		{"managment", "list"},
	}
	for _, args := range bad {
		if _, _, found := route(args); found {
			t.Fatalf("non-whitelisted command accepted: %v", args)
		}
	}
	if len(commands) != 29 {
		t.Fatalf("expected 29 commands, got %d", len(commands))
	}
}
