package billing

import (
	"context"
	"github.com/arangodb/go-driver"
	"github.com/google/uuid"
	pb "github.com/slntopp/nocloud-proto/billing"
	driver_mocks "github.com/slntopp/nocloud/mocks/github.com/arangodb/go-driver"
	"github.com/slntopp/nocloud/pkg/nocloud/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
	"testing"
	"time"
)

func currMigrationMock(t *testing.T, db *driver_mocks.MockDatabase, col *driver_mocks.MockCollection) {
	col.On("Database").Return(db)
	col.On("Name").Return("Records")
	cur := driver_mocks.NewMockCursor(t)
	db.On("Query", mock.Anything, mock.Anything, mock.Anything).Return(cur, nil).Times(1)
	cur.On("ReadDocument", mock.Anything, mock.Anything).Return(driver.DocumentMeta{}, nil).Maybe()
	cur.On("Close").Return(nil).Maybe()
}

func TestConsumeMock(t *testing.T) {
	ctx := context.TODO()

	db := driver_mocks.NewMockDatabase(t)
	col := driver_mocks.NewMockCollection(t)
	cursor := driver_mocks.NewMockCursor(t)

	currMigrationMock(t, db, col)
	db.On("CollectionExists", ctx, "Records").Return(true, nil)
	db.On("Collection", ctx, "Records").Return(col, nil)
	ctrl := NewRecordsServiceServer(zap.NewExample(), nil, db)

	now := time.Now().Unix()

	var record = pb.Record{
		Start:    now,
		End:      now + 24*60*60,
		Exec:     now,
		Priority: pb.Priority_NORMAL,
		Instance: uuid.New().String(),
		Resource: uuid.New().String(),
		Total:    1,
		Currency: &pb.Currency{Id: 0, Name: "NCU"},
	}

	conf := CurrencyConf{
		Currency: &pb.Currency{Id: 0, Name: "NCU"},
	}

	db.On("Query", ctx, checkOverlap, map[string]interface{}{
		"record":   &record,
		"@records": schema.RECORDS_COL,
	}).Return(cursor, nil)
	cursor.On("Close").Return(nil)

	cursor.On("ReadDocument", ctx, mock.Anything).Return(driver.DocumentMeta{}, nil)

	err := ctrl.ProcessRecord(ctx, &record, conf, now)
	assert.Nil(t, err)
}

func TestConsumeMock_NotNormal(t *testing.T) {
	ctx := context.TODO()

	db := driver_mocks.NewMockDatabase(t)
	col := driver_mocks.NewMockCollection(t)
	cursor2 := driver_mocks.NewMockCursor(t)

	currMigrationMock(t, db, col)
	db.On("CollectionExists", ctx, "Records").Return(true, nil)
	db.On("Collection", ctx, "Records").Return(col, nil)
	ctrl := NewRecordsServiceServer(zap.NewExample(), nil, db)

	now := time.Now().Unix()

	var record = pb.Record{
		Start:    now,
		End:      now + 24*60*60,
		Exec:     now,
		Priority: pb.Priority_ADDITIONAL,
		Instance: uuid.New().String(),
		Resource: uuid.New().String(),
		Total:    1,
		Currency: &pb.Currency{Id: 0, Name: "NCU"},
	}

	conf := CurrencyConf{
		Currency: &pb.Currency{Id: 0, Name: "NCU"},
	}

	col.On("CreateDocument", ctx, &record).Return(driver.DocumentMeta{}, nil)
	db.On("Query", ctx, generateUrgentTransactions, map[string]interface{}{
		"@transactions": schema.TRANSACTIONS_COL,
		"@instances":    schema.INSTANCES_COL,
		"instances":     schema.INSTANCES_COL,
		"@services":     schema.SERVICES_COL,
		"@records":      schema.RECORDS_COL,
		"@accounts":     schema.ACCOUNTS_COL,
		"permissions":   schema.PERMISSIONS_GRAPH.Name,
		"priority":      record.Priority,
		"now":           now,
		"graph":         schema.BILLING_GRAPH.Name,
		"currencies":    schema.CUR_COL,
		"currency":      conf.Currency,
		"billing_plans": schema.BILLING_PLANS_COL,
	}).Return(cursor2, nil)
	cursor2.On("HasMore").Return(true)
	docId := driver.NewDocumentID(schema.TRANSACTIONS_COL, uuid.New().String())
	cursor2.On("ReadDocument", ctx, mock.Anything).Return(driver.DocumentMeta{
		ID: docId,
	}, nil)
	db.On("Query", ctx, processUrgentTransactions, map[string]interface{}{
		"tr":            docId,
		"@transactions": schema.TRANSACTIONS_COL,
		"@accounts":     schema.ACCOUNTS_COL,
		"accounts":      schema.ACCOUNTS_COL,
		"@records":      schema.RECORDS_COL,
		"now":           now,
		"graph":         schema.BILLING_GRAPH.Name,
		"currencies":    schema.CUR_COL,
		"currency":      conf.Currency,
	}).Return(driver_mocks.NewMockCursor(t), nil)

	err := ctrl.ProcessRecord(ctx, &record, conf, now)
	assert.Nil(t, err)
}

func TestConsumeMock_ZeroTotal(t *testing.T) {
	ctx := context.TODO()

	db := driver_mocks.NewMockDatabase(t)
	col := driver_mocks.NewMockCollection(t)

	currMigrationMock(t, db, col)
	db.On("CollectionExists", ctx, "Records").Return(true, nil)
	db.On("Collection", ctx, "Records").Return(col, nil)
	ctrl := NewRecordsServiceServer(zap.NewExample(), nil, db)

	now := time.Now().Unix()

	var record = pb.Record{
		Start:    now,
		End:      now + 24*60*60,
		Exec:     now,
		Priority: pb.Priority_ADDITIONAL,
		Instance: uuid.New().String(),
		Resource: uuid.New().String(),
		Total:    0,
		Currency: &pb.Currency{Id: 0, Name: "NCU"},
	}

	conf := CurrencyConf{
		Currency: &pb.Currency{Id: 0, Name: "NCU"},
	}

	err := ctrl.ProcessRecord(ctx, &record, conf, now)
	assert.Nil(t, err)
}

const checkOverlap = `
let n = @record
RETURN COUNT(FOR r IN @@records
FILTER r.instance == n.instance
FILTER r.resource == n.resource
FILTER (n.start < r.end && n.start >= r.start) || (n.start <= r.start && r.start > n.end)
    RETURN r) == 0
`
