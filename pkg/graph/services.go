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
	"encoding/json"

	"github.com/arangodb/go-driver"
	pb "github.com/slntopp/nocloud/pkg/services/proto"
	"go.uber.org/zap"
)

const (
	SERVICES_COL = "Services"
	NS2SERV = NAMESPACES_COL + "2" + SERVICES_COL
)

type Service struct {
	Version string `json:"version"`
	Title string `json:"title"`
	Context map[string]interface{} `json:"context"`
	InstancesGroups map[string]InstancesGroup `json:"instances_groups"`
	State string `json:"state"`

	Hash string `json:"hash"`

	driver.DocumentMeta
}

type ServicesController struct {
	col driver.Collection // Services Collection
	ig_ctrl InstancesGroupsController

	log *zap.Logger
}

func NewServicesController(log *zap.Logger, db driver.Database) ServicesController {
	col, _ := db.Collection(nil, SERVICES_COL)
	return ServicesController{log: log, col: col, ig_ctrl: NewInstancesGroupsController(log, db)}
}

func (s *Service) ToServiceMessage() (res *pb.Service, err error) {
	b, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(b, res)
	return res, err
}

func MakeServiceFromMessage(req *pb.Service) (res *Service, err error) {
	b, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(b, res)
	return res, err
}

func (ctrl *ServicesController) Create(ctx context.Context, service Service) (error) {
	for _, ig := range service.InstancesGroups {
		ctrl.ig_ctrl.Create(ctx, ig)
	}
	return nil
}