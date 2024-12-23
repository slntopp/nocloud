package billing

import (
	"context"
	"errors"
	pb "github.com/slntopp/nocloud-proto/billing"
	"github.com/slntopp/nocloud/pkg/graph"
	"github.com/slntopp/nocloud/pkg/nocloud/payments/types"
	"github.com/slntopp/nocloud/pkg/nocloud/payments/whmcs_gateway"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"
	"reflect"
)

func (s *BillingServiceServer) WhmcsInvoicesSyncerCronJob(ctx context.Context, log *zap.Logger) {
	log = log.Named("WhmcsInvoicesSyncerCronJob")
	log.Info("Starting WHMCS Invoices syncer cron job")

	ncInvoices, err := s.invoices.List(ctx, "")
	if err != nil {
		log.Error("Error listing nocloud invoices", zap.Error(err))
		return
	}
	log.Info("Size of NC invoices slice", zap.Any("bytes_count_for_invoice", reflect.TypeOf(graph.Invoice{}).Size()), zap.Int("len", len(ncInvoices)))

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
		if ok && int(whmcsId.GetNumberValue()) != 0 {
			whmcsIdToInvoice[int(whmcsId.GetNumberValue())] = struct{}{}
		}
	}

	whmcsInvoices, err := s.whmcsGateway.GetInvoices(ctx)
	if err != nil {
		log.Error("Error listing whmcs invoices", zap.Error(err))
		return
	}
	log.Info("Size of WHMCS invoices slice", zap.Any("bytes_count_for_invoice", reflect.TypeOf(whmcs_gateway.Invoice{}).Size()), zap.Int("len", len(whmcsInvoices)))

	ctx = context.WithValue(ctx, types.GatewayCallback, true) // Prevent whmcs event cycling
	createdCount := 0
	for _, whmcsInvoice := range whmcsInvoices {
		if whmcsInvoice.Id == 0 {
			log.Warn("Whmcs invoice id is zero value")
			continue
		}
		if _, ok := whmcsIdToInvoice[int(whmcsInvoice.Id)]; ok {
			continue
		}
		inv, err := s.whmcsGateway.GetInvoice(ctx, int(whmcsInvoice.Id))
		if err != nil {
			log.Error("Failed to get body of whmcs invoice", zap.Error(err))
			continue
		}
		if err = s.whmcsGateway.CreateFromWhmcsInvoice(ctx, log, inv); err != nil {
			log.Error("Failed to create whmcs invoice", zap.Error(err))
			continue
		}
		createdCount++
	}

	log.Info("Finished WHMCS Invoices syncer cron job", zap.Int("deleted", delCount), zap.Int("created", createdCount))
}
