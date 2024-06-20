package migrations

import (
	"context"

	"github.com/arangodb/go-driver"
	"go.uber.org/zap"
)

var ensureCostAndTotal = `
FOR el IN @@collection
	FILTER !el.cost
	UPDATE el WITH { total: 1, cost: el.total } IN @@collection
`

func UpdateTotalAndCostFields(log *zap.Logger, col driver.Collection) {
	colName := col.Name()
	log.Info("Migrating records to ensure cost and total fields: " + colName)
	_, err := col.Database().Query(context.TODO(), ensureCostAndTotal, map[string]interface{}{
		"@collection": colName,
	})
	if err != nil {
		log.Fatal("Error migrating cost and total fields", zap.Error(err), zap.String("collection", colName))
	}
	log.Info("Migrated cost and total fields", zap.String("collection", colName))
}
