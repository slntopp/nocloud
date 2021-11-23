/*
Copyright © 2021 Nikita Ivanovski info@slnt-opp.xyz

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
	"net"
	"net/http"

	"github.com/gorilla/handlers"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/viper"

	"github.com/slntopp/nocloud/pkg/accounting/accountspb"
	"github.com/slntopp/nocloud/pkg/accounting/namespacespb"
	apipb "github.com/slntopp/nocloud/pkg/api/apipb"
	"github.com/slntopp/nocloud/pkg/health/healthpb"
	"github.com/slntopp/nocloud/pkg/nocloud"
	servicespb "github.com/slntopp/nocloud/pkg/services/proto"
	sppb "github.com/slntopp/nocloud/pkg/services_providers/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

var (
	log 			*zap.Logger

	healthHost 		string
	registryHost 	string
	servicesHost	string
	spRegistryHost  string

	corsAllowed 	[]string

	SIGNING_KEY		[]byte
)

type server struct{}

func NewServer() *server {
	return &server{}
}

func init() {
	viper.AutomaticEnv()
	log = nocloud.NewLogger()

	viper.SetDefault("CORS_ALLOWED", []string{"*"})

	viper.SetDefault("HEALTH_HOST", "health:8080")
	viper.SetDefault("REGISTRY_HOST", "accounts:8080")
	viper.SetDefault("SP_REGISTRY_HOST", "sp-registry:8080")
	viper.SetDefault("SERVICES_HOST", "services-registry:8080")
	
	viper.SetDefault("SIGNING_KEY", "seeeecreet")

	corsAllowed 	= viper.GetStringSlice("CORS_ALLOWED")

	healthHost 		= viper.GetString("HEALTH_HOST")
	registryHost 	= viper.GetString("REGISTRY_HOST")
	servicesHost 	= viper.GetString("SERVICES_HOST")
	spRegistryHost 	= viper.GetString("SP_REGISTRY_HOST")

	SIGNING_KEY 	= []byte(viper.GetString("SIGNING_KEY"))
}

func main() {
	defer func() {
		_ = log.Sync()
	}()

	log.Info("Connecting to HealthService", zap.String("host", healthHost))
	healthConn, err := grpc.Dial(healthHost, grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	healthClient := healthpb.NewHealthServiceClient(healthConn)

	log.Info("Connecting to AccountsService", zap.String("host", registryHost))
	registryConn, err := grpc.Dial(registryHost, grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	accountsClient := accountspb.NewAccountsServiceClient(registryConn)
	namespacesClient := namespacespb.NewNamespacesServiceClient(registryConn)

	log.Info("Connecting to ServicesProvidersService", zap.String("host", spRegistryHost))
	spRegistryConn, err := grpc.Dial(spRegistryHost, grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	spRegistryClient := sppb.NewServicesProvidersServiceClient(spRegistryConn)

	log.Info("Connecting to ServicesService", zap.String("host", spRegistryHost))
	servicesConn, err := grpc.Dial(servicesHost, grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	servicesClient := servicespb.NewServicesServiceClient(servicesConn)

	// Create a listener on TCP port
	lis, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatal("Failed to listen:", zap.Error(err))
	}

	// Create a gRPC server object
	s := grpc.NewServer(grpc.UnaryInterceptor(JWT_AUTH_INTERCEPTOR),)
	apipb.RegisterHealthServiceServer(s, &healthAPI{client: healthClient})

	apipb.RegisterAccountsServiceServer(s, &accountsAPI{client: accountsClient})
	apipb.RegisterNamespacesServiceServer(s, &namespacesAPI{client: namespacesClient})

	apipb.RegisterServicesProvidersServiceServer(s, &spRegistryAPI{client: spRegistryClient, log: log.Named("ServicesProvidersRegistry")})

	apipb.RegisterServicesServiceServer(s, &servicesAPI{client: servicesClient, log: log.Named("ServicesRegistry")})

	// Serve gRPC Server
	log.Info("Serving gRPC on 0.0.0.0:8080", zap.Skip())
	go func() {
		log.Fatal("Error", zap.Error(s.Serve(lis)))
	}()

	// Set up REST API server
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
	err = apipb.RegisterAccountsServiceHandler(context.Background(), gwmux, conn)
	if err != nil {
		log.Fatal("Failed to register AccountsService gateway", zap.Error(err))
	}
	err = apipb.RegisterNamespacesServiceHandler(context.Background(), gwmux, conn)
	if err != nil {
		log.Fatal("Failed to register NamespacesService gateway", zap.Error(err))
	}
	err = apipb.RegisterServicesProvidersServiceHandler(context.Background(), gwmux, conn)
	if err != nil {
		log.Fatal("Failed to register ServicesProvidersService gateway", zap.Error(err))
	}
	err = apipb.RegisterServicesServiceHandler(context.Background(), gwmux, conn)
	if err != nil {
		log.Fatal("Failed to register ServicesService gateway", zap.Error(err))
	}

	handler := handlers.CORS(
		handlers.AllowedOrigins(corsAllowed),
		handlers.AllowedHeaders([]string{"Content-Type"}),
		handlers.AllowedMethods([]string{"GET", "POST", "DELETE", "PUT", "PATCH", "OPTIONS", "HEAD"}),
	)(gwmux)

	log.Info("Serving gRPC-Gateway on http://0.0.0.0:8000")
	log.Fatal("Failed to Listen and Serve Gateway-Server", zap.Error(http.ListenAndServe(":8000", handler)))
}