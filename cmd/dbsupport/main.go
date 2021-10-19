package main

import (
	"os"
	"os/signal"
	"syscall"

	inflog "github.com/infinimesh/infinimesh/pkg/log"
	"github.com/slntopp/nocloud/pkg/graph"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var (
	log 			*zap.Logger

	arangodbHost 	string
	arangodbCred 	string
	passwd 			string
)

func init() {
	logger, err := inflog.NewProdOrDev()
	if err != nil {
		panic(err)
	}
	log = logger

	viper.AutomaticEnv()
	viper.SetDefault("DB_HOST", "db:8529")
	viper.SetDefault("DB_CRED", "root:openSesame")
	viper.SetDefault("PASSWD", "root")

	arangodbHost = viper.GetString("DB_HOST")
	arangodbCred = viper.GetString("DB_CRED")
	passwd 		 = viper.GetString("PASSWD")
}

func main() {
	graph.InitDB(log, arangodbHost, arangodbCred, passwd)
	sig := make(chan os.Signal)
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT)

	log.Info("Awaiting SIGTERM or SIGINIT")
	s := <- sig
	log.Info("Got signal, exiting", zap.String("signal", s.String()))
}