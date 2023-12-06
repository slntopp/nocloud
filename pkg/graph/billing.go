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
package graph

import (
	"context"
	"errors"
	statuspb "github.com/slntopp/nocloud-proto/statuses"

	"github.com/arangodb/go-driver"
	pb "github.com/slntopp/nocloud-proto/billing"
	"github.com/slntopp/nocloud/pkg/nocloud/schema"
	"go.uber.org/zap"
)

type BillingPlan struct {
	*pb.Plan
	driver.DocumentMeta
}

type BillingPlansController struct {
	log   *zap.Logger
	col   driver.Collection // Billing Plans collection
	graph driver.Graph
}

func NewBillingPlansController(logger *zap.Logger, db driver.Database) BillingPlansController {
	ctx := context.TODO()
	log := logger.Named("BillingPlansController")
	graph := GraphGetEnsure(log, ctx, db, schema.BILLING_GRAPH.Name)
	plans := GetEnsureCollection(log, ctx, db, schema.BILLING_PLANS_COL)
	GraphGetEdgeEnsure(log, ctx, graph, schema.SP2BP, schema.SERVICES_PROVIDERS_COL, schema.BILLING_PLANS_COL)
	return BillingPlansController{
		log: log, col: plans, graph: graph,
	}
}

func (ctrl *BillingPlansController) Create(ctx context.Context, plan *pb.Plan) (*BillingPlan, error) {
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
	meta, err := ctrl.col.ReplaceDocument(ctx, plan.Uuid, plan)
	if err != nil {
		return nil, err
	}
	plan.Uuid = meta.ID.Key()
	return &BillingPlan{
		Plan: plan, DocumentMeta: meta,
	}, nil
}

func (ctrl *BillingPlansController) Delete(ctx context.Context, plan *pb.Plan) error {
	if plan.Uuid == "" {
		return errors.New("uuid is empty")
	}

	plan.Status = statuspb.NoCloudStatus_DEL

	_, err := ctrl.col.UpdateDocument(ctx, plan.GetUuid(), plan)

	if err != nil {
		return err
	}

	/*db := ctrl.col.Database()
	bpId := driver.NewDocumentID(schema.BILLING_PLANS_COL, plan.GetUuid())
	_, err = db.Query(ctx, deleteFromEdgeBillingBlans, map[string]interface{}{
		"plan":                bpId,
		"permissions":         schema.PERMISSIONS_GRAPH.Name,
		"@services_providers": schema.SERVICES_PROVIDERS_COL,
		"@sp_to_bp":           schema.SP2BP,
	})*/

	return nil
}

/*const deleteFromEdgeBillingBlans = `
LET sp2bp = (
    FOR node, edge IN INBOUND @plan GRAPH @permissions
        FILTER IS_SAME_COLLECTION(node, @@services_providers)
        RETURN edge
)

FOR item IN sp2bp
    REMOVE item IN @@sp_to_bp
`*/

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

const getInstancesCount = `
FOR i IN Instances
    FILTER i.billing_plan.uuid == @plan
    FILTER i.status != @status
    RETURN i
`

func (ctrl *BillingPlansController) InstancesCount(ctx context.Context, plan *pb.Plan) (int, error) {
	if plan.Uuid == "" {
		return 0, errors.New("uuid is empty")
	}

	cur, err := ctrl.col.Database().Query(ctx, getInstancesCount, map[string]interface{}{
		"plan":   plan.Uuid,
		"status": statuspb.NoCloudStatus_DEL,
	})

	if err != nil {
		return 0, err
	}

	result := 0
	for cur.HasMore() {
		result += 1
	}

	return result, nil
}

func (ctrl *BillingPlansController) List(ctx context.Context, spUuid string) ([]*BillingPlan, error) {
	var query string
	bindVars := make(map[string]interface{}, 0)

	if spUuid == "" {
		query = `FOR plan IN BillingPlans RETURN plan`
	} else {
		query = `FOR node, edge IN 1 OUTBOUND @sp GRAPH Billing RETURN Document(edge._to)`
		spDocId := driver.NewDocumentID(schema.SERVICES_PROVIDERS_COL, spUuid)
		bindVars["sp"] = spDocId
	}
	ctrl.log.Debug("Ready to build query", zap.Any("bindVars", bindVars))

	c, err := ctrl.col.Database().Query(ctx, query, bindVars)
	if err != nil {
		return nil, err
	}
	defer c.Close()
	var r []*BillingPlan
	for c.HasMore() {
		var s pb.Plan
		meta, err := c.ReadDocument(ctx, &s)
		if err != nil {
			return nil, err
		}
		ctrl.log.Debug("Got document", zap.Any("plan", &s))
		s.Uuid = meta.ID.Key()
		r = append(r, &BillingPlan{&s, meta})
	}

	return r, nil
}
