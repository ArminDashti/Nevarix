// Package grpcapi serves the agent write path over gRPC.
package grpcapi

import (
	"context"
	"encoding/json"
	"time"

	agentv1 "github.com/ArminDashti/nevarix-core/agent-connectors/pb/nevarix/agent/v1"
	"github.com/ArminDashti/nevarix-core/core/model"
	"github.com/ArminDashti/nevarix-core/core/store"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	agentv1.UnimplementedAgentServiceServer
	store         *store.Store
	historyPoints int
}

func New(st *store.Store, historyPoints int) *Server {
	return &Server{store: st, historyPoints: historyPoints}
}

// IngestMetrics returns OK only after the snapshot is committed, matching the REST 200 contract.
func (s *Server) IngestMetrics(ctx context.Context, in *agentv1.AgentSnapshot) (*agentv1.IngestAck, error) {
	if in.GetAgentId() == "" {
		return nil, status.Error(codes.InvalidArgument, "agentId is required")
	}
	if err := s.store.IngestAgent(ctx, toModel(in), s.historyPoints); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &agentv1.IngestAck{Status: "ok"}, nil
}

func toModel(in *agentv1.AgentSnapshot) model.AgentSnapshot {
	snap := model.AgentSnapshot{
		AgentID:              in.GetAgentId(),
		Hostname:             in.GetHostname(),
		CPUPercent:           in.GetCpuPercent(),
		MemoryTotalBytes:     in.GetMemoryTotalBytes(),
		MemoryUsedBytes:      in.GetMemoryUsedBytes(),
		MemoryAvailableBytes: in.GetMemoryAvailableBytes(),
		MemoryUsedPercent:    in.GetMemoryUsedPercent(),
		BandwidthRxBps:       in.GetBandwidthRxBps(),
		BandwidthTxBps:       in.GetBandwidthTxBps(),
		UptimeSeconds:        in.GetUptimeSeconds(),
	}
	for _, c := range in.GetCores() {
		snap.Cores = append(snap.Cores, model.CoreSample{CoreIndex: int(c.GetCoreIndex()), UsagePercent: c.GetUsagePercent()})
	}
	for _, d := range in.GetDisks() {
		snap.Disks = append(snap.Disks, model.DiskSample{
			MountPoint:  d.GetMountPoint(),
			Fstype:      d.GetFstype(),
			TotalBytes:  d.GetTotalBytes(),
			UsedBytes:   d.GetUsedBytes(),
			FreeBytes:   d.GetFreeBytes(),
			UsedPercent: d.GetUsedPercent(),
		})
	}
	if in.GetSystemJson() != "" {
		snap.System = json.RawMessage(in.GetSystemJson())
	}
	if ts := in.GetCollectedAt(); ts != nil {
		snap.CollectedAt = ts.AsTime().UTC()
	} else {
		snap.CollectedAt = time.Time{}
	}
	return snap
}
