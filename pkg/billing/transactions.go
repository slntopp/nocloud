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
	"github.com/slntopp/nocloud/pkg/graph"
	"github.com/slntopp/nocloud/pkg/nocloud"
	"github.com/slntopp/nocloud/pkg/nocloud/access"
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
		node := driver.NewDocumentID(schema.ACCOUNTS_COL, acc).String()
		if !graph.HasAccess(ctx, s.db, requestor, node, access.ADMIN) {
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
		node := driver.NewDocumentID(schema.SERVICES_COL, service).String()
		if !graph.HasAccess(ctx, s.db, requestor, node, access.ADMIN) {
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

	ns := driver.NewDocumentID(schema.NAMESPACES_COL, schema.ROOT_NAMESPACE_KEY).String()
	ok := graph.HasAccess(ctx, s.db, requestor, ns, access.SUDO)
	if !ok {
		return nil, status.Error(codes.PermissionDenied, "Not enough Access Rights")
	}

	r, err := s.transactions.Create(ctx, t)
	if err != nil {
		log.Error("Failed to create transaction", zap.Error(err))
		return nil, status.Error(codes.Internal, "Failed to create transaction")
	}
	return r.Transaction, nil
}

const reprocessTransactions = `
LET account = UNSET(DOCUMENT(@account), "balance")
LET transactions = (
FOR t IN @@transactions // Iterate over Transactions
FILTER t.exec <= @now
FILTER t.account == account._key
    UPDATE t WITH { processed: true, proc: @now } IN @@transactions RETURN NEW )

UPDATE account WITH { balance: -SUM(transactions[*].total) } IN @@accounts
FOR t IN transactions
    RETURN t
`

func (s *BillingServiceServer) Reprocess(ctx context.Context, req *pb.ReprocessTransactionsRequest) (*pb.Transactions, error) {
	log := s.log.Named("Reprocess")
	requestor := ctx.Value(nocloud.NoCloudAccount).(string)
	log.Debug("Request received", zap.Any("request", req), zap.String("requestor", requestor))

	ns := driver.NewDocumentID(schema.NAMESPACES_COL, schema.ROOT_NAMESPACE_KEY).String()
	ok := graph.HasAccess(ctx, s.db, requestor, ns, access.SUDO)
	if !ok {
		return nil, status.Error(codes.PermissionDenied, "Not enough Access Rights")
	}

	acc := driver.NewDocumentID(schema.ACCOUNTS_COL, req.Account)
	c, err := s.db.Query(ctx, reprocessTransactions, map[string]interface{}{
		"@accounts":     schema.ACCOUNTS_COL,
		"@transactions": schema.TRANSACTIONS_COL,
		"account":       acc.String(),
		"now":           time.Now().Unix(),
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
