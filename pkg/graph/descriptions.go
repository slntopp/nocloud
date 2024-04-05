package graph

import (
	"context"
	"fmt"
	"github.com/arangodb/go-driver"
	pb "github.com/slntopp/nocloud-proto/billing/descriptions"
	"github.com/slntopp/nocloud/pkg/nocloud/schema"
	"go.uber.org/zap"
)

type DescriptionsController struct {
	log *zap.Logger
	col driver.Collection
}

func NewDescriptionsController(logger *zap.Logger, db driver.Database) *DescriptionsController {
	ctx := context.TODO()
	log := logger.Named("DescriptionsController")
	descriptions := GetEnsureCollection(log, ctx, db, schema.DESCRIPTIONS_COL)
	return &DescriptionsController{
		log: log, col: descriptions,
	}
}

func (c *DescriptionsController) Create(ctx context.Context, description *pb.Description) (*pb.Description, error) {
	log := c.log.Named("Create")

	document, err := c.col.CreateDocument(ctx, description)
	if err != nil {
		log.Error("Failed to create document", zap.Error(err))
		return nil, err
	}

	description.Uuid = document.Key
	return description, nil
}

func (c *DescriptionsController) Update(ctx context.Context, description *pb.Description) (*pb.Description, error) {
	log := c.log.Named("Update")

	_, err := c.col.ReplaceDocument(ctx, description.GetUuid(), description)
	if err != nil {
		log.Error("Failed to update document", zap.Error(err))
		return nil, err
	}

	return description, nil
}

func (c *DescriptionsController) Delete(ctx context.Context, uuid string) error {
	log := c.log.Named("Update")

	_, err := c.col.RemoveDocument(ctx, uuid)
	if err != nil {
		log.Error("Failed to update document", zap.Error(err))
		return err
	}

	return nil
}

func (c *DescriptionsController) Get(ctx context.Context, uuid string) (*pb.Description, error) {
	log := c.log.Named("Get")

	var description pb.Description

	meta, err := c.col.ReadDocument(ctx, uuid, &description)
	if err != nil {
		log.Error("Failed to get document", zap.Error(err))
		return nil, err
	}

	description.Uuid = meta.Key

	return &description, nil
}

func (c *DescriptionsController) List(ctx context.Context, req *pb.ListDescriptionsRequest) ([]*pb.Description, error) {
	log := c.log.Named("Get")

	query := "LET descs = (FOR d in @@descriptions "
	vars := map[string]any{
		"@descriptions": schema.DESCRIPTIONS_COL,
	}

	if req.Field != nil && req.Sort != nil {
		subQuery := ` SORT a.%s %s`
		field, sort := req.GetField(), req.GetSort()

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

	query += " RETURN merge(d, {uuid: d._key})) RETURN descs"

	log.Debug("Query", zap.String("q", query))

	cur, err := c.col.Database().Query(ctx, query, vars)
	if err != nil {
		log.Error("Failed to get documents", zap.Error(err))
		return nil, err
	}

	var descriptions []*pb.Description
	_, err = cur.ReadDocument(ctx, &descriptions)
	if err != nil {
		log.Error("Failed to read documents", zap.Error(err))
		return nil, err
	}

	return descriptions, nil
}

func (c *DescriptionsController) Count(ctx context.Context) ([]*pb.Description, error) {
	log := c.log.Named("Get")

	query := "LET descs = (FOR d in @@descriptions RETURN merge(d, {uuid: d._key})) RETURN descs"
	vars := map[string]any{
		"@descriptions": schema.DESCRIPTIONS_COL,
	}

	cur, err := c.col.Database().Query(ctx, query, vars)
	if err != nil {
		log.Error("Failed to get documents", zap.Error(err))
		return nil, err
	}

	var descriptions []*pb.Description
	_, err = cur.ReadDocument(ctx, &descriptions)
	if err != nil {
		log.Error("Failed to read documents", zap.Error(err))
		return nil, err
	}

	return descriptions, nil
}
