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
}

const numericToObjectCurrency = `
      FOR el IN @@collection
		FILTER el.currency != null
		FILTER IS_NUMBER(el.currency)
        LET newCurr = { id: el.currency, title: @names[TO_STRING(el.currency)] }
		UPDATE el WITH { currency: newCurr } IN @@collection
`

const objectToObjectCurrency = `
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
	_, err := col.Database().Query(context.TODO(), numericToObjectCurrency, map[string]interface{}{
		"@collection": colName,
		"names":       namesMap,
	})
	if err != nil {
		log.Fatal("Error migrating currency: numericToObject", zap.Error(err), zap.String("collection", colName))
	}
	_, err = col.Database().Query(context.TODO(), objectToObjectCurrency, map[string]interface{}{
		"@collection": colName,
		"names":       namesMap,
	})
	if err != nil {
		log.Fatal("Error migrating currency: objectToObject", zap.Error(err), zap.String("collection", colName))
	}
	log.Info("Migrated currency", zap.String("collection", colName))
}