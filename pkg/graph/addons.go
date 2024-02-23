package graph

import (
	"context"
	"fmt"
	"github.com/arangodb/go-driver"
	pb "github.com/slntopp/nocloud-proto/billing/addons"
	"github.com/slntopp/nocloud/pkg/nocloud/schema"
	"go.uber.org/zap"
)

type AddonsController struct {
	log *zap.Logger
	col driver.Collection
}

func NewAddonsController(logger *zap.Logger, db driver.Database) *AddonsController {
	ctx := context.TODO()
	log := logger.Named("AddonsController")
	addons := GetEnsureCollection(log, ctx, db, schema.ADDONS_COL)
	return &AddonsController{
		log: log, col: addons,
	}
}

func (c *AddonsController) Create(ctx context.Context, addon *pb.Addon) (*pb.Addon, error) {
	log := c.log.Named("Create")

	document, err := c.col.CreateDocument(ctx, addon)
	if err != nil {
		log.Error("Failed to create document", zap.Error(err))
		return nil, err
	}

	addon.Uuid = document.Key
	return addon, nil
}

func (c *AddonsController) Update(ctx context.Context, addon *pb.Addon) (*pb.Addon, error) {
	log := c.log.Named("Update")

	_, err := c.col.ReplaceDocument(ctx, addon.GetUuid(), addon)
	if err != nil {
		log.Error("Failed to update document", zap.Error(err))
		return nil, err
	}

	return addon, nil
}

func (c *AddonsController) Delete(ctx context.Context, uuid string) error {
	log := c.log.Named("Update")

	_, err := c.col.RemoveDocument(ctx, uuid)
	if err != nil {
		log.Error("Failed to update document", zap.Error(err))
		return err
	}

	return nil
}

func (c *AddonsController) Get(ctx context.Context, uuid string) (*pb.Addon, error) {
	log := c.log.Named("Get")

	var addon pb.Addon

	meta, err := c.col.ReadDocument(ctx, uuid, &addon)
	if err != nil {
		log.Error("Failed to get document", zap.Error(err))
		return nil, err
	}

	addon.Uuid = meta.Key

	return &addon, nil
}

func (c *AddonsController) List(ctx context.Context, req *pb.ListAddonsRequest) ([]*pb.Addon, error) {
	log := c.log.Named("Get")

	query := "LET adds = (FOR a in @@addons "
	vars := map[string]any{
		"@addons": schema.ADDONS_COL,
	}

	if req.GetFilters() != nil {
		for key, value := range req.GetFilters() {
			values := value.GetListValue().AsSlice()
			if len(values) == 0 {
				continue
			}
			query += fmt.Sprintf(` FILTER a["%s"] in @%s`, key, key)
			vars[key] = values
		}
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

	query += " RETURN merge(a, {uuid: a._key})) RETURN adds"

	log.Debug("Query", zap.String("q", query))

	cur, err := c.col.Database().Query(ctx, query, vars)
	if err != nil {
		log.Error("Failed to get documents", zap.Error(err))
		return nil, err
	}

	var addons []*pb.Addon
	_, err = cur.ReadDocument(ctx, &addons)
	if err != nil {
		log.Error("Failed to read documents", zap.Error(err))
		return nil, err
	}

	return addons, nil
}

func (c *AddonsController) Count(ctx context.Context, req *pb.CountAddonsRequest) ([]*pb.Addon, error) {
	log := c.log.Named("Get")

	query := "LET adds = (FOR a in @@addons "
	vars := map[string]any{
		"@addons": schema.ADDONS_COL,
	}

	if req.GetFilters() != nil {
		for key, value := range req.GetFilters() {
			values := value.GetListValue().AsSlice()
			if len(values) == 0 {
				continue
			}
			query += fmt.Sprintf(` FILTER a["%s"] in @%s`, key, key)
			vars[key] = values
		}
	}

	log.Debug("Query", zap.String("q", query))

	query += " RETURN merge(a, {uuid: a._key})) RETURN adds"

	cur, err := c.col.Database().Query(ctx, query, vars)
	if err != nil {
		log.Error("Failed to get documents", zap.Error(err))
		return nil, err
	}

	var addons []*pb.Addon
	_, err = cur.ReadDocument(ctx, &addons)
	if err != nil {
		log.Error("Failed to read documents", zap.Error(err))
		return nil, err
	}

	return addons, nil
}

const uniqueAddonQuery = `
LET groups = (
	FOR a in @@addons
		RETURN a.group
)

RETURN {
	groups: UNIQUE(groups)
}
`

func (c *AddonsController) GetUnique(ctx context.Context) (map[string]any, error) {
	cur, err := c.col.Database().Query(ctx, uniqueAddonQuery, map[string]interface{}{
		"@addons": schema.ADDONS_COL,
	})
	if err != nil {
		return nil, err
	}

	var result = map[string]any{}

	_, err = cur.ReadDocument(ctx, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}
