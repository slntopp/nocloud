package graph

import (
	"context"

	"github.com/arangodb/go-driver"
	"github.com/slntopp/nocloud/pkg/nocloud"
	"github.com/slntopp/nocloud/pkg/nocloud/access"
	"go.uber.org/zap"
)

const (
	NAMESPACES_COL = "Namespaces"
	NS2ACC = NAMESPACES_COL + "2" + ACCOUNTS_COL
)

type Namespace struct {
	Title string `json:"title"`
	driver.DocumentMeta
}

type NamespacesController struct {
	col driver.Collection
	log *zap.Logger
}

func NewNamespacesController(log *zap.Logger, col driver.Collection) NamespacesController {
	return NamespacesController{log: log, col: col}
}

func (ctrl *NamespacesController) Get(ctx context.Context, id string) (Namespace, error) {
	var r Namespace
	_, err := ctrl.col.ReadDocument(nil, id, &r)
	return r, err
}

func (ctrl *NamespacesController) Create(ctx context.Context, title string) (Namespace, error) {
	ns := Namespace{
		Title: title,
	}
	meta, err := ctrl.col.CreateDocument(ctx, ns)
	if err != nil {
		return Namespace{}, err
	}
	
	ns.DocumentMeta = meta

	acc := Account{
		DocumentMeta: driver.DocumentMeta {
			ID: driver.NewDocumentID(ACCOUNTS_COL, ctx.Value(nocloud.NoCloudAccount).(string)),
		},
	}

	edge, _ := ctrl.col.Database().Collection(ctx, ACC2NS)
	err = acc.LinkNamespace(ctx, edge, ns, access.ADMIN)
	return ns, err
}