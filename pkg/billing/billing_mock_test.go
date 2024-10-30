package billing_test

import (
	"context"
	"github.com/arangodb/go-driver"
	amqp "github.com/rabbitmq/amqp091-go"
	driver_mocks "github.com/slntopp/nocloud/mocks/github.com/arangodb/go-driver"
	rabbitmq_mocks "github.com/slntopp/nocloud/mocks/github.com/slntopp/nocloud/pkg/nocloud/rabbitmq"
	redisdb_mocks "github.com/slntopp/nocloud/mocks/github.com/slntopp/nocloud/pkg/nocloud/redis"
	"github.com/slntopp/nocloud/pkg/billing"
	"github.com/slntopp/nocloud/pkg/graph"
	"github.com/slntopp/nocloud/pkg/graph/migrations"
	"github.com/slntopp/nocloud/pkg/nocloud"
	"github.com/slntopp/nocloud/pkg/nocloud/schema"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"os"
	"testing"
)

type billingServiceServerFixture struct {
	srv *billing.BillingServiceServer

	mocks struct {
		db *driver_mocks.MockDatabase

		nsCol       *driver_mocks.MockCollection
		bilPlansCol *driver_mocks.MockCollection
		trCol       *driver_mocks.MockCollection
		recCol      *driver_mocks.MockCollection
		curCol      *driver_mocks.MockCollection
		cur2cur     *driver_mocks.MockCollection
		accCol      *driver_mocks.MockCollection
		credCol     *driver_mocks.MockCollection
		invCol      *driver_mocks.MockCollection
		descrCol    *driver_mocks.MockCollection
		srvCol      *driver_mocks.MockCollection
		igCol       *driver_mocks.MockCollection
		instCol     *driver_mocks.MockCollection
		addCol      *driver_mocks.MockCollection
		spCol       *driver_mocks.MockCollection

		permGraph *driver_mocks.MockGraph
		bilGraph  *driver_mocks.MockGraph
		credGraph *driver_mocks.MockGraph

		redisclient *redisdb_mocks.MockClient
		rbmqconn    *rabbitmq_mocks.MockConnection
		rbmqch      *rabbitmq_mocks.MockChannel
	}

	data struct {
		rootCtx context.Context
		ctx     context.Context
	}
}

func newBillingServiceServerFixture(t *testing.T) *billingServiceServerFixture {
	f := &billingServiceServerFixture{}

	// Initializing mocks
	f.mocks.db = driver_mocks.NewMockDatabase(t)
	f.mocks.permGraph = driver_mocks.NewMockGraph(t)
	f.mocks.bilGraph = driver_mocks.NewMockGraph(t)
	f.mocks.credGraph = driver_mocks.NewMockGraph(t)

	f.mocks.nsCol = driver_mocks.NewMockCollection(t)
	f.mocks.bilPlansCol = driver_mocks.NewMockCollection(t)
	f.mocks.trCol = driver_mocks.NewMockCollection(t)
	f.mocks.recCol = driver_mocks.NewMockCollection(t)
	f.mocks.curCol = driver_mocks.NewMockCollection(t)
	f.mocks.cur2cur = driver_mocks.NewMockCollection(t)
	f.mocks.accCol = driver_mocks.NewMockCollection(t)
	f.mocks.credCol = driver_mocks.NewMockCollection(t)
	f.mocks.invCol = driver_mocks.NewMockCollection(t)
	f.mocks.descrCol = driver_mocks.NewMockCollection(t)
	f.mocks.srvCol = driver_mocks.NewMockCollection(t)
	f.mocks.igCol = driver_mocks.NewMockCollection(t)
	f.mocks.instCol = driver_mocks.NewMockCollection(t)
	f.mocks.addCol = driver_mocks.NewMockCollection(t)
	f.mocks.spCol = driver_mocks.NewMockCollection(t)

	f.mocks.redisclient = redisdb_mocks.NewMockClient(t)
	f.mocks.rbmqconn = rabbitmq_mocks.NewMockConnection(t)
	f.mocks.rbmqch = rabbitmq_mocks.NewMockChannel(t)

	// Mock functions that being called from controllers initialization
	f.mocks.rbmqconn.On("Channel").Return(f.mocks.rbmqch, nil)

	f.mocks.db.On("CollectionExists", mock.Anything, mock.Anything).Return(true, nil)
	f.mocks.db.On("GraphExists", mock.Anything, mock.Anything).Return(true, nil)
	f.mocks.db.On("Graph", mock.Anything, schema.PERMISSIONS_GRAPH.Name).Return(f.mocks.permGraph, nil)
	f.mocks.db.On("Graph", mock.Anything, schema.BILLING_GRAPH.Name).Return(f.mocks.bilGraph, nil)
	f.mocks.db.On("Graph", mock.Anything, schema.CREDENTIALS_GRAPH.Name).Return(f.mocks.credGraph, nil)

	f.mocks.bilGraph.On("EdgeCollection", mock.Anything, schema.CUR2CUR).Return(f.mocks.cur2cur, driver.VertexConstraints{}, nil)
	graphs := []*driver_mocks.MockGraph{f.mocks.bilGraph, f.mocks.permGraph, f.mocks.credGraph}
	for _, g := range graphs {
		g.On("EdgeCollection", mock.Anything, mock.Anything).Return(driver_mocks.NewMockCollection(t), driver.VertexConstraints{}, nil)
		g.On("SetVertexConstraints", mock.Anything, mock.Anything, mock.Anything).Return(nil)
		g.On("EdgeCollectionExists", mock.Anything, mock.Anything).Return(true, nil)
	}
	cols := []*driver_mocks.MockCollection{f.mocks.nsCol, f.mocks.trCol,
		f.mocks.recCol, f.mocks.bilPlansCol, f.mocks.curCol,
		f.mocks.cur2cur, f.mocks.accCol, f.mocks.credCol, f.mocks.invCol, f.mocks.descrCol}
	for _, col := range cols {
		col.On("Database").Return(f.mocks.db).Maybe()
		col.On("Name").Return("not_name").Maybe()
	}
	f.mocks.db.On("Collection", mock.Anything, schema.NAMESPACES_COL).Return(f.mocks.nsCol, nil).Maybe()
	f.mocks.db.On("Collection", mock.Anything, schema.BILLING_PLANS_COL).Return(f.mocks.bilPlansCol, nil).Maybe()
	f.mocks.db.On("Collection", mock.Anything, schema.TRANSACTIONS_COL).Return(f.mocks.trCol, nil).Maybe()
	f.mocks.db.On("Collection", mock.Anything, schema.RECORDS_COL).Return(f.mocks.recCol, nil).Maybe()
	f.mocks.db.On("Collection", mock.Anything, schema.CUR_COL).Return(f.mocks.curCol, nil).Maybe()
	f.mocks.db.On("Collection", mock.Anything, schema.ACCOUNTS_COL).Return(f.mocks.accCol, nil).Maybe()
	f.mocks.db.On("Collection", mock.Anything, schema.CREDENTIALS_COL).Return(f.mocks.credCol, nil).Maybe()
	f.mocks.db.On("Collection", mock.Anything, schema.INVOICES_COL).Return(f.mocks.invCol, nil).Maybe()
	f.mocks.db.On("Collection", mock.Anything, schema.DESCRIPTIONS_COL).Return(f.mocks.descrCol, nil).Maybe()
	f.mocks.db.On("Collection", mock.Anything, schema.SERVICES_COL).Return(f.mocks.srvCol, nil).Maybe()
	f.mocks.db.On("Collection", mock.Anything, schema.INSTANCES_GROUPS_COL).Return(f.mocks.igCol, nil).Maybe()
	f.mocks.db.On("Collection", mock.Anything, schema.INSTANCES_COL).Return(f.mocks.instCol, nil).Maybe()
	f.mocks.db.On("Collection", mock.Anything, schema.ADDONS_COL).Return(f.mocks.addCol, nil).Maybe()
	f.mocks.db.On("Collection", mock.Anything, schema.SERVICES_PROVIDERS_COL).Return(f.mocks.spCol, nil).Maybe()
	//

	// specify how many times each controller was created. TODO: refactor controller constructors
	_currencyCalls := 3
	_instancesCalls := 2

	// Declare exchange in instances collection for ansible hooks
	f.mocks.rbmqch.EXPECT().ExchangeDeclare("hooks",
		"topic",
		true,
		false,
		false,
		false,
		amqp.Table(nil)).Return(nil).Times(1 * _instancesCalls)

	// Declare exchange in instances collection for instances creating for invoices
	f.mocks.rbmqch.EXPECT().ExchangeDeclare("instances",
		"topic",
		true,
		false,
		false,
		false,
		amqp.Table(nil)).Return(nil).Times(1 * _instancesCalls)

	// Ensuring hash index on currency title
	f.mocks.curCol.EXPECT().EnsureHashIndex(mock.Anything, []string{"title"}, &driver.EnsureHashIndexOptions{Unique: true}).
		Return(nil, true, nil).Times(1 * _currencyCalls)

	// Ensuring persistent index on services collection
	f.mocks.srvCol.EXPECT().EnsurePersistentIndex(mock.Anything, []string{"status"}, &driver.EnsurePersistentIndexOptions{
		Unique: false, Sparse: true, InBackground: true, Name: "service-status",
	}).Return(nil, true, nil).Once()

	// Ensuring persistent index on instances groups collection
	f.mocks.igCol.EXPECT().EnsurePersistentIndex(mock.Anything, []string{"type"}, &driver.EnsurePersistentIndexOptions{
		Unique: false, Sparse: true, InBackground: true, Name: "sp-type",
	}).Return(nil, true, nil).Once()

	// Creating default currencies 8 times(8 currencies)
	f.mocks.curCol.On("DocumentExists", mock.Anything, mock.Anything).Return(true, nil).Times(8 * _currencyCalls)

	// Updating old currencies to new. Cursor must not be used. Accounts, Invoices, Transactions, Records controllers
	// Should be called but don't check exact number of calls due to bad code in controllers
	f.mocks.db.On("Query", mock.Anything, migrations.NumericToObjectCurrency, mock.Anything).Return(driver.Cursor(nil), nil)
	f.mocks.db.On("Query", mock.Anything, migrations.ObjectToObjectCurrency, mock.Anything).Return(driver.Cursor(nil), nil)

	// Migrating currencies to dynamic. Cursor must not be used
	f.mocks.db.On("Query", mock.Anything, graph.MigrateToDynamicEdges, mock.Anything).Return(driver.Cursor(nil), nil).Times(1 * _currencyCalls)
	f.mocks.db.On("Query", mock.Anything, graph.MigrateToDynamicVertex, mock.Anything).Return(driver.Cursor(nil), nil).Times(1 * _currencyCalls)

	// Updating total and cost fields for records. Cursor must not be used
	f.mocks.db.On("Query", mock.Anything, migrations.EnsureCostAndTotal, mock.Anything).Return(driver.Cursor(nil), nil).Once()

	// migrate() mocks from BillingServiceServer
	cur := driver_mocks.NewMockCursor(t)
	f.mocks.db.On("Query", mock.Anything, "FOR plan IN BillingPlans RETURN plan", mock.Anything).Return(cur, nil).Once()
	cur.On("HasMore").Return(false).Once()
	cur.On("Close").Return(nil).Once()

	// Create empty csv file which have to be passed in invoices migration function
	filename := "./whmcs_invoices.csv"
	_, err := os.Create(filename)
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = os.Remove(filename)
	})

	// data
	// -----------------------------------------
	// -----------------------------------------
	// -----------------------------------------
	f.data.rootCtx = context.WithValue(context.Background(), nocloud.NoCloudAccount, schema.ROOT_ACCOUNT_KEY)
	f.data.ctx = context.WithValue(context.Background(), nocloud.NoCloudAccount, "account_test")

	f.srv = billing.NewBillingServiceServer(zap.NewExample(), f.mocks.db, f.mocks.rbmqconn, f.mocks.redisclient)
	return f
}

// Run this test to make sure all controller initialization actions runs as expected
// If this test fails, you must fix it FIRSTLY
func TestBillingServer_InitializationOfEverything(t *testing.T) {
	_ = newBillingServiceServerFixture(t)
}

//func TestBillingServer_Invoices_GetInvoices(t *testing.T) {
//	f := newBillingServiceServerFixture(t)
//
//	cur := driver_mocks.NewMockCursor(t)
//	f.mocks.db.EXPECT().Query(mock.Anything, "FOR t IN @@invoices RETURN merge(t, {uuid: t._key})", mock.Anything).Return(cur, nil).Once()
//	cur.EXPECT().ReadDocument(mock.Anything, mock.Anything).Return(driver.DocumentMeta{Key: "12"}, nil).Once()
//	cur.EXPECT().ReadDocument(mock.Anything, mock.Anything).Return(driver.DocumentMeta{}, driver.NoMoreDocumentsError{}).Once()
//	cur.EXPECT().Close().Return(nil).Once()
//
//	res, err := f.srv.GetInvoices(f.data.rootCtx, connect.NewRequest(&pb.GetInvoicesRequest{}))
//	assert.NoError(t, err)
//	assert.NotNil(t, res)
//	assert.Equal(t, 1, len(res.Msg.GetPool()))
//	assert.Equal(t, "12", res.Msg.GetPool()[0].Uuid)
//	assert.Equal(t, 1, 1)
//}
