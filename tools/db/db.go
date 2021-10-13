package main

import (
	inflog "github.com/infinimesh/infinimesh/pkg/log"
	"github.com/slntopp/nocloud/pkg/graph"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var (
	log *zap.Logger

	dbHost string
	dbPort string
	dbUser string
	dbPass string

	rootPass string
)

func init() {
	logger, err := inflog.NewProdOrDev()
	if err != nil {
		panic(err)
	}
	log = logger

	viper.AutomaticEnv()
	viper.SetDefault("DB_HOST", "localhost")
	viper.SetDefault("DB_PORT", "8529")
	viper.SetDefault("DB_USER", "root")
	viper.SetDefault("DB_PASS", "openSesame")

	viper.SetDefault("ROOT_PASS", "root")

	dbHost = viper.GetString("DB_HOST")
	dbPort = viper.GetString("DB_PORT")
	dbUser = viper.GetString("DB_USER")
	dbPass = viper.GetString("DB_PASS")
}

func main() {
	graph.InitDB(log, dbHost, dbPort, dbUser, dbPass)
}