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
package billing

import (
	"connectrpc.com/connect"
	"context"
	"fmt"
	"time"

	"google.golang.org/protobuf/types/known/structpb"

	epb "github.com/slntopp/nocloud-proto/events"
	"github.com/slntopp/nocloud-proto/registry/accounts"

	"github.com/arangodb/go-driver"
	"github.com/slntopp/nocloud-proto/access"
	pb "github.com/slntopp/nocloud-proto/billing"
	"github.com/slntopp/nocloud/pkg/graph"
	"github.com/slntopp/nocloud/pkg/nocloud"
	"github.com/slntopp/nocloud/pkg/nocloud/schema"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *BillingServiceServer) _HandleGetSingleTransaction(ctx context.Context, acc, uuid string) (*connect.Response[pb.Transactions], error) {
	tr, err := s.transactions.Get(ctx, uuid)
	if err != nil {
		return nil, status.Error(codes.NotFound, "Transaction doesn't exist")
	}

	ok := graph.HasAccess(ctx, s.db, acc, driver.NewDocumentID(schema.ACCOUNTS_COL, tr.Account), access.Level_ADMIN)

	if !ok {
		return nil, status.Error(codes.PermissionDenied, "Not enoguh Access Rights")
	}

	resp := connect.NewResponse(&pb.Transactions{Pool: []*pb.Transaction{tr}})

	return resp, nil
}

func (s *BillingServiceServer) GetTransactions(ctx context.Context, r *connect.Request[pb.GetTransactionsRequest]) (*connect.Response[pb.Transactions], error) {
	log := s.log.Named("GetTransactions")
	requestor := ctx.Value(nocloud.NoCloudAccount).(string)
	req := r.Msg

	log.Debug("Request received", zap.Any("request", req), zap.String("requestor", requestor))

	acc := requestor

	query := `FOR t IN @@transactions`
	vars := map[string]interface{}{
		"@transactions": schema.TRANSACTIONS_COL,
	}

	if req.GetUuid() != "" {
		return s._HandleGetSingleTransaction(ctx, acc, req.GetUuid())
	}

	if req.Account != nil {
		acc = *req.Account
		node := driver.NewDocumentID(schema.ACCOUNTS_COL, acc)
		if !graph.HasAccess(ctx, s.db, requestor, node, access.Level_ADMIN) {
			return nil, status.Error(codes.PermissionDenied, "Not enough Access Rights")
		}
		query += ` FILTER t.account == @acc`
		vars["acc"] = acc
	} else {
		if acc != schema.ROOT_ACCOUNT_KEY {
			return nil, status.Error(codes.PermissionDenied, "Not enough Access Rights")
		}
	}

	if req.Service != nil {
		service := *req.Service
		node := driver.NewDocumentID(schema.SERVICES_COL, service)
		if !graph.HasAccess(ctx, s.db, requestor, node, access.Level_ADMIN) {
			return nil, status.Error(codes.PermissionDenied, "Not enough Access Rights")
		}
		if req.Account == nil {
			query += ` FILTER t.service == @service`
		} else {
			query += ` && t.service == @service`
		}
		vars["service"] = service
	}

	if req.Type != nil {
		transactionType := req.GetType()

		if req.Account == nil && req.Service == nil {
			query += ` FILTER t.meta.type == @type`
		} else {
			query += ` && t.meta.type == @type`
		}
		vars["type"] = transactionType
	}

	if req.Field != nil && req.Sort != nil {
		subQuery := ` SORT t.%s %s`
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
			vars["offset"] = offset
			vars["count"] = limit
		}
	}
	query += ` RETURN t`

	log.Debug("Ready to retrieve transactions", zap.String("query", query), zap.Any("vars", vars))

	cursor, err := s.db.Query(ctx, query, vars)
	if err != nil {
		log.Error("Failed to retrieve transactions", zap.Error(err))
		return nil, status.Error(codes.Internal, "Failed to retrieve transactions")
	}
	defer cursor.Close()

	var transactions []*pb.Transaction
	for {
		transaction := &pb.Transaction{}
		meta, err := cursor.ReadDocument(ctx, transaction)
		if err != nil {
			if driver.IsNoMoreDocuments(err) {
				break
			}
			log.Error("Failed to retrieve transactions", zap.Error(err))
			return nil, status.Error(codes.Internal, "Failed to retrieve transactions")
		}
		transaction.Uuid = meta.Key
		transactions = append(transactions, transaction)
	}

	log.Debug("Transactions retrieved", zap.Any("transactions", transactions))

	resp := connect.NewResponse(&pb.Transactions{Pool: transactions})

	return resp, nil
}

func (s *BillingServiceServer) CreateTransaction(ctx context.Context, req *connect.Request[pb.Transaction]) (*connect.Response[pb.Transaction], error) {
	log := s.log.Named("CreateTransaction")
	requestor := ctx.Value(nocloud.NoCloudAccount).(string)
	t := req.Msg
	log.Debug("Request received", zap.Any("transaction", t), zap.String("requestor", requestor))

	ns := driver.NewDocumentID(schema.NAMESPACES_COL, schema.ROOT_NAMESPACE_KEY)
	ok := graph.HasAccess(ctx, s.db, requestor, ns, access.Level_ROOT)
	if !ok {
		return nil, status.Error(codes.PermissionDenied, "Not enough Access Rights")
	}

	if t.Meta == nil {
		t.Meta = map[string]*structpb.Value{}
		t.Meta["type"] = structpb.NewStringValue("transaction")
	}

	/*var baseRec, prevRec string

	if t.Base != nil {
		query, err := s.db.Query(ctx, getTransactionRecord, map[string]interface{}{
			"transactionKey": driver.NewDocumentID(schema.TRANSACTIONS_COL, t.GetBase()),
		})
		if err != nil {
			log.Error("Failed get base record", zap.Error(err))
			return nil, err
		}
		if query.HasMore() {
			_, err := query.ReadDocument(ctx, &baseRec)
			if err != nil {
				log.Error("Failed read base record", zap.Error(err))
				return nil, err
			}
		}
	}

	if t.Previous != nil {
		query, err := s.db.Query(ctx, getTransactionRecord, map[string]interface{}{
			"transactionKey": driver.NewDocumentID(schema.TRANSACTIONS_COL, t.GetBase()),
		})
		if err != nil {
			log.Error("Failed get base record", zap.Error(err))
			return nil, err
		}
		if query.HasMore() {
			_, err := query.ReadDocument(ctx, &prevRec)
			if err != nil {
				log.Error("Failed read base record", zap.Error(err))
				return nil, err
			}
		}
	}*/

	recBody := &pb.Record{
		Start:     time.Now().Unix(),
		End:       time.Now().Unix() + 1,
		Exec:      time.Now().Unix(),
		Processed: true,
		Priority:  t.GetPriority(),
		Total:     t.GetTotal(),
		Currency:  t.GetCurrency(),
		Service:   t.GetService(),
		Account:   t.GetAccount(),
		Meta:      t.GetMeta(),
		Cost:      t.GetTotal(),
	}

	if t.GetBase() != "" {
		recBody.Base = t.Base
	}

	if t.GetPrevious() != "" {
		recBody.Previous = t.Previous
	}

	rec := s.records.Create(ctx, recBody)

	if t.GetRecords() == nil {
		t.Records = []string{}
	}

	t.Records = append(t.Records, rec.Key())

	t.Created = time.Now().Unix()

	r, err := s.transactions.Create(ctx, t)
	if err != nil {
		log.Error("Failed to create transaction", zap.Error(err))
		return nil, status.Error(codes.Internal, "Failed to create transaction")
	}

	eventsClient.Publish(ctx, &epb.Event{
		Type: "email",
		Uuid: t.GetAccount(),
		Key:  "transaction_created",
	})

	if r.Transaction.Priority == pb.Priority_URGENT && r.Transaction.GetExec() != 0 {
		acc := driver.NewDocumentID(schema.ACCOUNTS_COL, r.Transaction.Account)
		transaction := driver.NewDocumentID(schema.TRANSACTIONS_COL, r.Transaction.Uuid)
		currencyConf := MakeCurrencyConf(ctx, log)
		suspConf := MakeSuspendConf(ctx, log)

		_, err := s.db.Query(ctx, processUrgentTransaction, map[string]interface{}{
			"@accounts":      schema.ACCOUNTS_COL,
			"@transactions":  schema.TRANSACTIONS_COL,
			"@records":       schema.RECORDS_COL,
			"accountKey":     acc.String(),
			"transactionKey": transaction.String(),
			"currency":       currencyConf.Currency,
			"currencies":     schema.CUR_COL,
			"now":            time.Now().Unix(),
			"graph":          schema.BILLING_GRAPH.Name,
		})
		if err != nil {
			log.Error("Failed to process transaction", zap.String("err", err.Error()))
			return nil, status.Error(codes.Internal, err.Error())
		}

		dbAcc, err := accClient.Get(ctx, &accounts.GetRequest{Uuid: r.Transaction.Account, Public: false})

		if err != nil {
			log.Error("Failed to get account", zap.String("err", err.Error()))
			return nil, status.Error(codes.Internal, err.Error())
		}

		var cur *pb.Currency

		if dbAcc.Currency == nil {
			cur = currencyConf.Currency
		} else {
			cur = dbAcc.Currency
		}

		var rate float64 = 1

<<<<<<< HEAD
		if cur.GetId() != currencyConf.Currency.GetId() {
			rate, err = s.currencies.GetExchangeRate(ctx, cur, currencyConf.Currency)
=======
		if cur != pb.Currency(currencyConf.Currency) {
			rate, _, err = s.currencies.GetExchangeRate(ctx, cur, pb.Currency(currencyConf.Currency))
>>>>>>> dev

			if err != nil {
				log.Error("Failed to get exchange rate", zap.String("err", err.Error()))
				return nil, status.Error(codes.Internal, err.Error())
			}
		}

		var balance = 0.0
		if dbAcc.Balance != nil {
			balance = *dbAcc.Balance
		}

		balance = balance * rate

		var isSuspended = false

		if dbAcc.Suspended != nil {
			isSuspended = *dbAcc.Suspended
		}

		if !isSuspended && balance < suspConf.Limit {
			_, err := accClient.Suspend(ctx, &accounts.SuspendRequest{Uuid: r.Transaction.Account})
			if err != nil {
				log.Error("Failed to suspend account", zap.String("err", err.Error()))
				return nil, status.Error(codes.Internal, err.Error())
			}
		} else if isSuspended && balance > suspConf.Limit {
			_, err := accClient.Unsuspend(ctx, &accounts.UnsuspendRequest{Uuid: r.Transaction.Account})
			if err != nil {
				log.Error("Failed to unsuspend account", zap.String("err", err.Error()))
				return nil, status.Error(codes.Internal, err.Error())
			}
		}

	} else {
		acc := driver.NewDocumentID(schema.ACCOUNTS_COL, r.Transaction.Account)
		transaction := driver.NewDocumentID(schema.TRANSACTIONS_COL, r.Transaction.Uuid)
		currencyConf := MakeCurrencyConf(ctx, log)

		_, err := s.db.Query(ctx, updateTransactionWithCurrency, map[string]interface{}{
			"@transactions":  schema.TRANSACTIONS_COL,
			"@records":       schema.RECORDS_COL,
			"accountKey":     acc.String(),
			"transactionKey": transaction.String(),
			"currency":       currencyConf.Currency,
			"currencies":     schema.CUR_COL,
			"graph":          schema.BILLING_GRAPH.Name,
		})
		if err != nil {
			log.Error("Failed to process transaction", zap.String("err", err.Error()))
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	resp := connect.NewResponse(r.Transaction)

	return resp, nil
}

func (s *BillingServiceServer) GetTransactionsCount(ctx context.Context, r *connect.Request[pb.GetTransactionsCountRequest]) (*connect.Response[pb.GetTransactionsCountResponse], error) {
	log := s.log.Named("GetTransactionsCount")
	requestor := ctx.Value(nocloud.NoCloudAccount).(string)
	req := r.Msg
	log.Debug("Request received", zap.Any("request", req), zap.String("requestor", requestor))

	acc := requestor

	query := `FOR t IN @@transactions`
	vars := map[string]interface{}{
		"@transactions": schema.TRANSACTIONS_COL,
	}

	if req.Account != nil {
		acc = *req.Account
		node := driver.NewDocumentID(schema.ACCOUNTS_COL, acc)
		if !graph.HasAccess(ctx, s.db, requestor, node, access.Level_ADMIN) {
			return nil, status.Error(codes.PermissionDenied, "Not enough Access Rights")
		}
		query += ` FILTER t.account == @acc`
		vars["acc"] = acc
	} else {
		if acc != schema.ROOT_ACCOUNT_KEY {
			return nil, status.Error(codes.PermissionDenied, "Not enough Access Rights")
		}
	}

	if req.Service != nil {
		service := *req.Service
		node := driver.NewDocumentID(schema.SERVICES_COL, service)
		if !graph.HasAccess(ctx, s.db, requestor, node, access.Level_ADMIN) {
			return nil, status.Error(codes.PermissionDenied, "Not enough Access Rights")
		}
		if req.Account == nil {
			query += ` FILTER t.service == @service`
		} else {
			query += ` && t.service == @service`
		}
		vars["service"] = service
	}

	if req.Type != nil {
		transactionType := req.GetType()

		if req.Account == nil && req.Service == nil {
			query += ` FILTER t.meta.type == @type`
		} else {
			query += ` && t.meta.type == @type`
		}
		vars["type"] = transactionType
	}

	query += ` RETURN t`

	log.Debug("Ready to retrieve transactions", zap.String("query", query), zap.Any("vars", vars))

	queryContext := driver.WithQueryCount(ctx)

	cursor, err := s.db.Query(queryContext, query, vars)
	if err != nil {
		log.Error("Failed to retrieve transactions", zap.Error(err))
		return nil, status.Error(codes.Internal, "Failed to retrieve transactions")
	}
	defer cursor.Close()

	log.Info("transactions count", zap.Int64("count", cursor.Count()))

	resp := connect.NewResponse(&pb.GetTransactionsCountResponse{Total: uint64(cursor.Count())})

	return resp, nil
}

func (s *BillingServiceServer) UpdateTransaction(ctx context.Context, r *connect.Request[pb.Transaction]) (*connect.Response[pb.UpdateTransactionResponse], error) {
	log := s.log.Named("UpdateTransaction")
	requestor := ctx.Value(nocloud.NoCloudAccount).(string)
	req := r.Msg
	log.Debug("Request received", zap.Any("transaction", req), zap.String("requestor", requestor))

	ns := driver.NewDocumentID(schema.NAMESPACES_COL, schema.ROOT_NAMESPACE_KEY)
	ok := graph.HasAccess(ctx, s.db, requestor, ns, access.Level_ROOT)
	if !ok {
		return nil, status.Error(codes.PermissionDenied, "Not enough Access Rights")
	}

	t, err := s.transactions.Get(ctx, req.GetUuid())
	if err != nil {
		log.Error("Failed to get transaction", zap.Error(err))
		return nil, status.Error(codes.Internal, "Failed to get transaction")
	}

	exec := t.GetExec()
	if exec != 0 {
		log.Error("Transaction has exec timestamp")
		return nil, status.Error(codes.Internal, "Transaction has exec timestamp")
	}
	if req.GetExec() != 0 {
		t.Exec = req.GetExec()
	}
	t.Uuid = req.GetUuid()
	t.Meta = req.GetMeta()

	_, err = s.transactions.Update(ctx, t)
	if err != nil {
		log.Error("Failed to update transaction", zap.Error(err))
		return nil, status.Error(codes.Internal, "Failed to update transaction")
	}

	_, err = s.db.Query(ctx, updateRecordsMeta, map[string]interface{}{
		"@records":       schema.RECORDS_COL,
		"transactionKey": driver.NewDocumentID(schema.TRANSACTIONS_COL, t.Uuid).String(),
	})

	if err != nil {
		log.Error("Failed to update record", zap.Error(err))
		return nil, err
	}

	if t.GetExec() != 0 {
		acc := driver.NewDocumentID(schema.ACCOUNTS_COL, t.Account)
		transaction := driver.NewDocumentID(schema.TRANSACTIONS_COL, t.Uuid)
		currencyConf := MakeCurrencyConf(ctx, log)
		suspConf := MakeSuspendConf(ctx, log)

		_, err := s.db.Query(ctx, processUrgentTransaction, map[string]interface{}{
			"@accounts":      schema.ACCOUNTS_COL,
			"@transactions":  schema.TRANSACTIONS_COL,
			"@records":       schema.RECORDS_COL,
			"accountKey":     acc.String(),
			"transactionKey": transaction.String(),
			"currency":       currencyConf.Currency,
			"currencies":     schema.CUR_COL,
			"now":            time.Now().Unix(),
			"graph":          schema.BILLING_GRAPH.Name,
		})
		if err != nil {
			log.Error("Failed to process transaction", zap.String("err", err.Error()))
			return nil, status.Error(codes.Internal, err.Error())
		}

		dbAcc, err := accClient.Get(ctx, &accounts.GetRequest{Uuid: t.Account, Public: false})

		if err != nil {
			log.Error("Failed to get account", zap.String("err", err.Error()))
			return nil, status.Error(codes.Internal, err.Error())
		}

		var cur *pb.Currency

		if dbAcc.Currency == nil {
			cur = currencyConf.Currency
		} else {
			cur = dbAcc.Currency
		}

		var rate float64 = 1

<<<<<<< HEAD
		if cur.GetId() != currencyConf.Currency.GetId() {
			rate, err = s.currencies.GetExchangeRate(ctx, cur, currencyConf.Currency)
=======
		if cur != pb.Currency(currencyConf.Currency) {
			rate, _, err = s.currencies.GetExchangeRate(ctx, cur, pb.Currency(currencyConf.Currency))
>>>>>>> dev

			if err != nil {
				log.Error("Failed to get exchange rate", zap.String("err", err.Error()))
				return nil, status.Error(codes.Internal, err.Error())
			}
		}

		var balance = 0.0
		if dbAcc.Balance != nil {
			balance = *dbAcc.Balance
		}

		balance = balance * rate

		var isSuspended = false

		if dbAcc.Suspended != nil {
			isSuspended = *dbAcc.Suspended
		}

		if !isSuspended && balance < suspConf.Limit {
			_, err := accClient.Suspend(ctx, &accounts.SuspendRequest{Uuid: t.Account})
			if err != nil {
				log.Error("Failed to suspend account", zap.String("err", err.Error()))
				return nil, status.Error(codes.Internal, err.Error())
			}
		} else if isSuspended && balance > suspConf.Limit {
			_, err := accClient.Unsuspend(ctx, &accounts.UnsuspendRequest{Uuid: t.Account})
			if err != nil {
				log.Error("Failed to unsuspend account", zap.String("err", err.Error()))
				return nil, status.Error(codes.Internal, err.Error())
			}
		}
	}

	resp := connect.NewResponse(&pb.UpdateTransactionResponse{Result: true})

	return resp, nil
}

const processUrgentTransaction = `
LET account = DOCUMENT(@accountKey)
LET transaction = DOCUMENT(@transactionKey)

LET currency = account.currency != null ? account.currency : @currency
LET rate = PRODUCT(
	FOR vertex, edge IN OUTBOUND
	SHORTEST_PATH DOCUMENT(CONCAT(@currencies, "/", TO_NUMBER(transaction.currency.id)))
	TO DOCUMENT(CONCAT(@currencies, "/", currency.id))
	GRAPH @graph
	FILTER edge
		RETURN edge.rate + (TO_NUMBER(edge.commission) / 100) * edge.rate
)

LET total = transaction.total * rate

FOR r in transaction.records
	UPDATE r WITH {cost: total, currency: currency, meta: MERGE(transaction.meta, {transaction: transaction._key, payment_date: @now}), exec: transaction.exec} in @@records

UPDATE transaction WITH {processed: true, proc: @now, currency: currency, total: total} IN @@transactions
UPDATE account WITH { balance: account.balance - total } IN @@accounts

return account
`

const updateTransactionWithCurrency = `
LET account = DOCUMENT(@accountKey)
LET transaction = DOCUMENT(@transactionKey)

LET currency = account.currency != null ? account.currency : @currency
LET rate = PRODUCT(
	FOR vertex, edge IN OUTBOUND
	SHORTEST_PATH DOCUMENT(CONCAT(@currencies, "/", TO_NUMBER(transaction.currency.id)))
	TO DOCUMENT(CONCAT(@currencies, "/", currency.id))
	GRAPH @graph
	FILTER edge
		RETURN edge.rate + (TO_NUMBER(edge.commission) / 100) * edge.rate
)

LET total = transaction.total * rate

FOR r in transaction.records
	UPDATE r WITH {cost: total, currency: currency, meta: MERGE(transaction.meta, {transaction: transaction._key})} in @@records

UPDATE transaction WITH {currency: currency, total: total} IN @@transactions
RETURN transaction
`

const updateRecordsMeta = `
LET transaction = DOCUMENT(@transactionKey)
FOR r in transaction.records
	UPDATE r WITH {meta: MERGE(transaction.meta, {transaction: transaction._key})} in @@records
`

const reprocessTransactions = `
LET account = UNSET(DOCUMENT(@account), "balance")
LET currency = account.currency != null ? account.currency : @currency
LET transactions = (
FOR t IN @@transactions // Iterate over Transactions
FILTER t.exec != null
FILTER t.exec <= @now
FILTER t.account == account._key
	LET rate = PRODUCT(
		FOR vertex, edge IN OUTBOUND
		SHORTEST_PATH DOCUMENT(CONCAT(@currencies, "/", TO_NUMBER(t.currency.id)))
		TO DOCUMENT(CONCAT(@currencies, "/", currency.id))
		GRAPH @graph
		FILTER edge
			RETURN edge.rate + (TO_NUMBER(edge.commission) / 100) * edge.rate
	)
    UPDATE t WITH { processed: true, proc: @now, total: t.total * rate, currency: currency } IN @@transactions RETURN NEW )

UPDATE account WITH { balance: -SUM(transactions[*].total) } IN @@accounts
FOR t IN transactions
    RETURN t
`

func (s *BillingServiceServer) Reprocess(ctx context.Context, r *connect.Request[pb.ReprocessTransactionsRequest]) (*connect.Response[pb.Transactions], error) {
	log := s.log.Named("Reprocess")
	requestor := ctx.Value(nocloud.NoCloudAccount).(string)

	req := r.Msg
	log.Debug("Request received", zap.Any("request", req), zap.String("requestor", requestor))

	currencyConf := MakeCurrencyConf(ctx, log)

	ns := driver.NewDocumentID(schema.NAMESPACES_COL, schema.ROOT_NAMESPACE_KEY)
	ok := graph.HasAccess(ctx, s.db, requestor, ns, access.Level_ROOT)
	if !ok {
		return nil, status.Error(codes.PermissionDenied, "Not enough Access Rights")
	}

	acc := driver.NewDocumentID(schema.ACCOUNTS_COL, req.Account)
	c, err := s.db.Query(ctx, reprocessTransactions, map[string]interface{}{
		"@accounts":     schema.ACCOUNTS_COL,
		"@transactions": schema.TRANSACTIONS_COL,
		"account":       acc.String(),
		"now":           time.Now().Unix(),
		"currency":      currencyConf.Currency,
		"currencies":    schema.CUR_COL,
		"graph":         schema.BILLING_GRAPH.Name,
	})
	if err != nil {
		log.Error("Error Reprocessing Transactions", zap.Error(err))
		return nil, status.Error(codes.Internal, "Error Reprocessing Transactions")
	}
	defer c.Close()

	var transactions []*pb.Transaction
	for {
		transaction := &pb.Transaction{}
		meta, err := c.ReadDocument(ctx, transaction)
		if err != nil {
			if driver.IsNoMoreDocuments(err) {
				break
			}
			log.Error("Failed to retrieve transactions", zap.Error(err))
			return nil, status.Error(codes.Internal, "Failed to retrieve transactions")
		}
		transaction.Uuid = meta.Key
		transactions = append(transactions, transaction)
	}

	log.Debug("Transactions retrieved", zap.Any("transactions", transactions))
	resp := connect.NewResponse(&pb.Transactions{Pool: transactions})
	return resp, nil
}
