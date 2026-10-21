package collector

import (
	"context"
	"time"

	"github.com/ArminDashti/nevarix-agent/internal/dockermetrics"
	"github.com/ArminDashti/nevarix-agent/internal/hardware"
	"github.com/ArminDashti/nevarix-agent/internal/history"
	"github.com/ArminDashti/nevarix-agent/internal/inventory"
	"github.com/ArminDashti/nevarix-agent/internal/nics"
	"github.com/ArminDashti/nevarix-agent/internal/storage"
	"github.com/ArminDashti/nevarix-agent/internal/sysreport"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/mem"
)

type CoreSample = sysreport.CoreSample

type DiskSample struct {
	MountPoint  string  `json:"mountPoint"`
	Fstype      string  `json:"fstype,omitempty"`
	TotalBytes  uint64  `json:"totalBytes"`
	UsedBytes   uint64  `json:"usedBytes"`
	FreeBytes   uint64  `json:"freeBytes"`
	UsedPercent float64 `json:"usedPercent"`
}

type Snapshot struct {
	AgentID              string           `json:"agentId"`
	Hostname             string           `json:"hostname"`
	CPUPercent           float64          `json:"cpuPercent"`
	Cores                []CoreSample     `json:"cores"`
	MemoryTotalBytes     uint64           `json:"memoryTotalBytes"`
	MemoryUsedBytes      uint64           `json:"memoryUsedBytes"`
	MemoryAvailableBytes uint64           `json:"memoryAvailableBytes"`
	MemoryUsedPercent    float64          `json:"memoryUsedPercent"`
	Disks                []DiskSample     `json:"disks"`
	BandwidthRxBps       float64          `json:"bandwidthRxBps"`
	BandwidthTxBps       float64          `json:"bandwidthTxBps"`
	UptimeSeconds        uint64           `json:"uptimeSeconds"`
	System               sysreport.Report `json:"system"`
	CollectedAt          time.Time        `json:"collectedAt"`
}

type Collector struct {
	agentID    string
	hostname   string
	dockerMax  int
	inv        *inventory.Collector
	nics       *nics.Collector
	storage    *storage.Collector
	hw         *hardware.Collector
	hist       *history.Store
}

func New(agentID, hostname string, historyPoints, dockerMax int) *Collector {
	return &Collector{
		agentID:   agentID,
		hostname:  hostname,
		dockerMax: dockerMax,
		inv:       inventory.New(),
		nics:      nics.New(),
		storage:   storage.New(),
		hw:        hardware.New(),
		hist:      history.New(historyPoints),
	}
}

func (c *Collector) Collect() (Snapshot, error) {
	now := time.Now().UTC()
	total, err := cpu.Percent(250*time.Millisecond, false)
	if err != nil {
		return Snapshot{}, err
	}
	perCore, err := cpu.Percent(0, true)
	if err != nil {
		return Snapshot{}, err
	}
	cores := make([]CoreSample, 0, len(perCore))
	for i, v := range perCore {
		cores = append(cores, CoreSample{CoreIndex: i, UsagePercent: round2(v)})
	}

	vm, err := mem.VirtualMemory()
	if err != nil {
		return Snapshot{}, err
	}

	parts, err := disk.Partitions(false)
	if err != nil {
		return Snapshot{}, err
	}
	disks := make([]DiskSample, 0, len(parts))
	seen := map[string]struct{}{}
	for _, p := range parts {
		if _, ok := seen[p.Mountpoint]; ok {
			continue
		}
		seen[p.Mountpoint] = struct{}{}
		u, err := disk.Usage(p.Mountpoint)
		if err != nil {
			continue
		}
		disks = append(disks, DiskSample{
			MountPoint:  p.Mountpoint,
			Fstype:      p.Fstype,
			TotalBytes:  u.Total,
			UsedBytes:   u.Used,
			FreeBytes:   u.Free,
			UsedPercent: round2(u.UsedPercent),
		})
	}

	uptime, _ := host.Uptime()
	cpuPct := 0.0
	if len(total) > 0 {
		cpuPct = round2(total[0])
	}

	base := c.inv.Collect()
	temps := sysreport.BucketTemps(base.Temperatures)
	nicList, rx, tx := c.nics.Collect(now)
	store := c.storage.Collect(now, temps)
	hw := c.hw.Collect()
	dkr := dockermetrics.Collect(context.Background(), nil, c.dockerMax)

	c.hist.Push(sysreport.KeyCPU, sysreport.Sample{Timestamp: now, CPUPercent: cpuPct, TemperatureC: temps.CPU})
	c.hist.Push(sysreport.KeyMemory, sysreport.Sample{Timestamp: now, MemPercent: round2(vm.UsedPercent), MemBytes: vm.Used, TemperatureC: temps.Memory})
	for i := range nicList {
		n := &nicList[i]
		n.TemperatureC = temps.NIC
		c.hist.Push(sysreport.KeyNIC+n.Name, sysreport.Sample{Timestamp: now, RxBps: n.RxBps, TxBps: n.TxBps, TemperatureC: n.TemperatureC})
		n.History = c.hist.Get(sysreport.KeyNIC + n.Name)
	}
	attachDiskHistory(c.hist, now, &store)
	for i := range dkr.Containers {
		ctr := &dkr.Containers[i]
		key := sysreport.KeyDocker + ctr.ID
		c.hist.Push(key, sysreport.Sample{
			Timestamp: now, CPUPercent: ctr.CPUPercent, MemBytes: ctr.MemoryBytes,
			RxBps: ctr.RxBps, TxBps: ctr.TxBps, ReadBps: ctr.ReadBps, WriteBps: ctr.WriteBps,
		})
		ctr.History = c.hist.Get(key)
	}

	cpuInfo := firstCPU(hw.Full.CPUs)
	logical, _ := cpu.Counts(true)
	cpuInfo.Logical = logical

	report := sysreport.Report{
		SchemaVersion:   sysreport.SchemaVersion,
		OS:              base.OS,
		Platform:        base.Platform,
		PlatformVersion: base.PlatformVersion,
		KernelVersion:   base.KernelVersion,
		Arch:            base.Arch,
		CPU: sysreport.CPUReport{
			Info:         cpuInfo,
			Current:      sysreport.CPUCurrent{Percent: cpuPct, Cores: cores},
			History:      c.hist.Get(sysreport.KeyCPU),
			TemperatureC: temps.CPU,
		},
		Memory: sysreport.MemoryReport{
			Info: sysreport.MemoryInfo{TotalBytes: vm.Total, Modules: hw.Full.Memory},
			Current: sysreport.MemoryCurrent{
				TotalBytes: vm.Total, UsedBytes: vm.Used, AvailableBytes: vm.Available, UsedPercent: round2(vm.UsedPercent),
			},
			History:      c.hist.Get(sysreport.KeyMemory),
			TemperatureC: temps.Memory,
		},
		Storage: store,
		Motherboard: sysreport.MotherboardReport{
			Vendor: base.Motherboard.Vendor, Product: base.Motherboard.Product, Serial: base.Motherboard.Serial,
			TemperatureC: temps.Motherboard,
		},
		Network:      sysreport.NetworkReport{NICs: nicList},
		Docker:       dkr,
		Hardware:     hw,
		Firewall:     base.Firewall,
		Temperatures: base.Temperatures,
	}

	return Snapshot{
		AgentID:              c.agentID,
		Hostname:             c.hostname,
		CPUPercent:           cpuPct,
		Cores:                cores,
		MemoryTotalBytes:     vm.Total,
		MemoryUsedBytes:      vm.Used,
		MemoryAvailableBytes: vm.Available,
		MemoryUsedPercent:    round2(vm.UsedPercent),
		Disks:                disks,
		BandwidthRxBps:       rx,
		BandwidthTxBps:       tx,
		UptimeSeconds:        uptime,
		System:               report,
		CollectedAt:          now,
	}, nil
}

func attachDiskHistory(h *history.Store, now time.Time, store *sysreport.StorageReport) {
	for i := range store.SSDs {
		d := &store.SSDs[i]
		key := sysreport.KeyDisk + d.ID
		h.Push(key, sysreport.Sample{Timestamp: now, ReadBps: d.ReadBps, WriteBps: d.WriteBps, TemperatureC: d.TemperatureC})
		d.History = h.Get(key)
	}
	for i := range store.HDDs {
		d := &store.HDDs[i]
		key := sysreport.KeyDisk + d.ID
		h.Push(key, sysreport.Sample{Timestamp: now, ReadBps: d.ReadBps, WriteBps: d.WriteBps, TemperatureC: d.TemperatureC})
		d.History = h.Get(key)
	}
}

func firstCPU(cpus []sysreport.CPUInfo) sysreport.CPUInfo {
	if len(cpus) == 0 {
		return sysreport.CPUInfo{}
	}
	return cpus[0]
}

func round2(v float64) float64 {
	return float64(int(v*100+0.5)) / 100
}
