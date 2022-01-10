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

	driverpb "github.com/slntopp/nocloud/pkg/drivers/instance/vanilla"
	"github.com/slntopp/nocloud/pkg/nocloud"
	"github.com/slntopp/nocloud/pkg/nocloud/auth"
	"github.com/slntopp/nocloud/pkg/nocloud/connectdb"
	sp "github.com/slntopp/nocloud/pkg/services_providers"
	sppb "github.com/slntopp/nocloud/pkg/services_providers/proto"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	"google.golang.org/grpc"
)

var (
	port string
	log *zap.Logger

	arangodbHost 	string
	arangodbCred 	string
	drivers 		[]string
	ext_servers 	[]string
	SIGNING_KEY		[]byte
)

func init() {
	viper.AutomaticEnv()
	log = nocloud.NewLogger()

	viper.SetDefault("PORT", "8080")

	viper.SetDefault("DB_HOST", "db:8529")
	viper.SetDefault("DB_CRED", "root:openSesame")
	viper.SetDefault("DRIVERS", "")
	viper.SetDefault("EXTENTION_SERVERS", "")
	viper.SetDefault("SIGNING_KEY", "seeeecreet")

	port = viper.GetString("PORT")

	arangodbHost 	= viper.GetString("DB_HOST")
	arangodbCred 	= viper.GetString("DB_CRED")
	drivers 		= viper.GetStringSlice("DRIVERS")
	ext_servers 	= viper.GetStringSlice("EXTENTION_SERVERS")
	SIGNING_KEY 	= []byte(viper.GetString("SIGNING_KEY"))
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
	
	server := sp.NewServicesProviderServer(log, db)

	for _, driver := range drivers {
		log.Info("Registering Driver", zap.String("driver", driver))
		conn, err := grpc.Dial(driver, grpc.WithInsecure())
		if err != nil {
			log.Error("Error registering driver", zap.String("driver", driver), zap.Error(err))
			continue
		}
		client := driverpb.NewDriverServiceClient(conn)
		driver_type, err := client.GetType(context.Background(), &driverpb.GetTypeRequest{})
		if err != nil {
			log.Error("Error dialing driver and getting its type", zap.String("driver", driver), zap.Error(err))
		}
		server.RegisterDriver(driver_type.GetType(), client)
		log.Info("Registered Driver", zap.String("driver", driver), zap.String("type", driver_type.GetType()))
	}

	for _, ext_server := range ext_servers {
		log.Info("Registering Extention Server", zap.String("ext_server", ext_server))
		conn, err := grpc.Dial(ext_server, grpc.WithInsecure())
		if err != nil {
			log.Error("Error registering Extention Server", zap.String("ext_server", ext_server), zap.Error(err))
			continue
		}
		client := sppb.NewServicesProvidersExtentionsServiceClient(conn)
		ext_srv_type, err := client.GetType(context.Background(),&sppb.GetTypeRequest{})
		if err != nil {
			log.Error("Error dialing Extention Server and getting its type", zap.String("ext_server", ext_server), zap.Error(err))
		}
		server.RegisterExtentionServer(ext_srv_type.GetType(), client)
		log.Info("Registered Extention Server", zap.String("ext_server", ext_server), zap.String("type", ext_srv_type.GetType()))
	}

	sppb.RegisterServicesProvidersServiceServer(s, server)

	log.Info(fmt.Sprintf("Serving gRPC on 0.0.0.0:%v", port), zap.Skip())
	log.Fatal("Failed to serve gRPC", zap.Error(s.Serve(lis)))
}