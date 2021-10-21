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
package main

import (
	"context"

	"github.com/slntopp/nocloud/pkg/api/apipb"
	"github.com/slntopp/nocloud/pkg/health/healthpb"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type healthAPI struct {
	client healthpb.HealthServiceClient
	apipb.UnimplementedHealthServiceServer
}

func (h *healthAPI) mustEmbedUnimplementedHealthServiceServer() {
	log.Info("Method missing")
}

func (h *healthAPI) Probe(ctx context.Context, request *healthpb.ProbeRequest) (response *healthpb.ProbeResponse, err error) {
	log.Info("IONe Health", zap.String("Probe", request.ProbeType))

	res, err := h.client.Probe(ctx, request)
	if err != nil {
		log.Error("Probe Failed", zap.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}
	return res, nil
}
