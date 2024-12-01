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
	"connectrpc.com/connect"
	"connectrpc.com/grpchealth"
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
	ccb "github.com/slntopp/nocloud-proto/billing/billingconnect"
	"github.com/slntopp/nocloud-proto/health/healthconnect"
	ic "github.com/slntopp/nocloud-proto/instances/instancesconnect"
	cc "github.com/slntopp/nocloud-proto/services/servicesconnect"
	"github.com/slntopp/nocloud/pkg/graph"
	"github.com/slntopp/nocloud/pkg/nocloud/invoices_manager"
	"github.com/slntopp/nocloud/pkg/nocloud/payments"
	"github.com/slntopp/nocloud/pkg/nocloud/payments/whmcs_gateway"
	"github.com/slntopp/nocloud/pkg/nocloud/rabbitmq"
	"github.com/slntopp/nocloud/pkg/nocloud/sync"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/grpc/metadata"
	"net/http"
	go_sync "sync"

	amqp "github.com/rabbitmq/amqp091-go"
	bpb "github.com/slntopp/nocloud-proto/billing"
	driverpb "github.com/slntopp/nocloud-proto/drivers/instance/vanilla"
	stpb "github.com/slntopp/nocloud-proto/settings"
	"github.com/slntopp/nocloud/pkg/instances"
	"github.com/slntopp/nocloud/pkg/nocloud"
	auth "github.com/slntopp/nocloud/pkg/nocloud/connect_auth"
	"github.com/slntopp/nocloud/pkg/nocloud/connectdb"
	"github.com/slntopp/nocloud/pkg/nocloud/schema"
	"github.com/slntopp/nocloud/pkg/services"
	"github.com/slntopp/nocloud/pkg/states"
	"github.com/spf13/viper"
	"go.uber.org/zap"

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
	SIGNING_KEY  []byte
	rbmq         string
	settingsHost string
	billingHost  string

	whmcsTaxExcluded bool
)

func init() {
	viper.AutomaticEnv()
	log = nocloud.NewLogger()

	viper.SetDefault("PORT", "8000")

	viper.SetDefault("REDIS_HOST", "redis:6379")
	viper.SetDefault("DB_HOST", "db:8529")
	viper.SetDefault("DB_CRED", "root:openSesame")
	viper.SetDefault("DB_NAME", schema.DB_NAME)
	viper.SetDefault("DRIVERS", "")
	viper.SetDefault("SIGNING_KEY", "seeeecreet")
	viper.SetDefault("RABBITMQ_CONN", "amqp://nocloud:secret@rabbitmq:5672/")
	viper.SetDefault("SETTINGS_HOST", "settings:8000")
	viper.SetDefault("BILLING_HOST", "billing:8000")
	viper.SetDefault("WHMCS_PRICES_TAX_EXCLUDED", true)

	port = viper.GetString("PORT")

	arangodbHost = viper.GetString("DB_HOST")
	arangodbCred = viper.GetString("DB_CRED")
	arangodbName = viper.GetString("DB_NAME")
	drivers = viper.GetStringSlice("DRIVERS")
	SIGNING_KEY = []byte(viper.GetString("SIGNING_KEY"))
	rbmq = viper.GetString("RABBITMQ_CONN")
	settingsHost = viper.GetString("SETTINGS_HOST")
	billingHost = viper.GetString("BILLING_HOST")
	redisHost = viper.GetString("REDIS_HOST")
	whmcsTaxExcluded = viper.GetBool("WHMCS_PRICES_TAX_EXCLUDED")
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

	authInterceptor := auth.NewInterceptor(log, rdb, SIGNING_KEY)
	interceptors := connect.WithInterceptors(authInterceptor)

	router := mux.NewRouter()
	router.Use(func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Debug("Request", zap.String("method", r.Method), zap.String("path", r.URL.Path))
			h.ServeHTTP(w, r)
		})
	})

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

	server := services.NewServicesServer(log, db, ps, rabbitmq.NewRabbitMQConnection(rbmq))
	iserver := instances.NewInstancesServiceServer(log, db, rabbitmq.NewRabbitMQConnection(rbmq), rdb)

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
	token, err := authInterceptor.MakeToken(schema.ROOT_ACCOUNT_KEY)
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

	log.Info("Registering Services Server")
	path, handler := cc.NewServicesServiceHandler(server, interceptors, connect.WithReadMaxBytes(100*1024*1024))
	router.PathPrefix(path).Handler(handler)
	log.Info("Services Server registered", zap.String("path", path))

	log.Info("Registering Instances Server")
	path, handler = ic.NewInstancesServiceHandler(iserver, interceptors, connect.WithReadMaxBytes(100*1024*1024))
	router.PathPrefix(path).Handler(handler)
	log.Info("Instances Server registered", zap.String("path", path))

	checker := grpchealth.NewStaticChecker()
	path, handler = grpchealth.NewHandler(checker)
	router.PathPrefix(path).Handler(handler)

	health := NewHealthServer(log, server, iserver)
	log.Info("Registering health server")
	path, handler = healthconnect.NewInternalProbeServiceHandler(health)
	router.PathPrefix(path).Handler(handler)

	// Register payments gateways (nocloud, whmcs)
	bClient := ccb.NewBillingServiceClient(http.DefaultClient, "http://"+billingHost)
	whmcsData, err := whmcs_gateway.GetWhmcsCredentials(rdb)
	if err != nil {
		log.Fatal("Can't get whmcs credentials", zap.Error(err))
	}
	manager := invoices_manager.NewInvoicesManager(bClient, graph.NewInvoicesController(log, db), authInterceptor)
	payments.RegisterGateways(whmcsData, graph.NewAccountsController(log, db), graph.NewCurrencyController(log, db), manager, whmcsTaxExcluded)

	log.Debug("Opening every sp syncer")
	ctrl := graph.NewServicesProvidersController(log.Named("Main"), db)
	sps, err := ctrl.List(context.Background(), schema.ROOT_NAMESPACE_KEY, true)
	if err != nil {
		log.Fatal("Failed to list services providers", zap.Error(err))
	}
	wg := &go_sync.WaitGroup{}
	wg.Add(len(sps))
	for _, sp := range sps {
		sp := sp
		go func() {
			if err := sync.NewDataSyncer(log.With(zap.String("caller", "Main")), rdb, sp.GetUuid(), -1).Open(); err != nil {
				log.Fatal("Failed to open sp syncer", zap.Error(err))
			}
			wg.Done()
		}()
	}
	wg.Wait()

	ctx := metadata.AppendToOutgoingContext(
		context.Background(), "authorization", "bearer "+token,
	)
	go iserver.MonitoringRoutine(ctx)

	host := fmt.Sprintf("0.0.0.0:%s", port)

	handler = cors.New(cors.Options{
		AllowedOrigins:      []string{"*"},
		AllowedMethods:      []string{"GET", "POST", "OPTIONS", "PUT", "DELETE"},
		AllowedHeaders:      []string{"*", "Connect-Protocol-Version"},
		AllowCredentials:    true,
		AllowPrivateNetwork: true,
	}).Handler(h2c.NewHandler(router, &http2.Server{}))

	log.Info("Serving", zap.String("host", host))
	err = http.ListenAndServe(host, handler)
	if err != nil {
		log.Fatal("Failed to start server", zap.Error(err))
	}
}
