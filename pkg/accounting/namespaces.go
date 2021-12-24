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
	"github.com/slntopp/nocloud/pkg/nocloud/access"
	"github.com/slntopp/nocloud/pkg/nocloud/roles"
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
	return &NamespacesServiceServer{
		log: log, db: db,
		ctrl: graph.NewNamespacesController(
			log.Named("NamespacesController"), db,
		),
		acc_ctrl: graph.NewAccountsController(
			log.Named("AccountsController"), db,
		),
	}
}

func (s *NamespacesServiceServer) Create(ctx context.Context, request *namespacespb.CreateRequest) (*namespacespb.CreateResponse, error) {
	log := s.log.Named("CreateNamespace")
	log.Debug("Request received", zap.Any("request", request), zap.Any("context", ctx))

	ns, err := s.ctrl.Create(ctx, request.Title)
	if err != nil {
		s.log.Debug("Error creating namespace", zap.Error(err))
		return  &namespacespb.CreateResponse{}, status.Error(codes.Internal, "Can't create Namespace")
	}

	return &namespacespb.CreateResponse{ Id: ns.Key }, nil
}

func (s *NamespacesServiceServer) List(ctx context.Context, request *namespacespb.ListRequest) (*namespacespb.ListResponse, error) {
	log := s.log.Named("ListNamespaces")
	log.Debug("List request received", zap.Any("request", request), zap.Any("context", ctx))

	requestor := ctx.Value(nocloud.NoCloudAccount).(string)
	log.Debug("Requestor", zap.String("id", requestor))

	acc, err := s.acc_ctrl.Get(ctx, requestor)
	if err != nil {
		s.log.Debug("Error getting account", zap.Any("error", err))
		return nil, status.Error(codes.PermissionDenied, "Requestor Account not found")
	}
	log.Debug("Requestor", zap.Any("account", acc))

	var pool []graph.Namespace
	pool, err = s.ctrl.List(ctx, acc, request.Depth)
	if err != nil {
		s.log.Debug("Error listing namespaces", zap.Any("error", err))
		return nil, status.Error(codes.Internal, "Error listing namespaces")
	}
	log.Debug("List result", zap.Any("pool", pool))

	result := make([]*namespacespb.Namespace, len(pool))
	for i, ns := range pool {
		result[i] = &namespacespb.Namespace{Id: ns.Key, Title:  ns.Title }
	}
	log.Debug("Convert result", zap.Any("pool", result))

	return &namespacespb.ListResponse{Pool: result}, nil
}

func (s *NamespacesServiceServer) Join(ctx context.Context, request *namespacespb.JoinRequest) (*namespacespb.JoinResponse, error) {
	log := s.log.Named("JoinNamespace")
	log.Debug("Request received", zap.Any("request", request), zap.Any("context", ctx))

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
	var level int32
	ok, level = graph.AccessLevel(ctx, s.db, ctx.Value(nocloud.NoCloudAccount).(string), acc.ID.String())
	if !ok || level < access.ADMIN {
		return nil, status.Error(codes.PermissionDenied, "Not enough access rights to Account")
	}

	ok, level = graph.AccessLevel(ctx, s.db, ctx.Value(nocloud.NoCloudAccount).(string), ns.ID.String())
	if !ok || level < access.ADMIN {
		return nil, status.Error(codes.PermissionDenied, "Not enough access rights to Namespace")
	}

	if request.Access == nil {
		level = level - 1
	} else {
		level = *request.Access
	}

	err = s.ctrl.Join(ctx, acc, ns, level, roles.DEFAULT)
	if err != nil {
		s.log.Debug("Error while joining account", zap.Error(err))
		return nil, status.Error(codes.Internal, "Error while joining account")
	}
	return &namespacespb.JoinResponse{Result: true}, nil
}

func (s *NamespacesServiceServer) Link(ctx context.Context, request *namespacespb.LinkRequest) (*namespacespb.LinkResponse, error) {
	log := s.log.Named("LinkNamespace")
	log.Debug("Request received", zap.Any("request", request), zap.Any("context", ctx))

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

	ok, level := graph.AccessLevel(ctx, s.db, ctx.Value(nocloud.NoCloudAccount).(string), ns.ID.String())
	if !ok || level < access.ADMIN {
		return nil, status.Error(codes.PermissionDenied, "Not enough access rights to Namespace")
	}

	if request.Access == nil {
		level = level - 1
	} else {
		level = *request.Access
	}

	err = s.ctrl.Link(ctx, acc, ns, level, roles.DEFAULT)
	if err != nil {
		s.log.Debug("Error while linking account", zap.Error(err))
		return nil, status.Error(codes.Internal, "Error while linking account to namespace")
	}
	return &namespacespb.LinkResponse{Result: true}, nil
}

func (s *NamespacesServiceServer) Delete(ctx context.Context, request *namespacespb.DeleteRequest) (*namespacespb.DeleteResponse, error) {
	log := s.log.Named("Delete")
	log.Debug("Request received", zap.Any("request", request), zap.Any("context", ctx))

	requestor := ctx.Value(nocloud.NoCloudAccount).(string)
	log.Debug("Requestor", zap.String("id", requestor))

	ns, err := s.ctrl.Get(ctx, request.Id)
	if err != nil {
		s.log.Debug("Error getting account", zap.Any("error", err))
		return nil, status.Error(codes.NotFound, "Account not found")
	}

	if !graph.HasAccess(ctx, s.db, requestor, ns.ID.String(), access.ADMIN) {
		return nil, status.Error(codes.PermissionDenied, "NoAccess")
	}

	err = ns.Delete(ctx, s.db)
	if err != nil {
		s.log.Debug("Error deleting account and it's children", zap.Error(err))
		return nil, status.Error(codes.Internal, "Error deleting account")
	}

	return &namespacespb.DeleteResponse{Result: true}, nil
}