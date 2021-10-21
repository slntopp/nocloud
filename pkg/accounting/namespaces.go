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
	"github.com/slntopp/nocloud/pkg/nocloud"
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

func (s *NamespacesServiceServer) Join(ctx context.Context, request *namespacespb.JoinRequest) (*namespacespb.JoinResponse, error) {
	log := s.log.Named("JoinNamespace")
	log.Debug("Request received", zap.Any("request", request), zap.Any("context", ctx))

	ctx, err := ValidateMetadata(ctx, log)
	if err != nil {
		return nil, err
	}

	acc, err := s.acc_ctrl.Get(ctx, request.Account)
	if err != nil {
		s.log.Debug("Error getting account", zap.Any("error", err))
		return nil, status.Error(codes.NotFound, "Account not found")
	}
	ns, err := s.ctrl.Get(ctx, request.Namespace)
	if err != nil {
		s.log.Debug("Error getting namespace", zap.Any("error", err))
		return nil, status.Error(codes.NotFound, "Namespace not found")
	}

	var ok bool
	ok = graph.HasAccess(ctx, s.db, ctx.Value(nocloud.NoCloudAccount).(string), ns.ID.String(), 3)
	if !ok {
		return nil, status.Error(codes.PermissionDenied, "Not enough access rights to Namespace")
	}

	ok = graph.HasAccess(ctx, s.db, ctx.Value(nocloud.NoCloudAccount).(string), acc.ID.String(), 3)
	if !ok {
		return nil, status.Error(codes.PermissionDenied, "Not enough access rights to Account")
	}

	err = s.ctrl.Join(ctx, acc, ns, *request.Access)
	if err != nil {
		s.log.Debug("Error while joining account", zap.Error(err))
		return nil, status.Error(codes.Internal, "Error while joining account")
	}
	return &namespacespb.JoinResponse{Result: true}, nil
}

func (s *NamespacesServiceServer) Link(ctx context.Context, request *namespacespb.LinkRequest) (*namespacespb.LinkResponse, error) {
	log := s.log.Named("LinkNamespace")
	log.Debug("Request received", zap.Any("request", request), zap.Any("context", ctx))

	ctx, err := ValidateMetadata(ctx, log)
	if err != nil {
		return nil, err
	}

	acc, err := s.acc_ctrl.Get(ctx, request.Account)
	if err != nil {
		s.log.Debug("Error getting account", zap.Any("error", err))
		return nil, status.Error(codes.NotFound, "Account not found")
	}
	ns, err := s.ctrl.Get(ctx, request.Namespace)
	if err != nil {
		s.log.Debug("Error getting namespace", zap.Any("error", err))
		return nil, status.Error(codes.NotFound, "Namespace not found")
	}

	ok := graph.HasAccess(ctx, s.db, ctx.Value(nocloud.NoCloudAccount).(string), ns.ID.String(), 3)
	if !ok {
		return nil, status.Error(codes.PermissionDenied, "Not enough access rights to Namespace")
	}

	err = s.ctrl.Link(ctx, acc, ns, *request.Access)
	if err != nil {
		s.log.Debug("Error while linking account", zap.Error(err))
		return nil, status.Error(codes.Internal, "Error while linking account to namespace")
	}
	return &namespacespb.LinkResponse{Result: true}, nil
}