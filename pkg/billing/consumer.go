package billing

import (
	"connectrpc.com/connect"
	"context"
	"fmt"
	bpb "github.com/slntopp/nocloud-proto/billing"
	pb "github.com/slntopp/nocloud-proto/billing/promocodes"
	epb "github.com/slntopp/nocloud-proto/events"
	healthpb "github.com/slntopp/nocloud-proto/health"
	"github.com/slntopp/nocloud/pkg/graph"
	"github.com/slntopp/nocloud/pkg/nocloud/schema"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
	"time"
)

func (s *BillingServiceServer) ProcessInstanceCreation(log *zap.Logger, ctx context.Context, event *epb.Event, currencyConf CurrencyConf, now int64) error {
	log = s.log.Named("ProcessInstanceCreation")
	log = log.With(zap.String("instance", event.Uuid))

	instance, err := s.instances.Get(ctx, event.Uuid)
	if err != nil {
		log.Error("Failed to get instance", zap.Error(err))
		return err
	}
	if instance == nil {
		log.Error("Failed to get instance. Instance is nil")
		return fmt.Errorf("failed to get instance. Instance is nil")
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
	costNoOneTime, _ := s.promocodes.GetDiscountPriceByInstance(instance.Instance, true, true)
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

	inv := &bpb.Invoice{
		Status: bpb.BillingStatus_UNPAID,
		Items: []*bpb.Item{
			{
				Description: fmt.Sprintf("Instance «%s» start payment", instance.GetTitle()),
				Amount:      1,
				Unit:        "Pcs",
				Price:       cost,
				Instance:    instance.GetUuid(),
			},
		},
		Meta: map[string]*structpb.Value{
			"creator":           structpb.NewStringValue("nocloud.billing.ProcessInstanceCreation"),
			"no_discount_price": structpb.NewStringValue(fmt.Sprintf("%.2f %s", initCost, currencyConf.Currency.GetTitle())),
		},
		Total:    cost,
		Type:     bpb.ActionType_INSTANCE_START,
		Created:  now,
		Deadline: now + (int64(time.Hour.Seconds()) * 24),
		Account:  acc.GetUuid(),
		Currency: accCurrency,
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

	// Create fixing transaction
	delta := costNoOneTime - cost
	if delta == 0 {
		return nil
	}
	if delta < 0 {
		log.Warn("Delta is less than 0. Price after promocode is higher than before promocode", zap.Float64("delta", delta))
		return nil
	}
	log.Info("Creating fixing transaction")
	delta, err = s.currencies.Convert(ctx, currencyConf.Currency, accCurrency, delta)
	if err != nil {
		log.Error("Failed to convert cost to create fixing transaction", zap.Error(err))
		return fmt.Errorf("failed to convert cost to create fixing transaction: %w", err)
	}
	newInv, err := s.invoices.Get(ctxWithRoot(ctx), invResp.Msg.GetUuid())
	if err != nil {
		log.Error("Failed to get invoice", zap.Error(err))
		return fmt.Errorf("failed to get invoice to create fixing transaction: %w", err)
	}
	newTr, err := s.CreateTransaction(ctxWithRoot(ctx), connect.NewRequest(&bpb.Transaction{
		Priority: bpb.Priority_NORMAL,
		Account:  acc.GetUuid(),
		Currency: accCurrency,
		Total:    -delta,
		Exec:     0,
	}))
	if err != nil {
		log.Error("Failed to create transaction", zap.Error(err))
		return fmt.Errorf("failed to create fixing transaction: %w", err)
	}
	if newInv.Transactions == nil {
		newInv.Transactions = []string{}
	}
	newInv.Transactions = append(newInv.Transactions, newTr.Msg.GetUuid())
	_, err = s.invoices.Update(ctxWithRoot(ctx), newInv)
	if err != nil {
		log.Error("Failed to update invoice to save fixing transaction", zap.Error(err))
		return fmt.Errorf("failed to update invoice to save fixing transaction: %w", err)
	}
	log.Info("Successfully created fixing transaction")
	return nil
}

func (s *BillingServiceServer) ConsumeCreatedInstances(ctx context.Context) {
	log := s.log.Named("ConsumeCreatedInstances")
init:
	ch, err := s.rbmq.Channel()
	if err != nil {
		log.Error("Failed to open a channel", zap.Error(err))
		time.Sleep(time.Second)
		goto init
	}

	queue, _ := ch.QueueDeclare(
		"created_instance",
		true, false, false, true, nil,
	)

	if err = ch.QueueBind("created_instance", "created_instance", "instances", false, nil); err != nil {
		log.Error("Failed to bind queue", zap.Error(err))
		time.Sleep(time.Second)
		goto init
	}

	records, err := ch.Consume(queue.Name, "", false, false, false, false, nil)
	if err != nil {
		log.Error("Failed to register a consumer", zap.Error(err))
		time.Sleep(time.Second)
		goto init
	}

	s.InstancesConsumerStatus.Status.Status = healthpb.Status_RUNNING
	currencyConf := MakeCurrencyConf(ctx, log, &s.settingsClient)

	for msg := range records {
		log.Debug("Received a message")
		var event epb.Event
		err = proto.Unmarshal(msg.Body, &event)
		log.Debug("Message unmarshalled", zap.Any("event", &event))
		if err != nil {
			log.Error("Failed to unmarshal event", zap.Error(err))
			if err = msg.Ack(false); err != nil {
				log.Warn("Failed to Acknowledge the delivery while unmarshal message", zap.Error(err))
			}
			continue
		}
		err := s.ProcessInstanceCreation(log, ctx, &event, currencyConf, time.Now().Unix())
		if err != nil {
			log.Error("Failed to process event", zap.Error(err))
		}
		if err = msg.Ack(false); err != nil {
			log.Warn("Failed to Acknowledge the delivery while unmarshal message", zap.Error(err))
		}
		continue
	}
}
