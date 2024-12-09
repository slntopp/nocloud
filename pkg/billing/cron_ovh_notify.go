package billing

import (
	"context"
	elpb "github.com/slntopp/nocloud-proto/events_logging"
	"github.com/slntopp/nocloud/pkg/nocloud"
	"go.uber.org/zap"
	"time"
)

func (s *BillingServiceServer) NotifyToUpdateOvhPricesCronJob(_ context.Context, log *zap.Logger) {
	log = log.Named("NotifyToUpdateOvhPricesCronJob")
	nocloud.Log(log, &elpb.Event{
		Uuid:      "no_uuid",
		Entity:    "ovh",
		Action:    "notify_to_update_prices",
		Scope:     "ovh",
		Rc:        0,
		Ts:        time.Now().Unix(),
		Snapshot:  &elpb.Snapshot{Diff: ""},
		Requestor: "system",
	})
}
