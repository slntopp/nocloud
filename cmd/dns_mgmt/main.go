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
	pb "github.com/slntopp/nocloud-proto/dns"
	healthpb "github.com/slntopp/nocloud-proto/health"
	"github.com/slntopp/nocloud/pkg/dns"
	"github.com/slntopp/nocloud/pkg/nocloud"
	auth "github.com/slntopp/nocloud/pkg/nocloud/admin_auth"
	grpc_server "github.com/slntopp/nocloud/pkg/nocloud/grpc"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

var (
	port string
	log  *zap.Logger

	redisHost   string
	SIGNING_KEY []byte
)

func init() {
	viper.AutomaticEnv()
	log = nocloud.NewLogger()

	viper.SetDefault("PORT", "8000")
	viper.SetDefault("REDIS_HOST", "redis:6379")
	viper.SetDefault("SIGNING_KEY", "seeeecreet")

	port = viper.GetString("PORT")
	redisHost = viper.GetString("REDIS_HOST")
	SIGNING_KEY = []byte(viper.GetString("SIGNING_KEY"))
}

func main() {
	defer func() {
		_ = log.Sync()
	}()

	log.Info("Setting up RedisDB Connection")
	rdb := redis.NewClient(&redis.Options{
		Addr: redisHost,
		DB:   0, // use default DB
	})
	log.Info("RedisDB connection established")

	auth.SetContext(log, rdb, SIGNING_KEY)
	s := grpc.NewServer(grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
		grpc_zap.UnaryServerInterceptor(log),
		grpc.UnaryServerInterceptor(auth.JWT_AUTH_INTERCEPTOR),
	)))

	server := dns.NewDNSServer(log, rdb)
	pb.RegisterDNSServer(s, server)
	healthpb.RegisterInternalProbeServiceServer(s, NewHealthServer(log))

	grpc_server.ServeGRPC(log, s, port)
}
