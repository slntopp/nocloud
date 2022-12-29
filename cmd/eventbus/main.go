package main

import (
	"fmt"
	"net"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	amqp "github.com/rabbitmq/amqp091-go"

	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	pb "github.com/slntopp/nocloud-proto/events"
	healthpb "github.com/slntopp/nocloud-proto/health"
	"github.com/slntopp/nocloud/pkg/eventbus"
	"github.com/slntopp/nocloud/pkg/nocloud"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

var (
	log  *zap.Logger
	port string
	rbmq string
)

func init() {

	viper.AutomaticEnv()
	log = nocloud.NewLogger()

	viper.SetDefault("PORT", "8000")
	viper.SetDefault("RABBITMQ_CONN", "amqp://nocloud:secret@rabbitmq:5672/")

	rbmq = viper.GetString("RABBITMQ_CONN")
	port = viper.GetString("PORT")
}

func main() {

	defer func() {
		_ = log.Sync()
	}()

	log.Info("setting up eventbus server")

	s := grpc.NewServer(
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			grpc_zap.UnaryServerInterceptor(log),
		)),
	)

	log.Info("dialing rbmq")
	conn, err := amqp.Dial(rbmq)
	if err != nil {
		log.Fatal("failed to connect to RabbitMQ", zap.Error(err))
	}
	defer conn.Close()

	server := eventbus.NewServer(log, conn)
	pb.RegisterEventsServiceServer(s, server)

	healthpb.RegisterInternalProbeServiceServer(s, NewHealthServer(log))

	lis, err := net.Listen("tcp", fmt.Sprintf(":%v", port))
	if err != nil {
		log.Fatal("failed to listen", zap.String("address", port), zap.Error(err))
	}

	log.Info("serving server")
	log.Fatal("failed to serve gRPC", zap.Error(s.Serve(lis)))
}
