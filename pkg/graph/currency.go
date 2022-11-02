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
	pb.Currency_NCU,
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
	edge := map[string]interface{}{}
	_, err := c.edges.ReadDocument(ctx, fmt.Sprintf("%d-%d", from, to), &edge)
	if err != nil {
		return 0, err
	}
	rate := edge["rate"].(float64)
	return rate, err
}

func (c *CurrencyController) CreateExchangeRate(ctx context.Context, from pb.Currency, to pb.Currency, rate float64) error {
	edge := map[string]interface{}{
		"_key":  fmt.Sprintf("%d-%d", from, to),
		"_from": fmt.Sprintf("%s/%d", schema.CUR_COL, from),
		"_to":   fmt.Sprintf("%s/%d", schema.CUR_COL, to),
		"rate":  rate,
	}
	_, err := c.edges.CreateDocument(ctx, &edge)

	return err
}

func (c *CurrencyController) UpdateExchangeRate(ctx context.Context, from pb.Currency, to pb.Currency, rate float64) error {
	key := fmt.Sprintf("%d-%d", from, to)

	edge := map[string]interface{}{
		"rate": rate,
	}
	_, err := c.edges.UpdateDocument(ctx, key, &edge)

	return err
}

func (c *CurrencyController) DeleteExchangeRate(ctx context.Context, from pb.Currency, to pb.Currency) error {
	key := fmt.Sprintf("%d-%d", from, to)

	_, err := c.edges.RemoveDocument(ctx, key)

	return err
}

func (c *CurrencyController) Convert(ctx context.Context, from pb.Currency, to pb.Currency, amount float64) (float64, error) {

	if from == to {
		return amount, nil
	}

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
	defer cursor.Close()

	for cursor.HasMore() {
		doc := &driver.DocumentMeta{}
		cursor.ReadDocument(ctx, doc)

		// specify bitsize to parse int32 exactly
		id, err := strconv.ParseInt(doc.Key, 10, 32)
		if err != nil {
			return currencies, err
		}

		currencies = append(currencies, pb.Currency(int32(id)))
	}

	return currencies, nil
}
