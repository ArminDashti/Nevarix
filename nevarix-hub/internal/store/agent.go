package store

import (
	"context"
	"encoding/json"
	"errors"
	"time"
)

// AgentSnapshot is the payload an agent posts (same JSON as its collector.Snapshot).
type AgentSnapshot struct {
	AgentID              string          `json:"agentId"`
	Hostname             string          `json:"hostname"`
	CPUPercent           float64         `json:"cpuPercent"`
	Cores                []CoreSample    `json:"cores"`
	MemoryTotalBytes     uint64          `json:"memoryTotalBytes"`
	MemoryUsedBytes      uint64          `json:"memoryUsedBytes"`
	MemoryAvailableBytes uint64          `json:"memoryAvailableBytes"`
	MemoryUsedPercent    float64         `json:"memoryUsedPercent"`
	Disks                []DiskSample    `json:"disks"`
	BandwidthRxBps       float64         `json:"bandwidthRxBps"`
	BandwidthTxBps       float64         `json:"bandwidthTxBps"`
	UptimeSeconds        uint64          `json:"uptimeSeconds"`
	System               json.RawMessage `json:"system"`
	CollectedAt          time.Time       `json:"collectedAt"`
}

type CoreSample struct {
	CoreIndex    int     `json:"coreIndex"`
	UsagePercent float64 `json:"usagePercent"`
}

type DiskSample struct {
	MountPoint  string  `json:"mountPoint"`
	Fstype      string  `json:"fstype,omitempty"`
	TotalBytes  uint64  `json:"totalBytes"`
	UsedBytes   uint64  `json:"usedBytes"`
	FreeBytes   uint64  `json:"freeBytes"`
	UsedPercent float64 `json:"usedPercent"`
}

type AgentSummary struct {
	AgentID           string    `json:"agentId"`
	Hostname          string    `json:"hostname"`
	CPUPercent        float64   `json:"cpuPercent"`
	MemoryUsedPercent float64   `json:"memoryUsedPercent"`
	LastSeenAt        time.Time `json:"lastSeenAt"`
}

// MetricPoint is one history sample; the same row feeds CPU, memory and bandwidth charts.
type MetricPoint struct {
	Timestamp  time.Time `json:"timestamp"`
	CPUPercent float64   `json:"cpuPercent"`
	MemPercent float64   `json:"memPercent"`
	RxBps      float64   `json:"rxBps"`
	TxBps      float64   `json:"txBps"`
}

// IngestAgent stores one snapshot in a single transaction. Replays of the same
// (agentId, collectedAt) are no-ops, so an agent can resend its outbox safely.
func (s *Store) IngestAgent(ctx context.Context, p AgentSnapshot, historyLimit int) error {
	if p.AgentID == "" {
		return errors.New("agentId is required")
	}
	collectedAt := p.CollectedAt.UTC()
	if collectedAt.IsZero() {
		collectedAt = time.Now().UTC()
	}
	if historyLimit < 10 {
		historyLimit = 10
	}
	coresJSON, err := marshalOr(p.Cores, "[]")
	if err != nil {
		return err
	}
	disksJSON, err := marshalOr(p.Disks, "[]")
	if err != nil {
		return err
	}
	systemJSON := []byte(p.System)
	if len(systemJSON) == 0 {
		systemJSON = []byte("{}")
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `
		INSERT INTO agent_hosts (agent_id, hostname, last_seen_at) VALUES ($1, $2, NOW())
		ON CONFLICT (agent_id) DO UPDATE SET hostname = EXCLUDED.hostname, last_seen_at = NOW()
	`, p.AgentID, p.Hostname); err != nil {
		return err
	}

	// Only move "latest" forward: a late replay of an old snapshot must not overwrite newer data.
	if _, err := tx.Exec(ctx, `
		INSERT INTO agent_metric_latest (
			agent_id, hostname, cpu_percent, cores_json,
			memory_total_bytes, memory_used_bytes, memory_available_bytes, memory_used_percent,
			disks_json, bandwidth_rx_bps, bandwidth_tx_bps, uptime_seconds, system_json, collected_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)
		ON CONFLICT (agent_id) DO UPDATE SET
			hostname = EXCLUDED.hostname,
			cpu_percent = EXCLUDED.cpu_percent,
			cores_json = EXCLUDED.cores_json,
			memory_total_bytes = EXCLUDED.memory_total_bytes,
			memory_used_bytes = EXCLUDED.memory_used_bytes,
			memory_available_bytes = EXCLUDED.memory_available_bytes,
			memory_used_percent = EXCLUDED.memory_used_percent,
			disks_json = EXCLUDED.disks_json,
			bandwidth_rx_bps = EXCLUDED.bandwidth_rx_bps,
			bandwidth_tx_bps = EXCLUDED.bandwidth_tx_bps,
			uptime_seconds = EXCLUDED.uptime_seconds,
			system_json = EXCLUDED.system_json,
			collected_at = EXCLUDED.collected_at
		WHERE agent_metric_latest.collected_at <= EXCLUDED.collected_at
	`, p.AgentID, p.Hostname, p.CPUPercent, coresJSON,
		int64(p.MemoryTotalBytes), int64(p.MemoryUsedBytes), int64(p.MemoryAvailableBytes), p.MemoryUsedPercent,
		disksJSON, p.BandwidthRxBps, p.BandwidthTxBps, int64(p.UptimeSeconds), systemJSON, collectedAt); err != nil {
		return err
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO agent_metric_points (agent_id, ts, cpu_percent, mem_percent, rx_bps, tx_bps)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (agent_id, ts) DO NOTHING
	`, p.AgentID, collectedAt, p.CPUPercent, p.MemoryUsedPercent, p.BandwidthRxBps, p.BandwidthTxBps); err != nil {
		return err
	}

	if _, err := tx.Exec(ctx, `
		DELETE FROM agent_metric_points
		WHERE agent_id = $1
		  AND id NOT IN (
			SELECT id FROM agent_metric_points WHERE agent_id = $1 ORDER BY ts DESC LIMIT $2
		  )
	`, p.AgentID, historyLimit); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func marshalOr[T any](v []T, empty string) ([]byte, error) {
	if v == nil {
		return []byte(empty), nil
	}
	return json.Marshal(v)
}

func (s *Store) ListAgents(ctx context.Context) ([]AgentSummary, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT h.agent_id, h.hostname, COALESCE(l.cpu_percent, 0), COALESCE(l.memory_used_percent, 0), h.last_seen_at
		FROM agent_hosts h
		LEFT JOIN agent_metric_latest l ON l.agent_id = h.agent_id
		ORDER BY h.last_seen_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []AgentSummary{}
	for rows.Next() {
		var a AgentSummary
		if err := rows.Scan(&a.AgentID, &a.Hostname, &a.CPUPercent, &a.MemoryUsedPercent, &a.LastSeenAt); err != nil {
			return nil, err
		}
		a.LastSeenAt = a.LastSeenAt.UTC()
		out = append(out, a)
	}
	return out, rows.Err()
}

// AgentLatest mirrors AgentSnapshot plus the agent id; System stays raw JSON.
type AgentLatest struct {
	AgentSnapshot
}

// LatestAgent returns the most recently seen agent's latest snapshot, or nil when none exist.
func (s *Store) LatestAgent(ctx context.Context) (*AgentLatest, error) {
	var a AgentSnapshot
	var coresRaw, disksRaw []byte
	var memTotal, memUsed, memAvail, uptime int64
	err := s.pool.QueryRow(ctx, `
		SELECT l.agent_id, l.hostname, l.cpu_percent, l.cores_json,
		       l.memory_total_bytes, l.memory_used_bytes, l.memory_available_bytes, l.memory_used_percent,
		       l.disks_json, l.bandwidth_rx_bps, l.bandwidth_tx_bps, l.uptime_seconds, l.system_json, l.collected_at
		FROM agent_metric_latest l
		JOIN agent_hosts h ON h.agent_id = l.agent_id
		ORDER BY h.last_seen_at DESC
		LIMIT 1
	`).Scan(
		&a.AgentID, &a.Hostname, &a.CPUPercent, &coresRaw,
		&memTotal, &memUsed, &memAvail, &a.MemoryUsedPercent,
		&disksRaw, &a.BandwidthRxBps, &a.BandwidthTxBps, &uptime, &a.System, &a.CollectedAt,
	)
	if isNoRows(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	a.MemoryTotalBytes = uint64(memTotal)
	a.MemoryUsedBytes = uint64(memUsed)
	a.MemoryAvailableBytes = uint64(memAvail)
	a.UptimeSeconds = uint64(uptime)
	a.CollectedAt = a.CollectedAt.UTC()
	a.Cores = []CoreSample{}
	a.Disks = []DiskSample{}
	_ = json.Unmarshal(coresRaw, &a.Cores)
	_ = json.Unmarshal(disksRaw, &a.Disks)
	return &AgentLatest{AgentSnapshot: a}, nil
}

// AgentHistory returns ascending samples for one agent.
func (s *Store) AgentHistory(ctx context.Context, agentID string) ([]MetricPoint, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT ts, cpu_percent, mem_percent, rx_bps, tx_bps
		FROM agent_metric_points WHERE agent_id = $1 ORDER BY ts ASC
	`, agentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []MetricPoint{}
	for rows.Next() {
		var p MetricPoint
		if err := rows.Scan(&p.Timestamp, &p.CPUPercent, &p.MemPercent, &p.RxBps, &p.TxBps); err != nil {
			return nil, err
		}
		p.Timestamp = p.Timestamp.UTC()
		out = append(out, p)
	}
	return out, rows.Err()
}
