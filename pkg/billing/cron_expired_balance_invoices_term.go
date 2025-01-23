package billing

import (
	"connectrpc.com/connect"
	"context"
	pb "github.com/slntopp/nocloud-proto/billing"
	"go.uber.org/zap"
	"time"
)

const deleteExpiredBalanceInvoiceAfterHours = 72

func (s *BillingServiceServer) DeleteExpiredBalanceInvoicesCronJob(ctx context.Context, log *zap.Logger) {
	log = log.Named("DeleteExpiredBalanceInvoicesCronJob")
	log.Info("Starting delete expired balance invoices cron job")

	ncInvoices, err := s.invoices.List(ctx, "")
	if err != nil {
		log.Error("Error listing nocloud invoices", zap.Error(err))
		return
	}

	now := time.Now().Unix()
	count := 0
	for _, inv := range ncInvoices {
		if inv.GetType() != pb.ActionType_BALANCE {
			continue
		}
		if inv.GetStatus() != pb.BillingStatus_UNPAID {
			continue
		}
		if inv.Deadline > now {
			continue
		}
		if now-inv.Created < deleteExpiredBalanceInvoiceAfterHours*60*60 {
			continue
		}
		req := connect.NewRequest(&pb.UpdateInvoiceStatusRequest{
			Uuid:   inv.GetUuid(),
			Status: pb.BillingStatus_TERMINATED,
		})
		if _, err = s.UpdateInvoiceStatus(ctx, req); err != nil {
			log.Error("Failed to terminate expired balance invoice", zap.Error(err))
			continue
		}
		delaySeconds(20)
		count++
	}

	log.Info("Finished delete expired balance invoices cron job", zap.Int("deleted", count))
}
