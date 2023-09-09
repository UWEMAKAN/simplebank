package gapi

import (
	"context"

	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
)

const (
	grpcGatewayUserAgentHeader = "grpcgateway-user-agent"
	xForwardedForHeader = "x-forwarded-for"
	grpcUserAgentHeader = "user-agent"
)

type Metadata struct {
	UserAgent string
	ClientIP  string
}

func (server *Server) extractMetadata(ctx context.Context) *Metadata {
	m := &Metadata{}
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if userAgents := md.Get(grpcGatewayUserAgentHeader); len(userAgents) > 0 {
			m.UserAgent = userAgents[0]
		}
		if userAgents := md.Get(grpcUserAgentHeader); len(userAgents) > 0 {
			m.UserAgent = userAgents[0]
		}
		if clientIPs := md.Get(xForwardedForHeader); len(clientIPs) > 0 {
			m.ClientIP = clientIPs[0]
		}
		if p, ok := peer.FromContext(ctx); ok {
			m.ClientIP = p.Addr.String()
		}
	}
	return m
}
