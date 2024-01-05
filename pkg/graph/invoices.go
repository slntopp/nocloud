package graph

import (
	"context"
	"errors"
	"github.com/arangodb/go-driver"
	pb "github.com/slntopp/nocloud-proto/billing"
	"github.com/slntopp/nocloud/pkg/nocloud/schema"
	"go.uber.org/zap"
)

type InvoicesController struct {
	col driver.Collection // Billing Plans collection

	log *zap.Logger
}

func NewInvoicesController(logger *zap.Logger, db driver.Database) InvoicesController {
	ctx := context.TODO()
	log := logger.Named("InvoicesController")
	col := GetEnsureCollection(log, ctx, db, schema.INVOICES_COL)
	return InvoicesController{
		log: log, col: col,
	}
}

func (ctrl *InvoicesController) Create(ctx context.Context, tx *pb.Invoice) (*pb.Invoice, error) {
	if tx.GetAccount() == "" {
		return nil, errors.New("account is required")
	}
	if tx.Total == 0 {
		return nil, errors.New("total is required")
	}
	meta, err := ctrl.col.CreateDocument(ctx, tx)
	if err != nil {
		ctrl.log.Error("Failed to create transaction", zap.Error(err))
		return nil, err
	}
	tx.Uuid = meta.Key
	return tx, nil
}

func (ctrl *InvoicesController) Get(ctx context.Context, uuid string) (*pb.Invoice, error) {
	var tx pb.Invoice
	_, err := ctrl.col.ReadDocument(ctx, uuid, &tx)
	if err != nil {
		ctrl.log.Error("Failed to read transaction", zap.Error(err))
		return nil, err
	}
	return &tx, nil
}

func (ctrl *InvoicesController) Update(ctx context.Context, tx *pb.Invoice) (*pb.Invoice, error) {
	_, err := ctrl.col.UpdateDocument(ctx, tx.GetUuid(), tx)
	if err != nil {
		ctrl.log.Error("Failed to update transaction", zap.Error(err))
		return nil, err
	}
	return tx, nil
}
