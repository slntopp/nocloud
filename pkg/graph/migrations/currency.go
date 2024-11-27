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
	{Id: 1, Title: "USD"},
	{Id: 2, Title: "EUR"},
	{Id: 3, Title: "BYN"},
	{Id: 4, Title: "PLN"},
	{Id: 5, Title: "EUR_COLO"},
	{Id: 6, Title: "EUR_TECH"},
	{Id: 7, Title: "RUB"},
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
