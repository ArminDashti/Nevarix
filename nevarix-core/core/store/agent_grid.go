package store

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/ArminDashti/nevarix-core/core/model"
)

// shortcut: status uses last-seen age (90s / 5m), upgrade when core stores a real heartbeat class.
func fillAgentGrid(a *model.AgentSummary, disksRaw, systemRaw []byte, now time.Time) {
	age := now.Sub(a.LastSeenAt)
	switch {
	case age <= 90*time.Second:
		a.Status = "connected"
	case age <= 5*time.Minute:
		a.Status = "unstable"
	default:
		a.Status = "disconnected"
	}
	a.Disks = diskSummary(disksRaw)
	osName, ip, desc := systemBits(systemRaw)
	a.OS = osName
	a.IP = ip
	a.Description = desc
}

func diskSummary(raw []byte) string {
	var disks []model.DiskSample
	if err := json.Unmarshal(raw, &disks); err != nil || len(disks) == 0 {
		return ""
	}
	parts := make([]string, 0, len(disks))
	for _, d := range disks {
		name := d.MountPoint
		if name == "" {
			name = "?"
		}
		parts = append(parts, name+" "+strconv.FormatFloat(d.UsedPercent, 'f', 0, 64)+"%")
	}
	return strings.Join(parts, ", ")
}

func systemBits(raw []byte) (osName, ip, desc string) {
	var sys struct {
		OS              string `json:"os"`
		Platform        string `json:"platform"`
		PlatformVersion string `json:"platformVersion"`
		Network         struct {
			NICs []struct {
				Addrs []string `json:"addrs"`
			} `json:"nics"`
		} `json:"network"`
	}
	if err := json.Unmarshal(raw, &sys); err != nil {
		return "", "", ""
	}
	osName = sys.OS
	if osName == "" {
		osName = sys.Platform
	}
	desc = strings.TrimSpace(sys.Platform + " " + sys.PlatformVersion)
	ip = firstIP(sys.Network.NICs)
	return osName, ip, desc
}

func firstIP(nics []struct {
	Addrs []string `json:"addrs"`
}) string {
	for _, nic := range nics {
		for _, addr := range nic.Addrs {
			host := addr
			if i := strings.IndexByte(host, '/'); i >= 0 {
				host = host[:i]
			}
			if host == "" || host == "127.0.0.1" || host == "::1" {
				continue
			}
			return host
		}
	}
	return ""
}
