package main

import (
	"context"
	"github.com/go-redis/redis/v8"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	amqp "github.com/rabbitmq/amqp091-go"
	pb "github.com/slntopp/nocloud-proto/events"
	healthpb "github.com/slntopp/nocloud-proto/health"
	"github.com/slntopp/nocloud/pkg/eventbus"
	"github.com/slntopp/nocloud/pkg/nocloud"
	"github.com/slntopp/nocloud/pkg/nocloud/auth"
	"github.com/slntopp/nocloud/pkg/nocloud/connectdb"
	grpc_server "github.com/slntopp/nocloud/pkg/nocloud/grpc"
	"github.com/slntopp/nocloud/pkg/nocloud/rabbitmq"
	"github.com/slntopp/nocloud/pkg/nocloud/schema"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"sync"
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
	workers := &sync.WaitGroup{}

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
	ctx, cancel := context.WithCancel(ctx)

	server := eventbus.NewServer(log, rabbitmq.NewRabbitMQConnection(conn), db)
	go server.ListenBusQueue(ctx, worker(workers))
	pb.RegisterEventsServiceServer(s, server)

	healthpb.RegisterInternalProbeServiceServer(s, NewHealthServer(log))

	grpc_server.ServeGRPC(log, s, port)
	log.Info("Stopping workers.")
	cancel()
	workers.Wait()
	log.Info("All workers were stopped.")
}

func worker(wg *sync.WaitGroup) *sync.WaitGroup {
	wg.Add(1)
	return wg
}
