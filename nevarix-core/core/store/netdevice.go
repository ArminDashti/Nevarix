package store

import (
	"context"
	"errors"
	"time"

	"github.com/ArminDashti/nevarix-core/core/model"
	"github.com/jackc/pgx/v5"
)

func (s *Store) ListNetDevices(ctx context.Context) ([]model.NetDeviceSummary, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT d.id, d.name, host(d.mgmt_ip), d.vendor, d.enabled,
		       COALESCE(l.cpu_percent, 0), COALESCE(l.memory_used_percent, 0),
		       d.last_ok_at, d.last_error
		FROM net_devices d
		LEFT JOIN net_device_latest l ON l.device_id = d.id
		ORDER BY d.name, host(d.mgmt_ip)
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.NetDeviceSummary{}
	for rows.Next() {
		var d model.NetDeviceSummary
		var lastOk *time.Time
		if err := rows.Scan(&d.ID, &d.Name, &d.MgmtIP, &d.Vendor, &d.Enabled,
			&d.CPUPercent, &d.MemoryUsedPercent, &lastOk, &d.LastError); err != nil {
			return nil, err
		}
		if lastOk != nil {
			t := lastOk.UTC()
			d.LastOkAt = &t
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

func (s *Store) GetNetDevice(ctx context.Context, id string) (*model.NetDevice, error) {
	var d model.NetDevice
	var lastOk *time.Time
	err := s.pool.QueryRow(ctx, `
		SELECT id, name, host(mgmt_ip), vendor, snmp_version, snmp_community, ssh_username,
		       poll_interval_sec, enabled, sys_descr, sys_object_id, last_ok_at, last_error, created_at
		FROM net_devices WHERE id = $1
	`, id).Scan(
		&d.ID, &d.Name, &d.MgmtIP, &d.Vendor, &d.SnmpVersion, &d.SnmpCommunity, &d.SshUsername,
		&d.PollIntervalSec, &d.Enabled, &d.SysDescr, &d.SysObjectID, &lastOk, &d.LastError, &d.CreatedAt,
	)
	if isNoRows(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if lastOk != nil {
		t := lastOk.UTC()
		d.LastOkAt = &t
	}
	d.CreatedAt = d.CreatedAt.UTC()
	return &d, nil
}

type CreateNetDeviceInput struct {
	Name            string
	MgmtIP          string
	SnmpCommunity   string
	PollIntervalSec int
	Enabled         bool
	SshUsername     string
	SshPassword     string
}

func (s *Store) CreateNetDevice(ctx context.Context, in CreateNetDeviceInput) (string, error) {
	if in.PollIntervalSec < 10 {
		in.PollIntervalSec = 60
	}
	var id string
	err := s.pool.QueryRow(ctx, `
		INSERT INTO net_devices (name, mgmt_ip, snmp_community, poll_interval_sec, enabled, ssh_username, ssh_password)
		VALUES ($1, $2::inet, $3, $4, $5, $6, $7)
		RETURNING id::text
	`, in.Name, in.MgmtIP, in.SnmpCommunity, in.PollIntervalSec, in.Enabled, in.SshUsername, in.SshPassword).Scan(&id)
	return id, err
}

type UpdateNetDeviceInput struct {
	Name            string
	SnmpCommunity   string
	PollIntervalSec int
	Enabled         bool
}

func (s *Store) UpdateNetDevice(ctx context.Context, id string, in UpdateNetDeviceInput) error {
	if in.PollIntervalSec < 10 {
		in.PollIntervalSec = 60
	}
	tag, err := s.pool.Exec(ctx, `
		UPDATE net_devices SET
			name = $2, snmp_community = $3, poll_interval_sec = $4, enabled = $5
		WHERE id = $1
	`, id, in.Name, in.SnmpCommunity, in.PollIntervalSec, in.Enabled)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (s *Store) DeleteNetDevice(ctx context.Context, id string) error {
	tag, err := s.pool.Exec(ctx, `DELETE FROM net_devices WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (s *Store) ListNetDevicesForPoll(ctx context.Context) ([]model.NetDevice, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id::text, name, host(mgmt_ip), vendor, snmp_version, snmp_community,
		       poll_interval_sec, enabled, sys_descr, sys_object_id, last_ok_at, last_error, created_at
		FROM net_devices WHERE enabled = TRUE
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.NetDevice{}
	for rows.Next() {
		var d model.NetDevice
		var lastOk *time.Time
		if err := rows.Scan(
			&d.ID, &d.Name, &d.MgmtIP, &d.Vendor, &d.SnmpVersion, &d.SnmpCommunity,
			&d.PollIntervalSec, &d.Enabled, &d.SysDescr, &d.SysObjectID, &lastOk, &d.LastError, &d.CreatedAt,
		); err != nil {
			return nil, err
		}
		if lastOk != nil {
			t := lastOk.UTC()
			d.LastOkAt = &t
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

func (s *Store) RecordNetPollFailure(ctx context.Context, deviceID string, errMsg string) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE net_devices SET last_error = $2 WHERE id = $1
	`, deviceID, errMsg)
	return err
}

func (s *Store) IngestNetPoll(ctx context.Context, sample model.NetPollSample, historyLimit int) error {
	if sample.DeviceID == "" {
		return errors.New("deviceId is required")
	}
	if historyLimit < 10 {
		historyLimit = 10
	}
	collectedAt := sample.CollectedAt.UTC()
	if collectedAt.IsZero() {
		collectedAt = time.Now().UTC()
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `
		UPDATE net_devices SET
			vendor = $2, sys_descr = $3, sys_object_id = $4,
			last_ok_at = $5, last_error = ''
		WHERE id = $1
	`, sample.DeviceID, sample.Vendor, sample.SysDescr, sample.SysObjectID, collectedAt); err != nil {
		return err
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO net_device_latest (
			device_id, cpu_percent, memory_used_percent, uptime_seconds,
			bandwidth_rx_bps, bandwidth_tx_bps, collected_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7)
		ON CONFLICT (device_id) DO UPDATE SET
			cpu_percent = EXCLUDED.cpu_percent,
			memory_used_percent = EXCLUDED.memory_used_percent,
			uptime_seconds = EXCLUDED.uptime_seconds,
			bandwidth_rx_bps = EXCLUDED.bandwidth_rx_bps,
			bandwidth_tx_bps = EXCLUDED.bandwidth_tx_bps,
			collected_at = EXCLUDED.collected_at
		WHERE net_device_latest.collected_at <= EXCLUDED.collected_at
	`, sample.DeviceID, sample.CPUPercent, sample.MemoryUsedPercent, int64(sample.UptimeSeconds),
		sample.BandwidthRxBps, sample.BandwidthTxBps, collectedAt); err != nil {
		return err
	}

	for _, iface := range sample.Interfaces {
		if _, err := tx.Exec(ctx, `
			INSERT INTO net_interfaces (
				device_id, if_index, name, oper_status, speed_bps,
				in_octets, out_octets, rx_bps, tx_bps, sampled_at
			) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
			ON CONFLICT (device_id, if_index) DO UPDATE SET
				name = EXCLUDED.name,
				oper_status = EXCLUDED.oper_status,
				speed_bps = EXCLUDED.speed_bps,
				in_octets = EXCLUDED.in_octets,
				out_octets = EXCLUDED.out_octets,
				rx_bps = EXCLUDED.rx_bps,
				tx_bps = EXCLUDED.tx_bps,
				sampled_at = EXCLUDED.sampled_at
		`, sample.DeviceID, iface.IfIndex, iface.Name, iface.OperStatus, int64(iface.SpeedBps),
			int64(iface.InOctets), int64(iface.OutOctets), iface.RxBps, iface.TxBps, collectedAt); err != nil {
			return err
		}
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO net_metric_points (device_id, ts, cpu_percent, mem_percent, rx_bps, tx_bps)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (device_id, ts) DO NOTHING
	`, sample.DeviceID, collectedAt, sample.CPUPercent, sample.MemoryUsedPercent,
		sample.BandwidthRxBps, sample.BandwidthTxBps); err != nil {
		return err
	}

	if _, err := tx.Exec(ctx, `
		DELETE FROM net_metric_points
		WHERE device_id = $1
		  AND id NOT IN (
			SELECT id FROM net_metric_points WHERE device_id = $1 ORDER BY ts DESC LIMIT $2
		  )
	`, sample.DeviceID, historyLimit); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (s *Store) ListNetInterfaces(ctx context.Context, deviceID string) ([]model.NetInterface, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT if_index, name, oper_status, speed_bps, rx_bps, tx_bps, sampled_at
		FROM net_interfaces WHERE device_id = $1 ORDER BY if_index
	`, deviceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.NetInterface{}
	for rows.Next() {
		var iface model.NetInterface
		var sampled *time.Time
		var speed int64
		if err := rows.Scan(&iface.IfIndex, &iface.Name, &iface.OperStatus, &speed,
			&iface.RxBps, &iface.TxBps, &sampled); err != nil {
			return nil, err
		}
		iface.SpeedBps = uint64(speed)
		if sampled != nil {
			t := sampled.UTC()
			iface.SampledAt = &t
		}
		out = append(out, iface)
	}
	return out, rows.Err()
}

func (s *Store) NetDeviceHistory(ctx context.Context, deviceID string) ([]model.MetricPoint, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT ts, cpu_percent, mem_percent, rx_bps, tx_bps
		FROM net_metric_points WHERE device_id = $1 ORDER BY ts ASC
	`, deviceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.MetricPoint{}
	for rows.Next() {
		var p model.MetricPoint
		if err := rows.Scan(&p.Timestamp, &p.CPUPercent, &p.MemPercent, &p.RxBps, &p.TxBps); err != nil {
			return nil, err
		}
		p.Timestamp = p.Timestamp.UTC()
		out = append(out, p)
	}
	return out, rows.Err()
}

func (s *Store) PriorInterfaceOctets(ctx context.Context, deviceID string, ifIndex int) (uint64, uint64, time.Time, bool) {
	var inOct, outOct int64
	var sampled *time.Time
	err := s.pool.QueryRow(ctx, `
		SELECT in_octets, out_octets, sampled_at FROM net_interfaces
		WHERE device_id = $1 AND if_index = $2
	`, deviceID, ifIndex).Scan(&inOct, &outOct, &sampled)
	if err != nil || sampled == nil {
		return 0, 0, time.Time{}, false
	}
	return uint64(inOct), uint64(outOct), sampled.UTC(), true
}
