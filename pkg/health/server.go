package health

import (
	"context"

	"github.com/slntopp/ione-go/pkg/health/healthpb"
)

type HealthServiceServer struct{}

func NewServer() *HealthServiceServer {
	return &HealthServiceServer{}
}

func (*HealthServiceServer) Probe(ctx context.Context, request *healthpb.ProbeRequest) (*healthpb.ProbeResponse, error) {
	if request.ProbeType == "PING" {
		return &healthpb.ProbeResponse{Response: "PONG"}, nil
	}

	return &healthpb.ProbeResponse{Response: "ok"}, nil
}
