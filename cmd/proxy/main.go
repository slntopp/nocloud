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
	"github.com/slntopp/nocloud/pkg/nocloud/schema"
	"net"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/slntopp/nocloud/pkg/graph"
	"github.com/slntopp/nocloud/pkg/nocloud"
	"github.com/slntopp/nocloud/pkg/nocloud/connectdb"
	"github.com/slntopp/nocloud/pkg/proxy"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	healthpb "github.com/slntopp/nocloud-proto/health"
)

var (
	log *zap.Logger

	arangodbHost string
	arangodbCred string
	arangodbName string

	SIGNING_KEY []byte
)

func init() {
	viper.AutomaticEnv()
	log = nocloud.NewLogger()

	viper.SetDefault("DB_HOST", "db:8529")
	viper.SetDefault("DB_CRED", "root:openSesame")
	viper.SetDefault("DB_NAME", schema.DB_NAME)

	viper.SetDefault("SIGNING_KEY", "seeeecreet")

	SIGNING_KEY = []byte(viper.GetString("SIGNING_KEY"))

	arangodbHost = viper.GetString("DB_HOST")
	arangodbCred = viper.GetString("DB_CRED")
	arangodbName = viper.GetString("DB_NAME")
}

func main() {
	defer func() {
		_ = log.Sync()
	}()

	log.Info("Setting up DB Connection")
	db := connectdb.MakeDBConnection(log, arangodbHost, arangodbCred, arangodbName)
	log.Info("DB connection established")

	ctrl := graph.NewServicesProvidersController(log, db)

	proxy.Setup(log, ctrl)

	r := mux.NewRouter()
	r.Use(AuthMiddleware)
	r.HandleFunc("/socket", proxy.Handler).Methods("GET")
	r.Use(mux.CORSMethodMiddleware(r))

	lis, err := net.Listen("tcp", ":8000")
	if err != nil {
		log.Fatal("Failed to listen", zap.String("address", ":8000"), zap.Error(err))
	}
	s := grpc.NewServer()
	healthpb.RegisterInternalProbeServiceServer(s, NewHealthServer(log))

	go s.Serve(lis)

	/* #nosec */
	log.Fatal("Failed to serve proxy", zap.Error(http.ListenAndServe(":8080", r)))
}
