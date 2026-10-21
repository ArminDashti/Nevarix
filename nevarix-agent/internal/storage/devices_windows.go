//go:build windows

package storage

import (
	"github.com/ArminDashti/nevarix-agent/internal/inventory"
	"github.com/ArminDashti/nevarix-agent/internal/sysreport"
)

func listDevices() []sysreport.DiskDevice {
	var rows []struct {
		FriendlyName string  `json:"FriendlyName"`
		SerialNumber string  `json:"SerialNumber"`
		MediaType    string  `json:"MediaType"`
		Size         float64 `json:"Size"`
		BusType      string  `json:"BusType"`
		DeviceId     string  `json:"DeviceId"`
	}
	const script = `ConvertTo-Json -Compress -InputObject @(Get-PhysicalDisk | Select-Object FriendlyName,SerialNumber,MediaType,Size,BusType,DeviceId)`
	if err := inventory.PowershellJSON(script, &rows); err != nil {
		return listWin32Disks()
	}
	out := make([]sysreport.DiskDevice, 0, len(rows))
	for _, r := range rows {
		id := r.DeviceId
		if id == "" {
			id = r.SerialNumber
		}
		if id == "" {
			id = r.FriendlyName
		}
		out = append(out, sysreport.DiskDevice{
			ID:        id,
			Name:      r.FriendlyName,
			Serial:    r.SerialNumber,
			Model:     r.FriendlyName,
			Kind:      KindFromMedia(r.MediaType),
			Bus:       r.BusType,
			SizeBytes: uint64(r.Size),
		})
	}
	return out
}

func listWin32Disks() []sysreport.DiskDevice {
	var rows []struct {
		Model         string  `json:"Model"`
		SerialNumber  string  `json:"SerialNumber"`
		Size          float64 `json:"Size"`
		InterfaceType string  `json:"InterfaceType"`
		MediaType     string  `json:"MediaType"`
		Index         int     `json:"Index"`
	}
	const script = `ConvertTo-Json -Compress -InputObject @(Get-CimInstance Win32_DiskDrive | Select-Object Model,SerialNumber,Size,InterfaceType,MediaType,Index)`
	if err := inventory.PowershellJSON(script, &rows); err != nil {
		return nil
	}
	out := make([]sysreport.DiskDevice, 0, len(rows))
	for _, r := range rows {
		out = append(out, sysreport.DiskDevice{
			ID:        r.SerialNumber,
			Name:      r.Model,
			Serial:    r.SerialNumber,
			Model:     r.Model,
			Kind:      KindFromMedia(r.MediaType),
			Bus:       r.InterfaceType,
			SizeBytes: uint64(r.Size),
		})
	}
	return out
}
