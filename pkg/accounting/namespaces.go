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
package accounting

import (
	"context"

	"github.com/arangodb/go-driver"
	"github.com/slntopp/nocloud/pkg/accounting/namespacespb"
	"github.com/slntopp/nocloud/pkg/graph"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type NamespacesServiceServer struct {
	namespacespb.UnimplementedNamespacesServiceServer
	db driver.Database
	ctrl graph.NamespacesController
	acc_ctrl graph.AccountsController

	log *zap.Logger
}

func NewNamespacesServer(log *zap.Logger, db driver.Database) *NamespacesServiceServer {
	nsCol, _ := db.Collection(nil, graph.NAMESPACES_COL)
	accCol, _ := db.Collection(nil, graph.ACCOUNTS_COL)
	credCol, _ := db.Collection(nil, graph.CREDENTIALS_COL)

	return &NamespacesServiceServer{
		log: log, db: db,
		ctrl: graph.NewNamespacesController(
			log.Named("NamespacesController"), nsCol,
		),
		acc_ctrl: graph.NewAccountsController(
			log.Named("AccountsController"), accCol, credCol,
		),
	}
}

func (s *NamespacesServiceServer) Create(ctx context.Context, request *namespacespb.CreateRequest) (*namespacespb.CreateResponse, error) {
	log := s.log.Named("CreateNamespace")
	log.Debug("Request received", zap.Any("request", request), zap.Any("context", ctx))
	ctx, err := ValidateMetadata(ctx, log)
	if err != nil {
		return nil, err
	}

	ns, err := s.ctrl.Create(ctx, request.Title)
	if err != nil {
		s.log.Debug("Error creating namespace", zap.Error(err))
		return  &namespacespb.CreateResponse{}, status.Error(codes.Internal, "Can't create Namespace")
	}

	return &namespacespb.CreateResponse{ Id: ns.Key }, nil
}