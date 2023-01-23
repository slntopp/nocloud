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
	"context"
	epb "github.com/slntopp/nocloud-proto/events"
	"time"

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

func (s *BillingServiceServer) GetTransactions(ctx context.Context, req *pb.GetTransactionsRequest) (*pb.Transactions, error) {
	log := s.log.Named("GetTransactions")
	requestor := ctx.Value(nocloud.NoCloudAccount).(string)
	log.Debug("Request received", zap.Any("request", req), zap.String("requestor", requestor))

	acc := requestor
	if req.Account != nil {
		acc = *req.Account
		node := driver.NewDocumentID(schema.ACCOUNTS_COL, acc)
		if !graph.HasAccess(ctx, s.db, requestor, node, access.Level_ADMIN) {
			return nil, status.Error(codes.PermissionDenied, "Not enough Access Rights")
		}
	}
	query := `FOR t IN @@transactions FILTER t.account == @acc`
	vars := map[string]interface{}{
		"@transactions": schema.TRANSACTIONS_COL,
		"acc":           acc,
	}
	if req.Service != nil {
		service := *req.Service
		node := driver.NewDocumentID(schema.SERVICES_COL, service)
		if !graph.HasAccess(ctx, s.db, requestor, node, access.Level_ADMIN) {
			return nil, status.Error(codes.PermissionDenied, "Not enough Access Rights")
		}
		query += ` && t.service == @service`
		vars["service"] = service
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

func (s *BillingServiceServer) CreateTransaction(ctx context.Context, t *pb.Transaction) (*pb.Transaction, error) {
	log := s.log.Named("CreateTransaction")
	requestor := ctx.Value(nocloud.NoCloudAccount).(string)
	log.Debug("Request received", zap.Any("transaction", t), zap.String("requestor", requestor))

	ns := driver.NewDocumentID(schema.NAMESPACES_COL, schema.ROOT_NAMESPACE_KEY)
	ok := graph.HasAccess(ctx, s.db, requestor, ns, access.Level_ROOT)
	if !ok {
		return nil, status.Error(codes.PermissionDenied, "Not enough Access Rights")
	}

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

	if r.Transaction.Priority == pb.Priority_URGENT {
		acc := driver.NewDocumentID(schema.ACCOUNTS_COL, r.Transaction.Account)
		transaction := driver.NewDocumentID(schema.TRANSACTIONS_COL, r.Transaction.Uuid)

		_, err := s.db.Query(ctx, processUrgentTransaction, map[string]interface{}{
			"@accounts":      schema.ACCOUNTS_COL,
			"@transactions":  schema.TRANSACTIONS_COL,
			"accountKey":     acc.String(),
			"transactionKey": transaction.String(),
			"now":            time.Now().Unix(),
		})
		if err != nil {
			log.Error("Failed to process transaction", zap.String("err", err.Error()))
			return nil, status.Error(codes.Internal, "Failed to process transaction")
		}
	}

	return r.Transaction, nil
}

const processUrgentTransaction = `
LET account = DOCUMENT(@accountKey)
LET transaction = DOCUMENT(@transactionKey)

UPDATE transaction WITH {processed: true, proc: @now} IN @@transactions
UPDATE account WITH { balance: account.balance - transaction.total } IN @@accounts

return account
`

const reprocessTransactions = `
LET account = UNSET(DOCUMENT(@account), "balance")
LET currency = account.currency != null ? account.currency : @currency
LET transactions = (
FOR t IN @@transactions // Iterate over Transactions
FILTER t.exec <= @now
FILTER t.account == account._key
	LET rate = PRODUCT(
		FOR vertex, edge IN OUTBOUND SHORTEST_PATH
		DOCUMENT(CONCAT(@currencies, "/", TO_NUMBER(t.currency))) TO
		DOCUMENT(CONCAT(@currencies, "/", currency)) GRAPH @graph
			RETURN edge.rate
	)
    UPDATE t WITH { processed: true, proc: @now, total: t.total * rate, currency: currency } IN @@transactions RETURN NEW )

UPDATE account WITH { balance: -SUM(transactions[*].total) } IN @@accounts
FOR t IN transactions
    RETURN t
`

func (s *BillingServiceServer) Reprocess(ctx context.Context, req *pb.ReprocessTransactionsRequest) (*pb.Transactions, error) {
	log := s.log.Named("Reprocess")
	requestor := ctx.Value(nocloud.NoCloudAccount).(string)
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
	return &pb.Transactions{Pool: transactions}, nil
}
