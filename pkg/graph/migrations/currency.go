package migrations

import (
	"context"
	"fmt"
	"github.com/arangodb/go-driver"
	pb "github.com/slntopp/nocloud-proto/billing"
	"github.com/slntopp/nocloud/pkg/nocloud/schema"
	"go.uber.org/zap"
)

var LEGACY_CURRENCIES = []*pb.Currency{
	{Id: schema.DEFAULT_CURRENCY_ID, Title: schema.DEFAULT_CURRENCY_NAME},
	{Id: 1, Title: "USD", Code: "USD", Public: true, Precision: 2},
	{Id: 2, Title: "EUR", Code: "EUR", Public: true, Precision: 2},
	{Id: 3, Title: "BYN", Code: "BYN", Public: true, Precision: 2},
	{Id: 4, Title: "PLN", Code: "PLN", Public: true, Precision: 2},
	{Id: 5, Title: "EUR_COLO", Code: "EUR_COLO", Public: true, Precision: 2},
	{Id: 6, Title: "EUR_TECH", Code: "EUR_TECH", Public: true, Precision: 2},
	{Id: 7, Title: "RUB", Code: "RUB", Public: true, Precision: 0},
}

const NumericToObjectCurrency = `
      FOR el IN @@collection
		FILTER el.currency != null
		FILTER IS_NUMBER(el.currency)
        LET newCurr = { id: el.currency, title: @names[TO_STRING(el.currency)] }
		UPDATE el WITH { currency: newCurr } IN @@collection
`

const ObjectToObjectCurrency = `
     FOR el IN @@collection
		FILTER el.currency != null
		FILTER (IS_DOCUMENT(el.currency) && el.currency.id != null) && (el.currency.title == null || el.currency.name != null)
        LET newCurr = { id: el.currency.id, title: @names[TO_STRING(el.currency.id)] }
		UPDATE el WITH { currency: newCurr } IN @@collection
`

func UpdateNumericCurrencyToDynamic(log *zap.Logger, col driver.Collection) {
	colName := col.Name()
	log.Info("Migrating currency to dynamic for collection: " + colName)
	namesMap := map[string]string{}
	for _, val := range LEGACY_CURRENCIES {
		namesMap[fmt.Sprintf("%d", val.GetId())] = val.GetTitle()
	}
	_, err := col.Database().Query(context.TODO(), NumericToObjectCurrency, map[string]interface{}{
		"@collection": colName,
		"names":       namesMap,
	})
	if err != nil {
		log.Fatal("Error migrating currency: numericToObject", zap.Error(err), zap.String("collection", colName))
	}
	_, err = col.Database().Query(context.TODO(), ObjectToObjectCurrency, map[string]interface{}{
		"@collection": colName,
		"names":       namesMap,
	})
	if err != nil {
		log.Fatal("Error migrating currency: objectToObject", zap.Error(err), zap.String("collection", colName))
	}
	log.Info("Migrated currency", zap.String("collection", colName))
}
