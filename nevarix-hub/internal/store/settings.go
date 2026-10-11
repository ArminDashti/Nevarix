package store

import (
	"context"
	"time"
)

type Settings struct {
	DefaultCIDR string `json:"defaultCidr"`
	AgentPort   int    `json:"agentPort"`
	TimeoutMs   int    `json:"timeoutMs"`
}

type Endpoint struct {
	IP       string    `json:"ip"`
	Hostname string    `json:"hostname"`
	AddedAt  time.Time `json:"addedAt"`
}

func DefaultSettings() Settings {
	return Settings{DefaultCIDR: "192.168.1.0/24", AgentPort: 9000, TimeoutMs: 500}
}

// Settings returns the saved settings, or the defaults when none were saved yet.
func (s *Store) Settings(ctx context.Context) (Settings, error) {
	st := DefaultSettings()
	err := s.pool.QueryRow(ctx,
		`SELECT default_cidr, agent_port, timeout_ms FROM hub_settings WHERE id = 1`,
	).Scan(&st.DefaultCIDR, &st.AgentPort, &st.TimeoutMs)
	if isNoRows(err) {
		return st, nil
	}
	return st, err
}

func (s *Store) SaveSettings(ctx context.Context, st Settings) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO hub_settings (id, default_cidr, agent_port, timeout_ms)
		VALUES (1, $1, $2, $3)
		ON CONFLICT (id) DO UPDATE SET
			default_cidr = EXCLUDED.default_cidr,
			agent_port = EXCLUDED.agent_port,
			timeout_ms = EXCLUDED.timeout_ms
	`, st.DefaultCIDR, st.AgentPort, st.TimeoutMs)
	return err
}

func (s *Store) Endpoints(ctx context.Context) ([]Endpoint, error) {
	rows, err := s.pool.Query(ctx, `SELECT ip, hostname, added_at FROM hub_endpoints ORDER BY added_at`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Endpoint{}
	for rows.Next() {
		var e Endpoint
		if err := rows.Scan(&e.IP, &e.Hostname, &e.AddedAt); err != nil {
			return nil, err
		}
		e.AddedAt = e.AddedAt.UTC()
		out = append(out, e)
	}
	return out, rows.Err()
}

// AddEndpoint upserts by IP so re-adding a scanned host does not duplicate it.
func (s *Store) AddEndpoint(ctx context.Context, e Endpoint) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO hub_endpoints (ip, hostname, added_at) VALUES ($1, $2, $3)
		ON CONFLICT (ip) DO UPDATE SET hostname = EXCLUDED.hostname, added_at = EXCLUDED.added_at
	`, e.IP, e.Hostname, time.Now().UTC())
	return err
}
