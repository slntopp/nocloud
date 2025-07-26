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
	"github.com/slntopp/nocloud-proto/health"
	"github.com/slntopp/nocloud/pkg/nocloud/sync"
	go_sync "sync"
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

func (s *InstancesServer) RoutinesState() []*health.RoutineStatus {
	return []*health.RoutineStatus{
		s.monitoring,
	}
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

type MonitoringContext struct {
	ctx       context.Context
	cancel    context.CancelFunc
	Sp        string
	Frequency int64
}

type ActiveMonitoring map[string]MonitoringContext

func (a ActiveMonitoring) Abort(sp string) {
	if mCtx, ok := a[sp]; ok {
		mCtx.cancel()
		delete(a, sp)
	}
}

func (a ActiveMonitoring) Get(sp string) (MonitoringContext, bool) {
	ctx, ok := a[sp]
	return ctx, ok
}

func (a ActiveMonitoring) Start(sp string, freq int64, action func()) {
	a.Abort(sp)
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	mCtx := MonitoringContext{
		ctx:       ctx,
		cancel:    cancel,
		Sp:        sp,
		Frequency: freq,
	}
	a[sp] = mCtx
	go func() {
		t := time.NewTicker(time.Duration(mCtx.Frequency) * time.Second)
		defer t.Stop()
		for {
			go action()
			select {
			case <-mCtx.ctx.Done():
				return
			case <-t.C:
			}
		}
	}()
}

var activeMonitoring = ActiveMonitoring{}

func (s *InstancesServer) MonitoringRoutine(_ctx context.Context, wg *go_sync.WaitGroup) {
	defer wg.Done()
	ctx := context.WithoutCancel(_ctx)
	log := s.log.Named("MonitoringRoutine")
	log.Info("Fetching Monitoring Configuration")

start:
	upd := make(chan bool, 1)
	conf := MakeConf(ctx, log, upd)
	log.Info("Got Monitoring Configuration", zap.Any("conf", conf))

	ticker := time.NewTicker(time.Second * time.Duration(conf.Frequency))
	tick := time.Now()
	for {
		log.Info("Entering new Iteration", zap.Time("ts", tick), zap.Int("active", len(activeMonitoring)))
		s.monitoring.Status.Status = health.Status_RUNNING

		spPool, err := s.sp_ctrl.List(ctx, schema.ROOT_ACCOUNT_KEY, true)
		if err != nil {
			log.Error("Failed to get ServicesProviders", zap.Error(err))
			continue
		}

		for _, sp := range spPool {
			spUuid := sp.GetUuid()
			if sp.GetStatus() == stpb.NoCloudStatus_DEL {
				activeMonitoring.Abort(spUuid)
				continue
			}
			var freq = int64(conf.Frequency)
			if sp.MonitoringFrequency != nil && *sp.MonitoringFrequency > 0 {
				freq = *sp.MonitoringFrequency
				log.Debug("Found custom monitoring frequency", zap.Int64("freq", freq), zap.String("sp", spUuid))
			}
			monitoring, ok := activeMonitoring.Get(spUuid)
			if !ok || monitoring.Frequency != freq {
				activeMonitoring.Start(spUuid, freq, func() {
					s.monitoringAction(ctx, log, spUuid)
				})
			}
		}

		s.monitoring.LastExecution = tick.Format("2006-01-02T15:04:05Z07:00")
		select {
		case <-_ctx.Done():
			log.Info("Context is done. Quitting")
			return
		case tick = <-ticker.C:
			continue
		case <-upd:
			log.Info("New Configuration Received, restarting Routine")
			goto start
		}
	}
}

func (s *InstancesServer) monitoringAction(ctx context.Context, _log *zap.Logger, spUuid string) {
	start := time.Now()
	sp, err := s.sp_ctrl.Get(ctx, spUuid)
	if err != nil {
		_log.Error("Couldn't get sp", zap.String("sp", spUuid), zap.Error(err))
		return
	}
	log := _log.With(zap.String("sp", sp.GetUuid()), zap.String("sp_title", sp.GetTitle()))
	log.Debug("Starting monitoring for service provider")
	syncer := sync.NewDataSyncer(log.With(zap.String("caller", "MonitoringRoutine")), s.rdb, sp.GetUuid(), 5)
	defer syncer.Open()
	_ = syncer.WaitUntilOpenedAndCloseAfter()

	iGroups, err := s.sp_ctrl.GetGroups(ctx, sp)
	if err != nil {
		log.Error("Failed to get services deployed to sp", zap.String("sp", sp.GetUuid()), zap.Error(err))
		return
	}

	var balance = make(map[string]float64, len(iGroups))

	for _, group := range iGroups {
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
	for _, ig := range iGroups {
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
	log.Debug("Got IGs for monitoring", zap.Int("length", len(iGroups)))
	log.Debug("Got addons for monitoring", zap.Int("addons", len(addons)))

	client, ok := s.drivers[sp.GetType()]
	if !ok {
		log.Error("Driver is not registered", zap.String("type", sp.GetType()))
		return
	}

	_, err = client.Monitoring(ctx, &driverpb.MonitoringRequest{
		Groups:           iGroups,
		ServicesProvider: sp.ServicesProvider,
		Scheduled:        true,
		Balance:          balance,
		Addons:           addons,
	})
	if err != nil {
		log.Error("Error Monitoring ServicesProvider", zap.String("sp", sp.GetUuid()), zap.Error(err))
	}

	elapsed := time.Since(start)
	log.Debug("Finished monitoring", zap.Float64("duration_seconds", elapsed.Seconds()))
}
