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
package services

import (
	"context"

	driverpb "github.com/slntopp/nocloud/pkg/drivers/instance/vanilla"
	instances "github.com/slntopp/nocloud/pkg/instances/proto"
	pb "github.com/slntopp/nocloud/pkg/services/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

//Gets statuses Instanses of Servce from pkg/statuses
func (s *ServicesServiceServer) GetStates(ctx context.Context, request *pb.GetStatesRequest) (*pb.GetStatesResponse, error) {
	log := s.log.Named("GetStates")

	service, err := s.Get(ctx, &pb.GetRequest{
		Uuid: request.Uuid,
	})
	if err != nil {
		return nil, err
	}

	resp, err := s.statuses.StateGet(ctx, service)
	if err != nil {
		log.Error("fail to get States", zap.Error(err))
		return nil, status.Error(codes.Internal, "fail to get States")
	}

	return resp, nil
}

//Gets statuses Instanses of Servce from pkg/statuses
func (s *ServicesServiceServer) UpdateStates(ctx context.Context, request *pb.GetStatesRequest) error {
	log := s.log.Named("UpdateStates")

	// s.ticker.Stop()
	//s.ticker.Reset()

	for range s.ticker.C {

		var ig_array []*instances.InstancesGroup

		sp_array, err := s.sp_ctrl.List(ctx, "") // gets service providers
		if err != nil {
			log.Error("fail to get ServicesProviders", zap.Error(err))
			return status.Error(codes.Internal, "fail to get ServicesProviders")
		}

		for isp := range sp_array {

			sp := sp_array[isp]
			s_array, err := s.sp_ctrl.ListDeployments(ctx, sp) // gets services
			if err != nil {
				log.Error("fail to get Services", zap.Error(err))
				return status.Error(codes.Internal, "fail to get Services")
				// continue? todo
			}

			for is := range s_array {
				g_array := s_array[is].GetInstancesGroups() // gets groups
				for _, ig := range g_array {
					ig_array = append(ig_array, ig)
				}

			}

			// 1. pool = List ServicesProviders
			// 2. for sp in pool {
			//    groups := sp.ListDeployments()
			//    driver.MonitorStates({sp, groups})
			// }

			_, err = s.drivers["key"].MonitorStates(ctx, &driverpb.StateUpdateRequest{//key todo
				Group:            ig_array,
				ServicesProvider: sp.ServicesProvider,
			})
			if err != nil {
				log.Error("fail to UpdateStates", zap.Error(err))
				return status.Error(codes.Internal, "fail to UpdateStates")
				// continue? todo
			}

		}

	}

	return nil
}
