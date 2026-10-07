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

	if _, err := CheckMember(member("ai"), log, create); err == nil {
		t.Error("an AI member creates instances")
	}
	if _, err := CheckMember(member("services.read"), log, create); err == nil {
		t.Error("a services.read member creates instances")
	}
	ctx, err := CheckMember(member("services.read"), log, list)
	if err != nil || ctx.Value(NoCloudAccount) != "o" || MemberOwner(ctx) != "" || ctx.Value(NoCloudActingMember) != "m" {
		t.Errorf("a services.read member does not list as the owner: %v", err)
	}
	ctx, err = CheckMember(member("services.manage", "services.read"), log, get)
	if err != nil || ctx.Value(NoCloudAccount) != "m" || MemberOwner(ctx) != "o" {
		t.Errorf("Get is not the member's own: %v", err)
	}
	for _, a := range []string{"billing.read", "billing.pay", "services.read", "services.manage"} {
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
	if got, _ := MemberAccess([]string{"services.manage", "billing.pay", "ai"}); !slices.Equal(got,
		[]string{"ai", "billing.pay", "billing.read", "services.manage", "services.read"}) {
		t.Errorf("implied access: %v", got)
	}
	if _, err := MemberAccess([]string{"root"}); err == nil {
		t.Error("unknown access accepted")
	}
	if got := MemberAccessOf(nil); !slices.Equal(got, []string{"ai"}) {
		t.Errorf("legacy member: %v", got)
	}
	none, _ := structpb.NewStruct(map[string]interface{}{"member_access": []interface{}{}})
	if got := MemberAccessOf(none); len(got) != 0 {
		t.Errorf("no access: %v", got)
	}
}
