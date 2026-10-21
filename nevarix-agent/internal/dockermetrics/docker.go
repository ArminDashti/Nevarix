package dockermetrics

import (
	"bytes"
	"context"
	"encoding/json"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/ArminDashti/nevarix-agent/internal/sysreport"
)

type Runner func(ctx context.Context, args ...string) (string, error)

func DefaultRunner() Runner {
	return func(ctx context.Context, args ...string) (string, error) {
		cmd := exec.CommandContext(ctx, "docker", args...)
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		if err := cmd.Run(); err != nil {
			return "", err
		}
		return stdout.String(), nil
	}
}

func Collect(ctx context.Context, run Runner, maxContainers int) sysreport.DockerReport {
	out := sysreport.DockerReport{
		Containers: []sysreport.DockerContainer{},
		Images:     []sysreport.DockerImage{},
		Volumes:    []sysreport.DockerVolume{},
	}
	if run == nil {
		run = DefaultRunner()
	}
	if maxContainers < 1 {
		maxContainers = 50
	}
	if _, err := run(ctx, "version"); err != nil {
		return out
	}
	out.Available = true
	out.Info = dockerInfo(ctx, run)
	out.Images = listImages(ctx, run)
	out.Volumes = listVolumes(ctx, run)
	out.Containers = listContainers(ctx, run, maxContainers)
	return out
}

func dockerInfo(ctx context.Context, run Runner) sysreport.DockerInfo {
	raw, err := run(ctx, "info", "--format", "{{json .}}")
	if err != nil {
		return sysreport.DockerInfo{}
	}
	var row struct {
		ServerVersion string `json:"ServerVersion"`
		OSType        string `json:"OSType"`
		Architecture  string `json:"Architecture"`
		NCPU          int    `json:"NCPU"`
		MemTotal      int64  `json:"MemTotal"`
		Driver        string `json:"Driver"`
		Containers    int    `json:"Containers"`
		Images        int    `json:"Images"`
	}
	_ = json.Unmarshal([]byte(strings.TrimSpace(raw)), &row)
	return sysreport.DockerInfo{
		ServerVersion: row.ServerVersion,
		OSType:        row.OSType,
		Architecture:  row.Architecture,
		NCPU:          row.NCPU,
		MemTotal:      row.MemTotal,
		Driver:        row.Driver,
		Containers:    row.Containers,
		Images:        row.Images,
	}
}

func listImages(ctx context.Context, run Runner) []sysreport.DockerImage {
	raw, err := run(ctx, "images", "--format", "{{json .}}")
	if err != nil {
		return []sysreport.DockerImage{}
	}
	out := []sysreport.DockerImage{}
	for _, line := range splitLines(raw) {
		var row struct {
			ID         string `json:"ID"`
			Repository string `json:"Repository"`
			Tag        string `json:"Tag"`
			Size       string `json:"Size"`
			CreatedAt  string `json:"CreatedAt"`
		}
		if err := json.Unmarshal([]byte(line), &row); err != nil {
			continue
		}
		tag := row.Repository
		if row.Tag != "" {
			tag = row.Repository + ":" + row.Tag
		}
		out = append(out, sysreport.DockerImage{
			ID:        row.ID,
			Tags:      []string{tag},
			SizeBytes: ParseSize(row.Size),
			CreatedAt: parseTime(row.CreatedAt),
		})
	}
	return out
}

func listVolumes(ctx context.Context, run Runner) []sysreport.DockerVolume {
	raw, err := run(ctx, "volume", "ls", "--format", "{{json .}}")
	if err != nil {
		return []sysreport.DockerVolume{}
	}
	out := []sysreport.DockerVolume{}
	for _, line := range splitLines(raw) {
		var row struct {
			Name       string `json:"Name"`
			Driver     string `json:"Driver"`
			Mountpoint string `json:"Mountpoint"`
		}
		if err := json.Unmarshal([]byte(line), &row); err != nil {
			continue
		}
		out = append(out, sysreport.DockerVolume{Name: row.Name, Driver: row.Driver, Mountpoint: row.Mountpoint})
	}
	return out
}

func listContainers(ctx context.Context, run Runner, max int) []sysreport.DockerContainer {
	raw, err := run(ctx, "ps", "-a", "--format", "{{json .}}")
	if err != nil {
		return []sysreport.DockerContainer{}
	}
	lines := splitLines(raw)
	out := []sysreport.DockerContainer{}
	n := 0
	for _, line := range lines {
		if n >= max {
			break
		}
		var row struct {
			ID    string `json:"ID"`
			Names string `json:"Names"`
			Image string `json:"Image"`
			State string `json:"State"`
		}
		if err := json.Unmarshal([]byte(line), &row); err != nil {
			continue
		}
		c := sysreport.DockerContainer{
			ID:    row.ID,
			Name:  row.Names,
			Image: row.Image,
			State: row.State,
		}
		if strings.EqualFold(row.State, "running") {
			fillStats(ctx, run, &c)
		}
		out = append(out, c)
		n++
	}
	return out
}

func fillStats(ctx context.Context, run Runner, c *sysreport.DockerContainer) {
	raw, err := run(ctx, "stats", "--no-stream", "--format", "{{json .}}", c.ID)
	if err != nil {
		return
	}
	var row struct {
		CPUPerc  string `json:"CPUPerc"`
		MemUsage string `json:"MemUsage"`
		NetIO    string `json:"NetIO"`
		BlockIO  string `json:"BlockIO"`
	}
	if err := json.Unmarshal([]byte(strings.TrimSpace(raw)), &row); err != nil {
		return
	}
	c.CPUPercent, _ = strconv.ParseFloat(strings.TrimSuffix(strings.TrimSpace(row.CPUPerc), "%"), 64)
	c.MemoryBytes = uint64(ParseSize(strings.Split(row.MemUsage, "/")[0]))
	rx, tx := splitPair(row.NetIO)
	c.RxBps = float64(ParseSize(rx))
	c.TxBps = float64(ParseSize(tx))
	rd, wr := splitPair(row.BlockIO)
	c.ReadBps = float64(ParseSize(rd))
	c.WriteBps = float64(ParseSize(wr))
	c.DiskBytes = uint64(ParseSize(rd) + ParseSize(wr))
}

func splitPair(s string) (string, string) {
	parts := strings.Split(s, "/")
	if len(parts) < 2 {
		return strings.TrimSpace(s), ""
	}
	return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
}

func ParseSize(s string) int64 {
	s = strings.TrimSpace(strings.ToUpper(s))
	mult := float64(1)
	switch {
	case strings.HasSuffix(s, "GIB"):
		mult = 1 << 30
		s = strings.TrimSuffix(s, "GIB")
	case strings.HasSuffix(s, "MIB"):
		mult = 1 << 20
		s = strings.TrimSuffix(s, "MIB")
	case strings.HasSuffix(s, "KIB"):
		mult = 1 << 10
		s = strings.TrimSuffix(s, "KIB")
	case strings.HasSuffix(s, "GB"):
		mult = 1e9
		s = strings.TrimSuffix(s, "GB")
	case strings.HasSuffix(s, "MB"):
		mult = 1e6
		s = strings.TrimSuffix(s, "MB")
	case strings.HasSuffix(s, "KB"):
		mult = 1e3
		s = strings.TrimSuffix(s, "KB")
	case strings.HasSuffix(s, "B"):
		s = strings.TrimSuffix(s, "B")
	}
	v, _ := strconv.ParseFloat(strings.TrimSpace(s), 64)
	return int64(v * mult)
}

func splitLines(s string) []string {
	raw := strings.Split(s, "\n")
	out := make([]string, 0, len(raw))
	for _, line := range raw {
		line = strings.TrimSpace(line)
		if line != "" {
			out = append(out, line)
		}
	}
	return out
}

func parseTime(s string) time.Time {
	for _, layout := range []string{time.RFC3339, "2006-01-02 15:04:05 -0700 MST"} {
		if t, err := time.Parse(layout, s); err == nil {
			return t.UTC()
		}
	}
	return time.Time{}
}
