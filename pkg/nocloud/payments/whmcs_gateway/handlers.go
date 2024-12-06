package whmcs_gateway

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	pb "github.com/slntopp/nocloud-proto/billing"
	"github.com/slntopp/nocloud/pkg/graph"
	"github.com/slntopp/nocloud/pkg/nocloud/payments/types"
	ps "github.com/slntopp/nocloud/pkg/pubsub"
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
	if !errors.Is(err, ErrNotFound) {
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

	cur, err := g.currencies.Get(ctx, acc.GetCurrency().GetId())
	if err != nil {
		return err
	}

	inv := &pb.Invoice{
		Status:       statusToNoCloud(whmcsInv.Status),
		Account:      acc.GetUuid(),
		Transactions: make([]string, 0),
		Instances:    make([]string, 0),
		Meta: map[string]*structpb.Value{
			"note":                  structpb.NewStringValue(""),
			invoiceIdField:          structpb.NewNumberValue(float64(data.InvoiceId)),
			graph.InvoiceTaxMetaKey: structpb.NewNumberValue(float64(whmcsInv.TaxRate / 100)),
			"creator":               structpb.NewStringValue("whmcs-gateway"),
		},
		Currency: acc.GetCurrency(),
		Type:     pb.ActionType_WHMCS_INVOICE,
	}

	tax := float64(whmcsInv.TaxRate / 100)

	var total float64
	newItems := make([]*pb.Item, 0)
	for _, item := range whmcsInv.Items.Items {
		var price float64
		whmcsAmount := float64(item.Amount)
		if g.taxExcluded {
			price = whmcsAmount + whmcsAmount*tax
		} else {
			price = whmcsAmount
		}
		price = graph.Round(price, cur.Precision, cur.Rounding)
		total += price

		newItems = append(newItems, &pb.Item{
			Description: item.Description,
			Amount:      1,
			Price:       float64(item.Amount),
			Unit:        "Pcs",
		})
	}
	inv.Items = newItems
	total = graph.Round(total, cur.Precision, cur.Rounding)
	inv.Total = total

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

func unmarshal[T any](b []byte) (T, error) {
	var res T
	if err := json.Unmarshal(b, &res); err != nil {
		return res, err
	}
	return res, nil
}

func (g *WhmcsGateway) HandleWhmcsEvent(log *zap.Logger, body []byte) error {
	log = log.Named("WhmcsHandler")

	resp := struct {
		Event string `json:"event"`
	}{}
	if err := json.Unmarshal(body, &resp); err != nil {
		log.Error("Error decoding request", zap.Error(err), zap.String("body", string(body)))
		return ps.NoNackErr(err)
	}

	log.Info("Event received", zap.String("event", resp.Event), zap.String("body", string(body)))
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
			return ps.NoNackErr(err)
		}
		log = log.With(zap.Int("invoice_id", data.InvoiceId))
		innerErr = g.syncWhmcsInvoice(ctx, data.InvoiceId)
	case "InvoiceModified":
		data, err := unmarshal[InvoiceModified](body)
		if err != nil {
			log.Error("Error decoding request", zap.Error(err))
			return ps.NoNackErr(err)
		}
		log = log.With(zap.Int("invoice_id", data.InvoiceId))
		innerErr = g.syncWhmcsInvoice(ctx, data.InvoiceId)
	case "InvoiceCancelled":
		data, err := unmarshal[InvoiceCancelled](body)
		if err != nil {
			log.Error("Error decoding request", zap.Error(err))
			return ps.NoNackErr(err)
		}
		log = log.With(zap.Int("invoice_id", data.InvoiceId))
		innerErr = g.syncWhmcsInvoice(ctx, data.InvoiceId)
	case "InvoiceRefunded":
		data, err := unmarshal[InvoiceRefunded](body)
		if err != nil {
			log.Error("Error decoding request", zap.Error(err))
			return ps.NoNackErr(err)
		}
		log = log.With(zap.Int("invoice_id", data.InvoiceId))
		innerErr = g.syncWhmcsInvoice(ctx, data.InvoiceId)
	case "InvoiceUnpaid":
		data, err := unmarshal[InvoiceUnpaid](body)
		if err != nil {
			log.Error("Error decoding request", zap.Error(err))
			return ps.NoNackErr(err)
		}
		log = log.With(zap.Int("invoice_id", data.InvoiceId))
		innerErr = g.syncWhmcsInvoice(ctx, data.InvoiceId)
	case "UpdateInvoiceTotal":
		data, err := unmarshal[UpdateInvoiceTotal](body)
		if err != nil {
			log.Error("Error decoding request", zap.Error(err))
			return ps.NoNackErr(err)
		}
		log = log.With(zap.Int("invoice_id", data.InvoiceId))
		innerErr = g.syncWhmcsInvoice(ctx, data.InvoiceId)
	case "InvoiceCreation", "InvoiceCreated":
		g.m.Lock()
		defer g.m.Unlock()
		data, err := unmarshal[InvoiceCreated](body)
		if err != nil {
			log.Error("Error decoding request", zap.Error(err))
			return ps.NoNackErr(err)
		}
		log = log.With(zap.Int("invoice_id", data.InvoiceId))
		innerErr = g.invoiceCreatedHandler(ctx, log, data)
	default:
		log.Warn("Unknown event", zap.String("event", resp.Event))
		return nil
	}
	if innerErr != nil {
		log.Error("Error handling event", zap.Error(innerErr))
		return innerErr
	}
	return nil
}
