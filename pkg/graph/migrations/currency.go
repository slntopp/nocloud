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
	{Id: schema.DEFAULT_CURRENCY_ID, Name: schema.DEFAULT_CURRENCY_NAME},
	{Id: 1, Name: "USD"},
	{Id: 2, Name: "EUR"},
	{Id: 3, Name: "BYN"},
	{Id: 4, Name: "PLN"},
}

const numericToObjectCurrency = `
	FOR el IN @@collection
		FILTER el.currency != null
		FILTER IS_NUMBER(el.currency)
		UPDATE el WITH { currency: { id: el.currency, name: @names[TO_STRING(el.currency)] } } IN @@collection
`

func UpdateNumericCurrencyToDynamic(log *zap.Logger, col driver.Collection) {
	namesMap := map[string]string{}
	for _, val := range LEGACY_CURRENCIES {
		namesMap[fmt.Sprintf("%d", val.GetId())] = val.GetName()
	}
	_, err := col.Database().Query(context.TODO(), numericToObjectCurrency, map[string]interface{}{
		"@collection": col.Name(),
		"names":       namesMap,
	})
	if err != nil {
		log.Fatal("Error migrating currency", zap.Error(err), zap.String("collection", col.Name()))
	}
	log.Info("Migrated currency", zap.String("collection", col.Name()))
}
