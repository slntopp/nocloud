package main

import (
	"context"

	pb "github.com/slntopp/nocloud/pkg/health/proto"
	"github.com/slntopp/nocloud/pkg/nocloud"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	log *zap.Logger
)

func init() {
	log = nocloud.NewLogger()
}

func main() {
	conn, err := grpc.Dial("localhost:8000", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal("Error dialing service", zap.Error(err))
	}

	client := pb.NewInternalProbeServiceClient(conn)
	r, err := client.Service(context.Background(), &pb.ProbeRequest{
		ProbeType: "services",
	})
	if err != nil {
		log.Fatal("Error performing probe check", zap.Error(err))
	}

	if r.Status != pb.Status_SERVING {
		log.Fatal("Status is not Serving", zap.Any("probe_response", r))
	}
}
