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
	"errors"

	"github.com/arangodb/go-driver"
	"github.com/gofrs/uuid"
	"go.uber.org/zap"
	"golang.org/x/net/context"

	pb "github.com/slntopp/nocloud/pkg/instances/proto"
)

const (
	INSTANCES_GROUPS_COL = "InstancesGroups"
)

type InstancesGroup struct {
	*pb.InstancesGroup
	driver.DocumentMeta
}

type InstancesGroupsController struct {
	col driver.Collection // Instances Collection
	inst_ctrl InstancesController
	
	log *zap.Logger
}

func NewInstancesGroupsController(log *zap.Logger, db driver.Database) InstancesGroupsController {
	col, _ := db.Collection(context.TODO(), INSTANCES_GROUPS_COL)
	return InstancesGroupsController{log: log.Named("InstancesGroupsController"), inst_ctrl: NewInstancesController(log, db), col: col}
}

func (ctrl *InstancesGroupsController) Create(ctx context.Context, group *pb.InstancesGroup) (error) {
	ctrl.log.Debug("Creating InstancesGroup", zap.Any("group", group))
	id, err := uuid.NewV4()
	if err != nil {
		ctrl.log.Debug("Error generating UUID", zap.Error(err))
		return errors.New("Error generating UUID")
	}
	for _, instance := range group.GetInstances() {
		err := ctrl.inst_ctrl.Create(ctx, instance)
		if err != nil {
			return err
		}
	}
	group.Uuid = id.String()
	return nil
}

func (ctrl *InstancesGroupsController) Update(ctx context.Context, group *pb.InstancesGroup) (error) {
	ctrl.log.Debug("Updating InstancesGroup", zap.Any("group", group))
	for _, instance := range group.GetInstances() {
		err := ctrl.inst_ctrl.Update(ctx, instance)
		if err != nil {
			return err
		}
	}
	return nil
}