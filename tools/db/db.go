/*
Copyright © 2021-2022 Nikita Ivanovski info@slnt-opp.xyz

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
	"github.com/slntopp/nocloud/pkg/graph"
	"github.com/slntopp/nocloud/pkg/nocloud"
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
	viper.AutomaticEnv()
	log = nocloud.NewLogger()

	viper.SetDefault("DB_HOST", "localhost")
	viper.SetDefault("DB_PORT", "8529")
	viper.SetDefault("DB_USER", "root")
	viper.SetDefault("DB_PASS", "openSesame")

	viper.SetDefault("ROOT_PASS", "secret")

	dbHost = viper.GetString("DB_HOST")
	dbPort = viper.GetString("DB_PORT")
	dbUser = viper.GetString("DB_USER")
	dbPass = viper.GetString("DB_PASS")
	
	rootPass = viper.GetString("ROOT_PASS")
}

func main() {
	defer func() {
		_ = log.Sync()
	}()

	graph.InitDB(log, dbHost + ":" + dbPort, dbUser + ":" + dbPass, rootPass)
}