package main

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/slntopp/nocloud/pkg/nocloud/schema"
	"net"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	amqp "github.com/rabbitmq/amqp091-go"

	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	pb "github.com/slntopp/nocloud-proto/events"
	healthpb "github.com/slntopp/nocloud-proto/health"
	"github.com/slntopp/nocloud/pkg/eventbus"
	"github.com/slntopp/nocloud/pkg/nocloud"
	"github.com/slntopp/nocloud/pkg/nocloud/auth"
	"github.com/slntopp/nocloud/pkg/nocloud/connectdb"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

var (
	log *zap.Logger

	arangodbHost string
	arangodbCred string
	arangodbName string

	SIGNING_KEY []byte

	redisHost string
	port      string
	rbmq      string
)

func init() {

	viper.AutomaticEnv()
	log = nocloud.NewLogger()

	viper.SetDefault("PORT", "8000")

	viper.SetDefault("DB_HOST", "db:8529")
	viper.SetDefault("DB_CRED", "root:openSesame")
	viper.SetDefault("DB_NAME", schema.DB_NAME)
	viper.SetDefault("REDIS_HOST", "redis:6379")
	viper.SetDefault("SIGNING_KEY", "seeeecreet")
	viper.SetDefault("RABBITMQ_CONN", "amqp://nocloud:secret@rabbitmq:5672/")

	port = viper.GetString("PORT")

	arangodbHost = viper.GetString("DB_HOST")
	arangodbCred = viper.GetString("DB_CRED")
	arangodbName = viper.GetString("DB_NAME")
	redisHost = viper.GetString("REDIS_HOST")

	SIGNING_KEY = []byte(viper.GetString("SIGNING_KEY"))
	rbmq = viper.GetString("RABBITMQ_CONN")

}

func main() {

	defer func() {
		_ = log.Sync()
	}()

	log.Info("Setting up DB Connection")
	db := connectdb.MakeDBConnection(log, arangodbHost, arangodbCred, arangodbName)
	log.Info("DB connection established")

	log.Info("Setting up RabbitMQ Connection")
	conn, err := amqp.Dial(rbmq)
	if err != nil {
		log.Fatal("failed to connect to RabbitMQ", zap.Error(err))
	}
	defer conn.Close()
	log.Info("RabbitMQ connection established")

	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		log.Fatal("Failed to listen", zap.String("address", port), zap.Error(err))
	}

	rdb := redis.NewClient(&redis.Options{
		Addr: redisHost,
		DB:   0,
	})

	auth.SetContext(log, rdb, SIGNING_KEY)
	s := grpc.NewServer(
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			grpc_zap.UnaryServerInterceptor(log),
			grpc.UnaryServerInterceptor(auth.JWT_AUTH_INTERCEPTOR),
		)),
		grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(
			grpc.StreamServerInterceptor(auth.JWT_STREAM_INTERCEPTOR),
		)),
	)

	ctx := context.WithValue(context.Background(), nocloud.NoCloudAccount, schema.ROOT_ACCOUNT_KEY)

	server := eventbus.NewServer(log, conn, db)
	go server.ListenBusQueue(ctx)
	pb.RegisterEventsServiceServer(s, server)

	healthpb.RegisterInternalProbeServiceServer(s, NewHealthServer(log))

	log.Info(fmt.Sprintf("Serving gRPC on 0.0.0.0:%s", port), zap.Skip())
	log.Fatal("Failed to serve gRPC", zap.Error(s.Serve(lis)))
}
