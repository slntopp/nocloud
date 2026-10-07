package nocloud

import (
	"context"
	"slices"
	"testing"

	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestCheckMember(t *testing.T) {
	log := zap.NewNop()
	member := func(access ...interface{}) context.Context {
		ctx := context.WithValue(context.Background(), NoCloudAccount, "m")
		return MemberClaims(ctx, "o", access)
	}
	create := "/nocloud.instances.InstancesService/Create"
	list := "/nocloud.instances.InstancesService/List"
	get := "/nocloud.registry.AccountsService/Get"
	update := "/nocloud.registry.AccountsService/Update"

	defer func(g string) { memberGate = g }(memberGate)
	memberGate = ""

	if _, err := CheckMember(member(), log, create); err == nil {
		t.Error("a member without access creates instances")
	}
	if _, err := CheckMember(member("services"), log, create); err == nil {
		t.Error("a services member creates instances")
	}
	if _, err := CheckMember(member("order", "services"), log, create); err != nil {
		t.Errorf("an order member does not create instances: %v", err)
	}
	if _, err := CheckMember(member("order", "services"), log, "/nocloud.instances.InstancesService/Delete"); err == nil {
		t.Error("an order member deletes instances")
	}
	ctx, err := CheckMember(member("services"), log, list)
	if err != nil || ctx.Value(NoCloudAccount) != "o" || MemberOwner(ctx) != "" || ctx.Value(NoCloudActingMember) != "m" {
		t.Errorf("a services member does not list as the owner: %v", err)
	}
	ctx, err = CheckMember(member("order", "services"), log, get)
	if err != nil || ctx.Value(NoCloudAccount) != "m" || MemberOwner(ctx) != "o" {
		t.Errorf("Get is not the member's own: %v", err)
	}
	for _, a := range []string{"order", "invoices", "support", "services"} {
		if _, err := CheckMember(member(a), log, update); err == nil {
			t.Errorf("a %s member updates accounts", a)
		}
	}
	owner := context.WithValue(context.Background(), NoCloudAccount, "o")
	if ctx, err := CheckMember(owner, log, create); err != nil || ctx != owner {
		t.Errorf("the owner is touched: %v", err)
	}

	memberGate = "log"
	if _, err := CheckMember(member(), log, create); err != nil {
		t.Errorf("log mode refuses: %v", err)
	}
}

func TestMemberAccess(t *testing.T) {
	if got, _ := MemberAccess([]string{"order", "support"}); !slices.Equal(got, []string{"order", "services", "support"}) {
		t.Errorf("implied access: %v", got)
	}
	if _, err := MemberAccess([]string{"ai"}); err == nil {
		t.Error("unknown access accepted")
	}
	old, _ := structpb.NewStruct(map[string]interface{}{"member_access": []interface{}{"ai", "billing.read"}})
	if got := MemberAccessOf(old); len(got) != 0 {
		t.Errorf("old access names: %v", got)
	}
	if got := MemberAccessOf(nil); len(got) != 0 {
		t.Errorf("legacy member: %v", got)
	}
	none, _ := structpb.NewStruct(map[string]interface{}{"member_access": []interface{}{}})
	if got := MemberAccessOf(none); len(got) != 0 {
		t.Errorf("no access: %v", got)
	}
}
