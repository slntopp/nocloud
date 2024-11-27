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
	"fmt"
	"github.com/go-redis/redis/v8"
	"net"

	"github.com/spf13/viper"
	"go.uber.org/zap"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	"google.golang.org/grpc"

	"google.golang.org/grpc/credentials/insecure"

	settingspb "github.com/slntopp/nocloud-proto/settings"
	"github.com/slntopp/nocloud/pkg/credentials"
	"github.com/slntopp/nocloud/pkg/nocloud"
	"github.com/slntopp/nocloud/pkg/nocloud/auth"
	"github.com/slntopp/nocloud/pkg/nocloud/connectdb"
	"github.com/slntopp/nocloud/pkg/nocloud/schema"
	accounting "github.com/slntopp/nocloud/pkg/registry"
	"github.com/slntopp/nocloud/pkg/sessions"

	healthpb "github.com/slntopp/nocloud-proto/health"
	pb "github.com/slntopp/nocloud-proto/registry"
	sspb "github.com/slntopp/nocloud-proto/sessions"
)

var (
	port string
	log  *zap.Logger

	arangodbHost    string
	arangodbCred    string
	arangodbName    string
	nocloudRootPass string
	settingsHost    string
	redisHost       string
	SIGNING_KEY     []byte
)

func init() {
	viper.AutomaticEnv()
	log = nocloud.NewLogger()

	viper.SetDefault("PORT", "8000")

	viper.SetDefault("DB_HOST", "db:8529")
	viper.SetDefault("DB_CRED", "root:openSesame")
	viper.SetDefault("DB_NAME", schema.DB_NAME)
	viper.SetDefault("NOCLOUD_ROOT_PASSWORD", "secret")
	viper.SetDefault("SETTINGS_HOST", "settings:8000")
	viper.SetDefault("REDIS_HOST", "redis:6379")

	viper.SetDefault("SIGNING_KEY", "seeeecreet")

	port = viper.GetString("PORT")

	arangodbHost = viper.GetString("DB_HOST")
	arangodbCred = viper.GetString("DB_CRED")
	arangodbName = viper.GetString("DB_NAME")
	nocloudRootPass = viper.GetString("NOCLOUD_ROOT_PASSWORD")
	settingsHost = viper.GetString("SETTINGS_HOST")
	redisHost = viper.GetString("REDIS_HOST")

	SIGNING_KEY = []byte(viper.GetString("SIGNING_KEY"))
}

func SetupSettingsClient() (settingspb.SettingsServiceClient, *grpc.ClientConn) {
	// Start settings client
	conn, err := grpc.Dial(settingsHost, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}

	return settingspb.NewSettingsServiceClient(conn), conn
}

func main() {
	defer func() {
		_ = log.Sync()
	}()

	log.Info("Setting up DB Connection")
	db := connectdb.MakeDBConnection(log, arangodbHost, arangodbCred, arangodbName)
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
	)

	token, err := auth.MakeToken(schema.ROOT_ACCOUNT_KEY)
	if err != nil {
		log.Fatal("Can't generate token", zap.Error(err))
	}

	sc, sconn := SetupSettingsClient()
	defer sconn.Close()

	sessions_server := sessions.NewSessionsServer(log, rdb, db)
	sspb.RegisterSessionsServiceServer(s, sessions_server)

	accounts_server := accounting.NewAccountsServer(log, db, rdb)
	accounts_server.SIGNING_KEY = SIGNING_KEY
	credentials.SetupSettingsClient(log.Named("Credentials"), sc, token)
	accounts_server.SetupSettingsClient(sc, token)
	err = accounts_server.EnsureRootExists(nocloudRootPass)
	if err != nil {
		log.Fatal("Couldn't ensure root Account(and Namespace) exist", zap.Error(err))
	}
	pb.RegisterAccountsServiceServer(s, accounts_server)

	namespaces_server := accounting.NewNamespacesServer(log, db)
	pb.RegisterNamespacesServiceServer(s, namespaces_server)

	healthpb.RegisterInternalProbeServiceServer(s, NewHealthServer(log))

	log.Info(fmt.Sprintf("Serving gRPC on 0.0.0.0:%v", port), zap.Skip())
	log.Fatal("Failed to serve gRPC", zap.Error(s.Serve(lis)))
}
