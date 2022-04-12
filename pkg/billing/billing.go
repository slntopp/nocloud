/*
Copyright © 2021-2022 Nikita Ivanovski info@slnt-opp.xyz

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
package billing

import (
	"context"

	"github.com/arangodb/go-driver"
	pb "github.com/slntopp/nocloud/pkg/billing/proto"
	"github.com/slntopp/nocloud/pkg/graph"
	healthpb "github.com/slntopp/nocloud/pkg/health/proto"
	"github.com/slntopp/nocloud/pkg/nocloud"
	"github.com/slntopp/nocloud/pkg/nocloud/access"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Routine struct {
	Name string
	LastExec string
	Running bool
}

type BillingServiceServer struct {
	pb.UnimplementedBillingServiceServer

	log *zap.Logger

	nss graph.NamespacesController
	plans graph.BillingPlansController
	transactions graph.TransactionsController
	records graph.RecordsController

	db driver.Database

	gen *healthpb.RoutineStatus
	proc *healthpb.RoutineStatus
}

func NewBillingServiceServer(logger *zap.Logger, db driver.Database) *BillingServiceServer {
	log := logger.Named("BillingService")
	return &BillingServiceServer{
		log: log,
		nss: graph.NewNamespacesController(log, db),
		plans: graph.NewBillingPlansController(log.Named("PlansController"), db),
		transactions: graph.NewTransactionsController(log.Named("TransactionsController"), db),
		records: graph.NewRecordsController(log.Named("RecordsController"), db),
		db: db, gen: &healthpb.RoutineStatus {
			Routine: "Generate Transactions",
			Status: &healthpb.ServingStatus{
				Service: "Billing Machine",
				Status: healthpb.Status_STOPPED,
			},
		}, proc: &healthpb.RoutineStatus {
			Routine: "Process Transactions",
			Status: &healthpb.ServingStatus{
				Service: "Billing Machine",
				Status: healthpb.Status_STOPPED,
			},
		},
	}
}

func (s *BillingServiceServer) CreatePlan(ctx context.Context, plan *pb.Plan) (*pb.Plan, error) {
	log := s.log.Named("CreatePlan")
	requestor := ctx.Value(nocloud.NoCloudAccount).(string)
	log.Debug("request", zap.Any("plan", plan), zap.String("requestor", requestor))

	ns, err := s.nss.Get(ctx, "0")
	if err != nil {
		return nil, err
	}
	ok := graph.HasAccess(ctx, s.db, requestor, ns.ID.String(), access.ADMIN)
	if !ok {
		return nil, status.Error(codes.PermissionDenied, "Not enough Access rights to manage BillingPlans")
	}

	res, err := s.plans.Create(ctx, plan)
	if err != nil {
		log.Error("Error creating plan", zap.Error(err))
		return nil, status.Error(codes.Internal, "Error creating plan")
	}

	return res.Plan, nil
}

func (s *BillingServiceServer) UpdatePlan(ctx context.Context, plan *pb.Plan) (*pb.Plan, error) {
	log := s.log.Named("UpdatePlan")
	requestor := ctx.Value(nocloud.NoCloudAccount).(string)
	log.Debug("request", zap.Any("plan", plan), zap.String("requestor", requestor))

	ns, err := s.nss.Get(ctx, "0")
	if err != nil {
		return nil, err
	}
	ok := graph.HasAccess(ctx, s.db, requestor, ns.ID.String(), access.ADMIN)
	if !ok {
		return nil, status.Error(codes.PermissionDenied, "Not enough Access rights to manage BillingPlans")
	}

	res, err := s.plans.Update(ctx, plan)
	if err != nil {
		log.Error("Error updating plan", zap.Error(err))
		return nil, status.Error(codes.Internal, "Error updating plan")
	}

	return res.Plan, nil
}

func (s *BillingServiceServer) DeletePlan(ctx context.Context, plan *pb.Plan) (*pb.Plan, error) {
	log := s.log.Named("DeletePlan")
	requestor := ctx.Value(nocloud.NoCloudAccount).(string)
	log.Debug("request", zap.Any("plan", plan), zap.String("requestor", requestor))

	ns, err := s.nss.Get(ctx, "0")
	if err != nil {
		return nil, err
	}
	ok := graph.HasAccess(ctx, s.db, requestor, ns.ID.String(), access.ADMIN)
	if !ok {
		return nil, status.Error(codes.PermissionDenied, "Not enough Access rights to manage BillingPlans")
	}

	err = s.plans.Delete(ctx, plan)
	if err != nil {
		log.Error("Error deleting plan", zap.Error(err))
		return nil, status.Error(codes.Internal, "Error deleting plan")
	}

	return plan, nil
}

func (s *BillingServiceServer) GetPlan(ctx context.Context, plan *pb.Plan) (*pb.Plan, error) {
	log := s.log.Named("GetPlan")
	requestor := ctx.Value(nocloud.NoCloudAccount).(string)
	log.Debug("request", zap.Any("plan", plan), zap.String("requestor", requestor))

	p, err := s.plans.Get(ctx, plan)
	if err != nil {
		log.Error("Error getting plan", zap.Error(err))
		return nil, status.Error(codes.Internal, "Error getting plan")
	}

	if p.Public {
		return p.Plan, nil
	}

	ok := graph.HasAccess(ctx, s.db, requestor, p.ID.String(), access.READ)
	if !ok {
		return nil, status.Error(codes.PermissionDenied, "Not enough Access rights to manage BillingPlans")
	}

	return p.Plan, nil
}

func (s *BillingServiceServer) ListPlans(ctx context.Context, _ *pb.ListRequest) (*pb.ListResponse, error) {
	log := s.log.Named("ListPlans")
	requestor := ctx.Value(nocloud.NoCloudAccount).(string)
	log.Debug("request", zap.String("requestor", requestor))

	plans, err := s.plans.List(ctx)
	if err != nil {
		log.Error("Error listing plans", zap.Error(err))
		return nil, status.Error(codes.Internal, "Error listing plans")
	}

	result := make([]*pb.Plan, 0)
	for _, plan := range plans {
		if plan.Public {
			result = append(result, plan.Plan)
			continue
		}
		ok := graph.HasAccess(ctx, s.db, requestor, plan.ID.String(), access.READ)
		if !ok {
			continue
		}
		result = append(result, plan.Plan)
	}

	return &pb.ListResponse{ Pool: result }, nil
}