package main

import (
	"net"

	apipb "github.com/slntopp/ione-go/pkg/api/apipb"
	"github.com/slntopp/ione-go/pkg/health/healthpb"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

var (
	log *zap.Logger

	healthHost string
)

type server struct{}

func NewServer() *server {
	return &server{}
}

func main() {

	healthConn, err := grpc.Dial(healthHost, grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	healthClient := healthpb.NewHealthServiceClient(healthConn)

	// Create a listener on TCP port
	lis, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatal("Failed to listen:", zap.Error(err))
	}

	// Create a gRPC server object
	s := grpc.NewServer()
	// Attach the Greeter service to the server
	apipb.RegisterHealthServiceServer(s, &healthAPI{client: healthClient})
	// Serve gRPC Server
	log.Info("Serving gRPC on 0.0.0.0:8080")
	log.Fatal("Error", zap.Error(s.Serve(lis)))
}