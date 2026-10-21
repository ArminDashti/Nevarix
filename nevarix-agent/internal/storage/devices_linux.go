//go:build linux

package storage

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/ArminDashti/nevarix-agent/internal/sysreport"
)

func listDevices() []sysreport.DiskDevice {
	ents, err := os.ReadDir("/sys/block")
	if err != nil {
		return nil
	}
	out := []sysreport.DiskDevice{}
	for _, e := range ents {
		name := e.Name()
		if SkipDevice(name) {
			continue
		}
		rot, _ := os.ReadFile(filepath.Join("/sys/block", name, "queue", "rotational"))
		model, _ := os.ReadFile(filepath.Join("/sys/block", name, "device", "model"))
		serial, _ := os.ReadFile(filepath.Join("/sys/block", name, "device", "serial"))
		sizeRaw, _ := os.ReadFile(filepath.Join("/sys/block", name, "size"))
		sectors, _ := strconv.ParseUint(strings.TrimSpace(string(sizeRaw)), 10, 64)
		kind := KindFromRotational(string(rot))
		if strings.HasPrefix(name, "nvme") {
			kind = "ssd"
		}
		out = append(out, sysreport.DiskDevice{
			ID:        name,
			Name:      name,
			Serial:    strings.TrimSpace(string(serial)),
			Model:     strings.TrimSpace(string(model)),
			Kind:      kind,
			SizeBytes: sectors * 512,
		})
	}
	return out
}
