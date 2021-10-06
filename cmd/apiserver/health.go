package main

import (
	"context"

	"github.com/slntopp/nocloud/pkg/api/apipb"
	"github.com/slntopp/nocloud/pkg/health/healthpb"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type healthAPI struct {
	client healthpb.HealthServiceClient
	apipb.UnimplementedHealthServiceServer
}

func (h *healthAPI) mustEmbedUnimplementedHealthServiceServer() {
	log.Info("Method missing")
}

func (h *healthAPI) Probe(ctx context.Context, request *healthpb.ProbeRequest) (response *healthpb.ProbeResponse, err error) {
	log.Info("IONe Health", zap.String("Probe", request.ProbeType))

	res, err := h.client.Probe(ctx, request)
	if err != nil {
		log.Error("Probe Failed", zap.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}
	return res, nil
}
