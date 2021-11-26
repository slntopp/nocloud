/*
Copyright © 2021 Nikita Ivanovski info@slnt-opp.xyz

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
	"errors"

	"github.com/arangodb/go-driver"
	"go.uber.org/zap"

	sppb "github.com/slntopp/nocloud/pkg/services_providers/proto"
)

const (
	SERVICES_PROVIDERS_COL = "ServicesProviders"
)

type ServicesProvider struct {
	*sppb.ServicesProvider
	driver.DocumentMeta
}

type ServicesProvidersController struct {
	col driver.Collection // Services Collection

	log *zap.Logger
}

func NewServicesProvidersController(log *zap.Logger, db driver.Database) ServicesProvidersController {
	col, _ := db.Collection(nil, SERVICES_PROVIDERS_COL)
	return ServicesProvidersController{log: log, col: col}
}

func (ctrl *ServicesProvidersController) Create(ctx context.Context, sp *ServicesProvider) (err error) {
	ctrl.log.Debug("Creating Document for ServicesProvider", zap.Any("config", sp))
	meta, err := ctrl.col.CreateDocument(ctx, *sp)
	sp.Uuid = meta.Key
	return err
}

func (ctrl *ServicesProvidersController) Get(ctx context.Context, id string) (r *ServicesProvider, err error) {
	ctrl.log.Debug("Getting ServicesProvider", zap.Any("sp", id))
	var sp sppb.ServicesProvider
	meta, err := ctrl.col.ReadDocument(ctx, id, &sp)
	if err != nil {
		ctrl.log.Debug("Error reading document(ServiceProvider)", zap.Error(err))
		return nil, errors.New("Error reading document")
	}
	ctrl.log.Debug("ReadDocument.Result", zap.Any("meta", meta), zap.Error(err), zap.Any("sp", &sp))
	sp.Uuid = meta.ID.Key()
	return &ServicesProvider{&sp, meta}, err
}

// List Services Providers in DB
func (ctrl *ServicesProvidersController) List(ctx context.Context, requestor string) ([]*ServicesProvider, error) {
	ctrl.log.Debug("Getting Services", zap.String("requestor", requestor))

	query := `FOR sp IN @@services RETURN sp`
	bindVars := map[string]interface{}{
		"@services": SERVICES_PROVIDERS_COL,
	}
	ctrl.log.Debug("Ready to build query", zap.Any("bindVars", bindVars))

	c, err := ctrl.col.Database().Query(ctx, query, bindVars)
	if err != nil {
		return nil, err
	}
	defer c.Close()
	var r []*ServicesProvider
	for {
		var s sppb.ServicesProvider
		meta, err := c.ReadDocument(ctx, &s)
		if driver.IsNoMoreDocuments(err) {
			break
		} else if err != nil {
			return nil, err
		}
		ctrl.log.Debug("Got document", zap.Any("service_provider", &s))
		s.Uuid = meta.ID.Key()
		r = append(r, &ServicesProvider{&s, meta})
	}

	return r,  nil
}