package graph

import (
	"context"
	"fmt"
	"github.com/arangodb/go-driver"
	bpb "github.com/slntopp/nocloud-proto/billing"
	"github.com/slntopp/nocloud/pkg/nocloud/schema"
	"go.uber.org/zap"
)

type PaymentGatewaysController interface {
	Create(ctx context.Context, cat *bpb.PaymentGateway) (*bpb.PaymentGateway, error)
	Update(ctx context.Context, cat *bpb.PaymentGateway) (*bpb.PaymentGateway, error)
	List(ctx context.Context, enabled bool) ([]*bpb.PaymentGateway, error)
	Get(ctx context.Context, key string) (*bpb.PaymentGateway, error)
	Delete(ctx context.Context, key string) error
}

type paymentGatewaysController struct {
	log *zap.Logger
	col driver.Collection

	db driver.Database
}

func NewPaymentGatewaysController(logger *zap.Logger, db driver.Database) PaymentGatewaysController {
	ctx := context.Background()
	log := logger.Named("PaymentGatewaysController")
	log.Debug("New Payment Gateways Controller Creating")

	col := GetEnsureCollection(log, ctx, db, schema.PAYMENT_GATEWAYS_COL)

	return &paymentGatewaysController{
		log: log,
		col: col,
		db:  db,
	}
}

func (ctrl *paymentGatewaysController) Create(ctx context.Context, cat *bpb.PaymentGateway) (*bpb.PaymentGateway, error) {
	ctrl.log.Debug("Creating Document for Payment Gateway", zap.Any("config", cat))

	if cat.Key == "" {
		return nil, fmt.Errorf("payment Gateway Key cannot be empty")
	}

	doc := struct {
		Key string `json:"_key,omitempty"`
	}{
		Key: cat.Key,
	}

	_, err := ctrl.col.CreateDocument(ctx, doc)
	if err != nil {
		return nil, err
	}

	_, err = ctrl.col.UpdateDocument(ctx, cat.Key, cat)
	if err != nil {
		return nil, err
	}

	return cat, err
}

func (ctrl *paymentGatewaysController) Update(ctx context.Context, cat *bpb.PaymentGateway) (*bpb.PaymentGateway, error) {
	ctrl.log.Debug("Updating PaymentGateway", zap.Any("sp", cat))

	meta, err := ctrl.col.ReplaceDocument(ctx, cat.GetKey(), cat)
	ctrl.log.Debug("ReplaceDocument.Result", zap.Any("meta", meta), zap.Error(err))
	return cat, err
}

const pgListQuery = `
FOR s IN @@pgs
	%s
	RETURN s
`

func (ctrl *paymentGatewaysController) List(ctx context.Context, enabled bool) ([]*bpb.PaymentGateway, error) {

	vars := map[string]interface{}{
		"@pgs": schema.PAYMENT_GATEWAYS_COL,
	}

	var query string
	var filters string

	if enabled {
		filters += ` FILTER s.enabled == true `
	}

	query = fmt.Sprintf(pgListQuery, "")
	c, err := ctrl.col.Database().Query(ctx, query, vars)
	if err != nil {
		return nil, err
	}

	defer c.Close()
	var r []*bpb.PaymentGateway
	for c.HasMore() {
		var s bpb.PaymentGateway

		_, err := c.ReadDocument(ctx, &s)

		if err != nil {
			return nil, err
		}

		r = append(r, &s)
	}

	return r, nil
}

func (ctrl *paymentGatewaysController) Get(ctx context.Context, uuid string) (*bpb.PaymentGateway, error) {
	var pg bpb.PaymentGateway
	_, err := ctrl.col.ReadDocument(ctx, uuid, &pg)
	return &pg, err
}

func (ctrl *paymentGatewaysController) Delete(ctx context.Context, uuid string) error {
	ctrl.log.Debug("Deleting PaymentGateway", zap.Any("uuid", uuid))

	meta, err := ctrl.col.RemoveDocument(ctx, uuid)
	ctrl.log.Debug("RemoveDocument.Result", zap.Any("meta", meta), zap.Error(err))
	return err
}
