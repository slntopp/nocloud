package graph

import (
	"context"
	"fmt"
	"strconv"

	"github.com/arangodb/go-driver"
	pb "github.com/slntopp/nocloud/pkg/billing/proto"
	"github.com/slntopp/nocloud/pkg/nocloud/schema"
	"go.uber.org/zap"
)

var currencies []pb.Currency = []pb.Currency{
	pb.Currency_USD,
	pb.Currency_EUR,
	pb.Currency_BYN,
	pb.Currency_PLN,
}

type CurrencyController struct {
	log   *zap.Logger
	col   driver.Collection
	edges driver.Collection
	graph driver.Graph
	db    driver.Database
}

func NewCurrencyController(log *zap.Logger, db driver.Database) *CurrencyController {
	ctx := context.TODO()
	graph := GraphGetEnsure(log, ctx, db, schema.BILLING_GRAPH.Name)
	col := GetEnsureCollection(log, ctx, db, schema.CUR_COL)
	edges := GraphGetEdgeEnsure(log, ctx, graph, schema.CUR2CUR, schema.CUR_COL, schema.CUR_COL)

	// Fill database with known currencies
	for _, currency := range currencies {
		key := fmt.Sprintf("%d", currency)
		exists, _ := col.DocumentExists(ctx, key)
		if !exists {
			col.CreateDocument(ctx, map[string]string{
				"_key": key,
			})
		}
	}

	return &CurrencyController{
		log:   log,
		col:   col,
		graph: graph,
		edges: edges,
		db:    db,
	}
}

const getCurrenciesQuery = `
FOR CURRENCY in @@currencies
	return CURRENCY
`

func (c *CurrencyController) GetExchangeRate(ctx context.Context, from pb.Currency, to pb.Currency) (float64, error) {
	edge := map[string]float64{
		"rate": 0.0,
	}
	_, err := c.edges.ReadDocument(ctx, fmt.Sprintf("%d2%d", from, to), &edge)
	if err != nil {
		return 0, err
	}

	return edge["rate"], nil
}

func (c *CurrencyController) Exchange(ctx context.Context, from pb.Currency, to pb.Currency, amount float64) (float64, error) {

	rate, err := c.GetExchangeRate(ctx, from, to)
	if err != nil {
		return 0, err
	}

	return amount * rate, nil
}

func (c *CurrencyController) Get(ctx context.Context) ([]pb.Currency, error) {
	currencies := []pb.Currency{}

	cursor, err := c.db.Query(ctx, getCurrenciesQuery, map[string]interface{}{
		"@currencies": schema.CUR_COL,
	})
	if err != nil {
		return nil, err
	}

	for cursor.HasMore() {
		doc := &driver.DocumentMeta{}
		cursor.ReadDocument(ctx, doc)

		id, err := strconv.Atoi(doc.Key)
		if err != nil {
			return currencies, err
		}

		currencies = append(currencies, pb.Currency(id))
	}

	return currencies, nil
}
