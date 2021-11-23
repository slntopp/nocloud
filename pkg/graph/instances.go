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
	"errors"

	"github.com/arangodb/go-driver"
	"github.com/gofrs/uuid"
	"go.uber.org/zap"

	pb "github.com/slntopp/nocloud/pkg/instances/proto"
)

const (
	INSTANCES_COL = "Instances"
)

type Instance struct {
	*pb.Instance
	driver.DocumentMeta
}


type InstancesController struct {
	col driver.Collection // Instances Collection

	log *zap.Logger
}

func NewInstancesController(log *zap.Logger, db driver.Database) InstancesController {
	col, _ := db.Collection(nil, INSTANCES_COL)
	return InstancesController{log: log.Named("InstancesController"), col: col}
}

func (ctrl *InstancesController) Create(ctx context.Context, instance *pb.Instance) (error) {
	ctrl.log.Debug("Creating Instance", zap.Any("instance", instance))
	id, err := uuid.NewV4()
	if err != nil {
		ctrl.log.Debug("Error generating UUID", zap.Error(err))
		return errors.New("Error generating UUID")
	}

	instance.Uuid = id.String()
	return nil
}

func (ctrl *InstancesController) Update(ctx context.Context, instance *pb.Instance) (error) {
	ctrl.log.Debug("Updating Instance", zap.Any("instance", instance))
	return nil
}