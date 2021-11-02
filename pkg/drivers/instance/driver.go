package driver

import (
	"context"

	pb "github.com/slntopp/nocloud/pkg/drivers/instance/proto"
)

var DRIVER_NAME string

func SetDriverName(name string) {
	DRIVER_NAME = name
}

type DriverServiceServer struct {
	pb.UnimplementedDriverServiceServer
}

func (s *DriverServiceServer) GetType(ctx context.Context, _ *pb.GetTypeRequest) (*pb.GetTypeResponse, error) {
	return &pb.GetTypeResponse{Type: DRIVER_NAME}, nil
}

func (s *DriverServiceServer) ValidateConfigSyntax(ctx context.Context, req *pb.ValidateConfigSyntaxRequest) (*pb.ValidateConfigSyntaxResponse, error) {
	return &pb.ValidateConfigSyntaxResponse{Result: true}, nil
}