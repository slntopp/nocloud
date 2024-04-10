package billing_test

import (
	"connectrpc.com/connect"
	"context"
	"github.com/arangodb/go-driver"
	pb "github.com/slntopp/nocloud-proto/billing"
	driver_mocks "github.com/slntopp/nocloud/mocks/github.com/arangodb/go-driver"
	"github.com/slntopp/nocloud/pkg/billing"
	"github.com/slntopp/nocloud/pkg/nocloud"
	"github.com/slntopp/nocloud/pkg/nocloud/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
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

		permGraph *driver_mocks.MockGraph
		bilGraph  *driver_mocks.MockGraph
		credGraph *driver_mocks.MockGraph
	}

	data struct {
		rootCtx        context.Context
		ctx            context.Context
		invoicesGetReq *pb.GetInvoicesRequest
	}
}

func newBillingServiceServerFixture(t *testing.T) *billingServiceServerFixture {
	f := &billingServiceServerFixture{}

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
	}
	f.mocks.db.On("Collection", mock.Anything, schema.NAMESPACES_COL).Return(f.mocks.nsCol, nil)
	f.mocks.db.On("Collection", mock.Anything, schema.BILLING_PLANS_COL).Return(f.mocks.bilPlansCol, nil)
	f.mocks.db.On("Collection", mock.Anything, schema.TRANSACTIONS_COL).Return(f.mocks.trCol, nil)
	f.mocks.db.On("Collection", mock.Anything, schema.RECORDS_COL).Return(f.mocks.recCol, nil)
	f.mocks.db.On("Collection", mock.Anything, schema.CUR_COL).Return(f.mocks.curCol, nil)
	f.mocks.db.On("Collection", mock.Anything, schema.ACCOUNTS_COL).Return(f.mocks.accCol, nil)
	f.mocks.db.On("Collection", mock.Anything, schema.CREDENTIALS_COL).Return(f.mocks.credCol, nil)
	f.mocks.db.On("Collection", mock.Anything, schema.INVOICES_COL).Return(f.mocks.invCol, nil)
	f.mocks.db.On("Collection", mock.Anything, schema.DESCRIPTIONS_COL).Return(f.mocks.descrCol, nil)

	// locked for 5 currencies
	f.mocks.curCol.On("DocumentExists", mock.Anything, mock.Anything).Return(true, nil).Times(5)

	// migrate() mocks
	cur := driver_mocks.NewMockCursor(t)
	f.mocks.db.On("Query", mock.Anything, "FOR plan IN BillingPlans RETURN plan", mock.Anything).Return(cur, nil).Once()
	cur.On("HasMore").Return(false).Once()
	cur.On("Close").Return(nil).Once()

	// -----------------------------------------
	// data
	f.data.rootCtx = context.WithValue(context.Background(), nocloud.NoCloudAccount, "0")
	f.data.ctx = context.WithValue(context.Background(), nocloud.NoCloudAccount, "account_test")

	f.data.invoicesGetReq = &pb.GetInvoicesRequest{}

	f.srv = billing.NewBillingServiceServer(zap.NewExample(), f.mocks.db)
	return f
}

func TestInvoices_OkWithEmptyReq(t *testing.T) {
	f := newBillingServiceServerFixture(t)

	cur := driver_mocks.NewMockCursor(t)
	f.mocks.db.EXPECT().Query(mock.Anything, "FOR t IN @@invoices RETURN merge(t, {uuid: t._key})", mock.Anything).Return(cur, nil).Once()
	cur.EXPECT().ReadDocument(mock.Anything, mock.Anything).Return(driver.DocumentMeta{Key: "12"}, nil).Once()
	cur.EXPECT().ReadDocument(mock.Anything, mock.Anything).Return(driver.DocumentMeta{}, driver.NoMoreDocumentsError{}).Once()
	cur.EXPECT().Close().Return(nil).Once()

	res, err := f.srv.GetInvoices(f.data.rootCtx, connect.NewRequest(f.data.invoicesGetReq))
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, 1, len(res.Msg.GetPool()))
	assert.Equal(t, "12", res.Msg.GetPool()[0].Uuid)
}
