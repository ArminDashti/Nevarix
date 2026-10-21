//go:build windows

package hardware

import (
	"strings"

	"github.com/ArminDashti/nevarix-agent/internal/inventory"
	"github.com/ArminDashti/nevarix-agent/internal/sysreport"
)

func fillOS(h *sysreport.HardwareFull) {
	var cs []struct {
		Manufacturer string `json:"Manufacturer"`
		Model        string `json:"Model"`
	}
	_ = inventory.PowershellJSON(`ConvertTo-Json -Compress -InputObject @(Get-CimInstance Win32_ComputerSystem | Select-Object Manufacturer,Model)`, &cs)
	if len(cs) > 0 {
		h.Product.Vendor = cs[0].Manufacturer
		h.Product.Name = cs[0].Model
	}
	var bios []struct {
		Manufacturer string `json:"Manufacturer"`
		SMBIOSBIOSVersion string `json:"SMBIOSBIOSVersion"`
		ReleaseDate string `json:"ReleaseDate"`
		SerialNumber string `json:"SerialNumber"`
	}
	_ = inventory.PowershellJSON(`ConvertTo-Json -Compress -InputObject @(Get-CimInstance Win32_BIOS | Select-Object Manufacturer,SMBIOSBIOSVersion,ReleaseDate,SerialNumber)`, &bios)
	if len(bios) > 0 {
		h.BIOS = sysreport.BIOSInfo{Vendor: bios[0].Manufacturer, Version: bios[0].SMBIOSBIOSVersion, Date: bios[0].ReleaseDate}
		h.Product.Serial = bios[0].SerialNumber
	}
	var chassis []struct {
		Manufacturer string `json:"Manufacturer"`
		ChassisTypes []int  `json:"ChassisTypes"`
		SerialNumber string `json:"SerialNumber"`
	}
	_ = inventory.PowershellJSON(`ConvertTo-Json -Compress -InputObject @(Get-CimInstance Win32_SystemEnclosure | Select-Object Manufacturer,ChassisTypes,SerialNumber)`, &chassis)
	if len(chassis) > 0 {
		h.Chassis = sysreport.ChassisInfo{Vendor: chassis[0].Manufacturer, Serial: chassis[0].SerialNumber}
	}
	var dimms []struct {
		BankLabel    string  `json:"BankLabel"`
		Manufacturer string  `json:"Manufacturer"`
		PartNumber   string  `json:"PartNumber"`
		SerialNumber string  `json:"SerialNumber"`
		Capacity     float64 `json:"Capacity"`
		Speed        uint32  `json:"Speed"`
	}
	_ = inventory.PowershellJSON(`ConvertTo-Json -Compress -InputObject @(Get-CimInstance Win32_PhysicalMemory | Select-Object BankLabel,Manufacturer,PartNumber,SerialNumber,Capacity,Speed)`, &dimms)
	h.Memory = make([]sysreport.DIMM, 0, len(dimms))
	for _, d := range dimms {
		h.Memory = append(h.Memory, sysreport.DIMM{
			Slot:         d.BankLabel,
			Manufacturer: d.Manufacturer,
			PartNumber:   strings.TrimSpace(d.PartNumber),
			Serial:       d.SerialNumber,
			SizeBytes:    uint64(d.Capacity),
			SpeedMHz:     d.Speed,
		})
	}
	var gpus []struct {
		Name                 string  `json:"Name"`
		AdapterRAM           float64 `json:"AdapterRAM"`
		DriverVersion        string  `json:"DriverVersion"`
		AdapterCompatibility string  `json:"AdapterCompatibility"`
	}
	_ = inventory.PowershellJSON(`ConvertTo-Json -Compress -InputObject @(Get-CimInstance Win32_VideoController | Select-Object Name,AdapterRAM,DriverVersion,AdapterCompatibility)`, &gpus)
	h.GPUs = make([]sysreport.GPUInfo, 0, len(gpus))
	for _, g := range gpus {
		h.GPUs = append(h.GPUs, sysreport.GPUInfo{
			Name:     g.Name,
			Vendor:   g.AdapterCompatibility,
			Driver:   g.DriverVersion,
			RAMBytes: uint64(g.AdapterRAM),
		})
	}
}
