// Package service holds the management use cases shared by the REST and gRPC connectors.
package service

import (
	"context"
	"log"
	"net"
	"time"

	"github.com/ArminDashti/nevarix-core/agent-connectors/discovery"
	"github.com/ArminDashti/nevarix-core/core/model"
	"github.com/ArminDashti/nevarix-core/core/store"
)

// ValidationError marks bad client input; connectors map it to 400 / InvalidArgument.
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string { return e.Message }

func invalid(msg string) error { return &ValidationError{Message: msg} }

type Management struct {
	store *store.Store
}

func New(st *store.Store) *Management {
	return &Management{store: st}
}

// EndpointView is one endpoints-list row: a scanned or added host, or an agent that reported.
type EndpointView struct {
	IP         string     `json:"ip,omitempty"`
	Hostname   string     `json:"hostname"`
	Source     string     `json:"source"`
	LastSeenAt *time.Time `json:"lastSeenAt,omitempty"`
}

func (m *Management) ListEndpoints(ctx context.Context) ([]EndpointView, error) {
	scanned, err := m.store.Endpoints(ctx)
	if err != nil {
		return nil, err
	}
	agents, err := m.store.ListAgents(ctx)
	if err != nil {
		return nil, err
	}
	out := []EndpointView{}
	for _, e := range scanned {
		out = append(out, EndpointView{IP: e.IP, Hostname: e.Hostname, Source: "scan"})
	}
	for _, a := range agents {
		seen := a.LastSeenAt
		out = append(out, EndpointView{Hostname: a.Hostname, Source: "agent", LastSeenAt: &seen})
	}
	return out, nil
}

func (m *Management) AddEndpoint(ctx context.Context, ip, hostname string) error {
	if net.ParseIP(ip) == nil {
		return invalid("ip is not a valid address")
	}
	return m.store.AddEndpoint(ctx, model.Endpoint{IP: ip, Hostname: hostname})
}

type ScanRequest struct {
	CIDR      string
	Port      int
	TimeoutMs int
}

type ScanResult struct {
	Scanned int
	Hits    []discovery.Hit
}

// Scan fills unset fields from saved settings, then probes every host for nevarix-agent.
func (m *Management) Scan(ctx context.Context, req ScanRequest) (ScanResult, error) {
	defaults, err := m.store.Settings(ctx)
	if err != nil {
		return ScanResult{}, err
	}
	if req.CIDR == "" {
		req.CIDR = defaults.DefaultCIDR
	}
	if req.Port == 0 {
		req.Port = defaults.AgentPort
	}
	if req.TimeoutMs == 0 {
		req.TimeoutMs = defaults.TimeoutMs
	}
	if err := validatePort(req.Port); err != nil {
		return ScanResult{}, err
	}
	hosts, err := discovery.Hosts(req.CIDR)
	if err != nil {
		return ScanResult{}, invalid(err.Error())
	}
	hits := discovery.Run(ctx, hosts, req.Port, time.Duration(req.TimeoutMs)*time.Millisecond)
	log.Printf("scan %s port=%d hosts=%d hits=%d", req.CIDR, req.Port, len(hosts), len(hits))
	return ScanResult{Scanned: len(hosts), Hits: hits}, nil
}

func (m *Management) GetSettings(ctx context.Context) (model.Settings, error) {
	return m.store.Settings(ctx)
}

func (m *Management) UpdateSettings(ctx context.Context, st model.Settings) (model.Settings, error) {
	if _, _, err := net.ParseCIDR(st.DefaultCIDR); err != nil {
		return model.Settings{}, invalid("defaultCidr is not a valid CIDR")
	}
	if err := validatePort(st.AgentPort); err != nil {
		return model.Settings{}, err
	}
	if st.TimeoutMs < 50 || st.TimeoutMs > 10000 {
		return model.Settings{}, invalid("timeoutMs must be between 50 and 10000")
	}
	if err := m.store.SaveSettings(ctx, st); err != nil {
		return model.Settings{}, err
	}
	return st, nil
}

func validatePort(p int) error {
	if p < 1 || p > 65535 {
		return invalid("port must be between 1 and 65535")
	}
	return nil
}
