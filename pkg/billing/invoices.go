package billing

import (
	"connectrpc.com/connect"
	"context"
	"fmt"
	"github.com/arangodb/go-driver"
	"github.com/slntopp/nocloud-proto/access"
	pb "github.com/slntopp/nocloud-proto/billing"
	epb "github.com/slntopp/nocloud-proto/events"
	"github.com/slntopp/nocloud/pkg/graph"
	"github.com/slntopp/nocloud/pkg/nocloud"
	"github.com/slntopp/nocloud/pkg/nocloud/schema"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"time"
)

func (s *BillingServiceServer) GetInvoices(ctx context.Context, r *connect.Request[pb.GetInvoicesRequest]) (*connect.Response[pb.Invoices], error) {
	log := s.log.Named("GetTransactions")
	requestor := ctx.Value(nocloud.NoCloudAccount).(string)

	req := r.Msg
	log.Debug("Request received", zap.Any("request", req), zap.String("requestor", requestor))

	acc := requestor

	query := `FOR t IN @@invoices`
	vars := map[string]interface{}{
		"@invoices": schema.INVOICES_COL,
	}

	if req.GetUuid() != "" {
		return s._HandleGetSingleInvoice(ctx, acc, req.GetUuid())
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
	query += ` RETURN merge(t, {uuid: t._key})`

	log.Debug("Ready to retrieve transactions", zap.String("query", query), zap.Any("vars", vars))

	cursor, err := s.db.Query(ctx, query, vars)
	if err != nil {
		log.Error("Failed to retrieve transactions", zap.Error(err))
		return nil, status.Error(codes.Internal, "Failed to retrieve transactions")
	}
	defer cursor.Close()

	var transactions []*pb.Invoice
	for {
		transaction := &pb.Invoice{}
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
	resp := connect.NewResponse(&pb.Invoices{Pool: transactions})
	return resp, nil
}

func (s *BillingServiceServer) CreateInvoice(ctx context.Context, req *connect.Request[pb.Invoice]) (*connect.Response[pb.Invoice], error) {
	log := s.log.Named("CreateTransaction")
	requestor := ctx.Value(nocloud.NoCloudAccount).(string)

	t := req.Msg
	log.Debug("Request received", zap.Any("transaction", t), zap.String("requestor", requestor))

	ns := driver.NewDocumentID(schema.NAMESPACES_COL, schema.ROOT_NAMESPACE_KEY)
	ok := graph.HasAccess(ctx, s.db, requestor, ns, access.Level_ROOT)
	if !ok {
		return nil, status.Error(codes.PermissionDenied, "Not enough Access Rights")
	}

	t.Created = time.Now().Unix()

	r, err := s.invoices.Create(ctx, t)
	if err != nil {
		log.Error("Failed to create transaction", zap.Error(err))
		return nil, status.Error(codes.Internal, "Failed to create transaction")
	}

	eventsClient.Publish(ctx, &epb.Event{
		Type: "email",
		Uuid: t.GetAccount(),
		Key:  "invoice_created",
	})

	resp := connect.NewResponse(r)

	return resp, nil
}

func (s *BillingServiceServer) GetInvoicesCount(ctx context.Context, r *connect.Request[pb.GetInvoicesCountRequest]) (*connect.Response[pb.GetInvoicesCountResponse], error) {
	log := s.log.Named("GetTransactionsCount")
	requestor := ctx.Value(nocloud.NoCloudAccount).(string)
	req := r.Msg
	log.Debug("Request received", zap.Any("request", req), zap.String("requestor", requestor))

	acc := requestor

	query := `FOR t IN @@invoices`
	vars := map[string]interface{}{
		"@invoices": schema.INVOICES_COL,
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

	resp := connect.NewResponse(&pb.GetInvoicesCountResponse{
		Total: uint64(cursor.Count()),
	})

	return resp, nil
}

func (s *BillingServiceServer) UpdateInvoice(ctx context.Context, r *connect.Request[pb.Invoice]) (*connect.Response[pb.Invoice], error) {
	log := s.log.Named("UpdateTransaction")
	requestor := ctx.Value(nocloud.NoCloudAccount).(string)
	req := r.Msg
	log.Debug("Request received", zap.Any("transaction", req), zap.String("requestor", requestor))

	ns := driver.NewDocumentID(schema.NAMESPACES_COL, schema.ROOT_NAMESPACE_KEY)
	ok := graph.HasAccess(ctx, s.db, requestor, ns, access.Level_ROOT)
	if !ok {
		return nil, status.Error(codes.PermissionDenied, "Not enough Access Rights")
	}

	t, err := s.invoices.Get(ctx, req.GetUuid())
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

	_, err = s.invoices.Update(ctx, t)
	if err != nil {
		log.Error("Failed to update transaction", zap.Error(err))
		return nil, status.Error(codes.Internal, "Failed to update transaction")
	}

	return connect.NewResponse(t), nil
}

func (s *BillingServiceServer) _HandleGetSingleInvoice(ctx context.Context, acc, uuid string) (*connect.Response[pb.Invoices], error) {
	tr, err := s.invoices.Get(ctx, uuid)
	if err != nil {
		return nil, status.Error(codes.NotFound, "Transaction doesn't exist")
	}

	ok := graph.HasAccess(ctx, s.db, acc, driver.NewDocumentID(schema.ACCOUNTS_COL, tr.Account), access.Level_ADMIN)

	if !ok {
		return nil, status.Error(codes.PermissionDenied, "Not enoguh Access Rights")
	}

	resp := connect.NewResponse(&pb.Invoices{Pool: []*pb.Invoice{tr}})

	return resp, nil
}
