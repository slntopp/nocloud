package graph

import (
	"context"
	"errors"
	"testing"

	bpb "github.com/slntopp/nocloud-proto/billing"
	pb "github.com/slntopp/nocloud-proto/instances"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

type fakePlans struct {
	BillingPlansController
	plans map[string]*bpb.Plan
}

func (f fakePlans) Get(_ context.Context, plan *bpb.Plan) (*BillingPlan, error) {
	stored, ok := f.plans[plan.GetUuid()]
	if !ok {
		return nil, errors.New("plan not found")
	}
	return &BillingPlan{Plan: stored}, nil
}

func lockedMeta(locked bool) map[string]*structpb.Value {
	return map[string]*structpb.Value{LockProductKey: structpb.NewBoolValue(locked)}
}

func instanceOn(plan, product string) *pb.Instance {
	inst := &pb.Instance{Product: &product}
	if plan != "" {
		inst.BillingPlan = &bpb.Plan{Uuid: plan}
	}
	return inst
}

func strPtr(s string) *string { return &s }

func TestLockedProductChange(t *testing.T) {
	plans := fakePlans{plans: map[string]*bpb.Plan{
		"ai":    {Uuid: "ai", Meta: lockedMeta(true)},
		"ai-2":  {Uuid: "ai-2", Meta: lockedMeta(true)},
		"vps":   {Uuid: "vps"},
		"off":   {Uuid: "off", Meta: lockedMeta(false)},
		"other": {Uuid: "other"},
	}}

	cases := []struct {
		name   string
		inst   *pb.Instance
		old    *pb.Instance
		locked bool
		err    bool
	}{
		{"product change on a locked plan", instanceOn("ai", "ai-50"), instanceOn("ai", "ai-20"), true, false},
		{"product change with the plan left out", &pb.Instance{Product: strPtr("ai-50")}, instanceOn("ai", "ai-20"), true, false},
		{"move off a locked plan", instanceOn("vps", "small"), instanceOn("ai", "ai-20"), true, false},
		{"move onto a locked plan", instanceOn("ai", "ai-50"), instanceOn("vps", "small"), true, false},
		{"move between locked plans", instanceOn("ai-2", "ai-20"), instanceOn("ai", "ai-20"), true, false},
		{"the locked flag in the request copy is ignored", &pb.Instance{
			Product:     strPtr("ai-50"),
			BillingPlan: &bpb.Plan{Uuid: "ai", Meta: lockedMeta(false)},
		}, instanceOn("ai", "ai-20"), true, false},
		{"same product and plan", instanceOn("ai", "ai-20"), instanceOn("ai", "ai-20"), false, false},
		{"product and plan left out", &pb.Instance{Title: "renamed"}, instanceOn("ai", "ai-20"), false, false},
		{"product change on an unlocked plan", instanceOn("vps", "big"), instanceOn("vps", "small"), false, false},
		{"lock_product false", instanceOn("off", "big"), instanceOn("off", "small"), false, false},
		{"move between unlocked plans", instanceOn("other", "small"), instanceOn("vps", "small"), false, false},
		{"unknown plan fails closed", instanceOn("missing", "x"), instanceOn("vps", "small"), false, true},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			locked, err := LockedProductChange(context.Background(), plans, c.inst, c.old)
			if (err != nil) != c.err {
				t.Fatalf("err = %v, want error %v", err, c.err)
			}
			if locked != c.locked {
				t.Fatalf("locked = %v, want %v", locked, c.locked)
			}
		})
	}
}

func withConfig(config map[string]any) *pb.Instance {
	value, err := structpb.NewStruct(config)
	if err != nil {
		panic(err)
	}
	return &pb.Instance{Config: value.GetFields()}
}

func TestKeepAdminConfig(t *testing.T) {
	skip := []any{"ai-20"}

	cases := []struct {
		name    string
		inst    *pb.Instance
		stored  *pb.Instance
		changed bool
		want    map[string]any
	}{
		{"dropped on a new instance", withConfig(map[string]any{"skip_next_payment": skip, "auto_start": true}), nil,
			true, map[string]any{"auto_start": true}},
		{"dropped when not stored", withConfig(map[string]any{"skip_next_payment": skip}), withConfig(map[string]any{}),
			true, map[string]any{}},
		{"stored value kept", withConfig(map[string]any{"skip_next_payment": skip}),
			withConfig(map[string]any{"skip_next_payment": []any{}}), true, map[string]any{"skip_next_payment": []any{}}},
		{"same value passes", withConfig(map[string]any{"skip_next_payment": skip}),
			withConfig(map[string]any{"skip_next_payment": skip}), false, map[string]any{"skip_next_payment": skip}},
		{"left out stays out", withConfig(map[string]any{"user": "u"}),
			withConfig(map[string]any{"skip_next_payment": skip}), false, map[string]any{"user": "u"}},
		{"no config", &pb.Instance{}, withConfig(map[string]any{"skip_next_payment": skip}), false, nil},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if changed := KeepAdminConfig(c.inst, c.stored); changed != c.changed {
				t.Fatalf("changed = %v, want %v", changed, c.changed)
			}
			if c.want == nil {
				if c.inst.GetConfig() != nil {
					t.Fatalf("config = %v, want nil", c.inst.GetConfig())
				}
				return
			}
			got := &structpb.Struct{Fields: c.inst.GetConfig()}
			if !proto.Equal(got, &structpb.Struct{Fields: withConfig(c.want).GetConfig()}) {
				t.Fatalf("config = %v, want %v", got.AsMap(), c.want)
			}
		})
	}
}
