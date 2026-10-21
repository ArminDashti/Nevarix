// Package sysreport is the v2 `system` object the agent embeds in each snapshot.
package sysreport

import (
	"time"

	"github.com/ArminDashti/nevarix-agent/internal/inventory"
)

const SchemaVersion = 2

// History keys: cpu, memory, nic:<name>, disk:<id>, docker:<containerId>
const (
	KeyCPU    = "cpu"
	KeyMemory = "memory"
	KeyNIC    = "nic:"
	KeyDisk   = "disk:"
	KeyDocker = "docker:"
)

type Sample struct {
	Timestamp    time.Time `json:"timestamp"`
	CPUPercent   float64   `json:"cpuPercent,omitempty"`
	MemPercent   float64   `json:"memPercent,omitempty"`
	MemBytes     uint64    `json:"memBytes,omitempty"`
	RxBps        float64   `json:"rxBps,omitempty"`
	TxBps        float64   `json:"txBps,omitempty"`
	ReadBps      float64   `json:"readBps,omitempty"`
	WriteBps     float64   `json:"writeBps,omitempty"`
	UsedPercent  float64   `json:"usedPercent,omitempty"`
	TemperatureC *float64  `json:"temperatureC,omitempty"`
}

type Report struct {
	SchemaVersion   int                      `json:"schemaVersion"`
	OS              string                   `json:"os"`
	Platform        string                   `json:"platform"`
	PlatformVersion string                   `json:"platformVersion"`
	KernelVersion   string                   `json:"kernelVersion"`
	Arch            string                   `json:"arch"`
	CPU             CPUReport                `json:"cpu"`
	Memory          MemoryReport             `json:"memory"`
	Storage         StorageReport            `json:"storage"`
	Motherboard     MotherboardReport        `json:"motherboard"`
	Network         NetworkReport            `json:"network"`
	Docker          DockerReport             `json:"docker"`
	Hardware        HardwareReport           `json:"hardware"`
	Firewall        []inventory.FirewallProfile `json:"firewall"`
	Temperatures    []inventory.Temperature  `json:"temperatures"`
}

type CPUReport struct {
	Info         CPUInfo   `json:"info"`
	Current      CPUCurrent `json:"current"`
	History      []Sample  `json:"history"`
	TemperatureC *float64  `json:"temperatureC"`
}

type CPUInfo struct {
	ModelName string  `json:"modelName"`
	Vendor    string  `json:"vendor"`
	Cores     int32   `json:"cores"`
	Logical   int     `json:"logical"`
	Mhz       float64 `json:"mhz"`
	CacheKB   int32   `json:"cacheKb"`
}

type CPUCurrent struct {
	Percent float64      `json:"percent"`
	Cores   []CoreSample `json:"cores"`
}

type CoreSample struct {
	CoreIndex    int     `json:"coreIndex"`
	UsagePercent float64 `json:"usagePercent"`
}

type MemoryReport struct {
	Info         MemoryInfo `json:"info"`
	Current      MemoryCurrent `json:"current"`
	History      []Sample   `json:"history"`
	TemperatureC *float64   `json:"temperatureC"`
}

type MemoryInfo struct {
	TotalBytes uint64 `json:"totalBytes"`
	Modules    []DIMM `json:"modules"`
}

type DIMM struct {
	Slot           string `json:"slot,omitempty"`
	Manufacturer   string `json:"manufacturer,omitempty"`
	PartNumber     string `json:"partNumber,omitempty"`
	Serial         string `json:"serial,omitempty"`
	SizeBytes      uint64 `json:"sizeBytes"`
	SpeedMHz       uint32 `json:"speedMhz,omitempty"`
}

type MemoryCurrent struct {
	TotalBytes     uint64  `json:"totalBytes"`
	UsedBytes      uint64  `json:"usedBytes"`
	AvailableBytes uint64  `json:"availableBytes"`
	UsedPercent    float64 `json:"usedPercent"`
}

type StorageReport struct {
	SSDs []DiskDevice `json:"ssds"`
	HDDs []DiskDevice `json:"hdds"`
}

type DiskDevice struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Serial       string   `json:"serial,omitempty"`
	Model        string   `json:"model,omitempty"`
	Kind         string   `json:"kind"`
	Bus          string   `json:"bus,omitempty"`
	SizeBytes    uint64   `json:"sizeBytes"`
	ReadBps      float64  `json:"readBps"`
	WriteBps     float64  `json:"writeBps"`
	TemperatureC *float64 `json:"temperatureC"`
	History      []Sample `json:"history"`
}

type MotherboardReport struct {
	Vendor       string   `json:"vendor"`
	Product      string   `json:"product"`
	Serial       string   `json:"serial"`
	TemperatureC *float64 `json:"temperatureC"`
}

type NetworkReport struct {
	NICs []NIC `json:"nics"`
}

type NIC struct {
	Name         string   `json:"name"`
	MAC          string   `json:"mac,omitempty"`
	MTU          int      `json:"mtu,omitempty"`
	Addrs        []string `json:"addrs"`
	RxBps        float64  `json:"rxBps"`
	TxBps        float64  `json:"txBps"`
	TemperatureC *float64 `json:"temperatureC"`
	History      []Sample `json:"history"`
}

type DockerReport struct {
	Available  bool              `json:"available"`
	Info       DockerInfo        `json:"info"`
	Containers []DockerContainer `json:"containers"`
	Images     []DockerImage     `json:"images"`
	Volumes    []DockerVolume    `json:"volumes"`
}

type DockerInfo struct {
	ServerVersion string `json:"serverVersion,omitempty"`
	OSType        string `json:"osType,omitempty"`
	Architecture  string `json:"architecture,omitempty"`
	NCPU          int    `json:"ncpu,omitempty"`
	MemTotal      int64  `json:"memTotal,omitempty"`
	Driver        string `json:"driver,omitempty"`
	Containers    int    `json:"containers,omitempty"`
	Images        int    `json:"images,omitempty"`
}

type DockerContainer struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Image        string   `json:"image"`
	State        string   `json:"state"`
	CPUPercent   float64  `json:"cpuPercent"`
	MemoryBytes  uint64   `json:"memoryBytes"`
	RxBps        float64  `json:"rxBps"`
	TxBps        float64  `json:"txBps"`
	ReadBps      float64  `json:"readBps"`
	WriteBps     float64  `json:"writeBps"`
	DiskBytes    uint64   `json:"diskBytes"`
	History      []Sample `json:"history"`
}

type DockerImage struct {
	ID        string    `json:"id"`
	Tags      []string  `json:"tags"`
	SizeBytes int64     `json:"sizeBytes"`
	CreatedAt time.Time `json:"createdAt,omitempty"`
}

type DockerVolume struct {
	Name       string `json:"name"`
	Driver     string `json:"driver,omitempty"`
	Mountpoint string `json:"mountpoint,omitempty"`
}

type HardwareReport struct {
	Full HardwareFull `json:"full"`
}

type HardwareFull struct {
	Product ProductInfo `json:"product"`
	BIOS    BIOSInfo    `json:"bios"`
	Chassis ChassisInfo `json:"chassis"`
	CPUs    []CPUInfo   `json:"cpus"`
	Memory  []DIMM      `json:"memory"`
	GPUs    []GPUInfo   `json:"gpus"`
}

type ProductInfo struct {
	Vendor string `json:"vendor,omitempty"`
	Name   string `json:"name,omitempty"`
	Serial string `json:"serial,omitempty"`
	SKU    string `json:"sku,omitempty"`
}

type BIOSInfo struct {
	Vendor  string `json:"vendor,omitempty"`
	Version string `json:"version,omitempty"`
	Date    string `json:"date,omitempty"`
}

type ChassisInfo struct {
	Vendor string `json:"vendor,omitempty"`
	Type   string `json:"type,omitempty"`
	Serial string `json:"serial,omitempty"`
}

type GPUInfo struct {
	Name     string `json:"name"`
	Vendor   string `json:"vendor,omitempty"`
	Driver   string `json:"driver,omitempty"`
	RAMBytes uint64 `json:"ramBytes,omitempty"`
}
