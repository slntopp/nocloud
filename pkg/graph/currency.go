package graph

import (
	"context"
	"fmt"
	"github.com/slntopp/nocloud/pkg/graph/migrations"
	"strconv"

	"github.com/arangodb/go-driver"
	pb "github.com/slntopp/nocloud-proto/billing"
	"github.com/slntopp/nocloud/pkg/nocloud/schema"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Currency struct {
	*pb.Currency
	driver.DocumentMeta
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

	_, _, err := col.EnsureHashIndex(ctx, []string{"name"}, &driver.EnsureHashIndexOptions{Unique: true})
	if err != nil {
		panic(err)
	}

	// Ensure default currencies exists
	for _, currency := range migrations.LEGACY_CURRENCIES {
		key := fmt.Sprintf("%d", currency.GetId())
		exists, _ := col.DocumentExists(ctx, key)
		if !exists {
			col.CreateDocument(ctx, Currency{Currency: currency, DocumentMeta: driver.DocumentMeta{Key: key}})
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

func (c *CurrencyController) CreateCurrency(ctx context.Context, currency *pb.Currency) error {
	key := fmt.Sprintf("%d", currency.GetId())
	_, err := c.col.CreateDocument(ctx, Currency{
		Currency: currency,
		DocumentMeta: driver.DocumentMeta{
			Key: key,
		},
	})
	return err
}

const getCurrenciesQuery = `
FOR CURRENCY in @@currencies
	return CURRENCY
`

func (c *CurrencyController) GetExchangeRateDirect(ctx context.Context, from *pb.Currency, to *pb.Currency) (float64, error) {
	edge := map[string]interface{}{}
	_, err := c.edges.ReadDocument(ctx, fmt.Sprintf("%d-%d", from.GetId(), to.GetId()), &edge)
	if err != nil {
		return 0, err
	}
	rate := edge["rate"].(float64)
	return rate, err
}

const getExchangeRateQuery = `
LET rates = (
 FOR vertex, edge IN OUTBOUND
 SHORTEST_PATH @from TO @to
 GRAPH @billing
 FILTER edge
    RETURN edge.rate
)

RETURN { len: LENGTH(rates), rate: PRODUCT(rates) }
`

func (c *CurrencyController) GetExchangeRate(ctx context.Context, from *pb.Currency, to *pb.Currency) (float64, error) {
	if from.GetId() == to.GetId() {
		return 1, nil
	}

	cursor, err := c.db.Query(ctx, getExchangeRateQuery, map[string]interface{}{
		"from":    fmt.Sprintf("%s/%d", schema.CUR_COL, from.GetId()),
		"to":      fmt.Sprintf("%s/%d", schema.CUR_COL, to.GetId()),
		"billing": schema.BILLING_GRAPH.Name,
	})
	if err != nil {
		return 0, err
	}
	defer cursor.Close()

	var res struct {
		Len  int     `json:"len"`
		Rate float64 `json:"rate"`
	}

	_, err = cursor.ReadDocument(ctx, &res)
	if err != nil {
		return 0, err
	}

	if res.Len == 0 {
		return 0, fmt.Errorf("no path or direct connection between %s and %s", from.GetName(), to.GetName())
	}

	return res.Rate, nil
}

func (c *CurrencyController) CreateExchangeRate(ctx context.Context, from *pb.Currency, to *pb.Currency, rate float64) error {
	edge := map[string]interface{}{
		"_key":  fmt.Sprintf("%d-%d", from.GetId(), to.GetId()),
		"_from": fmt.Sprintf("%s/%d", schema.CUR_COL, from.GetId()),
		"_to":   fmt.Sprintf("%s/%d", schema.CUR_COL, to.GetId()),
		"from":  from,
		"to":    to,
		"rate":  rate,
	}
	_, err := c.edges.CreateDocument(ctx, &edge)

	return err
}

func (c *CurrencyController) UpdateExchangeRate(ctx context.Context, from *pb.Currency, to *pb.Currency, rate float64) error {
	key := fmt.Sprintf("%d-%d", from.GetId(), to.GetId())

	edge := map[string]interface{}{
		"rate": rate,
	}
	_, err := c.edges.UpdateDocument(ctx, key, &edge)

	return err
}

func (c *CurrencyController) DeleteExchangeRate(ctx context.Context, from *pb.Currency, to *pb.Currency) error {
	key := fmt.Sprintf("%d-%d", from.GetId(), to.GetId())

	_, err := c.edges.RemoveDocument(ctx, key)

	return err
}

func (c *CurrencyController) Convert(ctx context.Context, from *pb.Currency, to *pb.Currency, amount float64) (float64, error) {

	if from.GetId() == to.GetId() {
		return amount, nil
	}

	rate, err := c.GetExchangeRate(ctx, from, to)
	if err != nil {
		return 0, status.Error(codes.NotFound, err.Error())
	}

	return amount * rate, nil
}

func (c *CurrencyController) GetCurrencies(ctx context.Context) ([]*pb.Currency, error) {
	currencies := []*pb.Currency{}

	cursor, err := c.db.Query(ctx, getCurrenciesQuery, map[string]interface{}{
		"@currencies": schema.CUR_COL,
	})
	if err != nil {
		return nil, err
	}
	defer cursor.Close()

	for cursor.HasMore() {
		doc := Currency{}
		_, err := cursor.ReadDocument(ctx, doc)
		if err != nil {
			return currencies, err
		}

		// specify bitsize to parse int32 exactly
		id, err := strconv.ParseInt(doc.Key, 10, 32)
		if err != nil {
			return currencies, err
		}

		currencies = append(currencies, &pb.Currency{
			Id:   id,
			Name: doc.Currency.Name,
		})
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
