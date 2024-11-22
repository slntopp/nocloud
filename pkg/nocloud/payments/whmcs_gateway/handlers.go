package whmcs_gateway

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/arangodb/go-driver"
	pb "github.com/slntopp/nocloud-proto/billing"
	"github.com/slntopp/nocloud/pkg/graph"
	"github.com/slntopp/nocloud/pkg/nocloud/payments/types"
	"go.uber.org/zap"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/structpb"
	"strings"
	"time"
)

func (g *WhmcsGateway) invoiceCreatedHandler(ctx context.Context, log *zap.Logger, data InvoiceCreated) error {
	log.Info("Got invoiceCreated event. Got data", zap.Any("data", data))
	var err error
	if _, err = g.getInvoiceByWhmcsId(data.InvoiceId); err == nil || data.Source == "api" {
		log.Info("Invoice already exists in NoCloud", zap.Int("invoice_id", data.InvoiceId))
		return nil
	}
	if !errors.Is(err, errNotFound) {
		log.Error("Error getting invoice", zap.Error(err))
		return err
	}

	whmcsInv, err := g.GetInvoice(ctx, data.InvoiceId)
	if err != nil {
		log.Error("Error getting invoice", zap.Error(err))
		return err
	}

	acc, err := g.GetAccountByWhmcsId(whmcsInv.UserId)
	if err != nil {
		log.Error("Error getting account", zap.Error(err))
		return err
	}

	inv := &pb.Invoice{
		Status:       statusToNoCloud(whmcsInv.Status),
		Account:      acc.GetUuid(),
		Transactions: make([]string, 0),
		Total:        float64(whmcsInv.Total),
		Meta: map[string]*structpb.Value{
			"note":         structpb.NewStringValue("=== CREATED THROUGH WHMCS ===\n" + whmcsInv.Notes),
			invoiceIdField: structpb.NewNumberValue(float64(data.InvoiceId)),
		},
		Currency: acc.GetCurrency(),
		Type:     pb.ActionType_WHMCS_INVOICE,
	}

	newItems := make([]*pb.Item, 0)
	for _, item := range whmcsInv.Items.Items {
		newItems = append(newItems, &pb.Item{
			Description: item.Description,
			Amount:      1,
			Price:       float64(item.Amount),
			Unit:        "Pcs",
		})
	}
	inv.Items = newItems

	if !strings.Contains(whmcsInv.DatePaid, "0000-00-00") {
		t, err := time.Parse("2006-01-02 15:04:05", whmcsInv.DatePaid)
		if err != nil {
			return fmt.Errorf("error syncWhmcsInvoice: %w", err)
		}
		inv.Payment = t.Unix()
		inv.Processed = inv.Payment
	}
	if !strings.Contains(whmcsInv.DueDate, "0000-00-00") {
		t, err := time.Parse("2006-01-02", whmcsInv.DueDate)
		if err != nil {
			return fmt.Errorf("error syncWhmcsInvoice: %w", err)
		}
		inv.Deadline = t.Unix()
	}
	if !strings.Contains(whmcsInv.Date, "0000-00-00") {
		t, err := time.Parse("2006-01-02", whmcsInv.Date)
		if err != nil {
			return fmt.Errorf("error syncWhmcsInvoice: %w", err)
		}
		inv.Created = t.Unix()
	}

	if err = g.invMan.CreateInvoice(ctx, inv); err != nil {
		log.Error("Error creating invoice", zap.Error(err))
		return err
	}

	return nil
}

func (g *WhmcsGateway) invoiceDeletedHandler(ctx context.Context, log *zap.Logger, data InvoiceDeleted) error {
	log.Info("Got invoiceCreated event. Got data", zap.Any("data", data))
	var err error
	inv, err := g.getInvoiceByWhmcsId(data.InvoiceId)
	if err != nil {
		if errors.Is(err, errNotFound) {
			log.Info("Invoice not found in NoCloud", zap.Int("invoice_id", data.InvoiceId))
			return nil
		}
		return err
	}
	if inv, err = g.invMan.UpdateInvoiceStatus(ctx, inv.GetUuid(), pb.BillingStatus_TERMINATED); err != nil {
		log.Error("Error updating invoice status to terminated", zap.Error(err))
		return err
	}
	delete(inv.Meta, invoiceIdField)
	inv.Meta["note"] = structpb.NewStringValue(fmt.Sprintf("DELETED THROUGH WHMCS %d\n", data.InvoiceId) + inv.Meta["note"].GetStringValue())
	if _, err = g.invMan.InvoicesController().Update(ctx, &graph.Invoice{Invoice: inv, DocumentMeta: driver.DocumentMeta{Key: inv.GetUuid()}}); err != nil {
		log.Error("Error updating invoice", zap.Error(err))
		return err
	}
	return nil
}

func unmarshal[T any](b []byte) (T, error) {
	var res T
	if err := json.Unmarshal(b, &res); err != nil {
		return res, err
	}
	return res, nil
}

func (g *WhmcsGateway) handleWhmcsEvent(log *zap.Logger, body []byte) {
	log = log.Named("WhmcsHandler")

	resp := struct {
		Event string `json:"event"`
	}{}
	if err := json.Unmarshal(body, &resp); err != nil {
		log.Error("Error decoding request", zap.Error(err), zap.String("body", string(body)))
		return
	}

	log.Info("Event received", zap.String("event", resp.Event))
	log = log.With(zap.String("event", resp.Event))

	ctx := context.WithValue(context.Background(), types.GatewayCallback, true)
	md := metadata.New(map[string]string{
		string(types.GatewayCallback): "true",
	})
	ctx = metadata.NewOutgoingContext(ctx, md)
	var innerErr error
	switch resp.Event {
	case "InvoicePaid":
		data, err := unmarshal[InvoicePaid](body)
		if err != nil {
			log.Error("Error decoding request", zap.Error(err))
			return
		}
		innerErr = g.syncWhmcsInvoice(ctx, data.InvoiceId)
	case "InvoiceModified":
		data, err := unmarshal[InvoiceModified](body)
		if err != nil {
			log.Error("Error decoding request", zap.Error(err))
			return
		}
		innerErr = g.syncWhmcsInvoice(ctx, data.InvoiceId)
	case "InvoiceCancelled":
		data, err := unmarshal[InvoiceCancelled](body)
		if err != nil {
			log.Error("Error decoding request", zap.Error(err))
			return
		}
		innerErr = g.syncWhmcsInvoice(ctx, data.InvoiceId)
	case "InvoiceRefunded":
		data, err := unmarshal[InvoiceRefunded](body)
		if err != nil {
			log.Error("Error decoding request", zap.Error(err))
			return
		}
		innerErr = g.syncWhmcsInvoice(ctx, data.InvoiceId)
	case "InvoiceUnpaid":
		data, err := unmarshal[InvoiceUnpaid](body)
		if err != nil {
			log.Error("Error decoding request", zap.Error(err))
			return
		}
		innerErr = g.syncWhmcsInvoice(ctx, data.InvoiceId)
	case "UpdateInvoiceTotal":
		data, err := unmarshal[UpdateInvoiceTotal](body)
		if err != nil {
			log.Error("Error decoding request", zap.Error(err))
			return
		}
		innerErr = g.syncWhmcsInvoice(ctx, data.InvoiceId)
	case "InvoiceCreation":
		data, err := unmarshal[InvoiceCreated](body)
		if err != nil {
			log.Error("Error decoding request", zap.Error(err))
			return
		}
		innerErr = g.invoiceCreatedHandler(ctx, log, data)
	case "InvoiceDeleted":
		data, err := unmarshal[InvoiceDeleted](body)
		if err != nil {
			log.Error("Error decoding request", zap.Error(err))
			return
		}
		innerErr = g.invoiceDeletedHandler(ctx, log, data)
	default:
		log.Error("Unknown event", zap.String("event", resp.Event))
		return
	}
	if innerErr != nil {
		log.Error("Error handling event", zap.Error(innerErr))
	}
}
