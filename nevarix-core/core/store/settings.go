package store

import (
	"context"
	"time"

	"github.com/ArminDashti/nevarix-core/core/model"
)

func DefaultSettings() model.Settings {
	return model.Settings{
		DefaultCIDR:        "192.168.1.0/24",
		AgentPort:          9000,
		TimeoutMs:          500,
		SnmpCommunity:      "public",
		NetPollIntervalSec: 60,
	}
}

// Settings returns the saved settings, or the defaults when none were saved yet.
func (s *Store) Settings(ctx context.Context) (model.Settings, error) {
	st := DefaultSettings()
	err := s.pool.QueryRow(ctx,
		`SELECT default_cidr, agent_port, timeout_ms, snmp_community, net_poll_interval_sec FROM core_settings WHERE id = 1`,
	).Scan(&st.DefaultCIDR, &st.AgentPort, &st.TimeoutMs, &st.SnmpCommunity, &st.NetPollIntervalSec)
	if isNoRows(err) {
		return st, nil
	}
	return st, err
}

func (s *Store) SaveSettings(ctx context.Context, st model.Settings) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO core_settings (id, default_cidr, agent_port, timeout_ms, snmp_community, net_poll_interval_sec)
		VALUES (1, $1, $2, $3, $4, $5)
		ON CONFLICT (id) DO UPDATE SET
			default_cidr = EXCLUDED.default_cidr,
			agent_port = EXCLUDED.agent_port,
			timeout_ms = EXCLUDED.timeout_ms,
			snmp_community = EXCLUDED.snmp_community,
			net_poll_interval_sec = EXCLUDED.net_poll_interval_sec
	`, st.DefaultCIDR, st.AgentPort, st.TimeoutMs, st.SnmpCommunity, st.NetPollIntervalSec)
	return err
}

func (s *Store) Endpoints(ctx context.Context) ([]model.Endpoint, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT ip, hostname, agent_port, ssh_username, added_at
		FROM core_endpoints ORDER BY added_at`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.Endpoint{}
	for rows.Next() {
		var e model.Endpoint
		if err := rows.Scan(&e.IP, &e.Hostname, &e.AgentPort, &e.Username, &e.AddedAt); err != nil {
			return nil, err
		}
		e.AddedAt = e.AddedAt.UTC()
		out = append(out, e)
	}
	return out, rows.Err()
}

// AddEndpoint upserts by IP so re-adding a scanned host does not duplicate it.
func (s *Store) AddEndpoint(ctx context.Context, e model.Endpoint) error {
	port := e.AgentPort
	_, err := s.pool.Exec(ctx, `
		INSERT INTO core_endpoints (ip, hostname, agent_port, ssh_username, ssh_password, added_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (ip) DO UPDATE SET
			hostname = EXCLUDED.hostname,
			agent_port = EXCLUDED.agent_port,
			ssh_username = EXCLUDED.ssh_username,
			ssh_password = CASE WHEN EXCLUDED.ssh_password <> '' THEN EXCLUDED.ssh_password ELSE core_endpoints.ssh_password END,
			added_at = EXCLUDED.added_at
	`, e.IP, e.Hostname, port, e.Username, e.Password, time.Now().UTC())
	return err
}
