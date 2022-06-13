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

type Record struct {
	*pb.Record
	driver.DocumentMeta
}

type RecordsController struct {
	col driver.Collection // Billing Plans collection
	db  driver.Database
	log *zap.Logger
}

func NewRecordsController(logger *zap.Logger, db driver.Database) RecordsController {
	ctx := context.TODO()
	log := logger.Named("RecordsController")
	col := GetEnsureCollection(log, ctx, db, schema.RECORDS_COL)
	return RecordsController{
		log: log, col: col, db: db,
	}
}

const checkOverlappingQuery = `
let n = @record
RETURN COUNT(FOR r IN @@records
FILTER r.instance == n.instance
FILTER r.resource == n.resource
FILTER (n.start < r.end && n.start >= r.start) || (n.start <= r.start && r.start > n.end)
    RETURN r) == 0
`

func (ctrl *RecordsController) CheckOverlapping(ctx context.Context, r *pb.Record) (ok bool) {
	c, err := ctrl.db.Query(ctx, checkOverlappingQuery, map[string]interface{}{
		"record":   r,
		"@records": schema.RECORDS_COL,
	})
	if err != nil {
		ctrl.log.Error("failed to check overlapping", zap.Error(err))
		return false
	}
	defer c.Close()

	_, err = c.ReadDocument(ctx, &ok)
	if err != nil {
		ctrl.log.Error("failed to read document", zap.Error(err))
		return false
	}
	return ok
}

func (ctrl *RecordsController) Create(ctx context.Context, r *pb.Record) {
	ok := ctrl.CheckOverlapping(ctx, r)
	ctrl.log.Debug("Pre-flight checks", zap.Bool("overlapping", ok))
	if !ok {
		ctrl.log.Warn("Skipping creating transactions: overlapping", zap.Any("record", r))
		return
	}

	_, err := ctrl.col.CreateDocument(ctx, r)
	if err != nil {
		ctrl.log.Error("failed to create record", zap.Error(err))
	}
}

const getRecordsQuery = `
LET T = DOCUMENT(@transaction)
FOR rec IN T.records
  RETURN DOCUMENT(CONCAT(@records, "/", rec))
`

func (ctrl *RecordsController) Get(ctx context.Context, tr string) (res []*pb.Record, err error) {
	c, err := ctrl.db.Query(ctx, getRecordsQuery, map[string]interface{}{
		"transaction": driver.NewDocumentID(schema.TRANSACTIONS_COL, tr),
		"records":     schema.RECORDS_COL,
	})
	if err != nil {
		ctrl.log.Error("failed to get records", zap.Error(err))
		return nil, err
	}
	defer c.Close()

	for {
		var r pb.Record
		_, err = c.ReadDocument(ctx, &r)
		if driver.IsNoMoreDocuments(err) {
			break
		} else if err != nil {
			return nil, err
		}
		res = append(res, &r)
	}

	return res, nil
}
