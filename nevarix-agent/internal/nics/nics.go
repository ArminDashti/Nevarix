package nics

import (
	"sync"
	"time"

	"github.com/ArminDashti/nevarix-agent/internal/sysreport"
	"github.com/shirou/gopsutil/v4/net"
)

type Collector struct {
	mu     sync.Mutex
	last   map[string]net.IOCountersStat
	lastAt time.Time
}

func New() *Collector {
	return &Collector{last: map[string]net.IOCountersStat{}}
}

func (c *Collector) Collect(now time.Time) ([]sysreport.NIC, float64, float64) {
	ifs, _ := net.Interfaces()
	byName := map[string]net.InterfaceStat{}
	for _, inf := range ifs {
		byName[inf.Name] = inf
	}
	counters, err := net.IOCounters(true)
	if err != nil {
		return []sysreport.NIC{}, 0, 0
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	sec := 0.0
	if !c.lastAt.IsZero() && now.After(c.lastAt) {
		sec = now.Sub(c.lastAt).Seconds()
	}

	out := make([]sysreport.NIC, 0, len(counters))
	var totRx, totTx float64
	for _, cur := range counters {
		if skipNIC(cur.Name) {
			continue
		}
		n := sysreport.NIC{Name: cur.Name, Addrs: []string{}}
		if inf, ok := byName[cur.Name]; ok {
			n.MAC = inf.HardwareAddr
			n.MTU = inf.MTU
			for _, a := range inf.Addrs {
				if a.Addr != "" {
					n.Addrs = append(n.Addrs, a.Addr)
				}
			}
		}
		if prev, ok := c.last[cur.Name]; ok && sec > 0 {
			n.RxBps = rate(cur.BytesRecv, prev.BytesRecv, sec)
			n.TxBps = rate(cur.BytesSent, prev.BytesSent, sec)
		}
		totRx += n.RxBps
		totTx += n.TxBps
		out = append(out, n)
		c.last[cur.Name] = cur
	}
	c.lastAt = now
	return out, round2(totRx), round2(totTx)
}

func Rate(cur, prev uint64, sec float64) float64 {
	return rate(cur, prev, sec)
}

func skipNIC(name string) bool {
	switch name {
	case "lo", "lo0", "Loopback Pseudo-Interface 1":
		return true
	}
	return false
}

func rate(cur, prev uint64, sec float64) float64 {
	if sec <= 0 || cur < prev {
		return 0
	}
	return round2(float64(cur-prev) / sec)
}

func round2(v float64) float64 {
	return float64(int(v*100+0.5)) / 100
}
