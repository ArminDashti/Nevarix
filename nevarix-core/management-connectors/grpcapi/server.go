// Package grpcapi serves the management panel API over gRPC.
package grpcapi

import (
	"context"
	"errors"

	"github.com/ArminDashti/nevarix-core/core/model"
	"github.com/ArminDashti/nevarix-core/core/store"
	managementv1 "github.com/ArminDashti/nevarix-core/management-connectors/pb/nevarix/management/v1"
	"github.com/ArminDashti/nevarix-core/management-connectors/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Server struct {
	managementv1.UnimplementedManagementServiceServer
	svc   *service.Management
	store *store.Store
}

func New(svc *service.Management, st *store.Store) *Server {
	return &Server{svc: svc, store: st}
}

func (s *Server) Health(context.Context, *managementv1.HealthRequest) (*managementv1.HealthResponse, error) {
	return &managementv1.HealthResponse{Status: "ok"}, nil
}

func (s *Server) ListAgents(ctx context.Context, _ *managementv1.ListAgentsRequest) (*managementv1.ListAgentsResponse, error) {
	agents, err := s.store.ListAgents(ctx)
	if err != nil {
		return nil, toStatus(err)
	}
	out := &managementv1.ListAgentsResponse{}
	for _, a := range agents {
		out.Agents = append(out.Agents, &managementv1.AgentSummary{
			AgentId:           a.AgentID,
			Hostname:          a.Hostname,
			CpuPercent:        a.CPUPercent,
			MemoryUsedPercent: a.MemoryUsedPercent,
			LastSeenAt:        timestamppb.New(a.LastSeenAt),
		})
	}
	return out, nil
}

func (s *Server) ListEndpoints(ctx context.Context, _ *managementv1.ListEndpointsRequest) (*managementv1.ListEndpointsResponse, error) {
	endpoints, err := s.svc.ListEndpoints(ctx)
	if err != nil {
		return nil, toStatus(err)
	}
	out := &managementv1.ListEndpointsResponse{}
	for _, e := range endpoints {
		ep := &managementv1.Endpoint{Ip: e.IP, Hostname: e.Hostname, Source: e.Source}
		if e.LastSeenAt != nil {
			ep.LastSeenAt = timestamppb.New(*e.LastSeenAt)
		}
		out.Endpoints = append(out.Endpoints, ep)
	}
	return out, nil
}

func (s *Server) AddEndpoint(ctx context.Context, in *managementv1.AddEndpointRequest) (*managementv1.AddEndpointResponse, error) {
	if err := s.svc.AddEndpoint(ctx, in.GetIp(), in.GetHostname()); err != nil {
		return nil, toStatus(err)
	}
	return &managementv1.AddEndpointResponse{Status: "ok"}, nil
}

func (s *Server) Scan(ctx context.Context, in *managementv1.ScanRequest) (*managementv1.ScanResponse, error) {
	res, err := s.svc.Scan(ctx, service.ScanRequest{CIDR: in.GetCidr(), Port: int(in.GetPort()), TimeoutMs: int(in.GetTimeoutMs())})
	if err != nil {
		return nil, toStatus(err)
	}
	out := &managementv1.ScanResponse{Scanned: int32(res.Scanned)}
	for _, h := range res.Hits {
		out.Hits = append(out.Hits, &managementv1.ScanHit{Ip: h.IP, Port: int32(h.Port), AgentId: h.AgentID, Hostname: h.Hostname})
	}
	return out, nil
}

func (s *Server) GetSettings(ctx context.Context, _ *managementv1.GetSettingsRequest) (*managementv1.Settings, error) {
	st, err := s.svc.GetSettings(ctx)
	if err != nil {
		return nil, toStatus(err)
	}
	return toProtoSettings(st), nil
}

func (s *Server) UpdateSettings(ctx context.Context, in *managementv1.Settings) (*managementv1.Settings, error) {
	st := model.Settings{DefaultCIDR: in.GetDefaultCidr(), AgentPort: int(in.GetAgentPort()), TimeoutMs: int(in.GetTimeoutMs())}
	saved, err := s.svc.UpdateSettings(ctx, st)
	if err != nil {
		return nil, toStatus(err)
	}
	return toProtoSettings(saved), nil
}

func (s *Server) ListSoftEtherUsers(ctx context.Context, _ *managementv1.ListSoftEtherUsersRequest) (*managementv1.ListSoftEtherUsersResponse, error) {
	users, err := s.store.ListUserStats(ctx)
	if err != nil {
		return nil, toStatus(err)
	}
	out := &managementv1.ListSoftEtherUsersResponse{}
	for _, u := range users {
		out.Users = append(out.Users, &managementv1.SoftEtherUser{
			Username:             u.Username,
			ClientIp:             u.ClientIP,
			Asn:                  u.ASN,
			DownloadBytes:        u.DownloadBytes,
			UploadBytes:          u.UploadBytes,
			UsageDurationSeconds: u.UsageDurationSeconds,
		})
	}
	return out, nil
}

func (s *Server) ListSoftEtherUserSessions(ctx context.Context, in *managementv1.ListSoftEtherUserSessionsRequest) (*managementv1.ListSoftEtherUserSessionsResponse, error) {
	sessions, err := s.store.ListUserSessions(ctx, in.GetUsername())
	if err != nil {
		return nil, toStatus(err)
	}
	out := &managementv1.ListSoftEtherUserSessionsResponse{}
	for _, l := range sessions {
		sess := &managementv1.SoftEtherSession{
			ConnectedAt:     timestamppb.New(l.ConnectedAt),
			ClientIp:        l.ClientIP,
			Asn:             l.ASN,
			DownloadBytes:   l.DownloadBytes,
			UploadBytes:     l.UploadBytes,
			DurationSeconds: l.DurationSeconds,
		}
		if l.DisconnectedAt != nil {
			sess.DisconnectedAt = timestamppb.New(*l.DisconnectedAt)
		}
		out.Sessions = append(out.Sessions, sess)
	}
	return out, nil
}

func toProtoSettings(st model.Settings) *managementv1.Settings {
	return &managementv1.Settings{DefaultCidr: st.DefaultCIDR, AgentPort: int32(st.AgentPort), TimeoutMs: int32(st.TimeoutMs)}
}

// toStatus maps validation failures to InvalidArgument and everything else to Internal.
func toStatus(err error) error {
	var ve *service.ValidationError
	if errors.As(err, &ve) {
		return status.Error(codes.InvalidArgument, ve.Message)
	}
	return status.Error(codes.Internal, err.Error())
}
