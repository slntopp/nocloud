/*
Copyright © 2021-2023 Nikita Ivanovski info@slnt-opp.xyz

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package main

import (
	"context"
	"github.com/go-redis/redis/v8"
	amqp "github.com/rabbitmq/amqp091-go"
	driverpb "github.com/slntopp/nocloud-proto/drivers/instance/vanilla"
	healthpb "github.com/slntopp/nocloud-proto/health"
	sppb "github.com/slntopp/nocloud-proto/services_providers"
	"github.com/slntopp/nocloud/pkg/nocloud"
	"github.com/slntopp/nocloud/pkg/nocloud/auth"
	"github.com/slntopp/nocloud/pkg/nocloud/connectdb"
	grpc_server "github.com/slntopp/nocloud/pkg/nocloud/grpc"
	"github.com/slntopp/nocloud/pkg/nocloud/rabbitmq"
	"github.com/slntopp/nocloud/pkg/nocloud/schema"
	sp "github.com/slntopp/nocloud/pkg/services_providers"
	"github.com/slntopp/nocloud/pkg/showcase_categories"
	"github.com/slntopp/nocloud/pkg/showcases"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	port string
	log  *zap.Logger

	arangodbHost string
	arangodbCred string
	arangodbName string
	redisHost    string
	drivers      []string
	ext_servers  []string
	SIGNING_KEY  []byte
	rbmq         string
)

func init() {
	viper.AutomaticEnv()
	log = nocloud.NewLogger()

	viper.SetDefault("PORT", "8000")

	viper.SetDefault("DB_HOST", "db:8529")
	viper.SetDefault("DB_CRED", "root:openSesame")
	viper.SetDefault("DB_NAME", schema.DB_NAME)
	viper.SetDefault("REDIS_HOST", "redis:6379")
	viper.SetDefault("DRIVERS", "")
	viper.SetDefault("EXTENTION_SERVERS", "")
	viper.SetDefault("SIGNING_KEY", "seeeecreet")
	viper.SetDefault("RABBITMQ_CONN", "amqp://nocloud:secret@rabbitmq:5672/")

	port = viper.GetString("PORT")

	arangodbHost = viper.GetString("DB_HOST")
	arangodbCred = viper.GetString("DB_CRED")
	arangodbName = viper.GetString("DB_NAME")
	redisHost = viper.GetString("REDIS_HOST")
	drivers = viper.GetStringSlice("DRIVERS")
	ext_servers = viper.GetStringSlice("EXTENTION_SERVERS")
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

	rdb := redis.NewClient(&redis.Options{
		Addr: redisHost,
		DB:   0,
	})
	if status := rdb.Ping(context.Background()); status.Err() != nil {
		log.Fatal("Failed to connect to redis", zap.Error(status.Err()))
	}
	log.Info("Redis connection established")

	auth.SetContext(log, rdb, SIGNING_KEY)
	s := grpc.NewServer(
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			grpc_zap.UnaryServerInterceptor(log),
			grpc.UnaryServerInterceptor(auth.JWT_AUTH_INTERCEPTOR),
		)),
		grpc.MaxRecvMsgSize(500*1024*1024),
		grpc.MaxSendMsgSize(500*1024*1024),
	)

	log.Info("Dialing RabbitMQ", zap.String("url", rbmq))
	conn, err := amqp.Dial(rbmq)
	if err != nil {
		log.Fatal("Failed to connect to RabbitMQ", zap.Error(err))
	}
	defer conn.Close()
	rabbitmq.FatalOnConnectionClose(log, conn)
	log.Info("RabbitMQ connection established")

	server := sp.NewServicesProviderServer(log, db, rabbitmq.NewRabbitMQConnection(conn), rdb)
	s_server := showcases.NewShowcasesServer(log, db)
	scCatServer := showcase_categories.NewCategoriesServer(log, db)

	log.Debug("Got drivers", zap.Strings("drivers", drivers))
	for _, driver := range drivers {
		log.Info("Registering Driver", zap.String("driver", driver))
		conn, err := grpc.Dial(driver, grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithDefaultCallOptions(
				grpc.MaxCallRecvMsgSize(500*1024*1024),
				grpc.MaxCallSendMsgSize(500*1024*1024),
			),
		)
		if err != nil {
			log.Fatal("Error registering driver", zap.String("driver", driver), zap.Error(err))
		}
		client := driverpb.NewDriverServiceClient(conn)
		driver_type, err := client.GetType(context.Background(), &driverpb.GetTypeRequest{})
		if err != nil {
			log.Fatal("Error dialing driver and getting its type", zap.String("driver", driver), zap.Error(err))
		}
		server.RegisterDriver(driver_type.GetType(), client)
		log.Info("Registered Driver", zap.String("driver", driver), zap.String("type", driver_type.GetType()))
	}

	log.Debug("Got extentions servers", zap.Strings("ext_servers", ext_servers))
	for _, ext_server := range ext_servers {
		log.Info("Registering Extention Server", zap.String("ext_server", ext_server))
		conn, err := grpc.Dial(ext_server, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Fatal("Error registering Extention Server", zap.String("ext_server", ext_server), zap.Error(err))
		}
		client := sppb.NewServicesProvidersExtentionsServiceClient(conn)
		ext_srv_type, err := client.GetType(context.Background(), &sppb.GetTypeRequest{})
		if err != nil {
			log.Fatal("Error dialing Extention Server and getting its type", zap.String("ext_server", ext_server), zap.Error(err))
		}
		server.RegisterExtentionServer(ext_srv_type.GetType(), client)
		log.Info("Registered Extention Server", zap.String("ext_server", ext_server), zap.String("type", ext_srv_type.GetType()))
	}
	sppb.RegisterServicesProvidersServiceServer(s, server)
	sppb.RegisterShowcasesServiceServer(s, s_server)
	sppb.RegisterShowcaseCategoriesServiceServer(s, scCatServer)

	healthpb.RegisterInternalProbeServiceServer(s, NewHealthServer(log, server))

	grpc_server.ServeGRPC(log, s, port)
}
