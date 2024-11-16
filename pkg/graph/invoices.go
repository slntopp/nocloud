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
	List(ctx context.Context, account string) ([]*Invoice, error)
	Transfer(ctx context.Context, uuid string, account string, resCurr *pb.Currency) (err error)
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

	transactions TransactionsController
	currencies   CurrencyController

	log *zap.Logger
}

func NewInvoicesController(logger *zap.Logger, db driver.Database) InvoicesController {
	ctx := context.TODO()
	log := logger.Named("InvoicesController")
	col := GetEnsureCollection(log, ctx, db, schema.INVOICES_COL)

	transactions := NewTransactionsController(log, db)
	currencies := NewCurrencyController(log, db)

	migrations.UpdateNumericCurrencyToDynamic(log, col)

	return &invoicesController{
		log: log, col: col, transactions: transactions, currencies: currencies,
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
	_, err := ctrl.col.UpdateDocument(ctx, tx.GetUuid(), tx)
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

func (ctrl *invoicesController) Transfer(ctx context.Context, uuid string, account string, resCurr *pb.Currency) (err error) {
	if resCurr == nil {
		return fmt.Errorf("currency is required")
	}

	_, hasTransaction := driver.HasTransactionID(ctx)
	var trId driver.TransactionID
	if !hasTransaction {
		db := ctrl.col.Database()
		trId, err = db.BeginTransaction(ctx, driver.TransactionCollections{
			Exclusive: []string{schema.INVOICES_COL, schema.TRANSACTIONS_COL},
		}, &driver.BeginTransactionOptions{})
		if err != nil {
			return err
		}
		ctx = driver.WithTransactionID(ctx, trId)
		defer func(err *error) {
			if panicErr := recover(); panicErr != nil {
				_ = db.AbortTransaction(ctx, trId, &driver.AbortTransactionOptions{})
				ctrl.log.Warn("Recovered from panic", zap.Any("error", panicErr))
				return
			}
			if err != nil {
				_ = db.AbortTransaction(ctx, trId, &driver.AbortTransactionOptions{})
				return
			}
		}(&err)
	}

	var inv *Invoice
	inv, err = ctrl.Get(ctx, uuid)
	if err != nil {
		return fmt.Errorf("failed to get invoice: %w", err)
	}
	oldAcc := inv.Account
	if oldAcc == account {
		err = fmt.Errorf("can't transfer to the same account")
		return err
	}
	inv.Account = account

	prevTotal := inv.Total
	newTotal, err := ctrl.currencies.Convert(ctx, inv.Currency, resCurr, inv.Total)
	if err != nil {
		return fmt.Errorf("failed to convert currency: %w", err)
	}
	for _, item := range inv.GetItems() {
		item.Price = item.Price * (newTotal / prevTotal)
	}
	inv.Total = newTotal
	inv.Currency = resCurr

	if _, err = ctrl.Update(ctx, inv); err != nil {
		return fmt.Errorf("failed to update invoice: %w", err)
	}
	for _, id := range inv.GetTransactions() {
		var tr *pb.Transaction
		tr, err = ctrl.transactions.Get(ctx, id)
		if err != nil {
			return fmt.Errorf("failed to get transaction: %w", err)
		}
		tr.Account = account
		if _, err = ctrl.transactions.Update(ctx, tr); err != nil {
			return fmt.Errorf("failed to update transaction: %w", err)
		}
	}

	if !hasTransaction {
		if err = ctrl.col.Database().CommitTransaction(ctx, trId, &driver.CommitTransactionOptions{}); err != nil {
			return err
		}
	}
	return nil
}

const listInvoices = `
    FOR doc IN @@collection
    %s
    RETURN MERGE(doc, {uuid: doc._key})
`

func (ctrl *invoicesController) List(ctx context.Context, account string) ([]*Invoice, error) {
	log := ctrl.log.Named("List")

	list := make([]*Invoice, 0)
	bindVars := map[string]interface{}{
		"@collection": schema.INVOICES_COL,
	}
	var filters = ""
	if account != "" {
		filters = ` FILTER doc.account == @account`
		bindVars["account"] = account
	}
	cur, err := ctrl.col.Database().Query(ctx, fmt.Sprintf(listInvoices, filters), bindVars)
	if err != nil {
		log.Error("Failed to list invoices", zap.Error(err))
		return nil, err
	}
	defer cur.Close()
	for cur.HasMore() {
		var tx = &Invoice{
			Invoice:           &pb.Invoice{},
			InvoiceNumberMeta: &InvoiceNumberMeta{},
		}
		result := map[string]interface{}{}
		meta, err := cur.ReadDocument(ctx, &result)
		if err != nil {
			log.Error("Failed to read invoice", zap.Error(err))
			return nil, err
		}
		if err = ctrl.DecodeInvoice(result, tx); err != nil {
			log.Error("Failed to decode invoice", zap.Error(err))
			return nil, err
		}
		tx.Uuid = meta.Key
		list = append(list, tx)
	}
	return list, err
}
