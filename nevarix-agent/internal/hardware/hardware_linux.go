//go:build linux

package hardware

import (
	"os"
	"strings"

	"github.com/ArminDashti/nevarix-agent/internal/inventory"
	"github.com/ArminDashti/nevarix-agent/internal/sysreport"
)

func fillOS(h *sysreport.HardwareFull) {
	h.Product = sysreport.ProductInfo{
		Vendor: inventory.ReadDMI("sys_vendor"),
		Name:   inventory.ReadDMI("product_name"),
		Serial: inventory.ReadDMI("product_serial"),
		SKU:    inventory.ReadDMI("product_sku"),
	}
	h.BIOS = sysreport.BIOSInfo{
		Vendor:  inventory.ReadDMI("bios_vendor"),
		Version: inventory.ReadDMI("bios_version"),
		Date:    inventory.ReadDMI("bios_date"),
	}
	h.Chassis = sysreport.ChassisInfo{
		Vendor: inventory.ReadDMI("chassis_vendor"),
		Type:   inventory.ReadDMI("chassis_type"),
		Serial: inventory.ReadDMI("chassis_serial"),
	}
	h.GPUs = linuxGPUs()
}

func linuxGPUs() []sysreport.GPUInfo {
	ents, err := os.ReadDir("/sys/class/drm")
	if err != nil {
		return []sysreport.GPUInfo{}
	}
	out := []sysreport.GPUInfo{}
	for _, e := range ents {
		name := e.Name()
		if !strings.HasPrefix(name, "card") || strings.Contains(name, "-") {
			continue
		}
		dev := "/sys/class/drm/" + name + "/device"
		vendor, _ := os.ReadFile(dev + "/vendor")
		device, _ := os.ReadFile(dev + "/device")
		out = append(out, sysreport.GPUInfo{
			Name:   name,
			Vendor: strings.TrimSpace(string(vendor)) + " " + strings.TrimSpace(string(device)),
		})
	}
	return out
}
