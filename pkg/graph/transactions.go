/*
Copyright © 2021-2023 Nikita Ivanovski info@slnt-opp.xyz

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
	"errors"
	"fmt"
	"github.com/arangodb/go-driver"
	pb "github.com/slntopp/nocloud-proto/billing"
	"github.com/slntopp/nocloud/pkg/graph/migrations"
	"github.com/slntopp/nocloud/pkg/nocloud/schema"
	"go.uber.org/zap"
)

type TransactionsController interface {
	Create(ctx context.Context, tx *pb.Transaction) (*Transaction, error)
	Get(ctx context.Context, uuid string) (*pb.Transaction, error)
	Update(ctx context.Context, tx *pb.Transaction) (*pb.Transaction, error)
	Transfer(ctx context.Context, uuid string, account string) error
}

type Transaction struct {
	*pb.Transaction
	driver.DocumentMeta
}

type transactionsController struct {
	col driver.Collection // Billing Plans collection

	records RecordsController

	log *zap.Logger
}

func NewTransactionsController(logger *zap.Logger, db driver.Database) TransactionsController {
	ctx := context.TODO()
	log := logger.Named("TransactionsController")
	col := GetEnsureCollection(log, ctx, db, schema.TRANSACTIONS_COL)

	log.Info("Creating Transaction controller")

	records := NewRecordsController(log, db)

	migrations.UpdateNumericCurrencyToDynamic(log, col)

	return &transactionsController{
		log: log, col: col, records: records,
	}
}

func (ctrl *transactionsController) Create(ctx context.Context, tx *pb.Transaction) (*Transaction, error) {
	if tx.GetAccount() == "" {
		return nil, errors.New("account is required")
	}
	if tx.Total == 0 {
		return nil, errors.New("total is required")
	}
	meta, err := ctrl.col.CreateDocument(ctx, tx)
	if err != nil {
		ctrl.log.Error("Failed to create transaction", zap.Error(err))
		return nil, err
	}
	tx.Uuid = meta.Key
	return &Transaction{tx, meta}, nil
}

func (ctrl *transactionsController) Get(ctx context.Context, uuid string) (*pb.Transaction, error) {
	var tx pb.Transaction
	meta, err := ctrl.col.ReadDocument(ctx, uuid, &tx)
	if err != nil {
		ctrl.log.Error("Failed to read transaction", zap.Error(err))
		return nil, err
	}
	tx.Uuid = meta.Key
	return &tx, nil
}

func (ctrl *transactionsController) Update(ctx context.Context, tx *pb.Transaction) (*pb.Transaction, error) {
	_, err := ctrl.col.UpdateDocument(ctx, tx.GetUuid(), tx)
	if err != nil {
		ctrl.log.Error("Failed to update transaction", zap.Error(err))
		return nil, err
	}
	return tx, nil
}

func (ctrl *transactionsController) Transfer(ctx context.Context, uuid string, account string) error {
	tr, err := ctrl.Get(ctx, uuid)
	if err != nil {
		return fmt.Errorf("failed to get transaction: %w", err)
	}
	tr.Account = account
	if _, err = ctrl.Update(ctx, tr); err != nil {
		return fmt.Errorf("failed to update transaction: %w", err)
	}
	recs, err := ctrl.records.Get(ctx, uuid)
	if err != nil {
		return fmt.Errorf("failed to get records: %w", err)
	}
	for _, rec := range recs {
		rec.Account = account
		if err = ctrl.records.Update(ctx, rec); err != nil {
			return fmt.Errorf("failed to update record: %w", err)
		}
	}
	return err
}
