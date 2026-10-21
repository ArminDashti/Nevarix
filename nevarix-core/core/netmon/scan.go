package netmon

import (
	"context"
	"sync"
	"time"

	"github.com/ArminDashti/nevarix-core/agent-connectors/discovery"
	"github.com/ArminDashti/nevarix-core/core/model"
)

const snmpWorkers = 32

func ScanSNMP(ctx context.Context, cidr, community string, timeout time.Duration) (int, []model.SnmpScanHit, error) {
	hosts, err := discovery.Hosts(cidr)
	if err != nil {
		return 0, nil, err
	}
	hits := make([]*model.SnmpScanHit, len(hosts))
	jobs := make(chan int)
	var wg sync.WaitGroup
	for w := 0; w < snmpWorkers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range jobs {
				if ctx.Err() != nil {
					return
				}
				descr, vendor, err := ProbeSysDescr(hosts[i], community)
				if err != nil || descr == "" {
					continue
				}
				hits[i] = &model.SnmpScanHit{IP: hosts[i], SysDescr: descr, Vendor: vendor}
			}
		}()
	}
	go func() {
		for i := range hosts {
			jobs <- i
		}
		close(jobs)
	}()
	wg.Wait()
	out := []model.SnmpScanHit{}
	for _, h := range hits {
		if h != nil {
			out = append(out, *h)
		}
	}
	return len(hosts), out, nil
}
