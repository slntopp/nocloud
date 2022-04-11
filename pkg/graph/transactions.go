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
package graph

import (
	"context"

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
	col driver.Collection // Billing Plans collection

	log *zap.Logger
}

func NewTransactionsController(log *zap.Logger, db driver.Database) TransactionsController {
	col, _ := db.Collection(context.TODO(), schema.TRANSACTIONS_COL)
	return TransactionsController{
		log: log, col: col,
	}
}