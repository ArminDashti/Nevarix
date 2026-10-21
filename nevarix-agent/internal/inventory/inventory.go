package inventory

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/sensors"
)

type Motherboard struct {
	Vendor  string `json:"vendor"`
	Product string `json:"product"`
	Serial  string `json:"serial"`
}

type FirewallProfile struct {
	Name            string `json:"name"`
	Enabled         bool   `json:"enabled"`
	DefaultInbound  string `json:"defaultInbound"`
	DefaultOutbound string `json:"defaultOutbound"`
}

type Temperature struct {
	Sensor  string  `json:"sensor"`
	Celsius float64 `json:"celsius"`
}

type System struct {
	OS              string            `json:"os"`
	Platform        string            `json:"platform"`
	PlatformVersion string            `json:"platformVersion"`
	KernelVersion   string            `json:"kernelVersion"`
	Arch            string            `json:"arch"`
	Motherboard     Motherboard       `json:"motherboard"`
	Firewall        []FirewallProfile `json:"firewall"`
	Temperatures    []Temperature     `json:"temperatures"`
}

// ponytail: firewall shells out to PowerShell, so it is cached for a minute.
// Motherboard is static and read once. Linux firewall is not collected yet.
const firewallTTL = time.Minute

type Collector struct {
	mu         sync.Mutex
	board      Motherboard
	boardRead  bool
	firewall   []FirewallProfile
	firewallAt time.Time
}

func New() *Collector {
	return &Collector{}
}

func (c *Collector) Collect() System {
	s := System{Firewall: []FirewallProfile{}, Temperatures: temperatures()}
	if info, err := host.Info(); err == nil {
		s.OS = info.OS
		s.Platform = info.Platform
		s.PlatformVersion = info.PlatformVersion
		s.KernelVersion = info.KernelVersion
		s.Arch = info.KernelArch
	}
	s.Motherboard = c.motherboard()
	s.Firewall = c.firewallProfiles(time.Now())
	return s
}

func (c *Collector) motherboard() Motherboard {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.boardRead {
		c.board = readMotherboard()
		c.boardRead = true
	}
	return c.board
}

func (c *Collector) firewallProfiles(now time.Time) []FirewallProfile {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.firewall == nil || now.Sub(c.firewallAt) >= firewallTTL {
		c.firewall = readFirewall()
		c.firewallAt = now
	}
	return c.firewall
}

func readMotherboard() Motherboard {
	switch runtime.GOOS {
	case "windows":
		var rows []struct {
			Manufacturer string `json:"Manufacturer"`
			Product      string `json:"Product"`
			SerialNumber string `json:"SerialNumber"`
		}
		const script = `ConvertTo-Json -Compress -InputObject @(Get-CimInstance Win32_BaseBoard | Select-Object Manufacturer,Product,SerialNumber)`
		if err := PowershellJSON(script, &rows); err != nil || len(rows) == 0 {
			return Motherboard{}
		}
		return Motherboard{Vendor: rows[0].Manufacturer, Product: rows[0].Product, Serial: rows[0].SerialNumber}
	case "linux":
		return Motherboard{
			Vendor:  ReadDMI("board_vendor"),
			Product: ReadDMI("board_name"),
			Serial:  ReadDMI("board_serial"),
		}
	}
	return Motherboard{}
}

func ReadDMI(name string) string {
	raw, err := os.ReadFile("/sys/class/dmi/id/" + name)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(raw))
}

func readFirewall() []FirewallProfile {
	if runtime.GOOS != "windows" {
		return []FirewallProfile{}
	}
	// ToString() keeps enum values readable instead of numeric.
	const script = `ConvertTo-Json -Compress -InputObject @(Get-NetFirewallProfile | ForEach-Object {
		[pscustomobject]@{
			Name = $_.Name
			Enabled = $_.Enabled.ToString()
			DefaultInbound = $_.DefaultInboundAction.ToString()
			DefaultOutbound = $_.DefaultOutboundAction.ToString()
		}
	})`
	var rows []struct {
		Name            string `json:"Name"`
		Enabled         string `json:"Enabled"`
		DefaultInbound  string `json:"DefaultInbound"`
		DefaultOutbound string `json:"DefaultOutbound"`
	}
	if err := PowershellJSON(script, &rows); err != nil {
		return []FirewallProfile{}
	}
	out := make([]FirewallProfile, 0, len(rows))
	for _, r := range rows {
		out = append(out, FirewallProfile{
			Name:            r.Name,
			Enabled:         r.Enabled == "True",
			DefaultInbound:  r.DefaultInbound,
			DefaultOutbound: r.DefaultOutbound,
		})
	}
	return out
}

func temperatures() []Temperature {
	stats, err := sensors.SensorsTemperatures()
	out := []Temperature{}
	if err != nil {
		return out
	}
	for _, s := range stats {
		if s.Temperature <= 0 {
			continue
		}
		out = append(out, Temperature{Sensor: s.SensorKey, Celsius: round1(s.Temperature)})
	}
	return out
}

func PowershellJSON(script string, out any) error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	raw, err := exec.CommandContext(ctx, "powershell", "-NoProfile", "-NonInteractive", "-Command", script).Output()
	if err != nil {
		return err
	}
	return json.Unmarshal(raw, out)
}

func round1(v float64) float64 {
	return float64(int(v*10+0.5)) / 10
}
