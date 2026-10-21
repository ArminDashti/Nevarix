package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	agentv1 "github.com/ArminDashti/nevarix-agent/api/agent/v1"
	"github.com/ArminDashti/nevarix-agent/internal/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Sender interface {
	PostMetrics(ctx context.Context, body []byte) error
	Close() error
}

type restClient struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

type grpcClient struct {
	token  string
	conn   *grpc.ClientConn
	stub   agentv1.AgentServiceClient
}

func New(cfg config.Config) Sender {
	switch cfg.Transport {
	case "rest":
		return newREST(cfg.CoreURL, cfg.AgentToken)
	case "grpc":
		c, err := newGRPC(cfg.GRPCAddr, cfg.AgentToken)
		if err != nil {
			log.Printf("grpc client: %v (falling back to REST)", err)
			return newREST(cfg.CoreURL, cfg.AgentToken)
		}
		return c
	default:
		c, err := newGRPC(cfg.GRPCAddr, cfg.AgentToken)
		if err != nil {
			log.Printf("grpc auto: %v; using REST", err)
			return newREST(cfg.CoreURL, cfg.AgentToken)
		}
		return c
	}
}

func newREST(baseURL, token string) *restClient {
	return &restClient{
		baseURL: baseURL,
		token:   token,
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

func (c *restClient) Close() error { return nil }

func (c *restClient) PostMetrics(ctx context.Context, body []byte) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/v1/ingest/metrics", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.token)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("ingest failed: status=%d body=%s", resp.StatusCode, string(respBody))
	}
	return nil
}

func newGRPC(addr, token string) (*grpcClient, error) {
	addr = strings.TrimPrefix(addr, "http://")
	addr = strings.TrimPrefix(addr, "https://")
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	return &grpcClient{token: token, conn: conn, stub: agentv1.NewAgentServiceClient(conn)}, nil
}

func (c *grpcClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func (c *grpcClient) PostMetrics(ctx context.Context, body []byte) error {
	in, err := snapshotToProto(body)
	if err != nil {
		return err
	}
	ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+c.token)
	ack, err := c.stub.IngestMetrics(ctx, in)
	if err != nil {
		return err
	}
	if ack.GetStatus() != "" && ack.GetStatus() != "ok" {
		return fmt.Errorf("ingest failed: status=%s", ack.GetStatus())
	}
	return nil
}

type wireSnapshot struct {
	AgentID              string          `json:"agentId"`
	Hostname             string          `json:"hostname"`
	CPUPercent           float64         `json:"cpuPercent"`
	Cores                []wireCore      `json:"cores"`
	MemoryTotalBytes     uint64          `json:"memoryTotalBytes"`
	MemoryUsedBytes      uint64          `json:"memoryUsedBytes"`
	MemoryAvailableBytes uint64          `json:"memoryAvailableBytes"`
	MemoryUsedPercent    float64         `json:"memoryUsedPercent"`
	Disks                []wireDisk      `json:"disks"`
	BandwidthRxBps       float64         `json:"bandwidthRxBps"`
	BandwidthTxBps       float64         `json:"bandwidthTxBps"`
	UptimeSeconds        uint64          `json:"uptimeSeconds"`
	System               json.RawMessage `json:"system"`
	CollectedAt          time.Time       `json:"collectedAt"`
}

type wireCore struct {
	CoreIndex    int     `json:"coreIndex"`
	UsagePercent float64 `json:"usagePercent"`
}

type wireDisk struct {
	MountPoint  string  `json:"mountPoint"`
	Fstype      string  `json:"fstype"`
	TotalBytes  uint64  `json:"totalBytes"`
	UsedBytes   uint64  `json:"usedBytes"`
	FreeBytes   uint64  `json:"freeBytes"`
	UsedPercent float64 `json:"usedPercent"`
}

func snapshotToProto(body []byte) (*agentv1.AgentSnapshot, error) {
	var snap wireSnapshot
	if err := json.Unmarshal(body, &snap); err != nil {
		return nil, err
	}
	out := &agentv1.AgentSnapshot{
		AgentId:              snap.AgentID,
		Hostname:             snap.Hostname,
		CpuPercent:           snap.CPUPercent,
		MemoryTotalBytes:     snap.MemoryTotalBytes,
		MemoryUsedBytes:      snap.MemoryUsedBytes,
		MemoryAvailableBytes: snap.MemoryAvailableBytes,
		MemoryUsedPercent:    snap.MemoryUsedPercent,
		BandwidthRxBps:       snap.BandwidthRxBps,
		BandwidthTxBps:       snap.BandwidthTxBps,
		UptimeSeconds:        snap.UptimeSeconds,
		SystemJson:           string(snap.System),
	}
	for _, c := range snap.Cores {
		out.Cores = append(out.Cores, &agentv1.CoreSample{CoreIndex: int32(c.CoreIndex), UsagePercent: c.UsagePercent})
	}
	for _, d := range snap.Disks {
		out.Disks = append(out.Disks, &agentv1.DiskSample{
			MountPoint: d.MountPoint, Fstype: d.Fstype,
			TotalBytes: d.TotalBytes, UsedBytes: d.UsedBytes, FreeBytes: d.FreeBytes, UsedPercent: d.UsedPercent,
		})
	}
	if !snap.CollectedAt.IsZero() {
		out.CollectedAt = timestamppb.New(snap.CollectedAt.UTC())
	}
	return out, nil
}
