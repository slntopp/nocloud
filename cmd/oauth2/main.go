package main

import (
	"github.com/go-redis/redis/v8"
	"github.com/slntopp/nocloud-proto/registry"
	"github.com/slntopp/nocloud/pkg/nocloud"
	"github.com/slntopp/nocloud/pkg/nocloud/auth"
	"github.com/slntopp/nocloud/pkg/oauth2"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"strings"
)

var (
	port string
	log  *zap.Logger

	corsAllowed  []string
	registryHost string
	redisHost    string
	SIGNING_KEY  []byte
)

func init() {
	viper.AutomaticEnv()
	log = nocloud.NewLogger()

	viper.SetDefault("PORT", "8000")

	viper.SetDefault("REGISTRY_HOST", "registry:8000")
	viper.SetDefault("REDIS_HOST", "redis:6379")

	viper.SetDefault("CORS_ALLOWED", []string{"*"})

	viper.SetDefault("SIGNING_KEY", "seeeecreet")

	port = viper.GetString("PORT")

	registryHost = viper.GetString("REGISTRY_HOST")
	redisHost = viper.GetString("REDIS_HOST")

	corsAllowed = strings.Split(viper.GetString("CORS_ALLOWED"), ",")

	SIGNING_KEY = []byte(viper.GetString("SIGNING_KEY"))
}

func main() {
	rdb := redis.NewClient(&redis.Options{
		Addr: redisHost,
		DB:   0,
	})

	auth.SetContext(log, rdb, SIGNING_KEY)

	registryConn, err := grpc.Dial(registryHost, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	defer registryConn.Close()

	registryClient := registry.NewAccountsServiceClient(registryConn)

	server := oauth2.NewOAuth2Server(log, registryHost)
	server.SetupRegistryClient(registryClient)
	server.Start(port, corsAllowed)
}
