package billing

import (
	"connectrpc.com/connect"
	"context"
	"fmt"
	"github.com/arangodb/go-driver"
	bpb "github.com/slntopp/nocloud-proto/billing"
	pb "github.com/slntopp/nocloud-proto/billing/promocodes"
	epb "github.com/slntopp/nocloud-proto/events"
	"github.com/slntopp/nocloud/pkg/graph"
	"github.com/slntopp/nocloud/pkg/nocloud/payments"
	"github.com/slntopp/nocloud/pkg/nocloud/schema"
	ps "github.com/slntopp/nocloud/pkg/pubsub"
	"github.com/slntopp/nocloud/pkg/pubsub/billing"
	"github.com/slntopp/nocloud/pkg/pubsub/services_registry"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
	"strings"
	"time"
)

func (s *BillingServiceServer) ProcessInvoiceWhmcsSync(log *zap.Logger, ctx context.Context, event *epb.Event) error {
	if event.GetData()["gw-callback"].GetBoolValue() {
		log.Info("skipped gateway callback event")
		return nil
	}
	inv, err := s.invoices.Get(ctx, event.GetUuid())
	if err != nil {
		return fmt.Errorf("failed to get invoice: %w", err)
	}
	log.Debug("getting account", zap.String("uuid", inv.GetAccount()))
	acc, err := s.accounts.Get(ctx, inv.GetAccount())
	if err != nil {
		return fmt.Errorf("failed to get account: %w", err)
	}
	if acc.GetPaymentsGateway() != "" && acc.GetPaymentsGateway() != "whmcs" {
		return nil
	}
	gw := payments.GetPaymentGateway(acc.GetPaymentsGateway())

	if event.GetKey() == billing.InvoiceCreated {
		if err = gw.CreateInvoice(ctx, inv.Invoice); err != nil {
			return fmt.Errorf("failed to create invoice on whmcs: %w", err)
		}
		return nil
	}

	if err = gw.UpdateInvoice(ctx, inv.Invoice); err != nil {
		return fmt.Errorf("failed to update invoice on whmcs: %w", err)
	}
	return nil
}

func (s *BillingServiceServer) ConsumeInvoiceWhmcsSync(log *zap.Logger, ctx context.Context, p *ps.PubSub[*epb.Event]) {
	log = log.Named("ConsumeInvoiceWhmcsSync")
	msgs, err := p.Consume("whmcs-sync", ps.DEFAULT_EXCHANGE, billing.Topic("invoices"))
	if err != nil {
		log.Fatal("Failed to start consumer")
		return
	}

	for msg := range msgs {
		var event epb.Event
		if err = proto.Unmarshal(msg.Body, &event); err != nil {
			log.Error("Failed to unmarshal event. Incorrect delivery", zap.Error(err))
			if err = msg.Ack(false); err != nil {
				log.Error("Failed to acknowledge the delivery", zap.Error(err))
			}
			continue
		}
		log.Debug("Pubsub event received", zap.Any("event", event))
		if err = s.ProcessInvoiceWhmcsSync(log, ctx, &event); err != nil {
			log.Error("Failed to process whmcs sync", zap.Error(err))
			if err = msg.Nack(false, false); err != nil {
				log.Error("Failed to nack the delivery", zap.Error(err))
			}
		}
		if err = msg.Ack(false); err != nil {
			log.Error("Failed to acknowledge the delivery", zap.Error(err))
		}
	}
}

func (s *BillingServiceServer) ProcessInvoiceStatusAction(log *zap.Logger, ctx context.Context, event *epb.Event) error {
	if event.GetKey() != billing.InvoicePaid && event.GetKey() != billing.InvoiceReturned {
		return nil
	}

	currConf := MakeCurrencyConf(log, &s.settingsClient)

	inv, err := s.invoices.Get(ctx, event.GetUuid())
	if err != nil {
		return fmt.Errorf("failed to get invoice: %w", err)
	}

	if event.GetKey() == billing.InvoicePaid {
		if inv, err = s.executePostPaidActions(ctx, log, inv, currConf.Currency); err != nil {
			return fmt.Errorf("failed to execute postpaid actions: %w", err)
		}
	}

	if event.GetKey() == billing.InvoiceReturned {
		if inv, err = s.executePostRefundActions(ctx, log, inv); err != nil {
			return fmt.Errorf("failed to execute postrefund actions: %w", err)
		}
	}

	inv.Processed = time.Now().Unix()
	if _, err = s.invoices.Update(ctx, inv); err != nil {
		return fmt.Errorf("failed to update invoice: %w", err)
	}

	return nil
}

func (s *BillingServiceServer) ConsumeInvoiceStatusActions(log *zap.Logger, ctx context.Context, p *ps.PubSub[*epb.Event]) {
	log = log.Named("ConsumeInvoiceStatusActions")
	msgs, err := p.Consume("postpaid-postrefund-actions", ps.DEFAULT_EXCHANGE, billing.Topic("invoices"))
	if err != nil {
		log.Fatal("Failed to start consumer")
		return
	}

	for msg := range msgs {
		var event epb.Event
		if err = proto.Unmarshal(msg.Body, &event); err != nil {
			log.Error("Failed to unmarshal event. Incorrect delivery", zap.Error(err))
			if err = msg.Ack(false); err != nil {
				log.Error("Failed to acknowledge the delivery", zap.Error(err))
			}
			continue
		}
		log.Debug("Pubsub event received", zap.String("key", event.Key), zap.String("type", event.Type))
		if err = s.ProcessInvoiceStatusAction(log, ctx, &event); err != nil {
			log.Error("Failed to process invoice action", zap.Error(err))
			if err = msg.Nack(false, false); err != nil {
				log.Error("Failed to nack the delivery", zap.Error(err))
			}
		}
		if err = msg.Ack(false); err != nil {
			log.Error("Failed to acknowledge the delivery", zap.Error(err))
		}
	}
}

func (s *BillingServiceServer) ProcessInstanceCreation(log *zap.Logger, ctx context.Context, event *epb.Event, currencyConf CurrencyConf, now int64) error {
	log = s.log.Named("ProcessInstanceCreation")
	log = log.With(zap.String("instance", event.Uuid))
	rootId := driver.NewDocumentID(schema.ACCOUNTS_COL, schema.ROOT_ACCOUNT_KEY)

	if event.GetKey() != services_registry.InstanceCreated {
		return nil
	}

	instance, err := s.instances.GetWithAccess(ctx, rootId, event.GetUuid())
	if err != nil {
		log.Error("Failed to get instance", zap.Error(err))
		return err
	}

	// Create promocode on newly created instance if it was passed on creation
	if promo, ok := event.GetData()["promocode"]; ok {
		if err = s.promocodes.AddEntry(ctx, promo.GetStringValue(), &pb.EntryResource{
			Instance: &event.Uuid,
		}); err != nil {
			log.Error("FATAL: Failed to link instance with promocode on instance creation", zap.Error(err), zap.String("promocode", promo.GetStringValue()))
		}
	}

	// Find owner account
	cur, err := s.db.Query(ctx, instanceOwner, map[string]interface{}{
		"instance":    instance.GetUuid(),
		"permissions": schema.PERMISSIONS_GRAPH.Name,
		"@instances":  schema.INSTANCES_COL,
		"@accounts":   schema.ACCOUNTS_COL,
	})
	if err != nil {
		log.Error("Error getting instance owner. Failed to execute query", zap.Error(err))
		return fmt.Errorf("error getting instance owner: %w", err)
	}
	var acc graph.Account
	_, err = cur.ReadDocument(ctx, &acc)
	if err != nil {
		log.Error("Error getting instance owner. Failed to read from cursor", zap.Error(err))
		return fmt.Errorf("failed to get instance owner: %w", err)
	}
	acc.Uuid = acc.Key
	if acc.GetUuid() == "" {
		log.Error("Instance owner not found. Uuid is empty")
		return fmt.Errorf("instance owner not found. Uuid is empty")
	}
	log.Debug("Instance owner found", zap.String("account", acc.GetUuid()))

	acc, err = s.accounts.GetAccountOrOwnerAccountIfPresent(ctxWithRoot(ctx), acc.GetUuid())
	if err != nil {
		log.Error("Failed to get account", zap.Error(err))
		return fmt.Errorf("failed to get account: %w", err)
	}
	var accCurrency = currencyConf.Currency
	if acc.Currency != nil {
		accCurrency = acc.Currency
	}
	initCost, _ := s.instances.CalculateInstanceEstimatePrice(instance.Instance, true)
	cost, err := s.promocodes.GetDiscountPriceByInstance(instance.Instance, true)
	if err != nil {
		log.Error("Failed to calculate instance cost", zap.Error(err))
		return fmt.Errorf("failed to calculate instance cost: %w", err)
	}
	cost, err = s.currencies.Convert(ctx, currencyConf.Currency, accCurrency, cost)
	if err != nil {
		log.Error("Failed to convert cost", zap.Error(err))
		return fmt.Errorf("failed to convert cost: %w", err)
	}

	bp := instance.GetBillingPlan()
	product, hasProduct := bp.GetProducts()[instance.GetProduct()]
	if !hasProduct {
		log.Warn("Product not found in billing plan", zap.String("product", instance.GetProduct()))
	}
	invoicePrefixVal, _ := bp.GetMeta()["prefix"]
	invoicePrefix := invoicePrefixVal.GetStringValue() + " "
	productTitle := product.GetTitle() + " "
	startDescription := fmt.Sprintf("%s%s", invoicePrefix, productTitle)
	startDescription = strings.TrimSpace(startDescription)

	tax := acc.GetTaxRate()
	invCost := cost + cost*tax

	inv := &bpb.Invoice{
		Status: bpb.BillingStatus_UNPAID,
		Items: []*bpb.Item{
			{
				Description: startDescription,
				Amount:      1,
				Unit:        "Pcs",
				Price:       invCost,
			},
		},
		Meta: map[string]*structpb.Value{
			"creator":               structpb.NewStringValue("system"),
			"no_discount_price":     structpb.NewStringValue(fmt.Sprintf("%.2f %s", initCost, currencyConf.Currency.GetTitle())),
			graph.InvoiceTaxMetaKey: structpb.NewNumberValue(tax),
		},
		Total:     invCost,
		Type:      bpb.ActionType_INSTANCE_START,
		Instances: []string{instance.GetUuid()},
		Created:   now,
		Deadline:  now + (int64(time.Hour.Seconds()) * 24 * 5),
		Account:   acc.GetUuid(),
		Currency:  accCurrency,
	}
	invResp, err := s.CreateInvoice(ctxWithRoot(ctx), connect.NewRequest(&bpb.CreateInvoiceRequest{
		Invoice:     inv,
		IsSendEmail: true,
	}))
	if err != nil {
		log.Error("Failed to create invoice", zap.Error(err))
		return fmt.Errorf("failed to create invoice: %w", err)
	}
	log.Info("Created invoice", zap.String("uuid", invResp.Msg.GetUuid()))
	return nil
}

func (s *BillingServiceServer) ConsumeCreatedInstances(log *zap.Logger, ctx context.Context, p *ps.PubSub[*epb.Event]) {
	log = s.log.Named("ConsumeCreatedInstances")
	msgs, err := p.Consume("created-instance-start-invoice", ps.DEFAULT_EXCHANGE, services_registry.Topic("instances"))
	if err != nil {
		log.Fatal("Failed to start consumer")
		return
	}

	for msg := range msgs {
		var event epb.Event
		if err = proto.Unmarshal(msg.Body, &event); err != nil {
			log.Error("Failed to unmarshal event. Incorrect delivery", zap.Error(err))
			if err = msg.Ack(false); err != nil {
				log.Error("Failed to acknowledge the delivery", zap.Error(err))
			}
			continue
		}
		curConf := MakeCurrencyConf(log, &s.settingsClient)
		log.Debug("Pubsub event received", zap.String("key", event.Key), zap.String("type", event.Type))
		if err = s.ProcessInstanceCreation(log, ctx, &event, curConf, time.Now().Unix()); err != nil {
			log.Error("Failed to process invoice action", zap.Error(err))
			if err = msg.Nack(false, false); err != nil {
				log.Error("Failed to nack the delivery", zap.Error(err))
			}
		}
		if err = msg.Ack(false); err != nil {
			log.Error("Failed to acknowledge the delivery", zap.Error(err))
		}
	}
}
