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
	"context"
	"fmt"
	"github.com/rs/cors"
	driverpb "github.com/slntopp/nocloud-proto/drivers/instance/vanilla"
	epb "github.com/slntopp/nocloud-proto/events"
	"github.com/slntopp/nocloud-proto/health/healthconnect"
	regpb "github.com/slntopp/nocloud-proto/registry"
	settingspb "github.com/slntopp/nocloud-proto/settings"
	"github.com/slntopp/nocloud/pkg/graph"
	"github.com/slntopp/nocloud/pkg/graph/migrations"
	"github.com/slntopp/nocloud/pkg/nocloud/invoices_manager"
	"github.com/slntopp/nocloud/pkg/nocloud/payments"
	"github.com/slntopp/nocloud/pkg/nocloud/payments/whmcs_gateway"
	"github.com/slntopp/nocloud/pkg/nocloud/rabbitmq"
	"github.com/slntopp/nocloud/pkg/nocloud/rest_auth"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"net/http"

	cc "github.com/slntopp/nocloud-proto/billing/billingconnect"
	billing "github.com/slntopp/nocloud/pkg/billing"
	"github.com/slntopp/nocloud/pkg/nocloud"
	auth "github.com/slntopp/nocloud/pkg/nocloud/connect_auth"
	"github.com/slntopp/nocloud/pkg/nocloud/connectdb"
	"github.com/slntopp/nocloud/pkg/nocloud/schema"

	"connectrpc.com/grpchealth"
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"google.golang.org/grpc/metadata"
)

var (
	port string
	log  *zap.Logger

	RabbitMQConn string
	redisHost    string
	arangodbHost string
	arangodbCred string
	arangodbName string
	SIGNING_KEY  []byte
	drivers      []string

	settingsHost string
	registryHost string
	eventsHost   string

	invoicesFile string
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
	viper.SetDefault("EXTENTION_SERVERS", "")
	viper.SetDefault("SIGNING_KEY", "seeeecreet")
	viper.SetDefault("INVOICES_MIGRATIONS_FILE", "./whmcs_invoices.csv")

	viper.SetDefault("SETTINGS_HOST", "settings:8000")
	viper.SetDefault("REGISTRY_HOST", "registry:8000")
	viper.SetDefault("EVENTS_HOST", "eventbus:8000")

	port = viper.GetString("PORT")

	arangodbHost = viper.GetString("DB_HOST")
	arangodbCred = viper.GetString("DB_CRED")
	arangodbName = viper.GetString("DB_NAME")
	redisHost = viper.GetString("REDIS_HOST")
	SIGNING_KEY = []byte(viper.GetString("SIGNING_KEY"))
	drivers = viper.GetStringSlice("DRIVERS")

	settingsHost = viper.GetString("SETTINGS_HOST")
	registryHost = viper.GetString("REGISTRY_HOST")
	eventsHost = viper.GetString("EVENTS_HOST")

	viper.SetDefault("RABBITMQ_CONN", "amqp://nocloud:secret@rabbitmq:5672/")
	RabbitMQConn = viper.GetString("RABBITMQ_CONN")

	invoicesFile = viper.GetString("INVOICES_MIGRATIONS_FILE")
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
	if res := rdb.Ping(context.Background()); res.Err() != nil {
		log.Fatal("Failed to connect to Redis", zap.Error(res.Err()))
	}

	conn, err := amqp.Dial(RabbitMQConn)
	if err != nil {
		log.Fatal("failed to connect to RabbitMQ", zap.Error(err))
	}
	defer conn.Close()
	rbmq := rabbitmq.NewRabbitMQConnection(conn)

	// Initialize controllers
	accountsCtrl := graph.NewAccountsController(log, db)
	addonsCtrl := graph.NewAddonsController(log, db)
	plansCtrl := graph.NewBillingPlansController(log, db)
	caCtrl := graph.NewCommonActionsController(log, db)
	currCtrl := graph.NewCurrencyController(log, db)
	descCtrl := graph.NewDescriptionsController(log, db)
	instCtrl := graph.NewInstancesController(log, db, rbmq)
	_ = graph.NewInstancesGroupsController(log, db, rbmq)
	invoicesCtrl := graph.NewInvoicesController(log, db)
	nssCtrl := graph.NewNamespacesController(log, db)
	promoCtrl := graph.NewPromocodesController(log, db, rbmq)
	recordsCtrl := graph.NewRecordsController(log, db)
	srvCtrl := graph.NewServicesController(log, db, rbmq)
	spCtrl := graph.NewServicesProvidersController(log, db)
	_ = graph.NewShowcasesController(log, db)
	transactCtrl := graph.NewTransactionsController(log, db)

	authInterceptor := auth.NewInterceptor(log, rdb, SIGNING_KEY)
	restInterceptor := rest_auth.NewInterceptor(log, rdb, SIGNING_KEY)
	interceptors := connect.WithInterceptors(authInterceptor)

	router := mux.NewRouter()
	router.Use(func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Debug("Request", zap.String("method", r.Method), zap.String("path", r.URL.Path))
			h.ServeHTTP(w, r)
		})
	})

	settingsConn, err := grpc.Dial(settingsHost, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	settingsClient := settingspb.NewSettingsServiceClient(settingsConn)

	accConn, err := grpc.Dial(registryHost, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	accClient := regpb.NewAccountsServiceClient(accConn)

	eventsConn, err := grpc.Dial(eventsHost, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	eventsClient := epb.NewEventsServiceClient(eventsConn)

	registeredDrivers := make(map[string]driverpb.DriverServiceClient)
	for _, driver := range drivers {
		log.Info("Registering Driver", zap.String("driver", driver))
		dconn, err := grpc.Dial(driver, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Fatal("Error registering driver", zap.String("driver", driver), zap.Error(err))
		}
		client := driverpb.NewDriverServiceClient(dconn)
		driver_type, err := client.GetType(context.Background(), &driverpb.GetTypeRequest{})
		if err != nil {
			log.Fatal("Error dialing driver and getting its type", zap.String("driver", driver), zap.Error(err))
		}
		registeredDrivers[driver_type.GetType()] = client
		log.Info("Registered Driver", zap.String("driver", driver), zap.String("type", driver_type.GetType()))
	}

	server := billing.NewBillingServiceServer(log, db, rbmq, rdb, registeredDrivers,
		settingsClient, accClient, eventsClient,
		nssCtrl, plansCtrl, transactCtrl, invoicesCtrl, recordsCtrl, currCtrl, accountsCtrl, descCtrl,
		instCtrl, spCtrl, srvCtrl, addonsCtrl, caCtrl, promoCtrl)
	currencies := billing.NewCurrencyServiceServer(log, db, currCtrl, caCtrl)
	log.Info("Starting Currencies Service")

	token, err := authInterceptor.MakeToken(schema.ROOT_ACCOUNT_KEY)
	if err != nil {
		log.Fatal("Can't generate token", zap.Error(err))
	}
	ctx := metadata.AppendToOutgoingContext(context.Background(), "authorization", "bearer "+token)

	log.Info("Check settings server")
	if _, err = settingsClient.Get(ctx, &settingspb.GetRequest{}); err != nil {
		log.Fatal("Can't check settings connection", zap.Error(err))
	}
	log.Info("Settings server connection established")

	log.Info("Starting Transaction Generator-Processor")
	go server.GenTransactionsRoutine(ctx)

	log.Info("Starting Account Suspension Routine")
	go server.SuspendAccountsRoutine(ctx)

	log.Info("Starting Invoices Routine")
	go server.IssueInvoicesRoutine(ctx)

	log.Info("Starting Instances Creation Consumer")
	go server.ConsumeCreatedInstances(ctx)

	log.Info("Registering BillingService Server")
	path, handler := cc.NewBillingServiceHandler(server, interceptors)
	router.PathPrefix(path).Handler(handler)

	records := billing.NewRecordsServiceServer(log, rbmq, db, settingsClient, recordsCtrl, plansCtrl, instCtrl, addonsCtrl, promoCtrl, caCtrl)
	log.Info("Starting Records Consumer")
	go records.Consume(ctx)

	log.Info("Registering CurrencyService Server")
	path, handler = cc.NewCurrencyServiceHandler(currencies, interceptors)
	router.PathPrefix(path).Handler(handler)

	addons := billing.NewAddonsServer(log, db, addonsCtrl, nssCtrl, caCtrl)
	log.Info("Registering AddonsService Server")
	path, handler = cc.NewAddonsServiceHandler(addons, interceptors)
	router.PathPrefix(path).Handler(handler)

	descriptions := billing.NewDescriptionsServer(log, db, descCtrl, nssCtrl, caCtrl)
	log.Info("Registering DescriptionsService Server")
	path, handler = cc.NewDescriptionsServiceHandler(descriptions, interceptors)
	router.PathPrefix(path).Handler(handler)

	promocodes := billing.NewPromocodesServer(log, db, promoCtrl, nssCtrl, caCtrl)
	log.Info("Registering PromocodesService Server")
	path, handler = cc.NewPromocodesServiceHandler(promocodes, interceptors)
	router.PathPrefix(path).Handler(handler)

	checker := grpchealth.NewStaticChecker()
	path, handler = grpchealth.NewHandler(checker)
	router.PathPrefix(path).Handler(handler)

	health := NewHealthServer(log, server, records, currencies)
	log.Info("Registering health server")
	path, handler = healthconnect.NewInternalProbeServiceHandler(health)
	router.PathPrefix(path).Handler(handler)

	migrations.MigrateOldInvoicesToNew(log, graph.GetEnsureCollection(log, ctx, db, schema.INVOICES_COL),
		graph.GetEnsureCollection(log, ctx, db, schema.TRANSACTIONS_COL), invoicesFile)

	// Register payments gateways (nocloud, whmcs)
	bClient := cc.NewBillingServiceClient(http.DefaultClient, "http://billing:8000")
	whmcsData, err := whmcs_gateway.GetWhmcsCredentials(rdb)
	if err != nil {
		log.Fatal("Can't get whmcs credentials", zap.Error(err))
	}
	manager := invoices_manager.NewInvoicesManager(bClient, invoicesCtrl, authInterceptor)
	payments.RegisterGateways(whmcsData, accountsCtrl, manager)

	// Register WHMCS hooks handler (hooks for invoices status e.g.)
	whmcsGw := whmcs_gateway.NewWhmcsGateway(whmcsData, accountsCtrl, manager)
	whmcsRouter := router.PathPrefix("/nocloud.billing.Whmcs").Subrouter()
	whmcsRouter.Use(restInterceptor.JwtMiddleWare)
	whmcsRouter.Path("/hooks").HandlerFunc(whmcsGw.BuildWhmcsHooksHandler(log))

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
