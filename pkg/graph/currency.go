package graph

import (
	"context"
	"fmt"
	"strconv"

	"github.com/arangodb/go-driver"
	pb "github.com/slntopp/nocloud-proto/billing"
	"github.com/slntopp/nocloud/pkg/nocloud/schema"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

func NewCurrencyController(log *zap.Logger, db driver.Database) CurrencyController {
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

	return CurrencyController{
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

func (c *CurrencyController) GetExchangeRateDirect(ctx context.Context, from pb.Currency, to pb.Currency) (float64, float64, error) {
	edge := map[string]interface{}{}
	_, err := c.edges.ReadDocument(ctx, fmt.Sprintf("%d-%d", from, to), &edge)
	if err != nil {
		return 0, 0, err
	}
	rate := edge["rate"].(float64)
	commission, ok := edge["commission"].(float64)
	if !ok {
		commission = 0
	}
	return rate, commission, err
}

const getExchangeRateQuery = `
 FOR vertex, edge IN OUTBOUND
 SHORTEST_PATH @from TO @to
 GRAPH @billing
 FILTER edge
    RETURN {rate: edge.rate, commission: TO_NUMBER(edge.commission)}
)`

func (c *CurrencyController) GetExchangeRate(ctx context.Context, from pb.Currency, to pb.Currency) (float64, float64, error) {
	if from == to {
		return 1, 0, nil
	}

	cursor, err := c.db.Query(ctx, getExchangeRateQuery, map[string]interface{}{
		"from":    fmt.Sprintf("%s/%d", schema.CUR_COL, from),
		"to":      fmt.Sprintf("%s/%d", schema.CUR_COL, to),
		"billing": schema.BILLING_GRAPH.Name,
	})
	if err != nil {
		c.log.Error("1", zap.Error(err))
		return 0, 0, err
	}
	defer cursor.Close()

	type Rate struct {
		Rate       float64 `json:"rate"`
		Commission float64 `json:"commission"`
	}

	rates := []Rate{}
	for cursor.HasMore() {
		obj := &Rate{}
		_, err := cursor.ReadDocument(ctx, obj)
		if err != nil {
			c.log.Error("2", zap.Error(err))
			return 0, 0, err
		}

		rates = append(rates, *obj)
	}

	if len(rates) == 0 {
		return 0, 0, fmt.Errorf("no path or direct connection between %s and %s", from.String(), to.String())
	}

	totalCommission := 0.0
	totalRate := 1.0
	for _, rate := range rates {
		totalRate *= rate.Rate + rate.Rate*(rate.Commission/100.0)
		totalCommission += rate.Commission
	}

	return totalRate, totalCommission, nil
}

func (c *CurrencyController) CreateExchangeRate(ctx context.Context, from pb.Currency, to pb.Currency, rate, commission float64) error {
	edge := map[string]interface{}{
		"_key":       fmt.Sprintf("%d-%d", from, to),
		"_from":      fmt.Sprintf("%s/%d", schema.CUR_COL, from),
		"_to":        fmt.Sprintf("%s/%d", schema.CUR_COL, to),
		"from":       from,
		"to":         to,
		"rate":       rate,
		"commission": commission,
	}
	_, err := c.edges.CreateDocument(ctx, &edge)

	return err
}

func (c *CurrencyController) UpdateExchangeRate(ctx context.Context, from pb.Currency, to pb.Currency, rate, commission float64) error {
	key := fmt.Sprintf("%d-%d", from, to)

	edge := map[string]interface{}{
		"rate":       rate,
		"commission": commission,
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

	rate, _, err := c.GetExchangeRate(ctx, from, to)
	if err != nil {
		return 0, status.Error(codes.NotFound, err.Error())
	}

	return amount * rate, nil
}

func (c *CurrencyController) GetCurrencies(ctx context.Context) ([]pb.Currency, error) {
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

const getRatesQuery = `
FOR RATE in @@rates
	return RATE
`

func (c *CurrencyController) GetExchangeRates(ctx context.Context) ([]*pb.GetExchangeRateResponse, error) {
	rates := []*pb.GetExchangeRateResponse{}

	cursor, err := c.db.Query(ctx, getRatesQuery, map[string]interface{}{
		"@rates": schema.CUR2CUR,
	})
	if err != nil {
		return nil, err
	}
	defer cursor.Close()

	for cursor.HasMore() {
		obj := &struct {
			driver.DocumentMeta
			*pb.GetExchangeRateResponse
		}{}
		_, err := cursor.ReadDocument(ctx, obj)
		if err != nil {
			return nil, err
		}

		rates = append(rates, obj.GetExchangeRateResponse)
	}

	return rates, nil
}
