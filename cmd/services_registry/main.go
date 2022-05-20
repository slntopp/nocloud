/*
Copyright © 2021-2022 Nikita Ivanovski info@slnt-opp.xyz

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
	"net"

	bpb "github.com/slntopp/nocloud/pkg/billing/proto"
	driverpb "github.com/slntopp/nocloud/pkg/drivers/instance/vanilla"
	healthpb "github.com/slntopp/nocloud/pkg/health/proto"
	"github.com/slntopp/nocloud/pkg/instances"
	ipb "github.com/slntopp/nocloud/pkg/instances/proto"
	"github.com/slntopp/nocloud/pkg/nocloud"
	"github.com/slntopp/nocloud/pkg/nocloud/auth"
	"github.com/slntopp/nocloud/pkg/nocloud/connectdb"
	"github.com/slntopp/nocloud/pkg/nocloud/schema"
	"github.com/slntopp/nocloud/pkg/services"
	pb "github.com/slntopp/nocloud/pkg/services/proto"
	stpb "github.com/slntopp/nocloud/pkg/settings/proto"
	"github.com/spf13/viper"
	"github.com/streadway/amqp"
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

	viper.SetDefault("DB_HOST", "db:8529")
	viper.SetDefault("DB_CRED", "root:openSesame")
	viper.SetDefault("DRIVERS", "")
	viper.SetDefault("SIGNING_KEY", "seeeecreet")
	viper.SetDefault("RABBITMQ_CONN", "amqp://nocloud:secret@rabbitmq:5672/")
	viper.SetDefault("SETTINGS_HOST", "settings:8080")
	viper.SetDefault("BILLING_HOST", "settings:8080")

	port = viper.GetString("PORT")

	arangodbHost = viper.GetString("DB_HOST")
	arangodbCred = viper.GetString("DB_CRED")
	drivers = viper.GetStringSlice("DRIVERS")
	SIGNING_KEY = []byte(viper.GetString("SIGNING_KEY"))
	rbmq = viper.GetString("RABBITMQ_CONN")
	settingsHost = viper.GetString("SETTINGS_HOST")
	billingHost = viper.GetString("BILLING_HOST")
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

	auth.SetContext(log, SIGNING_KEY)
	s := grpc.NewServer(
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			grpc_zap.UnaryServerInterceptor(log),
			grpc.UnaryServerInterceptor(auth.JWT_AUTH_INTERCEPTOR),
		)),
	)

	log.Info("Dialing RabbitMQ", zap.String("url", rbmq))
	rbmq, err := amqp.Dial(rbmq)
	if err != nil {
		log.Fatal("Failed to connect to RabbitMQ", zap.Error(err))
	}
	defer rbmq.Close()

	server := services.NewServicesServer(log, db)
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
