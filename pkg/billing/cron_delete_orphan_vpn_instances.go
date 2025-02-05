package billing

import (
	"connectrpc.com/connect"
	"context"
	instancespb "github.com/slntopp/nocloud-proto/instances"
	statuspb "github.com/slntopp/nocloud-proto/statuses"
	"github.com/slntopp/nocloud/pkg/graph"
	"github.com/slntopp/nocloud/pkg/nocloud"
	"go.uber.org/zap"
	"strings"
)

// TODO: Remove this. This solution is temporal because it violates project arch
func (s *BillingServiceServer) DeleteOrphanVPNInstances(ctx context.Context, log *zap.Logger) {
	log = log.Named("DeleteOrphanVPNInstances")
	log.Info("Starting DeleteOrphanVPNInstances")

	count := 0
	page, limit := uint64(1), uint64(0)
	req := connect.NewRequest(&instancespb.ListInstancesRequest{Page: &page, Limit: &limit})
	req.Header().Set("Authorization", "Bearer "+ctx.Value(nocloud.NoCloudToken).(string))
	resp, err := s.instancesClient.List(ctx, req)
	if err != nil {
		log.Error("Failed to list instances", zap.Error(err))
		return
	}
	log.Debug("Got instances", zap.Int64("count", resp.Msg.Count))

	plans := make(map[string]*graph.BillingPlan)
	_plans, err := s.plans.List(ctx, "")
	if err != nil {
		log.Error("Failed to list plans", zap.Error(err))
		return
	}
	for _, p := range _plans {
		if p == nil {
			continue
		}
		plans[p.Key] = p
	}

	for _, obj := range resp.Msg.Pool {
		if obj == nil || obj.Instance == nil {
			continue
		}
		inst := obj.Instance
		log := log.With(zap.String("instance", inst.GetUuid()))
		if inst.BillingPlan == nil || inst.Config == nil {
			continue
		}
		plan := plans[inst.GetBillingPlan().GetUuid()]
		if strings.ToLower(plan.Type) != "vpn" {
			continue
		}
		linked := inst.GetConfig()["instance"].GetStringValue()
		if linked == "" {
			continue
		}
		linkedInstance, err := s.instances.Get(ctx, linked)
		if err != nil || linkedInstance == nil {
			log.Error("Failed to get linked vpn instance", zap.Error(err))
			continue
		}
		if linkedInstance.GetStatus() == statuspb.NoCloudStatus_DEL {
			if err = s.instances.SetStatus(ctx, inst, statuspb.NoCloudStatus_DEL); err != nil {
				log.Error("Failed to delete orphan vpn instance", zap.Error(err))
				continue
			}
			log.Info("deleted", zap.String("deleted_instance", linked))
			count++
		}
	}

	log.Info("Finished DeleteOrphanVPNInstances", zap.Int("deleted", count))
}
