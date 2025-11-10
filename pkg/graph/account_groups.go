package graph

import (
	"context"
	"fmt"
	"github.com/arangodb/go-driver"
	accpb "github.com/slntopp/nocloud-proto/registry/accounts"
	"github.com/slntopp/nocloud/pkg/nocloud/schema"
	"go.uber.org/zap"
)

type AccountGroupsController interface {
	Create(ctx context.Context, cat *accpb.AccountGroup) (*accpb.AccountGroup, error)
	Update(ctx context.Context, cat *accpb.AccountGroup) (*accpb.AccountGroup, error)
	List(ctx context.Context, req *accpb.ListRequest) ([]*accpb.AccountGroup, error)
	Get(ctx context.Context, uuid string) (*accpb.AccountGroup, error)
	Delete(ctx context.Context, uuid string) error
}

type accountGroupsController struct {
	log *zap.Logger
	col driver.Collection

	db driver.Database
}

func NewAccountGroupsController(logger *zap.Logger, db driver.Database) AccountGroupsController {
	ctx := context.Background()
	log := logger.Named("AccountGroupsController")
	log.Debug("New Account Groups Controller Creating")

	col := GetEnsureCollection(log, ctx, db, schema.ACCOUNT_GROUPS_COL)

	return &accountGroupsController{
		log: log,
		col: col,
		db:  db,
	}
}

func (ctrl *accountGroupsController) Create(ctx context.Context, cat *accpb.AccountGroup) (*accpb.AccountGroup, error) {
	ctrl.log.Debug("Creating Document for Account Group", zap.Any("config", cat))

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

func (ctrl *accountGroupsController) Update(ctx context.Context, cat *accpb.AccountGroup) (*accpb.AccountGroup, error) {
	ctrl.log.Debug("Updating AccountGroup", zap.Any("sp", cat))

	meta, err := ctrl.col.ReplaceDocument(ctx, cat.GetUuid(), cat)
	ctrl.log.Debug("ReplaceDocument.Result", zap.Any("meta", meta), zap.Error(err))
	return cat, err
}

const groupListQuery = `
FOR s IN @@groups
	%s
	RETURN s
`

func (ctrl *accountGroupsController) List(ctx context.Context, req *accpb.ListRequest) ([]*accpb.AccountGroup, error) {

	vars := map[string]interface{}{
		"@groups": schema.ACCOUNT_GROUPS_COL,
	}

	var query string
	var filters string

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

	query = fmt.Sprintf(groupListQuery, filters)
	c, err := ctrl.col.Database().Query(ctx, query, vars)
	if err != nil {
		return nil, err
	}

	defer c.Close()
	var r []*accpb.AccountGroup
	for c.HasMore() {
		var s accpb.AccountGroup

		_, err := c.ReadDocument(ctx, &s)

		if err != nil {
			return nil, err
		}

		r = append(r, &s)
	}

	return r, nil
}

func (ctrl *accountGroupsController) Get(ctx context.Context, uuid string) (*accpb.AccountGroup, error) {
	var gr accpb.AccountGroup
	_, err := ctrl.col.ReadDocument(ctx, uuid, &gr)
	return &gr, err
}

func (ctrl *accountGroupsController) Delete(ctx context.Context, uuid string) error {
	ctrl.log.Debug("Deleting AccountGroup", zap.Any("uuid", uuid))

	meta, err := ctrl.col.RemoveDocument(ctx, uuid)
	ctrl.log.Debug("RemoveDocument.Result", zap.Any("meta", meta), zap.Error(err))
	return err
}
