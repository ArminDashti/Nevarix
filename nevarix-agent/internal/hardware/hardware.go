package hardware

import (
	"sync"

	"github.com/ArminDashti/nevarix-agent/internal/sysreport"
	"github.com/shirou/gopsutil/v4/cpu"
)

type Collector struct {
	mu   sync.Mutex
	once bool
	full sysreport.HardwareFull
}

func New() *Collector { return &Collector{} }

func (c *Collector) Collect() sysreport.HardwareReport {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.once {
		c.full = readFull()
		c.once = true
	}
	return sysreport.HardwareReport{Full: c.full}
}

func cpuInfos() []sysreport.CPUInfo {
	infos, err := cpu.Info()
	if err != nil || len(infos) == 0 {
		return []sysreport.CPUInfo{}
	}
	out := make([]sysreport.CPUInfo, 0, len(infos))
	for _, inf := range infos {
		out = append(out, sysreport.CPUInfo{
			ModelName: inf.ModelName,
			Vendor:    inf.VendorID,
			Cores:     inf.Cores,
			Mhz:       inf.Mhz,
			CacheKB:   inf.CacheSize,
		})
	}
	logical, _ := cpu.Counts(true)
	if len(out) > 0 {
		out[0].Logical = logical
	}
	return out
}

func readFull() sysreport.HardwareFull {
	h := sysreport.HardwareFull{
		CPUs:   cpuInfos(),
		Memory: []sysreport.DIMM{},
		GPUs:   []sysreport.GPUInfo{},
	}
	fillOS(&h)
	return h
}
