package main

import (
	"context"
	"net"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/viper"

	inflog "github.com/infinimesh/infinimesh/pkg/log"
	apipb "github.com/slntopp/nocloud/pkg/api/apipb"
	"github.com/slntopp/nocloud/pkg/health/healthpb"
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

func init() {
	logger, err := inflog.NewProdOrDev()
	if err != nil {
		panic(err)
	}
	log = logger

	viper.AutomaticEnv()
	viper.SetDefault("HEALTH_HOST", "health:8080")

	healthHost = viper.GetString("HEALTH_HOST")
}

func main() {

	defer func() {
		_ = log.Sync()
	}()

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
	log.Info("Serving gRPC on 0.0.0.0:8080", zap.Skip())
	go func() {
		log.Fatal("Error", zap.Error(s.Serve(lis)))
	}()

	conn, err := grpc.DialContext(
		context.Background(),
		"0.0.0.0:8080",
		grpc.WithBlock(),
		grpc.WithInsecure(),
	)
	if err != nil {
		log.Fatal("Failed to dial server:", zap.Error(err))
	}

	gwmux := runtime.NewServeMux()
	err = apipb.RegisterHealthServiceHandler(context.Background(), gwmux, conn)
	if err != nil {
		log.Fatal("Failed to register HealthService gateway", zap.Error(err))
	}

	gwServer := &http.Server{
		Addr:    ":8000",
		Handler: gwmux,
	}

	log.Info("Serving gRPC-Gateway on http://0.0.0.0:8000")
	log.Fatal("Failed to Listen and Serve Gateway-Server", zap.Error(gwServer.ListenAndServe()))
}