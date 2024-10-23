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
package instances

import (
	"context"
	pb "github.com/slntopp/nocloud-proto/billing/addons"
	"time"

	stpb "github.com/slntopp/nocloud-proto/statuses"

	"github.com/arangodb/go-driver"
	"github.com/slntopp/nocloud-proto/access"
	driverpb "github.com/slntopp/nocloud-proto/drivers/instance/vanilla"
	settingspb "github.com/slntopp/nocloud-proto/settings"
	sc "github.com/slntopp/nocloud/pkg/settings/client"
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
		Level:       access.Level_ADMIN,
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

func MakeConf(ctx context.Context, log *zap.Logger, upd chan bool) (conf MonitoringRoutineConf) {
	sc.Setup(log, ctx, &settingsClient)

	err := sc.Fetch(monFreqKey, &conf, defaultSetting)
	if err != nil {
		return defaultSetting.Value
	}

	go sc.Subscribe([]string{monFreqKey}, upd)

	return conf
}

func (s *InstancesServer) MonitoringRoutineState() Routine {
	return s.monitoring
}

const getAccsBalance = `
FOR node, edge, path IN 3
    INBOUND @ig
    GRAPH @permissions
    FILTER path.edges[*].role == ["owner","owner","owner"]
    FILTER IS_SAME_COLLECTION(node, @@accounts)
    RETURN LENGTH(node.account_owner) > 0 ? TO_NUMBER(DOCUMENT(@@accounts, node.account_owner)["balance"]) : TO_NUMBER(node.balance)
`

const getAddons = `
LET instances = (
    FOR node, edge IN 1
			OUTBOUND @ig
	        GRAPH @permissions
	        FILTER IS_SAME_COLLECTION(node, @@instances)
	        RETURN node
)

FILTER IS_ARRAY(instances)
FOR inst IN instances
FILTER IS_ARRAY(inst.addons)
FOR a IN inst.addons
    COLLECT uuid = a
    RETURN MERGE(DOCUMENT(CONCAT(@addons, "/", uuid)), { uuid })
`

func (s *InstancesServer) MonitoringRoutine(ctx context.Context) {
	log := s.log.Named("MonitoringRoutine")

	log.Info("Fetching Monitoring Configuration")

start:
	upd := make(chan bool, 1)
	conf := MakeConf(ctx, log, upd)

	log.Info("Got Monitoring Configuration", zap.Any("conf", conf))

	ticker := time.NewTicker(time.Second * time.Duration(conf.Frequency))
	tick := time.Now()
	for {
		s.monitoring.Running = true

		sp_pool, err := s.sp_ctrl.List(ctx, schema.ROOT_ACCOUNT_KEY, true)
		if err != nil {
			log.Error("Failed to get ServicesProviders", zap.Error(err))
			continue
		}
		log.Debug("Got ServicesProviders", zap.Int("length", len(sp_pool)))

		for _, sp := range sp_pool {
			if sp.GetStatus() == stpb.NoCloudStatus_DEL {
				continue
			}

			sp, err := s.sp_ctrl.Get(ctx, sp.Uuid)
			if err != nil {
				log.Error("Coudln't get ServicesProvider", zap.String("sp", sp.Uuid), zap.Error(err))
				continue
			}

			go func(sp *graph.ServicesProvider) {
				log := log.With(zap.String("sp", sp.GetUuid()), zap.String("sp_title", sp.GetTitle()))
				log.Debug("Starting MonitoringRoutine")
				//syncer := sync.NewDataSyncer(log.With(zap.String("caller", "MonitoringRoutine")), s.rdb, sp.GetUuid(), 5)
				//defer syncer.Open()
				//_ = syncer.WaitUntilOpenedAndCloseAfter()
				//log.Debug("Locking mutex")
				//m := s.spSyncers[sp.GetUuid()]
				//if m == nil {
				//	m = &go_sync.Mutex{}
				//	s.spSyncers[sp.GetUuid()] = m
				//}
				//m.Lock()
				//defer func() {
				//	log.Debug("Unlocking mutex")
				//	m.Unlock()
				//}()

				igroups, err := s.sp_ctrl.GetGroups(ctx, sp)
				if err != nil {
					log.Error("Failed to get Services deployed to ServiceProvider", zap.String("sp", sp.GetUuid()), zap.Error(err))
					return
				}

				var balance = make(map[string]float64, len(igroups))

				for _, group := range igroups {
					cur, err := s.db.Query(ctx, getAccsBalance, map[string]any{
						"ig":          driver.NewDocumentID(schema.INSTANCES_GROUPS_COL, group.GetUuid()),
						"permissions": schema.PERMISSIONS_GRAPH.Name,
						"@accounts":   schema.ACCOUNTS_COL,
					})
					if err != nil {
						log.Error("Failed to get cursor", zap.Error(err), zap.String("uuid", group.GetUuid()))
						continue
					}

					var result float64

					_, err = cur.ReadDocument(ctx, &result)
					if err != nil {
						log.Error("Failed to get balance", zap.Error(err), zap.String("uuid", group.GetUuid()))
						continue
					}
					balance[group.GetUuid()] = result
					cur.Close()
				}

				var addons = map[string]*pb.Addon{}
				for _, ig := range igroups {
					cur, err := s.db.Query(ctx, getAddons, map[string]any{
						"ig":          driver.NewDocumentID(schema.INSTANCES_GROUPS_COL, ig.GetUuid()),
						"addons":      schema.ADDONS_COL,
						"@instances":  schema.INSTANCES_COL,
						"permissions": schema.PERMISSIONS_GRAPH.Name,
					})
					if err != nil {
						log.Error("Failed to get addons", zap.Error(err))
						return
					}
					for cur.HasMore() {
						var addon = &pb.Addon{}
						_, err = cur.ReadDocument(ctx, addon)
						if err != nil {
							log.Error("Failed to get addons", zap.Error(err))
							continue
						}
						addons[addon.Uuid] = addon
					}
				}

				log.Debug("Got InstancesGroups", zap.Int("length", len(igroups)))
				log.Debug("Got Addons", zap.Any("addons", addons))

				client, ok := s.drivers[sp.GetType()]
				if !ok {
					log.Error("Driver is not registered", zap.String("sp", sp.GetUuid()), zap.String("type", sp.GetType()))
					return
				}

				_, err = client.Monitoring(ctx, &driverpb.MonitoringRequest{
					Groups:           igroups,
					ServicesProvider: sp.ServicesProvider,
					Scheduled:        true,
					Balance:          balance,
					Addons:           addons,
				})
				if err != nil {
					log.Error("Error Monitoring ServicesProvider", zap.String("sp", sp.GetUuid()), zap.Error(err))
				}

				log.Debug("Finished MonitoringRoutine")
			}(sp)
		}

		s.monitoring.LastExec = tick.Format("2006-01-02T15:04:05Z07:00")
		select {
		case tick = <-ticker.C:
			continue
		case <-upd:
			log.Info("New Configuration Received, restarting Routine")
			goto start
		}
	}
}
