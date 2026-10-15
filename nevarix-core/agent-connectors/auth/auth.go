// Package auth checks the shared agent ingest token for both REST and gRPC.
package auth

import (
	"context"
	"crypto/subtle"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const agentServicePrefix = "/nevarix.agent.v1.AgentService/"

// BearerMatches reports whether an Authorization header carries the token. An empty token never matches.
func BearerMatches(header, token string) bool {
	const prefix = "Bearer "
	if token == "" || !strings.HasPrefix(header, prefix) {
		return false
	}
	got := strings.TrimSpace(strings.TrimPrefix(header, prefix))
	return subtle.ConstantTimeCompare([]byte(got), []byte(token)) == 1
}

// UnaryInterceptor guards AgentService methods; every other gRPC method passes through.
func UnaryInterceptor(token string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if !strings.HasPrefix(info.FullMethod, agentServicePrefix) {
			return handler(ctx, req)
		}
		md, _ := metadata.FromIncomingContext(ctx)
		for _, v := range md.Get("authorization") {
			if BearerMatches(v, token) {
				return handler(ctx, req)
			}
		}
		return nil, status.Error(codes.Unauthenticated, "unauthorized")
	}
}
