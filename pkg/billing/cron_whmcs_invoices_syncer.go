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
	"time"
)

func (s *BillingServiceServer) WhmcsInvoicesSyncerCronJob(ctx context.Context, log *zap.Logger) {
	log = log.Named("WhmcsInvoicesSyncerCronJob")
	log.Info("Starting WHMCS Invoices syncer cron job")
	now := time.Now().Unix()

	ncInvoices, err := s.invoices.List(ctx, "")
	if err != nil {
		log.Error("Error listing nocloud invoices", zap.Error(err))
		return
	}
	log.Info("Size of NC invoices slice", zap.Any("bytes_count_for_invoice", reflect.TypeOf(graph.Invoice{}).Size()), zap.Int("len", len(ncInvoices)))

	whmcsInvoices, err := s.whmcsGateway.GetInvoices(ctx)
	if err != nil {
		log.Error("Error listing whmcs invoices", zap.Error(err))
		return
	}
	log.Info("Size of WHMCS invoices slice", zap.Any("bytes_count_for_invoice", reflect.TypeOf(whmcs_gateway.Invoice{}).Size()), zap.Int("len", len(whmcsInvoices)))
	ids := make(map[int]whmcs_gateway.InvoiceInList)
	for _, val := range whmcsInvoices {
		ids[int(val.Id)] = val
	}

	for _, inv := range ncInvoices {
		if inv.Meta == nil {
			continue
		}
		whmcsId, ok := inv.GetMeta()["whmcs_invoice_id"]
		if !ok {
			continue
		}
		whmcsInvoice, ok := ids[int(whmcsId.GetNumberValue())]
		if !ok {
			continue
		}
		if whmcsInvoice.PaymentMethod == "" {
			continue
		}
		if whmcsInvoice.PaymentMethod != "balancepay" && whmcsInvoice.PaymentMethod != "system" {
			continue
		}
		inv.Meta["paid_with_balance"] = structpb.NewBoolValue(true)
		if _, err = s.invoices.Update(ctx, inv); err != nil {
			log.Error("Error updating invoice", zap.Error(err))
			continue
		}
	}

	delCount := 0
	for _, inv := range ncInvoices {
		if inv.Meta == nil {
			continue
		}
		whmcsId, ok := inv.GetMeta()["whmcs_invoice_id"]
		if !ok {
			continue
		}
		if _, ok = ids[int(whmcsId.GetNumberValue())]; ok {
			continue
		}
		if _, err = s.whmcsGateway.GetInvoice(ctx, int(whmcsId.GetNumberValue())); err == nil {
			log.Warn("Invoice found but wasn't presented in whmcs invoices list", zap.Int("id", int(whmcsId.GetNumberValue())))
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

	log.Info("Finished WHMCS Invoices syncer cron job", zap.Int("deleted", delCount))
	return

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

	ctx = context.WithValue(ctx, types.GatewayCallback, true) // Prevent whmcs event cycling
	createdCount := 0
	for _, whmcsInvoice := range whmcsInvoices {
		if whmcsInvoice.Id == 0 {
			log.Warn("Whmcs invoice id is zero value")
			continue
		}
		logI := log.With(zap.Int("whmcs_id", int(whmcsInvoice.Id)))
		if _, ok := whmcsIdToInvoice[int(whmcsInvoice.Id)]; ok {
			continue
		}
		// Do not create invoice if it is younger than half a day (preventing accidental duplicate) and if too old
		const secondsInDay = 86400
		const tooOldDate = 1577836800 // GMT: Wednesday, 1 January 2020 г., 0:00:00
		dateCreated, parseCreatedErr := time.Parse(time.DateTime, whmcsInvoice.CreatedAt)
		if parseCreatedErr == nil && dateCreated.Unix() > 0 && (now-dateCreated.Unix() < secondsInDay/2 || dateCreated.Unix() < tooOldDate) {
			continue
		}
		inv, err := s.whmcsGateway.GetInvoice(ctx, int(whmcsInvoice.Id))
		if err != nil {
			logI.Error("Failed to get body of whmcs invoice", zap.Error(err))
			continue
		}
		if parseCreatedErr != nil || dateCreated.Unix() <= 0 {
			dateCreated, err = time.Parse(time.DateOnly, inv.Date)
			if err != nil || dateCreated.Unix() <= 0 || (now-dateCreated.Unix() < secondsInDay/2 || dateCreated.Unix() < tooOldDate) {
				continue
			}
		}
		if err = s.whmcsGateway.CreateFromWhmcsInvoice(ctx, log, inv); err != nil {
			logI.Error("Failed to create whmcs invoice", zap.Error(err))
			continue
		}
		createdCount++
	}

	log.Info("Finished WHMCS Invoices syncer cron job", zap.Int("deleted", delCount), zap.Int("created", createdCount))
}
