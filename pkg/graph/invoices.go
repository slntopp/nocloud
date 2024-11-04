package graph

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/arangodb/go-driver"
	pb "github.com/slntopp/nocloud-proto/billing"
	"github.com/slntopp/nocloud/pkg/graph/migrations"
	"github.com/slntopp/nocloud/pkg/nocloud/schema"
	"go.uber.org/zap"
	"strings"
	"time"
)

type InvoicesController interface {
	TransactionsAPI
	DecodeInvoice(source interface{}, dest *Invoice) error
	ParseNumberIntoTemplate(template string, number int, date time.Time) string
	Create(ctx context.Context, tx *Invoice) (*Invoice, error)
	Get(ctx context.Context, uuid string) (*Invoice, error)
	Update(ctx context.Context, tx *Invoice) (*Invoice, error)
	Patch(ctx context.Context, id string, patch map[string]interface{}) error
	List(ctx context.Context) ([]*Invoice, error)
}

type InvoiceNumberMeta struct {
	NumericNumber  int    `json:"numeric_number"`
	NumberTemplate string `json:"number_template"`
}

type Invoice struct {
	*pb.Invoice
	*InvoiceNumberMeta
	driver.DocumentMeta
}

type invoicesController struct {
	col driver.Collection // Billing Plans collection

	log *zap.Logger
}

func NewInvoicesController(logger *zap.Logger, db driver.Database) InvoicesController {
	ctx := context.TODO()
	log := logger.Named("InvoicesController")
	col := GetEnsureCollection(log, ctx, db, schema.INVOICES_COL)

	migrations.UpdateNumericCurrencyToDynamic(log, col)

	return &invoicesController{
		log: log, col: col,
	}
}

func (ctrl *invoicesController) BeginTransaction(ctx context.Context) (context.Context, error) {
	cols := driver.TransactionCollections{
		Write: []string{schema.INVOICES_COL},
	}
	if trID, ok := ctx.Value(AQLTransactionContextKey).(string); ok {
		if status, err := ctrl.col.Database().TransactionStatus(ctx, driver.TransactionID(trID)); err == nil && status.Status == driver.TransactionRunning {
			return ctx, errors.New("already processing another transaction")
		}
	}
	trID, err := ctrl.col.Database().BeginTransaction(ctx, cols, &driver.BeginTransactionOptions{})
	if err != nil {
		return ctx, fmt.Errorf("error while starting transaction: %w", err)
	}
	ctx = driver.WithTransactionID(ctx, trID)
	return context.WithValue(ctx, AQLTransactionContextKey, string(trID)), nil
}

func (ctrl *invoicesController) CommitTransaction(ctx context.Context) error {
	trID, _ := ctx.Value(AQLTransactionContextKey).(string)
	err := ctrl.col.Database().CommitTransaction(ctx, driver.TransactionID(trID), &driver.CommitTransactionOptions{})
	if err != nil {
		ctrl.log.Error("Failed to commit transaction", zap.Error(err))
	}
	return err
}

func (ctrl *invoicesController) AbortTransaction(ctx context.Context) error {
	trID, _ := ctx.Value(AQLTransactionContextKey).(string)
	err := ctrl.col.Database().AbortTransaction(ctx, driver.TransactionID(trID), &driver.AbortTransactionOptions{})
	if err != nil {
		ctrl.log.Error("Failed to abort transaction", zap.Error(err))
	}
	return err
}

func (ctrl *invoicesController) DecodeInvoice(source interface{}, dest *Invoice) error {
	bytes, err := json.Marshal(source)
	if err != nil {
		return err
	}
	if err = json.Unmarshal(bytes, dest.Invoice); err != nil {
		return err
	}
	if err = json.Unmarshal(bytes, dest.InvoiceNumberMeta); err != nil {
		return err
	}

	return nil
}

func (ctrl *invoicesController) ParseNumberIntoTemplate(template string, number int, date time.Time) string {
	year := date.Year()
	month := int(date.Month())
	day := date.Day()
	template = strings.Replace(template, "{YEAR}", fmt.Sprintf("%d", year), -1)
	template = strings.Replace(template, "{MONTH}", fmt.Sprintf("%d", month), -1)
	template = strings.Replace(template, "{DAY}", fmt.Sprintf("%d", day), -1)
	template = strings.Replace(template, "{NUMBER}", fmt.Sprintf("%d", number), -1)
	return template
}

func (ctrl *invoicesController) Create(ctx context.Context, tx *Invoice) (*Invoice, error) {
	if tx.GetAccount() == "" {
		return nil, errors.New("account is required")
	}
	if tx.Total == 0 {
		return nil, errors.New("total is required")
	}
	meta, err := ctrl.col.CreateDocument(driver.WithWaitForSync(ctx, true), tx)
	if err != nil {
		ctrl.log.Error("Failed to create transaction", zap.Error(err))
		return nil, err
	}
	tx.Uuid = meta.Key
	return tx, nil
}

func (ctrl *invoicesController) Get(ctx context.Context, uuid string) (*Invoice, error) {
	var tx = &Invoice{
		Invoice:           &pb.Invoice{},
		InvoiceNumberMeta: &InvoiceNumberMeta{},
	}
	result := map[string]interface{}{}
	meta, err := ctrl.col.ReadDocument(ctx, uuid, &result)
	if err != nil {
		ctrl.log.Error("Failed to read invoice", zap.Error(err))
		return nil, err
	}
	if err = ctrl.DecodeInvoice(result, tx); err != nil {
		ctrl.log.Error("Failed to decode invoice", zap.Error(err))
		return nil, err
	}

	tx.Uuid = meta.Key
	return tx, err
}

func (ctrl *invoicesController) Update(ctx context.Context, tx *Invoice) (*Invoice, error) {
	_, err := ctrl.col.UpdateDocument(driver.WithWaitForSync(ctx, true), tx.GetUuid(), tx)
	if err != nil {
		ctrl.log.Error("Failed to update invoice", zap.Error(err))
		return nil, err
	}
	return tx, nil
}

func (ctrl *invoicesController) Patch(ctx context.Context, id string, patch map[string]interface{}) error {
	_, err := ctrl.col.UpdateDocument(driver.WithWaitForSync(ctx, true), id, patch)
	return err
}

func (ctrl *invoicesController) List(ctx context.Context) ([]*Invoice, error) {
	result := make([]*Invoice, 0)
	cur, err := ctrl.col.Database().Query(ctx, "FOR doc IN "+ctrl.col.Name()+" RETURN MERGE(doc, {uuid: doc._key})", nil)
	if err != nil {
		ctrl.log.Error("Failed to list invoices", zap.Error(err))
		return nil, err
	}
	defer cur.Close()
	for cur.HasMore() {
		var tx = &Invoice{
			Invoice:           &pb.Invoice{},
			InvoiceNumberMeta: &InvoiceNumberMeta{},
		}
		meta, err := cur.ReadDocument(ctx, tx)
		if err != nil {
			ctrl.log.Error("Failed to read invoice", zap.Error(err))
			return nil, err
		}
		tx.Uuid = meta.Key
		result = append(result, tx)
	}
	return result, err
}
