/*
Copyright © 2021-2023 Nikita Ivanovski info@slnt-opp.xyz

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
package graph

import (
	"context"

	"github.com/arangodb/go-driver"
	"github.com/slntopp/nocloud-proto/access"
	"github.com/slntopp/nocloud-proto/registry/namespaces"
	"github.com/slntopp/nocloud/pkg/nocloud"
	"github.com/slntopp/nocloud/pkg/nocloud/roles"
	"github.com/slntopp/nocloud/pkg/nocloud/schema"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"
)

type NamespacesController interface {
	Get(ctx context.Context, id string) (Namespace, error)
	List(ctx context.Context, requestor Account, req_depth int32, offset, limit uint64, field, sort string, filters map[string]*structpb.Value) (*ListQueryResult[*Namespace], error)
	Create(ctx context.Context, title string) (Namespace, error)
	Patch(ctx context.Context, uuid, title string) error
	Link(ctx context.Context, acc Account, ns Namespace, access access.Level, role string) error
	Join(ctx context.Context, acc Account, ns Namespace, access access.Level, role string) error
	Delete(ctx context.Context, id string) error
}

type Namespace struct {
	*namespaces.Namespace
	driver.DocumentMeta
}

type namespacesController struct {
	col driver.Collection
	log *zap.Logger
}

func NewNamespacesController(logger *zap.Logger, db driver.Database) NamespacesController {
	ctx := context.TODO()
	log := logger.Named("NamespacesController")

	log.Debug("New Namespaces Controller Creating")

	graph := GraphGetEnsure(log, ctx, db, schema.PERMISSIONS_GRAPH.Name)
	col := GraphGetVertexEnsure(log, ctx, db, graph, schema.NAMESPACES_COL)

	GraphGetEdgeEnsure(log, ctx, graph, schema.NS2ACC, schema.NAMESPACES_COL, schema.ACCOUNTS_COL)
	GraphGetEdgeEnsure(log, ctx, graph, schema.ACC2NS, schema.ACCOUNTS_COL, schema.NAMESPACES_COL)

	return &namespacesController{log: log, col: col}
}

func (ctrl *namespacesController) Get(ctx context.Context, id string) (Namespace, error) {
	return GetWithAccess[Namespace](ctx, ctrl.col.Database(), driver.NewDocumentID(schema.NAMESPACES_COL, id))
}

func (ctrl *namespacesController) List(ctx context.Context, requestor Account, req_depth int32, offset, limit uint64, field, sort string, filters map[string]*structpb.Value) (*ListQueryResult[*Namespace], error) {
	return ListNamespaces[*Namespace](ctx, ctrl.log, ctrl.col.Database(), requestor.ID, schema.NAMESPACES_COL, req_depth, offset, limit, field, sort, filters)
}

func (ctrl *namespacesController) Create(ctx context.Context, title string) (Namespace, error) {
	ns := Namespace{
		Namespace: &namespaces.Namespace{Title: title},
	}

	meta, err := ctrl.col.CreateDocument(ctx, ns)
	if err != nil {
		return Namespace{}, err
	}

	ns.DocumentMeta = meta
	key := ctx.Value(nocloud.NoCloudAccount).(string)
	acc := Account{
		DocumentMeta: driver.DocumentMeta{
			Key: key,
			ID:  driver.NewDocumentID(schema.ACCOUNTS_COL, key),
		},
	}

	return ns, ctrl.Link(ctx, acc, ns, access.Level_ADMIN, roles.OWNER)
}

func (ctrl *namespacesController) Patch(ctx context.Context, uuid, title string) error {
	ns := Namespace{
		Namespace: &namespaces.Namespace{Title: title},
	}

	_, err := ctrl.col.UpdateDocument(ctx, uuid, ns)
	if err != nil {
		return err
	}

	return nil
}

func (ctrl *namespacesController) Link(ctx context.Context, acc Account, ns Namespace, access access.Level, role string) error {
	edge, _ := ctrl.col.Database().Collection(ctx, schema.ACC2NS)
	return acc.LinkNamespace(ctx, edge, ns, access, role)
}

func (ctrl *namespacesController) Join(ctx context.Context, acc Account, ns Namespace, access access.Level, role string) error {
	edge, _ := ctrl.col.Database().Collection(ctx, schema.NS2ACC)
	return acc.JoinNamespace(ctx, edge, ns, access, role)
}

func (ns *Namespace) Delete(ctx context.Context, db driver.Database) error {
	err := DeleteRecursive(ctx, db, ns.ID)
	if err != nil {
		return err
	}

	return nil
}

func (ctrl *namespacesController) Delete(ctx context.Context, id string) error {
	ns, err := ctrl.Get(ctx, id)
	if err != nil {
		return err
	}
	return ns.Delete(ctx, ctrl.col.Database())
}
