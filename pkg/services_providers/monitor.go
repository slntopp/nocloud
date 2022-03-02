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
	"encoding/json"
	"time"

	driverpb "github.com/slntopp/nocloud/pkg/drivers/instance/vanilla"
	instpb "github.com/slntopp/nocloud/pkg/instances/proto"
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
	defaultConf = MonitoringRoutineConf{
		Frequency: 60,
	}

	description = "ServicesProviders Monitoring Routing Configuration"
	public = false
)

func init() {
	viper.AutomaticEnv()
	viper.SetDefault("SETTINGS_HOST", "settings:8080")
    host := viper.GetString("SETTINGS_HOST")
    
	conn, err := grpc.Dial(host, grpc.WithTransportCredentials(insecure.NewCredentials()))
    if err != nil {
        panic(err)
    }

    settingsClient = settingspb.NewSettingsServiceClient(conn)
}

func MakeConf(ctx context.Context) MonitoringRoutineConf {

	var conf MonitoringRoutineConf
	var r_str string
	r, err := settingsClient.Get(ctx, &settingspb.GetRequest{Keys: []string{monFreqKey}})
	if err != nil {
		goto set_default
	}
	if _, ok := r.GetFields()[monFreqKey]; !ok {
		goto set_default
	}
	r_str = r.GetFields()[monFreqKey].GetStringValue()
	err = json.Unmarshal([]byte(r_str), &conf)
	if err != nil {
		goto set_default
	}
	return conf

	set_default:
	conf = defaultConf
	payload, err := json.Marshal(conf)
	if err == nil {
		settingsClient.Put(ctx, &settingspb.PutRequest{
			Key: monFreqKey,
			Value: string(payload),
			Description: &description,
			Public: &public,
		})
	}
	return conf
}

func (s *ServicesProviderServer) MonitoringRoutine(ctx context.Context) {
	log := s.log.Named("MonitoringRoutine")

	conf := MakeConf(ctx)
	ticker := time.NewTicker(time.Second * time.Duration(conf.Frequency))

	go func ()  {
		begin_sub:
		sub, err := settingsClient.Sub(ctx, &settingspb.SubRequest{Key: "sp-monitoring-routing"})
		if err != nil {
			log.Error("Could not subscribe to Settings update", zap.Error(err))
			goto final
		}

		for {
			data, err := sub.Recv()
			if err != nil {
				log.Error("Error reading from Settings Sub channel", zap.Error(err))
				break
			}
			err = json.Unmarshal([]byte(data.GetValue()), &conf)
			if err != nil {
				log.Error("Error unmarshall udpate from Settings Sub channel", zap.Error(err))
				break
			}
			ticker.Reset(time.Duration(conf.Frequency))
		}

		log.Info("Monitoring Conf Updated", zap.Any("conf", conf))
		final:
		time.Sleep(time.Duration(conf.Frequency))
		goto begin_sub
	}()

	for tick := range ticker.C {
		log.Info("Starting", zap.Time("tick", tick))

		sp_pool, err := s.ctrl.List(ctx, schema.ROOT_ACCOUNT_KEY)
		if err != nil {
			log.Error("Failed to get ServicesProviders", zap.Error(err))
			continue
		}
		log.Debug("Got ServicesProviders", zap.Int("length", len(sp_pool)))

		for _, sp := range sp_pool {
			go func(sp *graph.ServicesProvider) {
				services, err := s.ctrl.ListDeployments(ctx, sp)
				if err != nil {
					log.Error("Failed to get Services deployed to ServiceProvider", zap.String("sp", sp.GetUuid()), zap.Error(err))
					return
				}

				var igroups []*instpb.InstancesGroup
				for _, service := range services {
					for _, igroup := range service.GetInstancesGroups() {
						if igroup.GetType() == sp.GetType() {
							igroups = append(igroups, igroup)
						}
					}
				}
				log.Debug("Got InstancesGroups", zap.Int("length", len(igroups)))

				client, ok := s.drivers[sp.GetType()]
				if !ok {
					log.Error("Driver is not registered", zap.String("sp", sp.GetUuid()), zap.String("type", sp.GetType()))
					return
				}

				_, err = client.Monitoring(ctx, &driverpb.MonitoringRequest{
					Groups: igroups,
					ServicesProvider: sp.ServicesProvider,
				})
				if err != nil {
					log.Error("Error Monitoring ServicesProvider", zap.String("sp", sp.GetUuid()), zap.Error(err))
				}
			}(sp)
		}
	}
}