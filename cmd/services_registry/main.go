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
	"fmt"
	"github.com/go-redis/redis/v8"
	"net"

	amqp "github.com/rabbitmq/amqp091-go"
	bpb "github.com/slntopp/nocloud-proto/billing"
	driverpb "github.com/slntopp/nocloud-proto/drivers/instance/vanilla"
	healthpb "github.com/slntopp/nocloud-proto/health"
	ipb "github.com/slntopp/nocloud-proto/instances"
	pb "github.com/slntopp/nocloud-proto/services"
	stpb "github.com/slntopp/nocloud-proto/settings"
	"github.com/slntopp/nocloud/pkg/instances"
	"github.com/slntopp/nocloud/pkg/nocloud"
	"github.com/slntopp/nocloud/pkg/nocloud/auth"
	"github.com/slntopp/nocloud/pkg/nocloud/connectdb"
	"github.com/slntopp/nocloud/pkg/nocloud/schema"
	"github.com/slntopp/nocloud/pkg/services"
	"github.com/slntopp/nocloud/pkg/states"
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
	redisHost    string
	drivers      []string
	SIGNING_KEY  []byte
	rbmq         string
	settingsHost string
	billingHost  string
)

func init() {
	viper.AutomaticEnv()
	log = nocloud.NewLogger()

	viper.SetDefault("PORT", "8000")

	viper.SetDefault("REDIS_HOST", "redis:6379")
	viper.SetDefault("DB_HOST", "db:8529")
	viper.SetDefault("DB_CRED", "root:openSesame")
	viper.SetDefault("DRIVERS", "")
	viper.SetDefault("SIGNING_KEY", "seeeecreet")
	viper.SetDefault("RABBITMQ_CONN", "amqp://nocloud:secret@rabbitmq:5672/")
	viper.SetDefault("SETTINGS_HOST", "settings:8000")
	viper.SetDefault("BILLING_HOST", "billing:8000")

	port = viper.GetString("PORT")

	arangodbHost = viper.GetString("DB_HOST")
	arangodbCred = viper.GetString("DB_CRED")
	drivers = viper.GetStringSlice("DRIVERS")
	SIGNING_KEY = []byte(viper.GetString("SIGNING_KEY"))
	rbmq = viper.GetString("RABBITMQ_CONN")
	settingsHost = viper.GetString("SETTINGS_HOST")
	billingHost = viper.GetString("BILLING_HOST")
	redisHost = viper.GetString("REDIS_HOST")
}

func main() {
	defer func() {
		_ = log.Sync()
	}()

	log.Info("Setting up DB Connection")
	db := connectdb.MakeDBConnection(log, arangodbHost, arangodbCred)
	log.Info("DB connection established")

	lis, err := net.Listen("tcp", fmt.Sprintf(":%v", port))
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

	log.Info("Dialing RabbitMQ", zap.String("url", rbmq))
	rbmq, err := amqp.Dial(rbmq)
	if err != nil {
		log.Fatal("Failed to connect to RabbitMQ", zap.Error(err))
	}
	defer rbmq.Close()

	log.Info("Setting up Pub/Sub")
	ps, err := states.SetupStatesStreaming()
	if err != nil {
		log.Fatal("Failed to setup states streaming", zap.Error(err))
	}
	log.Info("Pub/Sub setted up")

	server := services.NewServicesServer(log, db, ps, rbmq)
	iserver := instances.NewInstancesServiceServer(log, db, rbmq)

	for _, driver := range drivers {
		log.Info("Registering Driver", zap.String("driver", driver))
		conn, err := grpc.Dial(driver, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Fatal("Error registering driver", zap.String("driver", driver), zap.Error(err))
		}
		client := driverpb.NewDriverServiceClient(conn)
		driver_type, err := client.GetType(context.Background(), &driverpb.GetTypeRequest{})
		if err != nil {
			log.Fatal("Error dialing driver and getting its type", zap.String("driver", driver), zap.Error(err))
		}
		server.RegisterDriver(driver_type.GetType(), client)
		iserver.RegisterDriver(driver_type.GetType(), client)
		log.Info("Registered Driver", zap.String("driver", driver), zap.String("type", driver_type.GetType()))
	}

	log.Info("Registering Settings Service", zap.String("url", settingsHost))
	setconn, err := grpc.Dial(settingsHost, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	defer setconn.Close()

	setc := stpb.NewSettingsServiceClient(setconn)
	token, err := auth.MakeToken(schema.ROOT_ACCOUNT_KEY)
	if err != nil {
		log.Fatal("Can't generate token", zap.Error(err))
	}
	server.SetupSettingsClient(setc, token)
	log.Info("Settings Service registered")

	log.Info("Registering Billing Service", zap.String("url", billingHost))
	billconn, err := grpc.Dial(billingHost, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	defer billconn.Close()

	billc := bpb.NewBillingServiceClient(billconn)
	server.SetupBillingClient(billc)
	log.Info("Billing Service registered")

	pb.RegisterServicesServiceServer(s, server)
	ipb.RegisterInstancesServiceServer(s, iserver)

	healthpb.RegisterInternalProbeServiceServer(s, NewHealthServer(log, server))

	log.Info(fmt.Sprintf("Serving gRPC on 0.0.0.0:%v", port), zap.Skip())
	log.Fatal("Failed to serve gRPC", zap.Error(s.Serve(lis)))
}
