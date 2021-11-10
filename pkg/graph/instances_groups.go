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
	"github.com/arangodb/go-driver"
	"go.uber.org/zap"
	"golang.org/x/net/context"
)

const (
	INSTANCES_GROUPS_COL = "InstancesGroups"
)

type InstancesGroup struct {
	Type string `json:"type"`
	Config map[string]interface{} `json:"config"`
	Instances []Instance `json:"instances"`

	driver.DocumentMeta
}

type InstancesGroupsController struct {
	col driver.Collection // Instances Collection
	inst_ctrl InstancesController
	
	log *zap.Logger
}

func NewInstancesGroupsController(log *zap.Logger, db driver.Database) InstancesGroupsController {
	col, _ := db.Collection(nil, INSTANCES_GROUPS_COL)
	return InstancesGroupsController{log: log, inst_ctrl: NewInstancesController(log, db), col: col}
}

func (ctrl *InstancesGroupsController) Create(ctx context.Context, group InstancesGroup) (error) {
	for _, instance := range group.Instances {
		ctrl.inst_ctrl.Create(ctx, instance)
	}
	return nil
}