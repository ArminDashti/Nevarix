// Package model holds the domain types shared by core and every connector.
package model

import (
	"encoding/json"
	"time"
)

// AgentSnapshot is the host snapshot an agent pushes (REST JSON or gRPC IngestMetrics).
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

// MetricPoint is one agent history sample; the same row feeds CPU, memory and bandwidth charts.
type MetricPoint struct {
	Timestamp  time.Time `json:"timestamp"`
	CPUPercent float64   `json:"cpuPercent"`
	MemPercent float64   `json:"memPercent"`
	RxBps      float64   `json:"rxBps"`
	TxBps      float64   `json:"txBps"`
}

// Point is one local-collector series sample.
type Point struct {
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
}

// BandwidthPoint is one local-collector bandwidth sample.
type BandwidthPoint struct {
	Timestamp time.Time `json:"timestamp"`
	RxBps     float64   `json:"rxBps"`
	TxBps     float64   `json:"txBps"`
}

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

// OnlineSession is one SoftEther session snapshot.
type OnlineSession struct {
	Username               string     `json:"username"`
	ClientIP               string     `json:"clientIp"`
	ASN                    *string    `json:"asn"`
	BandwidthBps           float64    `json:"bandwidthBps"`
	DownloadBytes          uint64     `json:"downloadBytes"`
	UploadBytes            uint64     `json:"uploadBytes"`
	SessionDurationSeconds int64      `json:"sessionDurationSeconds"`
	ConnectedAt            *time.Time `json:"connectedAt"`
	SessionKey             string     `json:"sessionKey"`
}

type UserStat struct {
	Username             string  `json:"username"`
	ClientIP             string  `json:"clientIp"`
	ASN                  *string `json:"asn"`
	DownloadBytes        uint64  `json:"downloadBytes"`
	UploadBytes          uint64  `json:"uploadBytes"`
	UsageDurationSeconds int64   `json:"usageDurationSeconds"`
}

type SessionLog struct {
	ConnectedAt     time.Time  `json:"connectedAt"`
	DisconnectedAt  *time.Time `json:"disconnectedAt"`
	ClientIP        string     `json:"clientIp"`
	ASN             *string    `json:"asn"`
	DownloadBytes   uint64     `json:"downloadBytes"`
	UploadBytes     uint64     `json:"uploadBytes"`
	DurationSeconds int64      `json:"durationSeconds"`
}
