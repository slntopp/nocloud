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
package main

import (
	"context"
	"github.com/slntopp/nocloud/pkg/instances"

	pb "github.com/slntopp/nocloud-proto/health"
	"github.com/slntopp/nocloud/pkg/services"
	"go.uber.org/zap"
)

const SERVICE = "Services Registry"

type HealthServer struct {
	pb.UnimplementedInternalProbeServiceServer
	log  *zap.Logger
	srv  *services.ServicesServer
	isrv *instances.InstancesServer
}

func NewHealthServer(log *zap.Logger, srv *services.ServicesServer) *HealthServer {
	return &HealthServer{
		log: log, srv: srv,
	}
}

func (s *HealthServer) Service(_ context.Context, _ *pb.ProbeRequest) (*pb.ServingStatus, error) {
	return &pb.ServingStatus{
		Service: SERVICE,
		Status:  pb.Status_SERVING,
	}, nil
}

func (s *HealthServer) Routine(_ context.Context, _ *pb.ProbeRequest) (*pb.RoutinesStatus, error) {
	state := s.isrv.MonitoringRoutineState()
	status := &pb.RoutineStatus{
		Routine: state.Name,
		Status: &pb.ServingStatus{
			Service: SERVICE,
			Status:  pb.Status_STOPPED,
		},
		LastExecution: state.LastExec,
	}
	if state.Running {
		status.Status.Status = pb.Status_RUNNING
	}

	return &pb.RoutinesStatus{
		Routines: []*pb.RoutineStatus{status},
	}, nil
}
