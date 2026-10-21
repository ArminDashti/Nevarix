package model

import "time"

type NetDevice struct {
	ID              string     `json:"id"`
	Name            string     `json:"name"`
	MgmtIP          string     `json:"mgmtIp"`
	Vendor          string     `json:"vendor"`
	SnmpVersion     string     `json:"snmpVersion"`
	SnmpCommunity   string     `json:"snmpCommunity,omitempty"`
	SshUsername     string     `json:"sshUsername,omitempty"`
	PollIntervalSec int        `json:"pollIntervalSec"`
	Enabled         bool       `json:"enabled"`
	SysDescr        string     `json:"sysDescr,omitempty"`
	SysObjectID     string     `json:"sysObjectId,omitempty"`
	LastOkAt        *time.Time `json:"lastOkAt,omitempty"`
	LastError       string     `json:"lastError,omitempty"`
	CreatedAt       time.Time  `json:"createdAt"`
}

type NetDeviceSummary struct {
	ID                string     `json:"id"`
	Name              string     `json:"name"`
	MgmtIP            string     `json:"mgmtIp"`
	Vendor            string     `json:"vendor"`
	Enabled           bool       `json:"enabled"`
	CPUPercent        float64    `json:"cpuPercent"`
	MemoryUsedPercent float64    `json:"memoryUsedPercent"`
	LastOkAt          *time.Time `json:"lastOkAt,omitempty"`
	LastError         string     `json:"lastError,omitempty"`
}

type NetInterface struct {
	IfIndex    int        `json:"ifIndex"`
	Name       string     `json:"name"`
	OperStatus int        `json:"operStatus"`
	SpeedBps   uint64     `json:"speedBps"`
	RxBps      float64    `json:"rxBps"`
	TxBps      float64    `json:"txBps"`
	SampledAt  *time.Time `json:"sampledAt,omitempty"`
}

type NetPollSample struct {
	DeviceID          string
	CPUPercent        float64
	MemoryUsedPercent float64
	UptimeSeconds     uint64
	BandwidthRxBps    float64
	BandwidthTxBps    float64
	CollectedAt       time.Time
	Interfaces        []NetInterfacePoll
	Vendor            string
	SysDescr          string
	SysObjectID       string
}

type NetInterfacePoll struct {
	IfIndex    int
	Name       string
	OperStatus int
	SpeedBps   uint64
	InOctets   uint64
	OutOctets  uint64
	RxBps      float64
	TxBps      float64
}

type SnmpScanHit struct {
	IP       string `json:"ip"`
	SysDescr string `json:"sysDescr"`
	Vendor   string `json:"vendor"`
}
