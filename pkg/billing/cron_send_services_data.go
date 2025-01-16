package billing

import (
	"connectrpc.com/connect"
	"context"
	"github.com/arangodb/go-driver"
	instancespb "github.com/slntopp/nocloud-proto/instances"
	accountspb "github.com/slntopp/nocloud-proto/registry/accounts"
	"github.com/slntopp/nocloud/pkg/graph"
	"github.com/slntopp/nocloud/pkg/nocloud"
	"github.com/slntopp/nocloud/pkg/nocloud/schema"
	"go.uber.org/zap"
)

func (s *BillingServiceServer) CollectSystemReport(ctx context.Context, log *zap.Logger) {
	log = log.Named("CollectSystemReport")
	log.Info("Starting CollectSystemReport")

	accounts, err := s.accounts.List(ctx, graph.Account{
		Account: &accountspb.Account{Uuid: schema.ROOT_ACCOUNT_KEY},
		DocumentMeta: driver.DocumentMeta{
			ID:  driver.NewDocumentID(schema.ACCOUNTS_COL, schema.ROOT_ACCOUNT_KEY),
			Key: schema.ROOT_ACCOUNT_KEY,
		},
	}, 10000)
	if err != nil {
		log.Error("Failed to get account", zap.Error(err))
		return
	}
	log.Debug("Got accounts", zap.Int("count", len(accounts)))

	page, limit := uint64(1), uint64(0)
	req := connect.NewRequest(&instancespb.ListInstancesRequest{Page: &page, Limit: &limit})
	req.Header().Set("Authorization", ctx.Value(nocloud.NoCloudToken).(string))
	resp, err := s.instancesClient.List(ctx, req)
	if err != nil {
		log.Error("Failed to list instances", zap.Error(err))
		return
	}
	log.Debug("Got instances", zap.Int64("count", resp.Msg.Count))

	accToWhmcsId := make(map[string]int)
	for _, acc := range accounts {
		data := acc.GetData().AsMap()
		idVal, _ := data["whmcs_id"].(float64)
		id := int(idVal)
		if id > 0 {
			accToWhmcsId[acc.Key] = id
		}
	}

	instances := make(map[int][]*instancespb.Instance)
	for _, obj := range resp.Msg.Pool {
		whmcsId, ok := accToWhmcsId[obj.GetAccount()]
		if ok {
			if val, ok := instances[whmcsId]; !ok || val == nil {
				instances[whmcsId] = make([]*instancespb.Instance, 0)
			}
			instances[whmcsId] = append(instances[whmcsId], obj.Instance)
		}
	}

	// Parse clients TODO

}
