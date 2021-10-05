package main

import (
	"fmt"
	"net"

	"github.com/spf13/viper"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	inflog "github.com/infinimesh/infinimesh/pkg/log"
	"github.com/slntopp/ione-go/pkg/health"
	"github.com/slntopp/ione-go/pkg/health/healthpb"
)

var (
	port string
	log *zap.Logger
)

func init() {
	logger, err := inflog.NewProdOrDev()
	if err != nil {
		panic(err)
	}
	log = logger

	viper.AutomaticEnv()
	viper.SetDefault("PORT", "8080")

	port = viper.GetString("PORT")
}

func main() {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%v", port))
	if err != nil {
		log.Fatal("Failed to listen", zap.String("address", port), zap.Error(err))
	}

	server := health.NewServer(log)
	s := grpc.NewServer()
	
	healthpb.RegisterHealthServiceServer(s, server)
	log.Info(fmt.Sprintf("Serving gRPC on 0.0.0.0:%v", port), zap.Skip())
	log.Fatal("Failed to serve gRPC", zap.Error(s.Serve(lis)))
}