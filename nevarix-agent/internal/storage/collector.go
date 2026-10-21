package storage

import (
	"sync"
	"time"

	"github.com/ArminDashti/nevarix-agent/internal/sysreport"
	"github.com/shirou/gopsutil/v4/disk"
)

type Collector struct {
	mu     sync.Mutex
	last   map[string]disk.IOCountersStat
	lastAt time.Time
}

func New() *Collector {
	return &Collector{last: map[string]disk.IOCountersStat{}}
}

func (c *Collector) Collect(now time.Time, temps sysreport.TempBuckets) sysreport.StorageReport {
	devs := listDevices()
	io, _ := disk.IOCounters()
	c.mu.Lock()
	defer c.mu.Unlock()
	sec := 0.0
	if !c.lastAt.IsZero() && now.After(c.lastAt) {
		sec = now.Sub(c.lastAt).Seconds()
	}
	out := sysreport.StorageReport{SSDs: []sysreport.DiskDevice{}, HDDs: []sysreport.DiskDevice{}}
	for i := range devs {
		d := &devs[i]
		if st, ok := io[d.ID]; ok {
			if prev, ok := c.last[d.ID]; ok && sec > 0 {
				d.ReadBps = rate(st.ReadBytes, prev.ReadBytes, sec)
				d.WriteBps = rate(st.WriteBytes, prev.WriteBytes, sec)
			}
			c.last[d.ID] = st
		}
		d.TemperatureC = sysreport.DiskTemp(temps, d.ID, d.Model)
		if d.Kind == "hdd" {
			out.HDDs = append(out.HDDs, *d)
		} else {
			// nvme/unknown treated as ssd-class unless rotational said hdd
			out.SSDs = append(out.SSDs, *d)
		}
	}
	c.lastAt = now
	return out
}
