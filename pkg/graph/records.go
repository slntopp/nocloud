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
	"fmt"
	"github.com/arangodb/go-driver"
	pb "github.com/slntopp/nocloud-proto/billing"
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

func (ctrl *RecordsController) Create(ctx context.Context, r *pb.Record) driver.DocumentID {
	if r.Priority != pb.Priority_ADDITIONAL {
		ok := ctrl.CheckOverlapping(ctx, r)
		ctrl.log.Debug("Pre-flight checks", zap.Bool("overlapping", ok))
		if !ok {
			ctrl.log.Warn("Skipping creating transactions: overlapping", zap.Any("record", r))
			return ""
		}
	}

	meta, err := ctrl.col.CreateDocument(ctx, r)
	if err != nil {
		ctrl.log.Error("failed to create record", zap.Error(err))
	}
	return meta.ID
}

const getRecordsQuery = `
LET T = DOCUMENT(@transaction)
LET recs = T.records ? T.records : []
FOR rec IN recs
  RETURN DOCUMENT(CONCAT(@records, "/", rec))
  
`

func (ctrl *RecordsController) Get(ctx context.Context, tr string) (res []*pb.Record, err error) {
	c, err := ctrl.db.Query(ctx, getRecordsQuery, map[string]interface{}{
		"transaction": driver.NewDocumentID(schema.TRANSACTIONS_COL, tr).String(),
		"records":     schema.RECORDS_COL,
	})
	if err != nil {
		ctrl.log.Error("failed to get records", zap.Error(err))
		return nil, err
	}
	defer c.Close()

	for {
		var r pb.Record
		m, err := c.ReadDocument(ctx, &r)
		if driver.IsNoMoreDocuments(err) {
			break
		} else if err != nil {
			return nil, err
		}
		r.Uuid = m.ID.Key()
		res = append(res, &r)
	}

	return res, nil
}

const getReportsQuery = `
FOR i in @@instances
	LET records = (
		FOR record in @@records 
			%s
			FILTER record.processed
			FILTER record.instance == i._key
			RETURN record
	)
    RETURN {uuid: i._key, total: SUM(records[*].total), currency: FIRST(records).currency ? FIRST(records).currency : 0}
`

func (ctrl *RecordsController) GetReports(ctx context.Context, req *pb.GetInstancesReportRequest) ([]*pb.InstanceReport, error) {
	query := getReportsQuery
	params := map[string]interface{}{
		"@records":   schema.RECORDS_COL,
		"@instances": schema.INSTANCES_COL,
	}

	if req.From != nil && req.To != nil {
		query = fmt.Sprintf(query, "FILTER record.exec >= @from AND record.exec <=@to")
		params["from"] = req.GetFrom()
		params["to"] = req.GetTo()
	} else {
		query = fmt.Sprintf(query, "")
	}

	cursor, err := ctrl.db.Query(ctx, query, params)
	if err != nil {
		return nil, err
	}

	var res []*pb.InstanceReport

	for cursor.HasMore() {
		var rep = pb.InstanceReport{}
		_, err := cursor.ReadDocument(ctx, &rep)
		if err != nil {
			return nil, err
		}
		res = append(res, &rep)
	}

	return res, nil
}

func (ctrl *RecordsController) GetReport(ctx context.Context, req *pb.GetDetailedInstanceReportRequest) (*pb.GetDetailedInstanceReportResponse, error) {
	query := "LET records = ( FOR record in @@records FILTER record.processed FILTER record.instance == @instance"
	params := map[string]interface{}{
		"@records": schema.RECORDS_COL,
		"instance": req.GetUuid(),
	}

	if req.From != nil && req.To != nil {
		query += "FILTER record.exec >= @from AND record.exec <=@to"
		params["from"] = req.GetFrom()
		params["to"] = req.GetTo()
	}

	if req.Field != nil && req.Sort != nil {
		subQuery := ` SORT record.%s %s`
		field, sort := req.GetField(), req.GetSort()

		if field == "total" {
			if sort == "asc" {
				sort = "desc"
			} else {
				sort = "asc"
			}
		}

		query += fmt.Sprintf(subQuery, field, sort)
	}

	if req.Page != nil && req.Limit != nil {
		if req.GetLimit() != 0 {
			limit, page := req.GetLimit(), req.GetPage()
			offset := (page - 1) * limit

			query += ` LIMIT @offset, @count`
			params["offset"] = offset
			params["count"] = limit
		}
	}

	query += " RETURN merge(record, {uuid: record._key})) RETURN {records: records, total: SUM(records[*].total), count: COUNT(records)}"

	ctrl.log.Debug("Final query", zap.String("query", query))

	cursor, err := ctrl.db.Query(ctx, query, params)
	if err != nil {
		return nil, err
	}

	var res pb.GetDetailedInstanceReportResponse

	_, err = cursor.ReadDocument(ctx, &res)
	if err != nil {
		return nil, err
	}

	return &res, nil
}

const reportsCountQuery = `RETURN LENGTH(@@instances)`

func (ctrl *RecordsController) GetReportsCount(ctx context.Context) (int64, error) {
	cur, err := ctrl.db.Query(ctx, reportsCountQuery, map[string]interface{}{
		"@instances": schema.INSTANCES_COL,
	})
	if err != nil {
		return 0, err
	}

	var result int64

	_, err = cur.ReadDocument(ctx, &result)
	if err != nil {
		return 0, err
	}
	return result, nil
}

const reportsInstanceCountQuery = `
LET records = ( 
	FOR record in @@records 
		%s
		FILTER record.processed 
		FILTER record.instance == @instance
		RETURN record
	) 

RETURN LENGTH(records)
`

func (ctrl *RecordsController) GetInstanceReportCountReport(ctx context.Context, req *pb.GetInstanceReportCountRequest) (int64, error) {
	query := reportsInstanceCountQuery
	params := map[string]interface{}{
		"@records": schema.RECORDS_COL,
		"instance": req.GetUuid(),
	}

	if req.From != nil && req.To != nil {
		query = fmt.Sprintf(query, "FILTER record.exec >= @from AND record.exec <=@to")
		params["from"] = req.GetFrom()
		params["to"] = req.GetTo()
	} else {
		query = fmt.Sprintf(query, "")
	}

	cursor, err := ctrl.db.Query(ctx, query, params)
	if err != nil {
		return 0, err
	}

	var result int64

	_, err = cursor.ReadDocument(ctx, &result)
	if err != nil {
		return 0, err
	}

	return result, nil
}
