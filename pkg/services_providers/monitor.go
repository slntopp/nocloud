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
package services_providers

import (
	"context"
	"time"

	driverpb "github.com/slntopp/nocloud/pkg/drivers/instance/vanilla"
	sc "github.com/slntopp/nocloud/pkg/settings/client"
	settingspb "github.com/slntopp/nocloud/pkg/settings/proto"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/slntopp/nocloud/pkg/graph"
	"github.com/slntopp/nocloud/pkg/nocloud/schema"
	"go.uber.org/zap"
)

var (
	settingsClient settingspb.SettingsServiceClient
)

// Settings Key storing monitoring routine conf
const monFreqKey string = "sp-monitoring-routine"

type MonitoringRoutineConf struct {
	Frequency int `json:"freq"` // Frequency in Seconds
}

var (
	defaultSetting = &sc.Setting[MonitoringRoutineConf]{
		Value: MonitoringRoutineConf{
			Frequency: 60,
		},
		Description: "ServicesProviders Monitoring Routine Configuration",
		Public:      false,
	}
)

func init() {
	viper.AutomaticEnv()
	viper.SetDefault("SETTINGS_HOST", "settings:8000")
	host := viper.GetString("SETTINGS_HOST")

	conn, err := grpc.Dial(host, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}

	settingsClient = settingspb.NewSettingsServiceClient(conn)
}

func MakeConf(ctx context.Context, log *zap.Logger) (conf MonitoringRoutineConf) {
	sc.Setup(log, ctx, &settingsClient)

	err := sc.Fetch(monFreqKey, &conf, defaultSetting)
	if err != nil {
		return defaultSetting.Value
	}

	return conf
}

func (s *ServicesProviderServer) MonitoringRoutineState() Routine {
	return s.monitoring
}

func (s *ServicesProviderServer) MonitoringRoutine(ctx context.Context) {
	log := s.log.Named("MonitoringRoutine")

	conf := MakeConf(ctx, log)
	log.Info("Got Monitoring Configuration", zap.Any("conf", conf))
	ticker := time.NewTicker(time.Second * time.Duration(conf.Frequency))
	tick := time.Now()
	for {
		s.monitoring.Running = true

		sp_pool, err := s.ctrl.List(ctx, schema.ROOT_ACCOUNT_KEY)
		if err != nil {
			log.Error("Failed to get ServicesProviders", zap.Error(err))
			continue
		}
		log.Debug("Got ServicesProviders", zap.Int("length", len(sp_pool)))

		for _, sp := range sp_pool {
			sp, err := s.ctrl.Get(ctx, sp.Uuid)
			if err != nil {
				log.Error("Coudln't get ServicesProvider", zap.String("sp", sp.Uuid), zap.Error(err))
				continue
			}
			go func(sp *graph.ServicesProvider) {
				igroups, err := s.ctrl.ListDeployments(ctx, sp, true)
				if err != nil {
					log.Error("Failed to get Services deployed to ServiceProvider", zap.String("sp", sp.GetUuid()), zap.Error(err))
					return
				}

				log.Debug("Got InstancesGroups", zap.Int("length", len(igroups)))

				client, ok := s.drivers[sp.GetType()]
				if !ok {
					log.Error("Driver is not registered", zap.String("sp", sp.GetUuid()), zap.String("type", sp.GetType()))
					return
				}

				_, err = client.Monitoring(ctx, &driverpb.MonitoringRequest{
					Groups:           igroups,
					ServicesProvider: sp.ServicesProvider,
					Scheduled:        true,
				})
				if err != nil {
					log.Error("Error Monitoring ServicesProvider", zap.String("sp", sp.GetUuid()), zap.Error(err))
				}
			}(sp)
		}

		s.monitoring.LastExec = tick.Format("2006-01-02T15:04:05Z07:00")
		tick = <-ticker.C
	}
}
