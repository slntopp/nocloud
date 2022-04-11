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
package graph

import (
	"context"
	"errors"

	"github.com/arangodb/go-driver"
	pb "github.com/slntopp/nocloud/pkg/billing/proto"
	"github.com/slntopp/nocloud/pkg/nocloud/schema"
	"go.uber.org/zap"
)

type BillingPlan struct {
	*pb.Plan
	driver.DocumentMeta
}

type BillingPlansController struct {
	log *zap.Logger
	col driver.Collection // Billing Plans collection
}

func NewBillingPlansController(log *zap.Logger, db driver.Database) BillingPlansController {
	plans, _ := db.Collection(context.TODO(), schema.BILLING_PLANS_COL)
	return BillingPlansController{
		log: log, col: plans,
	}
}

func (ctrl *BillingPlansController) Create(ctx context.Context, plan *pb.Plan) (*BillingPlan, error) {
	plan.LinkedInstances = []string{}
	meta, err := ctrl.col.CreateDocument(ctx, plan)
	if err != nil {
		return nil, err
	}
	plan.Uuid = meta.ID.Key()
	return &BillingPlan{
		Plan: plan, DocumentMeta: meta,
	}, nil
}

func (ctrl *BillingPlansController) Update(ctx context.Context, plan *pb.Plan) (*BillingPlan, error) {
	if plan.Uuid == "" {
		return nil, errors.New("uuid is empty")
	}
	plan.LinkedInstances = []string{}
	meta, err := ctrl.col.ReplaceDocument(ctx, plan.Uuid, plan)
	if err != nil {
		return nil, err
	}
	plan.Uuid = meta.ID.Key()
	return &BillingPlan{
		Plan: plan, DocumentMeta: meta,
	}, nil
}

func (ctrl *BillingPlansController) Delete(ctx context.Context, plan *pb.Plan) (err error) {
	if plan.Uuid == "" {
		return errors.New("uuid is empty")
	}

	if len(plan.LinkedInstances) > 0 {
		return errors.New("requested Billing Plan has linked Instances")
	}

	_, err = ctrl.col.RemoveDocument(ctx, plan.Uuid)
	return err
}

func (ctrl *BillingPlansController) Get(ctx context.Context, plan *pb.Plan) (*BillingPlan, error) {
	if plan.Uuid == "" {
		return nil, errors.New("uuid is empty")
	}
	meta, err := ctrl.col.ReadDocument(ctx, plan.Uuid, plan)
	if err != nil {
		return nil, err
	}

	return &BillingPlan{
		Plan: plan, DocumentMeta: meta,
	}, nil
}

func (ctrl *BillingPlansController) List(ctx context.Context) ([]*BillingPlan, error) {
	query := `FOR plan IN @@plans RETURN plan`
	bindVars := map[string]interface{}{
		"@plans": schema.BILLING_PLANS_COL,
	}
	ctrl.log.Debug("Ready to build query", zap.Any("bindVars", bindVars))

	c, err := ctrl.col.Database().Query(ctx, query, bindVars)
	if err != nil {
		return nil, err
	}
	defer c.Close()
	var r []*BillingPlan
	for {
		var s pb.Plan
		meta, err := c.ReadDocument(ctx, &s)
		if driver.IsNoMoreDocuments(err) {
			break
		} else if err != nil {
			return nil, err
		}
		ctrl.log.Debug("Got document", zap.Any("plan", &s))
		s.Uuid = meta.ID.Key()
		r = append(r, &BillingPlan{&s, meta})
	}

	return r,  nil
}

func (ctrl *BillingPlansController) Link(ctx context.Context, uuid string, instances []string) (error) {
	var plan pb.Plan
	_, err := ctrl.col.ReadDocument(ctx, uuid, &plan)
	if err != nil {
		return err
	}
	if plan.LinkedInstances == nil {
		plan.LinkedInstances = []string{}
	}
	plan.LinkedInstances = append(plan.LinkedInstances, instances...)
	_, err = ctrl.col.ReplaceDocument(ctx, uuid, &plan)
	return err
}