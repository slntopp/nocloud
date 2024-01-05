package billing

import (
	"context"
	"fmt"
	"github.com/arangodb/go-driver"
	"github.com/slntopp/nocloud-proto/access"
	pb "github.com/slntopp/nocloud-proto/billing"
	epb "github.com/slntopp/nocloud-proto/events"
	"github.com/slntopp/nocloud-proto/registry/accounts"
	"github.com/slntopp/nocloud/pkg/graph"
	"github.com/slntopp/nocloud/pkg/nocloud"
	"github.com/slntopp/nocloud/pkg/nocloud/schema"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
	"time"
)

func (s *BillingServiceServer) GetInvoices(ctx context.Context, req *pb.GetInvoicesRequest) (*pb.Invoices, error) {
	log := s.log.Named("GetTransactions")
	requestor := ctx.Value(nocloud.NoCloudAccount).(string)
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
	return &pb.Transactions{Pool: transactions}, nil
}

func (s *BillingServiceServer) CreateInvoice(ctx context.Context, t *pb.Invoice) (*pb.Invoice, error) {
	log := s.log.Named("CreateTransaction")
	requestor := ctx.Value(nocloud.NoCloudAccount).(string)
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

	var baseRec, prevRec string

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
	}

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
	}

	if baseRec != "" {
		recBody.Base = &baseRec
	}

	if prevRec != "" {
		recBody.Previous = &prevRec
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

		var cur pb.Currency

		if dbAcc.Currency == nil {
			cur = pb.Currency_NCU
		} else {
			cur = *dbAcc.Currency
		}

		var rate float64 = 1

		if cur != pb.Currency(currencyConf.Currency) {
			rate, err = s.currencies.GetExchangeRate(ctx, cur, pb.Currency(currencyConf.Currency))

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

	return r.Transaction, nil
}

func (s *BillingServiceServer) GetInvoicesCount(ctx context.Context, req *pb.GetInvoicesCountRequest) (*pb.GetInvoicesCountResponse, error) {
	log := s.log.Named("GetTransactionsCount")
	requestor := ctx.Value(nocloud.NoCloudAccount).(string)
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

	return &pb.GetTransactionsCountResponse{
		Total: uint64(cursor.Count()),
	}, nil
}

func (s *BillingServiceServer) UpdateInvoice(ctx context.Context, req *pb.Invoice) (*pb.Invoice, error) {
	log := s.log.Named("UpdateTransaction")
	requestor := ctx.Value(nocloud.NoCloudAccount).(string)
	log.Debug("Request received", zap.Any("transaction", req), zap.String("requestor", requestor))

	ns := driver.NewDocumentID(schema.NAMESPACES_COL, schema.ROOT_NAMESPACE_KEY)
	ok := graph.HasAccess(ctx, s.db, requestor, ns, access.Level_ROOT)
	if !ok {
		return nil, status.Error(codes.PermissionDenied, "Not enough Access Rights")
	}

	t, err := s.transactions.Get(ctx, req.GetUuid())
	if err != nil {
		return nil, err
	}

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

	if t.GetPriority() == pb.Priority_URGENT && t.GetExec() != 0 {
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

		var cur pb.Currency

		if dbAcc.Currency == nil {
			cur = pb.Currency_NCU
		} else {
			cur = *dbAcc.Currency
		}

		var rate float64 = 1

		if cur != pb.Currency(currencyConf.Currency) {
			rate, err = s.currencies.GetExchangeRate(ctx, cur, pb.Currency(currencyConf.Currency))

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

	return &pb.UpdateTransactionResponse{Result: true}, nil
}
