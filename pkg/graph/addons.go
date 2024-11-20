package graph

import (
	"context"
	"fmt"

	"github.com/arangodb/go-driver"
	pb "github.com/slntopp/nocloud-proto/billing/addons"
	"github.com/slntopp/nocloud/pkg/nocloud/schema"
	"go.uber.org/zap"
)

type AddonsController interface {
	Create(ctx context.Context, addon *pb.Addon) (*pb.Addon, error)
	Update(ctx context.Context, addon *pb.Addon) (*pb.Addon, error)
	CreateBulk(ctx context.Context, addons []*pb.Addon) ([]*pb.Addon, error)
	UpdateBulk(ctx context.Context, addons []*pb.Addon) ([]*pb.Addon, error)
	Delete(ctx context.Context, uuid string) error
	Get(ctx context.Context, uuid string) (*pb.Addon, error)
	List(ctx context.Context, req *pb.ListAddonsRequest) ([]*pb.Addon, error)
	Count(ctx context.Context, req *pb.CountAddonsRequest) ([]*pb.Addon, error)
	GetUnique(ctx context.Context) (map[string]any, error)
}

type addonsController struct {
	log *zap.Logger
	col driver.Collection
}

func NewAddonsController(logger *zap.Logger, db driver.Database) AddonsController {
	ctx := context.TODO()
	log := logger.Named("AddonsController")
	addons := GetEnsureCollection(log, ctx, db, schema.ADDONS_COL)
	return &addonsController{
		log: log, col: addons,
	}
}

func (c *addonsController) Create(ctx context.Context, addon *pb.Addon) (*pb.Addon, error) {
	log := c.log.Named("Create")

	document, err := c.col.CreateDocument(ctx, addon)
	if err != nil {
		log.Error("Failed to create document", zap.Error(err))
		return nil, err
	}

	addon.Uuid = document.Key
	return addon, nil
}

func (c *addonsController) Update(ctx context.Context, addon *pb.Addon) (*pb.Addon, error) {
	log := c.log.Named("Update")

	_, err := c.col.ReplaceDocument(ctx, addon.GetUuid(), addon)
	if err != nil {
		log.Error("Failed to update document", zap.Error(err))
		return nil, err
	}

	return addon, nil
}

func (c *addonsController) CreateBulk(ctx context.Context, addons []*pb.Addon) ([]*pb.Addon, error) {
	log := c.log.Named("Create")

	meta_slice, _, err := c.col.CreateDocuments(ctx, addons)
	if err != nil {
		log.Error("Failed to create document", zap.Error(err))
		return nil, err
	}

	for i, item := range meta_slice.Keys() {
		addons[i].Uuid = item
	}

	return addons, nil
}

func (c *addonsController) UpdateBulk(ctx context.Context, addons []*pb.Addon) ([]*pb.Addon, error) {
	log := c.log.Named("Update")

	var keys = make([]string, len(addons))

	for i, addon := range addons {
		keys[i] = addon.GetUuid()
	}

	_, _, err := c.col.ReplaceDocuments(ctx, keys, addons)
	if err != nil {
		log.Error("Failed to update document", zap.Error(err))
		return nil, err
	}

	return addons, nil
}

func (c *addonsController) Delete(ctx context.Context, uuid string) error {
	log := c.log.Named("Update")

	_, err := c.col.RemoveDocument(ctx, uuid)
	if err != nil {
		log.Error("Failed to update document", zap.Error(err))
		return err
	}

	return nil
}

func (c *addonsController) Get(ctx context.Context, uuid string) (*pb.Addon, error) {
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

func (c *addonsController) List(ctx context.Context, req *pb.ListAddonsRequest) ([]*pb.Addon, error) {
	log := c.log.Named("Get")

	query := "LET adds = (FOR a in @@addons "
	vars := map[string]any{
		"@addons": schema.ADDONS_COL,
	}

	if req.GetFilters() != nil {
		for key, value := range req.GetFilters() {
			if key == "title" {
				query += fmt.Sprintf(` FILTER a.title LIKE "%s"`, "%"+value.GetStringValue()+"%")
			} else if key == "system" {
				// TODO: make it easier to read
				query += fmt.Sprintf(` FILTER (%t && a.system == true) || (!%t && (!a.system || a.system == false))`, value.GetBoolValue(), value.GetBoolValue())
			} else if key == "search_param" {
				query += fmt.Sprintf(` FILTER LOWER(a.title) LIKE LOWER("%s") || LOWER(a.group) LIKE LOWER("%s") || a._key LIKE "%s"`, "%"+value.GetStringValue()+"%", "%"+value.GetStringValue()+"%", "%"+value.GetStringValue()+"%")
			} else {
				values := value.GetListValue().AsSlice()
				if len(values) == 0 {
					continue
				}
				query += fmt.Sprintf(` FILTER a["%s"] in @%s`, key, key)
				vars[key] = values
			}
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

func (c *addonsController) Count(ctx context.Context, req *pb.CountAddonsRequest) ([]*pb.Addon, error) {
	log := c.log.Named("Get")

	query := "LET adds = (FOR a in @@addons "
	vars := map[string]any{
		"@addons": schema.ADDONS_COL,
	}

	if req.GetFilters() != nil {
		for key, value := range req.GetFilters() {
			if key == "title" {
				query += fmt.Sprintf(` FILTER a.title LIKE "%s"`, "%"+value.GetStringValue()+"%")
			} else if key == "system" {
				// TODO: make it easier to read
				query += fmt.Sprintf(` FILTER (%t && a.system == true) || (!%t && (!a.system || a.system == false))`, value.GetBoolValue(), value.GetBoolValue())
			} else if key == "search_param" {
				query += fmt.Sprintf(` FILTER LOWER(a.title) LIKE LOWER("%s") || LOWER(a.group) LIKE LOWER("%s") || a._key LIKE "%s"`, "%"+value.GetStringValue()+"%", "%"+value.GetStringValue()+"%", "%"+value.GetStringValue()+"%")
			} else if key == "plan_uuid" {
				values := value.GetListValue().AsSlice()
				if len(values) == 0 {
					continue
				}
				query += fmt.Sprintf(` 
FILTER LENGTH(
  FOR p IN @@plans
	FILTER p.uuid IN @planUuids
    LET products = p.products
    LET presentInProducts = LENGTH(
        FILTER IS_OBJECT(products)
        FOR attr IN ATTRIBUTES(products)
            FILTER a._key IN products[attr].addons
            RETURN true
    ) > 0
    FILTER a._key IN p.addons || presentInProducts
	RETURN true
) > 0
`)
				vars["planUuids"] = values
				vars["@plans"] = schema.BILLING_PLANS_COL
			} else {
				values := value.GetListValue().AsSlice()
				if len(values) == 0 {
					continue
				}
				query += fmt.Sprintf(` FILTER a["%s"] in @%s`, key, key)
				vars[key] = values
			}
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

const uniqueAddonQuery = `
LET groups = (
	FOR a in @@addons
		RETURN a.group
)

RETURN {
	groups: UNIQUE(groups)
}
`

func (c *addonsController) GetUnique(ctx context.Context) (map[string]any, error) {
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
