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
	"fmt"
	"net"

	"github.com/spf13/viper"
	"go.uber.org/zap"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	"google.golang.org/grpc"

	"github.com/slntopp/nocloud/pkg/accounting"
	"github.com/slntopp/nocloud/pkg/accounting/accountspb"
	"github.com/slntopp/nocloud/pkg/accounting/namespacespb"
	"github.com/slntopp/nocloud/pkg/nocloud"
	"github.com/slntopp/nocloud/pkg/nocloud/auth"
	"github.com/slntopp/nocloud/pkg/nocloud/connectdb"
)

var (
	port 			string
	log 			*zap.Logger

	arangodbHost 	string
	arangodbCred 	string
	nocloudRootPass string

	SIGNING_KEY 	[]byte
)

func init() {
	viper.AutomaticEnv()
	log = nocloud.NewLogger()
	
	viper.SetDefault("PORT", "8080")

	viper.SetDefault("DB_HOST", "db:8529")
	viper.SetDefault("DB_CRED", "root:openSesame")
	viper.SetDefault("NOCLOUD_ROOT_PASSWORD", "secret")

	viper.SetDefault("SIGNING_KEY", "seeeecreet")

	port = viper.GetString("PORT")

	arangodbHost 	= viper.GetString("DB_HOST")
	arangodbCred 	= viper.GetString("DB_CRED")
	nocloudRootPass = viper.GetString("NOCLOUD_ROOT_PASSWORD")

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

	accounts_server := accounting.NewAccountsServer(log, db)
	accounts_server.SIGNING_KEY = SIGNING_KEY
	accounts_server.EnsureRootExists(nocloudRootPass)
	accountspb.RegisterAccountsServiceServer(s, accounts_server)

	namespaces_server := accounting.NewNamespacesServer(log, db)
	namespacespb.RegisterNamespacesServiceServer(s, namespaces_server)
	
	log.Info(fmt.Sprintf("Serving gRPC on 0.0.0.0:%v", port), zap.Skip())
	log.Fatal("Failed to serve gRPC", zap.Error(s.Serve(lis)))
}