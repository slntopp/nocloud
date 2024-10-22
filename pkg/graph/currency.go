package graph

import (
	"context"
	"fmt"
	"strconv"

	"github.com/slntopp/nocloud/pkg/graph/migrations"

	"github.com/arangodb/go-driver"
	pb "github.com/slntopp/nocloud-proto/billing"
	"github.com/slntopp/nocloud/pkg/nocloud/schema"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Currency struct {
	Id     int32  `json:"id"`
	Title  string `json:"title"`
	Public bool   `json:"public"`
	driver.DocumentMeta
}

func CurrencyFromPb(currency *pb.Currency) Currency {
	key := fmt.Sprintf("%d", currency.GetId())
	return Currency{
		Id:     currency.GetId(),
		Title:  currency.GetTitle(),
		Public: currency.GetPublic(),
		DocumentMeta: driver.DocumentMeta{
			Key: key,
			ID:  driver.NewDocumentID(schema.CUR_COL, key),
		},
	}
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

	log.Info("Creating Currency controller")

	log.Info("Creating default currencies")
	// Ensure default currencies exists
	for _, currency := range migrations.LEGACY_CURRENCIES {
		key := fmt.Sprintf("%d", currency.GetId())
		exists, _ := col.DocumentExists(ctx, key)
		if !exists {
			col.CreateDocument(ctx, CurrencyFromPb(currency))
		}
	}
	log.Info("Default currencies ensured")

	ctrl := CurrencyController{
		log:   log,
		col:   col,
		graph: graph,
		edges: edges,
		db:    db,
	}

	log.Info("Migrating old currency template to dynamic")
	ctrl.migrateToDynamic()

	log.Info("Ensuring hash index on currency title")
	_, _, err := col.EnsureHashIndex(ctx, []string{"title"}, &driver.EnsureHashIndexOptions{Unique: true})
	if err != nil {
		panic(err)
	}

	log.Info("Currency controller was created")
	return ctrl
}

const migrateToDynamicVertex = `
	FOR el IN @@collection
		FILTER el.id == null || el.title == null || el.title == ""
		UPDATE el WITH { id: TO_NUMBER(el._key), title: @names[el._key], name: null, public: true } IN @@collection
        OPTIONS { keepNull: false }
`
const migrateToDynamicEdges = `
	FOR edge IN @@collection
		LET from_doc = DOCUMENT(@@cur_collection, edge._from)
        LET to_doc = DOCUMENT(@@cur_collection, edge._to)
        FILTER from_doc != null && to_doc != null
		UPDATE edge WITH { from: from_doc, to: to_doc } IN @@collection
`

func (c *CurrencyController) migrateToDynamic() {
	c.log.Info("Migrating currency to dynamic")
	namesMap := map[string]string{}
	for _, val := range migrations.LEGACY_CURRENCIES {
		namesMap[fmt.Sprintf("%d", val.GetId())] = val.GetTitle()
	}
	_, err := c.col.Database().Query(context.TODO(), migrateToDynamicVertex, map[string]interface{}{
		"@collection": c.col.Name(),
		"names":       namesMap,
	})
	if err != nil {
		c.log.Fatal("Error migrating vertex currency", zap.Error(err), zap.String("collection", c.col.Name()))
	}
	_, err = c.col.Database().Query(context.TODO(), migrateToDynamicEdges, map[string]interface{}{
		"@collection":     schema.CUR2CUR,
		"@cur_collection": c.col.Name(),
	})
	if err != nil {
		c.log.Fatal("Error migrating edges currency", zap.Error(err), zap.String("collection", c.col.Name()))
	}
	c.log.Info("Migrated currency successfully", zap.String("collection", c.col.Name()))
}

func (c *CurrencyController) CreateCurrency(ctx context.Context, currency *pb.Currency) error {
	_, err := c.col.CreateDocument(ctx, CurrencyFromPb(currency))
	return err
}

const getCurrenciesQuery = `
FOR CURRENCY in @@currencies
	return CURRENCY
`

func (c *CurrencyController) GetExchangeRateDirect(ctx context.Context, from pb.Currency, to pb.Currency) (float64, float64, error) {
	edge := map[string]interface{}{}
	_, err := c.edges.ReadDocument(ctx, fmt.Sprintf("%d-%d", from.GetId(), to.GetId()), &edge)
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
    RETURN {rate: edge.rate, commission: TO_NUMBER(edge.commission)}`

func (c *CurrencyController) GetExchangeRate(ctx context.Context, from *pb.Currency, to *pb.Currency) (float64, float64, error) {
	if from.Id == to.Id {
		return 1, 0, nil
	}

	cursor, err := c.db.Query(ctx, getExchangeRateQuery, map[string]interface{}{
		"from":    fmt.Sprintf("%s/%d", schema.CUR_COL, from.GetId()),
		"to":      fmt.Sprintf("%s/%d", schema.CUR_COL, to.GetId()),
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
		"_key":       fmt.Sprintf("%d-%d", from.GetId(), to.GetId()),
		"_from":      fmt.Sprintf("%s/%d", schema.CUR_COL, from.GetId()),
		"_to":        fmt.Sprintf("%s/%d", schema.CUR_COL, to.GetId()),
		"from":       CurrencyFromPb(&from),
		"to":         CurrencyFromPb(&to),
		"rate":       rate,
		"commission": commission,
	}
	_, err := c.edges.CreateDocument(ctx, &edge)

	return err
}

func (c *CurrencyController) UpdateExchangeRate(ctx context.Context, from pb.Currency, to pb.Currency, rate, commission float64) error {
	key := fmt.Sprintf("%d-%d", from.GetId(), to.GetId())

	edge := map[string]interface{}{
		"rate":       rate,
		"commission": commission,
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

	rate, _, err := c.GetExchangeRate(ctx, from, to)
	if err != nil {
		return 0, status.Error(codes.NotFound, err.Error())
	}

	return amount * rate, nil
}

func (c *CurrencyController) GetCurrencies(ctx context.Context, isAdmin bool) ([]*pb.Currency, error) {
	currencies := []*pb.Currency{}

	cursor, err := c.db.Query(ctx, getCurrenciesQuery, map[string]interface{}{
		"@currencies": schema.CUR_COL,
	})
	if err != nil {
		return nil, err
	}
	defer cursor.Close()

	for cursor.HasMore() {
		doc := struct {
			driver.DocumentMeta
			Title  string `json:"title"`
			Public bool   `json:"public"`
		}{}
		_, err := cursor.ReadDocument(ctx, &doc)
		if err != nil {
			return currencies, err
		}

		// specify bitsize to parse int32 exactly
		id, err := strconv.ParseInt(doc.Key, 10, 32)
		if err != nil {
			return currencies, err
		}

		// Ignore private currency
		if !isAdmin && !doc.Public {
			continue
		}

		currencies = append(currencies, &pb.Currency{
			Id:     int32(id),
			Title:  doc.Title,
			Public: doc.Public,
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
		resp := struct {
			driver.DocumentMeta
			From       Currency `json:"from"`
			To         Currency `json:"to"`
			Rate       float64  `json:"rate"`
			Commission float64  `json:"commission"`
		}{}
		_, err := cursor.ReadDocument(ctx, &resp)
		if err != nil {
			return nil, err
		}

		idFrom := resp.From.Id
		idTo := resp.To.Id
		rates = append(rates, &pb.GetExchangeRateResponse{
			From: &pb.Currency{
				Id:    idFrom,
				Title: resp.From.Title,
			},
			To: &pb.Currency{
				Id:    idTo,
				Title: resp.To.Title,
			},
			Rate:       resp.Rate,
			Commission: resp.Commission,
		})
	}

	return rates, nil
}
