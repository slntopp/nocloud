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

/*
	THIS IS AN EXAMPLE DRIVER SERVER
*/

package driver

import (
	"context"

	pb "github.com/slntopp/nocloud/pkg/drivers/instance/proto"
)

var DRIVER_NAME string

func SetDriverName(name string) {
	DRIVER_NAME = name
}

type DriverServiceServer struct {
	pb.UnimplementedDriverServiceServer
}

func (s *DriverServiceServer) GetType(ctx context.Context, _ *pb.GetTypeRequest) (*pb.GetTypeResponse, error) {
	return &pb.GetTypeResponse{Type: DRIVER_NAME}, nil
}

func (s *DriverServiceServer) ValidateConfigSyntax(ctx context.Context, req *pb.ValidateConfigSyntaxRequest) (*pb.ValidateConfigSyntaxResponse, error) {
	return &pb.ValidateConfigSyntaxResponse{Result: true}, nil
}