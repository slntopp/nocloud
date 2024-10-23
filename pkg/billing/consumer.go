package billing

import (
	"connectrpc.com/connect"
	"context"
	"fmt"
	bpb "github.com/slntopp/nocloud-proto/billing"
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

	instance, err := s.instances.Get(ctx, event.Uuid)
	if err != nil {
		log.Error("Failed to get instance", zap.Error(err))
		return err
	}
	if instance == nil {
		log.Error("Failed to get instance. Instance is nil")
		return fmt.Errorf("failed to get instance. Instance is nil")
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
	cost, err := s.instances.CalculateInstanceEstimatePrice(instance.Instance, true)
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
				Description: fmt.Sprintf("Instance '%s' start payment", instance.GetUuid()),
				Amount:      1,
				Unit:        "Pcs",
				Price:       cost,
				Instance:    instance.GetUuid(),
			},
		},
		Meta: map[string]*structpb.Value{
			"auto_created": structpb.NewBoolValue(true),
			"creator":      structpb.NewStringValue("nocloud.billing.ProcessInstanceCreation"),
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
	currencyConf := MakeCurrencyConf(ctx, log)

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
