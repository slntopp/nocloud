package registry

import (
	"testing"

	accountspb "github.com/slntopp/nocloud-proto/registry/accounts"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

func TestMcpTokenIsStableUntilGenChanges(t *testing.T) {
	key := []byte("signing")
	a, err := mcpToken(key, "acc", 1)
	if err != nil {
		t.Fatal(err)
	}
	b, err := mcpToken(key, "acc", 1)
	if err != nil {
		t.Fatal(err)
	}
	if a != b {
		t.Fatal("token changed without a new generation")
	}
	c, err := mcpToken(key, "acc", 2)
	if err != nil {
		t.Fatal(err)
	}
	if c == a {
		t.Fatal("new generation must change the token")
	}
}

func TestApplyMcpMask(t *testing.T) {
	cur := defaultMcp()
	if cur.GetEnabled() {
		t.Fatal("MCP must stay off until the client turns it on")
	}
	cur.Enabled = true
	zero := 0.0
	off := false
	orders := int32(0)
	if err := applyMcp(cur, &accountspb.SetMcpRequest{
		UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"pay", "spendDay", "orders_day"}},
		Pay:        &off,
		SpendDay:   &zero,
		OrdersDay:  &orders,
	}); err != nil {
		t.Fatal(err)
	}
	if cur.GetPay() || cur.GetEnabled() != true || cur.GetOrdersDay() != 0 || cur.SpendDay == nil || *cur.SpendDay != 0 {
		t.Fatalf("partial update spilled: %+v", cur)
	}
	if err := applyMcp(cur, &accountspb.SetMcpRequest{
		UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"spend_day"}},
	}); err != nil {
		t.Fatal(err)
	}
	if cur.SpendDay != nil {
		t.Fatal("empty spend_day must clear the cap")
	}
	if err := applyMcp(cur, &accountspb.SetMcpRequest{UpdateMask: &fieldmaskpb.FieldMask{}}); err == nil {
		t.Fatal("empty mask must be rejected")
	}
}

func TestMcpDocumentKeepsZero(t *testing.T) {
	cur := defaultMcp()
	cur.OrdersDay = 0
	cur.Pay = false
	doc := mcpDocument(cur)
	if doc["orders_day"] != int32(0) || doc["pay"] != false {
		t.Fatalf("zero values dropped: %#v", doc)
	}
	if _, ok := doc["spend_day"]; ok {
		t.Fatal("unset spend cap must stay absent")
	}
}
