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
*/package main

import (
	"context"
	"crypto/tls"
	http_server "github.com/slntopp/nocloud/pkg/nocloud/http"
	"strings"

	"github.com/rs/cors"
	billingpb "github.com/slntopp/nocloud-proto/billing"
	dnspb "github.com/slntopp/nocloud-proto/dns"
	edgepb "github.com/slntopp/nocloud-proto/edge"
	eventspb "github.com/slntopp/nocloud-proto/events"
	healthpb "github.com/slntopp/nocloud-proto/health"
	instancespb "github.com/slntopp/nocloud-proto/instances"
	registrypb "github.com/slntopp/nocloud-proto/registry"
	servicespb "github.com/slntopp/nocloud-proto/services"
	sppb "github.com/slntopp/nocloud-proto/services_providers"
	sessionspb "github.com/slntopp/nocloud-proto/sessions"
	settingspb "github.com/slntopp/nocloud-proto/settings"
	"github.com/slntopp/nocloud/pkg/nocloud"
	"github.com/tmc/grpc-websocket-proxy/wsproxy"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	log *zap.Logger

	gatewayHost string
	adminUiHost string

	apiserver        string
	corsAllowed      []string
	insecure_enabled bool
	with_block       bool
)

func init() {
	viper.AutomaticEnv()
	log = nocloud.NewLogger()

	viper.SetDefault("CORS_ALLOWED", "*")
	viper.SetDefault("APISERVER_HOST", "proxy:8000")
	viper.SetDefault("GATEWAY_HOST", ":8000")
	viper.SetDefault("ADMIN_UI_HOST", ":8080")
	viper.SetDefault("INSECURE", true)
	viper.SetDefault("WITH_BLOCK", false)

	gatewayHost = viper.GetString("GATEWAY_HOST")
	adminUiHost = viper.GetString("ADMIN_UI_HOST")

	apiserver = viper.GetString("APISERVER_HOST")
	corsAllowed = strings.Split(viper.GetString("CORS_ALLOWED"), ",")
	insecure_enabled = viper.GetBool("INSECURE")
	with_block = viper.GetBool("WITH_BLOCK")
}

func main() {
	defer func() {
		_ = log.Sync()
	}()

	log.Info("Starting REST-API Server")
	log.Info("Registering Endpoints", zap.String("server", apiserver))
	var err error

	gwmux := runtime.NewServeMux(
		runtime.WithOutgoingHeaderMatcher(func(key string) (string, bool) {
			if strings.ToLower(key) == "set-cookie" {
				return "Set-Cookie", true
			}
			return runtime.DefaultHeaderMatcher(key)
		}))
	opts := []grpc.DialOption{}
	if insecure_enabled {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	} else {
		/* #nosec */
		creds := credentials.NewTLS(&tls.Config{InsecureSkipVerify: true})
		opts = append(opts, grpc.WithTransportCredentials(creds))
	}
	if with_block {
		opts = append(opts, grpc.WithBlock())
	}
	opts = append(opts,
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(500*1024*1024),
			grpc.MaxCallSendMsgSize(500*1024*1024),
		),
	)

	// Registering gRPC services

	// HealthService
	log.Info("Registering HealthService Gateway")
	err = healthpb.RegisterHealthServiceHandlerFromEndpoint(
		context.Background(),
		gwmux,
		apiserver,
		opts,
	)
	if err != nil {
		log.Fatal("Failed to register HealthService gateway", zap.Error(err))
	}

	// AccountsService
	log.Info("Registering AccountsService Gateway")
	err = sessionspb.RegisterSessionsServiceHandlerFromEndpoint(
		context.Background(),
		gwmux,
		apiserver,
		opts,
	)
	if err != nil {
		log.Fatal("Failed to register AccountsService gateway", zap.Error(err))
	}

	// AccountsService
	log.Info("Registering AccountsService Gateway")
	err = registrypb.RegisterAccountsServiceHandlerFromEndpoint(
		context.Background(),
		gwmux,
		apiserver,
		opts,
	)
	if err != nil {
		log.Fatal("Failed to register AccountsService gateway", zap.Error(err))
	}

	// NamespacesService
	log.Info("Registering NamespacesService Gateway")
	err = registrypb.RegisterNamespacesServiceHandlerFromEndpoint(
		context.Background(),
		gwmux,
		apiserver,
		opts,
	)
	if err != nil {
		log.Fatal("Failed to register NamespacesService gateway", zap.Error(err))
	}

	// ServicesProvidersService
	log.Info("Registering ServicesProvidersService Gateway")
	err = sppb.RegisterServicesProvidersServiceHandlerFromEndpoint(
		context.Background(),
		gwmux,
		apiserver,
		opts,
	)
	if err != nil {
		log.Fatal("Failed to register ServicesProvidersService gateway", zap.Error(err))
	}

	// ShowcasesService
	log.Info("Registering ShowcasesService Gateway")
	err = sppb.RegisterShowcasesServiceHandlerFromEndpoint(
		context.Background(),
		gwmux,
		apiserver,
		opts,
	)
	if err != nil {
		log.Fatal("Failed to register ServicesProvidersService gateway", zap.Error(err))
	}

	// ShowcaseCategoriesService
	log.Info("Registering ShowcaseCategoriesService Gateway")
	err = sppb.RegisterShowcaseCategoriesServiceHandlerFromEndpoint(
		context.Background(),
		gwmux,
		apiserver,
		opts,
	)
	if err != nil {
		log.Fatal("Failed to register ShowcaseCategoriesService gateway", zap.Error(err))
	}

	// AccountGroupsService
	log.Info("Registering AccountGroupsService Gateway")
	err = registrypb.RegisterAccountGroupsServiceHandlerFromEndpoint(
		context.Background(),
		gwmux,
		apiserver,
		opts,
	)
	if err != nil {
		log.Fatal("Failed to register AccountGroupsService gateway", zap.Error(err))
	}

	// ServicesService
	log.Info("Registering ServicesService Gateway")
	err = servicespb.RegisterServicesServiceHandlerFromEndpoint(
		context.Background(),
		gwmux,
		apiserver,
		opts,
	)
	if err != nil {
		log.Fatal("Failed to register ServicesService gateway", zap.Error(err))
	}

	// InstancesService
	log.Info("Registering InstancesService Gateway")
	err = instancespb.RegisterInstancesServiceHandlerFromEndpoint(
		context.Background(),
		gwmux,
		apiserver,
		opts,
	)
	if err != nil {
		log.Fatal("Failed to register InstancesService gateway", zap.Error(err))
	}

	// DNS
	log.Info("Registering DNS Gateway")
	err = dnspb.RegisterDNSHandlerFromEndpoint(context.Background(), gwmux, apiserver, opts)
	if err != nil {
		log.Fatal("Failed to register DNS gateway", zap.Error(err))
	}

	// SettingsService
	log.Info("Registering SettingsService Gateway")
	err = settingspb.RegisterSettingsServiceHandlerFromEndpoint(
		context.Background(),
		gwmux,
		apiserver,
		opts,
	)
	if err != nil {
		log.Fatal("Failed to register SettingsService gateway", zap.Error(err))
	}

	// BillingService
	log.Info("Registering BillingService Gateway")
	err = billingpb.RegisterBillingServiceHandlerFromEndpoint(
		context.Background(),
		gwmux,
		apiserver,
		opts,
	)
	if err != nil {
		log.Fatal("Failed to register BillingService gateway", zap.Error(err))
	}

	log.Info("Registering BillingService Gateway")
	err = billingpb.RegisterAddonsServiceHandlerFromEndpoint(
		context.Background(),
		gwmux,
		apiserver,
		opts,
	)
	if err != nil {
		log.Fatal("Failed to register BillingService gateway", zap.Error(err))
	}

	log.Info("Registering BillingService Gateway")
	err = billingpb.RegisterDescriptionsServiceHandlerFromEndpoint(
		context.Background(),
		gwmux,
		apiserver,
		opts,
	)
	if err != nil {
		log.Fatal("Failed to register BillingService gateway", zap.Error(err))
	}

	// CurrencyService
	log.Info("Registering CurrencyService Gateway")
	err = billingpb.RegisterCurrencyServiceHandlerFromEndpoint(
		context.Background(),
		gwmux,
		apiserver,
		opts,
	)
	if err != nil {
		log.Fatal("Failed to register CurrencyService gateway", zap.Error(err))
	}

	// EventsBus
	log.Info("Registering EventsBus Gateway")
	err = eventspb.RegisterEventsServiceHandlerFromEndpoint(
		context.Background(),
		gwmux,
		apiserver,
		opts,
	)
	if err != nil {
		log.Fatal("Failed to register CurrencyService gateway", zap.Error(err))
	}

	// Edge
	log.Info("Registering Edge Gateway")
	err = edgepb.RegisterEdgeServiceHandlerFromEndpoint(
		context.Background(),
		gwmux,
		apiserver,
		opts,
	)
	if err != nil {
		log.Fatal("Failed to register EdgeServices gateway", zap.Error(err))
	}

	log.Info("Allowed Origins", zap.Strings("hosts", corsAllowed))
	handler := cors.New(cors.Options{
		AllowedOrigins: corsAllowed,
		AllowedHeaders: []string{"*", "Connect-Protocol-Version", "grpc-metadata-nocloud-primary-currency-code", "NoCloud-Primary-Currency-Code", "NoCloud-Primary-Currency-Precision-Override",
			"grpc-metadata-nocloud-primary-currency-precision-override", "nocloud-primary-currency-precision-override"},
		AllowedMethods:   []string{"GET", "POST", "DELETE", "PUT", "PATCH", "OPTIONS", "HEAD"},
		AllowCredentials: true,
	}).Handler(gwmux)

	// AdminUI Handler
	ui_handler := cors.New(cors.Options{
		AllowedOrigins: corsAllowed,
		AllowedHeaders: []string{"*", "Connect-Protocol-Version", "grpc-metadata-nocloud-primary-currency-code", "NoCloud-Primary-Currency-Code", "NoCloud-Primary-Currency-Precision-Override",
			"grpc-metadata-nocloud-primary-currency-precision-override", "nocloud-primary-currency-precision-override"},
		AllowedMethods:   []string{"GET", "POST", "DELETE", "PUT", "PATCH", "OPTIONS", "HEAD"},
		AllowCredentials: true,
	}).Handler(AdminUIHandler())

	go func() {
		log.Info("Serving Admin-UI on " + adminUiHost)
		http_server.Serve(log, adminUiHost, ui_handler)
	}()
	log.Info("Serving gRPC-Gateway on " + gatewayHost)
	http_server.Serve(log, gatewayHost, wsproxy.WebsocketProxy(handler))
}
