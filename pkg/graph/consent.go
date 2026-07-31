package graph

import (
	"context"

	"github.com/arangodb/go-driver"
	pb "github.com/slntopp/nocloud-proto/consent"
	"github.com/slntopp/nocloud/pkg/nocloud/schema"
	"go.uber.org/zap"
)

type ConsentsController interface {
	Create(ctx context.Context, record *pb.Record) (*pb.Record, error)
	List(ctx context.Context, req *pb.ListRequest) ([]*pb.Record, error)
}

type consentsController struct {
	log *zap.Logger
	col driver.Collection
}

func NewConsentsController(logger *zap.Logger, db driver.Database) ConsentsController {
	ctx := context.TODO()
	log := logger.Named("ConsentsController")
	col := GetEnsureCollection(log, ctx, db, schema.CONSENTS_COL)
	return &consentsController{
		log: log, col: col,
	}
}

func (c *consentsController) Create(ctx context.Context, record *pb.Record) (*pb.Record, error) {
	log := c.log.Named("Create")

	document, err := c.col.CreateDocument(ctx, record)
	if err != nil {
		log.Error("Failed to create document", zap.Error(err))
		return nil, err
	}

	record.Uuid = document.Key
	return record, nil
}

func (c *consentsController) List(ctx context.Context, req *pb.ListRequest) ([]*pb.Record, error) {
	log := c.log.Named("List")

	query := "LET recs = (FOR r in @@consents"
	vars := map[string]any{
		"@consents": schema.CONSENTS_COL,
	}

	if req.Account != nil {
		query += " FILTER r.account == @account"
		vars["account"] = req.GetAccount()
	}

	if req.Domain != nil {
		query += " FILTER r.domain == @domain"
		vars["domain"] = req.GetDomain()
	}

	query += " SORT r.created_at DESC"

	if req.Page != nil && req.Limit != nil && req.GetLimit() != 0 {
		limit, page := req.GetLimit(), req.GetPage()
		offset := (page - 1) * limit

		query += " LIMIT @offset, @count"
		vars["offset"] = offset
		vars["count"] = limit
	}

	query += " RETURN merge(r, {uuid: r._key})) RETURN recs"

	cur, err := c.col.Database().Query(ctx, query, vars)
	if err != nil {
		log.Error("Failed to get documents", zap.Error(err))
		return nil, err
	}

	var records []*pb.Record
	_, err = cur.ReadDocument(ctx, &records)
	if err != nil {
		log.Error("Failed to read documents", zap.Error(err))
		return nil, err
	}

	return records, nil
}
