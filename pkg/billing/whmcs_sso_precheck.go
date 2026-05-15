package billing

import (
	"context"

	pb "github.com/slntopp/nocloud-proto/billing"
	settingspb "github.com/slntopp/nocloud-proto/settings"
	"github.com/slntopp/nocloud/pkg/graph"
	"go.uber.org/zap"
)

func NewWhmcsSSOPaymentPrecheck(log *zap.Logger, settings settingspb.SettingsServiceClient) func(context.Context, *pb.Invoice, graph.Account) error {
	sc := settings
	return func(_ context.Context, inv *pb.Invoice, acc graph.Account) error {
		invConf := MakeInvoicesConf(log, &sc)
		ginv := graph.Invoice{Invoice: inv}
		return checkAdditionalProperties(invConf, ginv, acc)
	}
}
