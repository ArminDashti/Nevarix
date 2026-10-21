package netmon

import (
	"context"
	"log"
	"time"

	"github.com/ArminDashti/nevarix-core/core/model"
	"github.com/ArminDashti/nevarix-core/core/store"
)

type Poller struct {
	store        *store.Store
	historyLimit int
	interval     time.Duration
}

func NewPoller(st *store.Store, historyLimit int, interval time.Duration) *Poller {
	if interval < 10*time.Second {
		interval = 60 * time.Second
	}
	return &Poller{store: st, historyLimit: historyLimit, interval: interval}
}

func (p *Poller) Run(ctx context.Context) {
	p.pollAll(ctx)
	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			p.pollAll(ctx)
		}
	}
}

func (p *Poller) pollAll(ctx context.Context) {
	devices, err := p.store.ListNetDevicesForPoll(ctx)
	if err != nil {
		log.Printf("netmon: list devices: %v", err)
		return
	}
	for _, d := range devices {
		p.pollOne(ctx, d)
	}
}

func (p *Poller) pollOne(ctx context.Context, device model.NetDevice) {
	prior := func(deviceID string, ifIndex int) (uint64, uint64, time.Time, bool) {
		return p.store.PriorInterfaceOctets(ctx, deviceID, ifIndex)
	}
	sample, err := PollDevice(device, prior)
	if err != nil {
		_ = p.store.RecordNetPollFailure(ctx, device.ID, err.Error())
		log.Printf("netmon: poll %s (%s): %v", device.Name, device.MgmtIP, err)
		return
	}
	if err := p.store.IngestNetPoll(ctx, sample, p.historyLimit); err != nil {
		log.Printf("netmon: ingest %s: %v", device.ID, err)
	}
}
