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

	pb "github.com/slntopp/nocloud-proto/health"
	"github.com/slntopp/nocloud/pkg/nocloud"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	log *zap.Logger
)

func init() {
	log = nocloud.NewLogger()
}

func main() {
	conn, err := grpc.Dial("localhost:8000", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal("Error dialing service", zap.Error(err))
	}

	client := pb.NewInternalProbeServiceClient(conn)
	r, err := client.Service(context.Background(), &pb.ProbeRequest{
		ProbeType: "services",
	})
	if err != nil {
		log.Fatal("Error performing probe check", zap.Error(err))
	}

	if r.Status != pb.Status_SERVING {
		log.Fatal("Status is not Serving", zap.Any("probe_response", r))
	}
}
