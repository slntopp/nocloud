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
*/package main

import (
	"context"
	"crypto/tls"
	"net/http"
	"os"
	"strings"

	"github.com/rs/cors"
	billingpb "github.com/slntopp/nocloud-proto/billing"
	dnspb "github.com/slntopp/nocloud-proto/dns"
	healthpb "github.com/slntopp/nocloud-proto/health"
	instancespb "github.com/slntopp/nocloud-proto/instances"
	registrypb "github.com/slntopp/nocloud-proto/registry"
	servicespb "github.com/slntopp/nocloud-proto/services"
	sppb "github.com/slntopp/nocloud-proto/services_providers"
	settingspb "github.com/slntopp/nocloud-proto/settings"
	"github.com/slntopp/nocloud/pkg/nocloud"
	"github.com/tmc/grpc-websocket-proxy/wsproxy"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/gabriel-vasile/mimetype"
)

var (
	log *zap.Logger

	gatewayHost      string
	apiserver        string
	corsAllowed      []string
	insecure_enabled bool
	with_block       bool
)

func init() {
	viper.AutomaticEnv()
	log = nocloud.NewLogger()

	viper.SetDefault("CORS_ALLOWED", []string{"*"})
	viper.SetDefault("APISERVER_HOST", "proxy:8000")
	viper.SetDefault("GATEWAY_HOST", ":8000")
	viper.SetDefault("INSECURE", true)
	viper.SetDefault("WITH_BLOCK", false)

	gatewayHost = viper.GetString("GATEWAY_HOST")
	apiserver = viper.GetString("APISERVER_HOST")
	corsAllowed = strings.Split(viper.GetString("CORS_ALLOWED"), ",")
	insecure_enabled = viper.GetBool("INSECURE")
	with_block = viper.GetBool("WITH_BLOCK")
}

func getContentType(path string) (mime string, ok bool) {
	chunks := strings.Split(path, ".")
	switch chunks[len(chunks)-1] {
	case "css":
		return "text/css; charset=utf-8", true
	default:
		return "", false
	}
}

func staticHandler(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
	log.Debug("Static request", zap.Any("path", pathParams))
	file := "index.html"
	if path, ok := pathParams["path"]; ok {
		file = path
	}
	if path, ok := pathParams["file"]; ok {
		file += "/" + path
	}
	index, err := os.ReadFile("/dist/" + file)
	if err != nil {
		log.Error("Error reading file", zap.Error(err))
		w.WriteHeader(404)
		return
	}

	mime, ok := getContentType(file)
	if !ok {
		mime = mimetype.Detect(index).String()
	}
	log.Debug("Content-Type", zap.String("mime", mime))
	w.Header().Set("Content-Type", mime)
	if _, err = w.Write(index); err != nil {
		log.Warn("Coulnd't write index.html", zap.Error(err))
	}
}

func main() {
	defer func() {
		_ = log.Sync()
	}()

	log.Info("Starting REST-API Server")
	log.Info("Registering Endpoints", zap.String("server", apiserver))
	var err error

	gwmux := runtime.NewServeMux()
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
	log.Info("Registering DNS Gateway")
	err = dnspb.RegisterDNSHandlerFromEndpoint(context.Background(), gwmux, apiserver, opts)
	if err != nil {
		log.Fatal("Failed to register DNS gateway", zap.Error(err))
	}
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

	for _, p := range []string{"/admin", "/admin/{path}", "/admin/{path}/{file}"} {
		err = gwmux.HandlePath("GET", p, staticHandler)
		if err != nil {
			log.Fatal(
				"Failed to register custom static handler",
				zap.String("path", p),
				zap.Error(err),
			)
		}
	}

	log.Info("Allowed Origins", zap.Strings("hosts", corsAllowed))
	handler := cors.New(cors.Options{
		AllowedOrigins:   corsAllowed,
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowedMethods:   []string{"GET", "POST", "DELETE", "PUT", "PATCH", "OPTIONS", "HEAD"},
		AllowCredentials: true,
	}).Handler(gwmux)

	log.Info("Serving gRPC-Gateway on " + gatewayHost)
	/* #nosec */
	log.Fatal(
		"Failed to Listen and Serve Gateway-Server",
		zap.Error(http.ListenAndServe(gatewayHost, wsproxy.WebsocketProxy(handler))),
	)
}
