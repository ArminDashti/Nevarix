package dockermetrics

import (
	"context"
	"strings"
	"testing"
)

func TestParseSize(t *testing.T) {
	if ParseSize("1.5GiB") != int64(1.5*(1<<30)) {
		t.Fatalf("gib = %d", ParseSize("1.5GiB"))
	}
	if ParseSize("2MB") != 2e6 {
		t.Fatalf("mb = %d", ParseSize("2MB"))
	}
}

func TestCollectUnavailable(t *testing.T) {
	run := func(ctx context.Context, args ...string) (string, error) {
		return "", errDocker
	}
	r := Collect(context.Background(), run, 10)
	if r.Available {
		t.Fatal("expected unavailable")
	}
}

func TestCollectMock(t *testing.T) {
	run := func(ctx context.Context, args ...string) (string, error) {
		switch args[0] {
		case "version":
			return "ok", nil
		case "info":
			return `{"ServerVersion":"27.0","NCPU":4,"Images":2}`, nil
		case "images":
			return `{"ID":"abc","Repository":"nginx","Tag":"latest","Size":"10MB"}`, nil
		case "volume":
			return `{"Name":"data","Driver":"local"}`, nil
		case "ps":
			return `{"ID":"c1","Names":"web","Image":"nginx","State":"running"}`, nil
		case "stats":
			return `{"CPUPerc":"1.5%","MemUsage":"10MiB / 1GiB","NetIO":"1kB / 2kB","BlockIO":"3MB / 4MB"}`, nil
		}
		return "", nil
	}
	r := Collect(context.Background(), run, 10)
	if !r.Available || r.Info.ServerVersion != "27.0" || len(r.Images) != 1 || len(r.Volumes) != 1 {
		t.Fatalf("report=%+v", r)
	}
	if r.Containers[0].CPUPercent != 1.5 {
		t.Fatalf("cpu=%v", r.Containers[0].CPUPercent)
	}
	if !strings.HasPrefix(r.Images[0].Tags[0], "nginx") {
		t.Fatalf("tag=%v", r.Images[0].Tags)
	}
}

type dockerErr string

func (e dockerErr) Error() string { return string(e) }

const errDocker dockerErr = "no docker"
