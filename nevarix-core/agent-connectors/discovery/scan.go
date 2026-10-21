// Package discovery finds nevarix-agent installs on the network by probing their /health endpoint.
package discovery

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"
)

// ponytail: synchronous scan, capped at /16 (65k hosts). A /16 at 500ms
// timeout and 64 workers takes minutes; upgrade path is an async job + polling.
// Only plain HTTP /health is probed; no TLS or auth on the agent listener yet.
const (
	maxHostBits = 16
	workers     = 64
)

type Hit struct {
	IP       string `json:"ip"`
	Port     int    `json:"port"`
	AgentID  string `json:"agentId"`
	Hostname string `json:"hostname"`
}

// identity is the agent's GET /health body. Only hosts answering as nevarix-agent count as hits.
type identity struct {
	Agent    string `json:"agent"`
	AgentID  string `json:"agentId"`
	Hostname string `json:"hostname"`
}

// Hosts expands an IPv4 CIDR into usable host addresses (no network/broadcast).
func Hosts(cidr string) ([]string, error) {
	_, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, err
	}
	base := ipnet.IP.To4()
	if base == nil {
		return nil, errors.New("only IPv4 CIDR is supported")
	}
	ones, bits := ipnet.Mask.Size()
	hostBits := bits - ones
	if hostBits > maxHostBits {
		return nil, fmt.Errorf("range too large, smallest allowed prefix is /%d", bits-maxHostBits)
	}
	start := binary.BigEndian.Uint32(base)
	count := uint32(1) << hostBits
	var hosts []string
	for i := uint32(0); i < count; i++ {
		if count > 2 && (i == 0 || i == count-1) {
			continue
		}
		addr := make(net.IP, 4)
		binary.BigEndian.PutUint32(addr, start+i)
		hosts = append(hosts, addr.String())
	}
	return hosts, nil
}

// Run probes GET /health on every host:port and returns the agents found, in input order.
func Run(ctx context.Context, hosts []string, port int, timeout time.Duration) []Hit {
	found := make([]*Hit, len(hosts))
	jobs := make(chan int)
	var wg sync.WaitGroup
	client := &http.Client{Timeout: timeout}
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range jobs {
				found[i] = probe(ctx, client, hosts[i], port)
			}
		}()
	}
	for i := range hosts {
		select {
		case jobs <- i:
		case <-ctx.Done():
		}
	}
	close(jobs)
	wg.Wait()

	hits := []Hit{}
	for _, h := range found {
		if h != nil {
			hits = append(hits, *h)
		}
	}
	return hits
}

// Probe checks a single host:port for nevarix-agent /health.
func Probe(ctx context.Context, ip string, port int, timeout time.Duration) *Hit {
	client := &http.Client{Timeout: timeout}
	return probe(ctx, client, ip, port)
}

func probe(ctx context.Context, client *http.Client, ip string, port int) *Hit {
	url := "http://" + net.JoinHostPort(ip, strconv.Itoa(port)) + "/health"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil
	}
	res, err := client.Do(req)
	if err != nil {
		return nil
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil
	}
	var id identity
	if err := json.NewDecoder(io.LimitReader(res.Body, 1<<12)).Decode(&id); err != nil || id.Agent != "nevarix-agent" {
		return nil
	}
	return &Hit{IP: ip, Port: port, AgentID: id.AgentID, Hostname: id.Hostname}
}
