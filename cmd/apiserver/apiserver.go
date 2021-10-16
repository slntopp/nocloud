package main

import (
	"context"
	"net"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/viper"

	inflog "github.com/infinimesh/infinimesh/pkg/log"
	"github.com/slntopp/nocloud/pkg/accounts/accountspb"
	apipb "github.com/slntopp/nocloud/pkg/api/apipb"
	"github.com/slntopp/nocloud/pkg/health/healthpb"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

var (
	log 			*zap.Logger

	healthHost 		string
	accountsHost 	string

	SIGNING_KEY		[]byte
)

type server struct{}

func NewServer() *server {
	return &server{}
}

func init() {
	logger, err := inflog.NewProdOrDev()
	if err != nil {
		panic(err)
	}
	log = logger

	viper.AutomaticEnv()
	viper.SetDefault("HEALTH_HOST", "health:8080")
	viper.SetDefault("ACCOUNTS_HOST", "accounts:8080")
	
	viper.SetDefault("SIGNING_KEY", "seeeecreet")

	healthHost 		= viper.GetString("HEALTH_HOST")
	accountsHost 	= viper.GetString("ACCOUNTS_HOST")

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

	log.Info("Connecting to AccountsService", zap.String("host", accountsHost))
	accountsConn, err := grpc.Dial(accountsHost, grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	accountsClient := accountspb.NewAccountsServiceClient(accountsConn)

	// Create a listener on TCP port
	lis, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatal("Failed to listen:", zap.Error(err))
	}

	// Create a gRPC server object
	s := grpc.NewServer(grpc.UnaryInterceptor(JWT_AUTH_INTERCEPTOR),)
	// Attach the Greeter service to the server
	apipb.RegisterHealthServiceServer(s, &healthAPI{client: healthClient})
	apipb.RegisterAccountsServiceServer(s, &accountsAPI{client: accountsClient})
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
	gwServer := &http.Server{
		Addr:    ":8000",
		Handler: gwmux,
	}

	log.Info("Serving gRPC-Gateway on http://0.0.0.0:8000")
	log.Fatal("Failed to Listen and Serve Gateway-Server", zap.Error(gwServer.ListenAndServe()))
}