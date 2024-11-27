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
	"connectrpc.com/connect"
	"context"
	"fmt"
	"github.com/arangodb/go-driver"
	pb "github.com/slntopp/nocloud-proto/billing"
	"github.com/slntopp/nocloud/pkg/graph/migrations"
	"github.com/slntopp/nocloud/pkg/nocloud/schema"
	"go.uber.org/zap"
)

type RecordsController interface {
	CheckOverlapping(ctx context.Context, r *pb.Record) (ok bool)
	Create(ctx context.Context, r *pb.Record) driver.DocumentID
	Get(ctx context.Context, tr string) (res []*pb.Record, err error)
	GetInstancesReports(ctx context.Context, req *pb.GetInstancesReportRequest) ([]*pb.InstanceReport, error)
	GetRecordsReports(ctx context.Context, req *pb.GetRecordsReportsRequest) (*connect.Response[pb.GetRecordsReportsResponse], error)
	GetInstancesReportsCount(ctx context.Context) (int64, error)
	GetRecordsReportsCount(ctx context.Context, req *pb.GetRecordsReportsCountRequest) (int64, error)
	GetUnique(ctx context.Context) (map[string]interface{}, error)
	Update(ctx context.Context, rec *pb.Record) error
}

type Record struct {
	*pb.Record
	driver.DocumentMeta
}

type recordsController struct {
	col driver.Collection // Billing Plans collection
	db  driver.Database
	log *zap.Logger
}

func NewRecordsController(logger *zap.Logger, db driver.Database) RecordsController {
	ctx := context.TODO()
	log := logger.Named("RecordsController")
	col := GetEnsureCollection(log, ctx, db, schema.RECORDS_COL)

	log.Info("Creating Records controller")

	migrations.UpdateNumericCurrencyToDynamic(log, col)
	migrations.UpdateTotalAndCostFields(log, col)

	return &recordsController{
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

func (ctrl *recordsController) CheckOverlapping(ctx context.Context, r *pb.Record) (ok bool) {
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

func (ctrl *recordsController) Create(ctx context.Context, r *pb.Record) driver.DocumentID {
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

func (ctrl *recordsController) Get(ctx context.Context, tr string) (res []*pb.Record, err error) {
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

func (ctrl *recordsController) Update(ctx context.Context, rec *pb.Record) error {
	_, err := ctrl.col.UpdateDocument(ctx, rec.GetUuid(), rec)
	return err
}

func (ctrl *recordsController) GetInstancesReports(ctx context.Context, req *pb.GetInstancesReportRequest) ([]*pb.InstanceReport, error) {
	query := "LET reports = (FOR i in @@instances LET records = ( FOR record in @@records  FILTER record.processed FILTER record.instance == i._key"
	params := map[string]interface{}{
		"@records":   schema.RECORDS_COL,
		"@instances": schema.INSTANCES_COL,
	}

	if req.From != nil {
		query += " FILTER record.exec >= @from"
		params["from"] = req.GetFrom()
	}

	if req.To != nil {
		query += " FILTER record.exec <=@to"
		params["to"] = req.GetTo()
	}

	params["def_currency"] = &pb.Currency{Id: schema.DEFAULT_CURRENCY_ID, Title: schema.DEFAULT_CURRENCY_NAME}
	query += " RETURN record) RETURN {uuid: i._key, total: SUM(records[*].total), currency: FIRST(records).currency ? FIRST(records).currency : @def_currency}) FOR r in reports"

	if req.Field != nil && req.Sort != nil {
		subQuery := ` SORT r.%s %s`
		field, sort := req.GetField(), req.GetSort()

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

	query += " RETURN r"

	ctrl.log.Debug("Final query", zap.String("query", query))

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

func (ctrl *recordsController) GetRecordsReports(ctx context.Context, req *pb.GetRecordsReportsRequest) (*connect.Response[pb.GetRecordsReportsResponse], error) {
	query := "LET records = ( FOR record in @@records FILTER record.processed"
	params := map[string]interface{}{
		"@records":    schema.RECORDS_COL,
		"@currencies": schema.CUR_COL,
	}

	if req.Account != nil {
		query += ` FILTER record.account == @acc`
		params["acc"] = req.GetAccount()
	}

	if req.Service != nil {
		query += ` FILTER record.service == @srv`
		params["srv"] = req.GetService()
	}

	if req.GetFilters() != nil {
		for key, value := range req.GetFilters() {
			if key == "transactionType" {
				values := value.GetListValue().AsSlice()
				if len(values) == 0 {
					continue
				}
				query += fmt.Sprintf(` FILTER record.meta["%s"] in @%s`, key, key)
				params[key] = values
			} else if key == "start" || key == "end" || key == "exec" || key == "total" {
				values := value.GetStructValue().AsMap()
				if val, ok := values["from"]; ok {
					from := val.(float64)
					query += fmt.Sprintf(` FILTER record["%s"] >= %f`, key, from)
				}

				if val, ok := values["to"]; ok {
					to := val.(float64)
					query += fmt.Sprintf(` FILTER record["%s"] <= %f`, key, to)
				}
			} else if key == "payment_date" {
				values := value.GetStructValue().AsMap()
				if val, ok := values["from"]; ok {
					from := val.(float64)
					query += fmt.Sprintf(` FILTER record.meta["%s"] >= %f`, key, from)
				}

				if val, ok := values["to"]; ok {
					to := val.(float64)
					query += fmt.Sprintf(` FILTER record.meta["%s"] <= %f`, key, to)
				}
			} else {
				values := value.GetListValue().AsSlice()
				if len(values) == 0 {
					continue
				}
				query += fmt.Sprintf(` FILTER record["%s"] in @%s`, key, key)
				params[key] = values
			}
		}
	}

	if req.Field != nil && req.Sort != nil {
		subQuery := ` SORT record.%s %s`
		field, sort := req.GetField(), req.GetSort()

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

	query += ` 
   LET currency = DOCUMENT(@@currencies, TO_STRING(record.currency.id))
   RETURN merge(record, {uuid: record._key, currency: currency})
) 
RETURN {records: records, total: SUM(records[*].total), count: COUNT(records)}`

	ctrl.log.Debug("Final query", zap.String("query", query))

	cursor, err := ctrl.db.Query(ctx, query, params)
	if err != nil {
		return nil, err
	}

	var res pb.GetRecordsReportsResponse

	_, err = cursor.ReadDocument(ctx, &res)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&res), nil
}

const reportsCountQuery = `RETURN LENGTH(@@instances)`

func (ctrl *recordsController) GetInstancesReportsCount(ctx context.Context) (int64, error) {
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

func (ctrl *recordsController) GetRecordsReportsCount(ctx context.Context, req *pb.GetRecordsReportsCountRequest) (int64, error) {
	query := "LET records = ( FOR record in @@records FILTER record.processed"
	params := map[string]interface{}{
		"@records": schema.RECORDS_COL,
	}

	if req.Account != nil {
		query += ` FILTER record.account == @acc`
		params["acc"] = req.GetAccount()
	}

	if req.Service != nil {
		query += ` FILTER record.service == @srv`
		params["srv"] = req.GetService()
	}

	if req.GetFilters() != nil {
		for key, value := range req.GetFilters() {
			if key == "transactionType" {
				values := value.GetListValue().AsSlice()
				if len(values) == 0 {
					continue
				}
				query += fmt.Sprintf(` FILTER record.meta["%s"] in @%s`, key, key)
				params[key] = values
			} else if key == "start" || key == "end" || key == "exec" || key == "total" {
				values := value.GetStructValue().AsMap()
				if val, ok := values["from"]; ok {
					from := val.(float64)
					query += fmt.Sprintf(` FILTER record["%s"] >= %f`, key, from)
				}

				if val, ok := values["to"]; ok {
					to := val.(float64)
					query += fmt.Sprintf(` FILTER record["%s"] <= %f`, key, to)
				}
			} else if key == "payment_date" {
				values := value.GetStructValue().AsMap()
				if val, ok := values["from"]; ok {
					from := val.(float64)
					query += fmt.Sprintf(` FILTER record.meta["%s"] >= %f`, key, from)
				}

				if val, ok := values["to"]; ok {
					to := val.(float64)
					query += fmt.Sprintf(` FILTER record.meta["%s"] <= %f`, key, to)
				}
			} else {
				values := value.GetListValue().AsSlice()
				if len(values) == 0 {
					continue
				}
				query += fmt.Sprintf(` FILTER record["%s"] in @%s`, key, key)
				params[key] = values
			}
		}
	}

	query += " RETURN record) RETURN LENGTH(records)"

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

const uniqueQuery = `
LET records = (
	FOR r in @@records
	FILTER r.meta != null
	FILTER r.meta.transactionType != null
	FILTER r.meta.transactionType != ""
	RETURN r
)

LET records_products = (
	FOR r in @@records
	FILTER r.product != null
	RETURN r
)

LET records_resources = (
	FOR r in @@records
	FILTER r.resource != null
	RETURN r
)

RETURN {
	types: UNIQUE(records[*].meta.transactionType),
	products: UNIQUE(records_products[*].product),
	resources: UNIQUE(records_resources[*].resource)
}
`

func (ctrl *recordsController) GetUnique(ctx context.Context) (map[string]interface{}, error) {

	cur, err := ctrl.db.Query(ctx, uniqueQuery, map[string]interface{}{
		"@records": schema.RECORDS_COL,
	})
	if err != nil {
		return nil, err
	}

	var result = map[string]interface{}{}

	_, err = cur.ReadDocument(ctx, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}
