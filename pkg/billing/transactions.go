/*
Copyright © 2021-2022 Nikita Ivanovski info@slnt-opp.xyz

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package billing

import (
	"context"
	"time"

	"github.com/arangodb/go-driver"
	pb "github.com/slntopp/nocloud/pkg/billing/proto"
	"github.com/slntopp/nocloud/pkg/nocloud/schema"
	"go.uber.org/zap"
)

type Transaction struct {
	*pb.Transaction
	driver.DocumentMeta
}

type TransactionsController struct {
	log *zap.Logger
	col driver.Collection // Transactions Collection
	
	db driver.Database
}

func NewTransactionsController(log *zap.Logger, db driver.Database) TransactionsController {
	col, _ := db.Collection(context.TODO(), schema.TRANSACTIONS_COL)
	return TransactionsController{
		log: log.Named("TransactionsController"),
		col: col, db: db,
	}
}

func (t *Transaction) MakeKey() {

}

const createCheckQuery = `
LET new = @t

FOR t IN @@transactions
FILTER t.instance == new.instance && t.resource == new.resource
FILTER ASSERT(((new.start < t.start && new.end < t.start) || (new.start > t.end && new.end > t.end)), "Transactions are overlapping each other")
    RETURN t
`

func (ctrl *TransactionsController) Create(ctx context.Context, t *Transaction) (*Transaction, error) {
	c, err := ctrl.db.Query(ctx, createCheckQuery, map[string]interface{}{
		"t": t.Transaction,
		"@transactions": schema.TRANSACTIONS_COL,
	})
	if err != nil {
		return nil, err
	}
	c.Close()

	t.Transaction.Processed = false

	meta, err := ctrl.col.CreateDocument(ctx, t.Transaction)
	if err != nil {
		return nil, err
	}
	t.Uuid = meta.ID.Key()

	return t, nil
}

func (ctrl *TransactionsController) Get(ctx context.Context, t *Transaction) (*Transaction, error) {
	var res pb.Transaction
	ctrl.log.Info("get", zap.String("uuid", t.Uuid))
	meta, err := ctrl.col.ReadDocument(ctx, t.Uuid, &res)
	if err != nil {
		return nil, err
	}
	res.Uuid = meta.Key
	return &Transaction{Transaction: &res}, nil
}

const processTransactionsQuery = `
LET now = @now
FOR t IN @@transactions
FILTER !t.processed && t.exec <= now
    UPDATE t._key WITH { processed: true } IN @@transactions RETURN NEW
`
func (ctrl *TransactionsController) ProcessorRoutine(p chan *pb.Transaction) {
	log := ctrl.log.Named("TransactionsProcessorRoutine")
	for range time.NewTicker(time.Second * 30).C {
		go func() {
			c, err := ctrl.db.Query(context.TODO(), processTransactionsQuery, map[string]interface{}{
				"now": time.Now().Unix(),
				"@transactions": schema.TRANSACTIONS_COL,
			})
			if err != nil {
				log.Error("Error executing query", zap.Error(err))
			}
			defer c.Close()
			
			for {
				var t pb.Transaction
				_, err := c.ReadDocument(context.TODO(), &t)
				if driver.IsNoMoreDocuments(err) {
					break
				} else if err != nil {
					log.Error("Error reading processed transaction", zap.Error(err))
					continue
				}
				p <- &t
			}
		}()
	}
}