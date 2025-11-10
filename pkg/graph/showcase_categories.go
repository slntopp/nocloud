package graph

import (
	"context"
	"fmt"
	"github.com/arangodb/go-driver"
	sppb "github.com/slntopp/nocloud-proto/services_providers"
	"github.com/slntopp/nocloud/pkg/nocloud/schema"
	"go.uber.org/zap"
)

type CategoriesController interface {
	Create(ctx context.Context, cat *sppb.ShowcaseCategory) (*sppb.ShowcaseCategory, error)
	Update(ctx context.Context, cat *sppb.ShowcaseCategory) (*sppb.ShowcaseCategory, error)
	List(ctx context.Context, requester string, root bool, req *sppb.ListRequest) ([]*sppb.ShowcaseCategory, error)
	Get(ctx context.Context, uuid string) (*sppb.ShowcaseCategory, error)
	Delete(ctx context.Context, uuid string) error
}

type categoriesController struct {
	log *zap.Logger
	col driver.Collection

	db driver.Database
}

func NewCategoriesController(logger *zap.Logger, db driver.Database) CategoriesController {
	ctx := context.Background()
	log := logger.Named("ShowcaseCategoriesController")
	log.Debug("New Showcase Categories Controller Creating")

	col := GetEnsureCollection(log, ctx, db, schema.SHOWCASE_CATEGORIES_COL)

	return &categoriesController{
		log: log,
		col: col,
		db:  db,
	}
}

func (ctrl *categoriesController) Create(ctx context.Context, cat *sppb.ShowcaseCategory) (*sppb.ShowcaseCategory, error) {
	ctrl.log.Debug("Creating Document for Category", zap.Any("config", cat))

	meta, err := ctrl.col.CreateDocument(ctx, cat)
	if err != nil {
		return nil, err
	}

	cat.Uuid = meta.Key

	_, err = ctrl.col.UpdateDocument(ctx, cat.GetUuid(), cat)
	if err != nil {
		return nil, err
	}

	return cat, err
}

func (ctrl *categoriesController) Update(ctx context.Context, cat *sppb.ShowcaseCategory) (*sppb.ShowcaseCategory, error) {
	ctrl.log.Debug("Updating ShowcaseCategory", zap.Any("sp", cat))

	meta, err := ctrl.col.ReplaceDocument(ctx, cat.GetUuid(), cat)
	ctrl.log.Debug("ReplaceDocument.Result", zap.Any("meta", meta), zap.Error(err))
	return cat, err
}

const catListQuery = `
FOR s IN @@categories
	%s
	RETURN s
`

func (ctrl *categoriesController) List(ctx context.Context, requester string, root bool, req *sppb.ListRequest) ([]*sppb.ShowcaseCategory, error) {

	vars := map[string]interface{}{
		"@categories": schema.SHOWCASE_CATEGORIES_COL,
	}

	var query string
	var filters string

	if len(req.GetExcludeUuids()) > 0 {
		filters += ` FILTER s._key NOT IN @excludeUuids`
		vars["excludeUuids"] = req.GetExcludeUuids()
	}

	if req.GetFilters() != nil {
		for key, value := range req.GetFilters() {
			if key == "" {
			} else {
				values := value.GetListValue().AsSlice()
				if len(values) == 0 {
					continue
				}
				filters += fmt.Sprintf(` FILTER s["%s"] in @%s`, key, key)
				vars[key] = values
			}
		}
	}

	if requester == "" {
		filters += " FILTER s.public == @public"
		vars["public"] = true
	} else {
		if !root {
			filters += " FILTER s.public == @public"
			vars["public"] = true
		}
	}

	query = fmt.Sprintf(catListQuery, filters)
	c, err := ctrl.col.Database().Query(ctx, query, vars)
	if err != nil {
		return nil, err
	}

	defer c.Close()
	var r []*sppb.ShowcaseCategory
	for c.HasMore() {
		var s sppb.ShowcaseCategory

		_, err := c.ReadDocument(ctx, &s)

		if err != nil {
			return nil, err
		}

		r = append(r, &s)
	}

	return r, nil
}

func (ctrl *categoriesController) Get(ctx context.Context, uuid string) (*sppb.ShowcaseCategory, error) {
	var showcase sppb.ShowcaseCategory
	_, err := ctrl.col.ReadDocument(ctx, uuid, &showcase)
	return &showcase, err
}

func (ctrl *categoriesController) Delete(ctx context.Context, uuid string) error {
	ctrl.log.Debug("Deleting ShowcaseCategory", zap.Any("uuid", uuid))

	meta, err := ctrl.col.RemoveDocument(ctx, uuid)
	ctrl.log.Debug("RemoveDocument.Result", zap.Any("meta", meta), zap.Error(err))
	return err
}
