package service

import (
	"context"
	"net"
	"time"

	"github.com/ArminDashti/nevarix-core/core/model"
	"github.com/ArminDashti/nevarix-core/core/netmon"
	"github.com/ArminDashti/nevarix-core/core/store"
	"github.com/jackc/pgx/v5"
)

func (m *Management) ListNetDevices(ctx context.Context) ([]model.NetDeviceSummary, error) {
	return m.store.ListNetDevices(ctx)
}

func (m *Management) GetNetDevice(ctx context.Context, id string) (*model.NetDevice, error) {
	return m.store.GetNetDevice(ctx, id)
}

type CreateNetDeviceRequest struct {
	Name            string `json:"name"`
	MgmtIP          string `json:"mgmtIp"`
	SnmpCommunity   string `json:"snmpCommunity"`
	PollIntervalSec int    `json:"pollIntervalSec"`
	Enabled         *bool  `json:"enabled"`
	SshUsername     string `json:"sshUsername"`
	SshPassword     string `json:"sshPassword"`
	VerifySnmp      bool   `json:"verifySnmp"`
}

func (m *Management) CreateNetDevice(ctx context.Context, req CreateNetDeviceRequest) (string, error) {
	if net.ParseIP(req.MgmtIP) == nil {
		return "", invalid("mgmtIp is not a valid address")
	}
	if req.SnmpCommunity == "" {
		st, err := m.store.Settings(ctx)
		if err != nil {
			return "", err
		}
		req.SnmpCommunity = st.SnmpCommunity
	}
	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}
	if req.Name == "" {
		req.Name = req.MgmtIP
	}
	if req.VerifySnmp {
		if _, err := netmon.ReachableSNMP(req.MgmtIP, req.SnmpCommunity); err != nil {
			return "", invalid("snmp unreachable: " + err.Error())
		}
	}
	id, err := m.store.CreateNetDevice(ctx, store.CreateNetDeviceInput{
		Name:            req.Name,
		MgmtIP:          req.MgmtIP,
		SnmpCommunity:   req.SnmpCommunity,
		PollIntervalSec: req.PollIntervalSec,
		Enabled:         enabled,
		SshUsername:     req.SshUsername,
		SshPassword:     req.SshPassword,
	})
	return id, err
}

type UpdateNetDeviceRequest struct {
	Name            string `json:"name"`
	SnmpCommunity   string `json:"snmpCommunity"`
	PollIntervalSec int    `json:"pollIntervalSec"`
	Enabled         bool   `json:"enabled"`
}

func (m *Management) UpdateNetDevice(ctx context.Context, id string, req UpdateNetDeviceRequest) error {
	if req.SnmpCommunity == "" {
		return invalid("snmpCommunity is required")
	}
	err := m.store.UpdateNetDevice(ctx, id, store.UpdateNetDeviceInput{
		Name:            req.Name,
		SnmpCommunity:   req.SnmpCommunity,
		PollIntervalSec: req.PollIntervalSec,
		Enabled:         req.Enabled,
	})
	if err != nil && err == pgx.ErrNoRows {
		return invalid("device not found")
	}
	return err
}

func (m *Management) DeleteNetDevice(ctx context.Context, id string) error {
	err := m.store.DeleteNetDevice(ctx, id)
	if err != nil && err == pgx.ErrNoRows {
		return invalid("device not found")
	}
	return err
}

func (m *Management) ListNetInterfaces(ctx context.Context, deviceID string) ([]model.NetInterface, error) {
	return m.store.ListNetInterfaces(ctx, deviceID)
}

func (m *Management) NetDeviceHistory(ctx context.Context, deviceID string) ([]model.MetricPoint, error) {
	return m.store.NetDeviceHistory(ctx, deviceID)
}

type SnmpScanRequest struct {
	CIDR      string
	TimeoutMs int
}

func (m *Management) ScanSNMP(ctx context.Context, req SnmpScanRequest) (int, []model.SnmpScanHit, error) {
	st, err := m.store.Settings(ctx)
	if err != nil {
		return 0, nil, err
	}
	if req.CIDR == "" {
		req.CIDR = st.DefaultCIDR
	}
	if req.TimeoutMs <= 0 {
		req.TimeoutMs = st.TimeoutMs
	}
	if _, _, err := net.ParseCIDR(req.CIDR); err != nil {
		return 0, nil, invalid("cidr is not valid")
	}
	return netmon.ScanSNMP(ctx, req.CIDR, st.SnmpCommunity, time.Duration(req.TimeoutMs)*time.Millisecond)
}
