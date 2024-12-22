package billing

import (
	"context"
	"errors"
	"github.com/DmitriyVTitov/size"
	pb "github.com/slntopp/nocloud-proto/billing"
	"github.com/slntopp/nocloud/pkg/nocloud/payments/types"
	"github.com/slntopp/nocloud/pkg/nocloud/payments/whmcs_gateway"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"
)

func (s *BillingServiceServer) WhmcsDeletedInvoicesSyncerCronJob(ctx context.Context, log *zap.Logger) {
	log = log.Named("WhmcsDeletedInvoicesSyncerCronJob")
	log.Info("Starting WHMCS Invoices syncer cron job")

	ncInvoices, err := s.invoices.List(ctx, "")
	if err != nil {
		log.Error("Error listing nocloud invoices", zap.Error(err))
		return
	}
	log.Info("Size of NC invoices slice", zap.Int("bytes_count", size.Of(ncInvoices)), zap.Int("len", len(ncInvoices)))

	delCount := 0
	for _, inv := range ncInvoices {
		if inv.Meta == nil {
			continue
		}
		whmcsId, ok := inv.GetMeta()["whmcs_invoice_id"]
		if !ok {
			continue
		}
		if _, err = s.whmcsGateway.GetInvoice(ctx, int(whmcsId.GetNumberValue())); err == nil {
			continue
		}
		if !errors.Is(err, whmcs_gateway.ErrNotFound) {
			log.Error("Error getting invoice from whmcs", zap.Error(err))
			continue
		}
		delete(inv.Meta, "whmcs_invoice_id")
		inv.Status = pb.BillingStatus_TERMINATED
		pNote := inv.Meta["note"].GetStringValue()
		inv.Meta["note"] = structpb.NewStringValue("DELETED ON WHMCS!\n" + pNote)
		if _, err = s.invoices.Replace(ctx, inv); err != nil {
			log.Error("Error updating invoice", zap.Error(err))
			continue
		}
		delCount++
	}

	whmcsIdToInvoice := make(map[int]struct{})
	for _, inv := range ncInvoices {
		if inv.Meta == nil {
			continue
		}
		whmcsId, ok := inv.GetMeta()["whmcs_invoice_id"]
		if ok {
			whmcsIdToInvoice[int(whmcsId.GetNumberValue())] = struct{}{}
		}
	}

	whmcsInvoices, err := s.whmcsGateway.GetInvoices(ctx)
	if err != nil {
		log.Error("Error listing whmcs invoices", zap.Error(err))
		return
	}
	log.Info("Size of WHMCS invoices slice", zap.Int("bytes_count", size.Of(whmcsInvoices)), zap.Int("len", len(whmcsInvoices)))

	ctx = context.WithValue(ctx, types.GatewayCallback, true) // Prevent whmcs event cycling
	createdCount := 0
	for _, whmcsInvoice := range whmcsInvoices {
		if _, ok := whmcsIdToInvoice[whmcsInvoice.InvoiceId]; ok {
			continue
		}
		if err = s.whmcsGateway.CreateFromWhmcsInvoice(ctx, log, whmcsInvoice); err != nil {
			log.Error("Failed to create whmcs invoice", zap.Error(err))
			continue
		}
		createdCount++
	}

	log.Info("Finished WHMCS Invoices syncer cron job", zap.Int("deleted", delCount), zap.Int("created", createdCount))
}
