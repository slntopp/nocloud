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
	"github.com/go-redis/redis/v8"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	"github.com/heltonmarx/goami/ami"
	grpc_server "github.com/slntopp/nocloud/pkg/nocloud/grpc"
	"github.com/spf13/viper"
	"go.uber.org/zap"
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

	amiUser     string
	amiHost     string
	amiSecret   string
	amiRequired bool
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
	viper.SetDefault("AMI_HOST", "127.0.0.1:5038")
	viper.SetDefault("AMI_USERNAME", "admin")
	viper.SetDefault("AMI_SECRET", "admin")
	viper.SetDefault("AMI_REQUIRED", "false")

	viper.SetDefault("SIGNING_KEY", "seeeecreet")

	port = viper.GetString("PORT")

	arangodbHost = viper.GetString("DB_HOST")
	arangodbCred = viper.GetString("DB_CRED")
	arangodbName = viper.GetString("DB_NAME")
	nocloudRootPass = viper.GetString("NOCLOUD_ROOT_PASSWORD")
	settingsHost = viper.GetString("SETTINGS_HOST")
	redisHost = viper.GetString("REDIS_HOST")

	SIGNING_KEY = []byte(viper.GetString("SIGNING_KEY"))

	amiHost = viper.GetString("AMI_HOST")
	amiUser = viper.GetString("AMI_USERNAME")
	amiSecret = viper.GetString("AMI_SECRET")
	amiRequired = viper.GetBool("AMI_REQUIRED")
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

	log.Debug("Initializing AMI")
	socket, err := ami.NewSocket(amiHost)
	if err != nil {
		if amiRequired {
			log.Fatal("AMI socket error", zap.Error(err))
		} else {
			log.Error("AMI socket error", zap.Error(err))
		}
	}
	if _, err := ami.Connect(socket); err != nil {
		if amiRequired {
			log.Fatal("AMI connect error", zap.Error(err))
		} else {
			log.Error("AMI connect error", zap.Error(err))
		}
	}
	uuid, _ := ami.GetUUID()
	if ok, err := ami.Login(socket, amiUser, amiSecret, "Off", uuid); err != nil || !ok {
		if amiRequired {
			log.Fatal("AMI login error", zap.Error(err))
		} else {
			log.Error("AMI login error", zap.Error(err))
		}
	}

	token, err := auth.MakeToken(schema.ROOT_ACCOUNT_KEY)
	if err != nil {
		log.Fatal("Can't generate token", zap.Error(err))
	}

	sc, sconn := SetupSettingsClient()
	defer sconn.Close()

	sessions_server := sessions.NewSessionsServer(log, rdb, db)
	sspb.RegisterSessionsServiceServer(s, sessions_server)

	accounts_server := accounting.NewAccountsServer(log, db, rdb, socket)
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

	grpc_server.ServeGRPC(log, s, port)
}
