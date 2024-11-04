package graph

import (
	"context"
	"github.com/arangodb/go-driver"
	"github.com/slntopp/nocloud-proto/access"
	"github.com/slntopp/nocloud/pkg/nocloud/schema"
	"go.uber.org/zap"
)

type CommonActionsController interface {
	HasAccess(ctx context.Context, account string, node driver.DocumentID, level access.Level) bool
	AccessLevel(ctx context.Context, account string, node driver.DocumentID) (bool, access.Level)
}

type Access struct {
	From  driver.DocumentID `json:"_from"`
	To    driver.DocumentID `json:"_to"`
	Level access.Level      `json:"level"`
	Role  string            `json:"role"`

	driver.DocumentMeta
}

type commonActionsController struct {
	log *zap.Logger
	db  driver.Database
}

func NewCommonActionsController(logger *zap.Logger, db driver.Database) CommonActionsController {
	log := logger.Named("CommonActionsController")
	return &commonActionsController{
		log: log, db: db,
	}
}

func (ctrl *commonActionsController) HasAccess(ctx context.Context, account string, node driver.DocumentID, level access.Level) bool {
	if (schema.ACCOUNTS_COL + "/" + account) == node.String() {
		return true
	}
	_, r := ctrl.AccessLevel(ctx, account, node)
	return r >= level
}

func (ctrl *commonActionsController) AccessLevel(ctx context.Context, account string, node driver.DocumentID) (bool, access.Level) {
	if driver.NewDocumentID(schema.ACCOUNTS_COL, account) == node {
		return true, access.Level_ROOT
	}
	query := `FOR path IN OUTBOUND K_SHORTEST_PATHS @account TO @node GRAPH @permissions RETURN path.edges[0].level`
	c, err := ctrl.db.Query(ctx, query, map[string]interface{}{
		"account":     schema.ACCOUNTS_COL + "/" + account,
		"node":        node,
		"permissions": schema.PERMISSIONS_GRAPH.Name,
	})
	if err != nil {
		return false, 0
	}
	defer c.Close()

	var accs access.Level = 0
	for {
		var level access.Level
		_, err := c.ReadDocument(ctx, &level)
		if driver.IsNoMoreDocuments(err) {
			break
		} else if err != nil {
			continue
		}
		if level > accs {
			accs = level
		}
	}
	return accs > 0, accs
}
