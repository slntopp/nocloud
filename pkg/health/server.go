package health

import (
	"context"

	"github.com/slntopp/nocloud/pkg/health/healthpb"
	"go.uber.org/zap"
)

type HealthServiceServer struct{
	healthpb.UnimplementedHealthServiceServer
	log *zap.Logger
}

func NewServer(log *zap.Logger) *HealthServiceServer {
	return &HealthServiceServer{log: log}
}

func (s *HealthServiceServer) Probe(ctx context.Context, request *healthpb.ProbeRequest) (*healthpb.ProbeResponse, error) {
	log := s.log.Named("Health Probe")
	log.Info("Probe received", zap.String("Type", request.ProbeType))
	if request.ProbeType == "PING" {
		return &healthpb.ProbeResponse{Response: "PONG"}, nil
	}

	return &healthpb.ProbeResponse{Response: "ok"}, nil
}
