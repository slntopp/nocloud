package main

import (
	"github.com/spf13/viper"
	"google.golang.org/grpc"

	"github.com/slntopp/ione-go/pkg/health"
	"github.com/slntopp/ione-go/pkg/health/healthpb"
)

var (
	port string
)

func init() {
	viper.AutomaticEnv()
	viper.SetDefault("PORT", "8080")

	port = viper.GetString("PORT")

	server := health.NewServer()
	s := grpc.NewServer()

	healthpb.RegisterHealthServiceServer(s, server)
}