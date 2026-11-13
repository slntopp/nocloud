package registry

import (
	"context"
	"strings"
	"testing"

	accountspb "github.com/slntopp/nocloud-proto/registry/accounts"
	"github.com/slntopp/nocloud/pkg/nocloud"
)

func ptr(s string) *string { return &s }

func TestGradeOverridePatch(t *testing.T) {
	cases := []struct {
		name      string
		current   string
		requested *string
		want      any
		wantOk    bool
		wantErr   bool
	}{
		{name: "absent in request", current: "A+", requested: nil},
		{name: "unchanged", current: "A+", requested: ptr(" A+ ")},
		{name: "set", current: "", requested: ptr("A+"), want: "A+", wantOk: true},
		{name: "change", current: "A+", requested: ptr("B"), want: "B", wantOk: true},
		{name: "clear", current: "A+", requested: ptr(""), want: nil, wantOk: true},
		{name: "already empty", current: "", requested: ptr("  ")},
		{name: "too long", current: "", requested: ptr(strings.Repeat("x", maxGradeOverrideLength+1)), wantErr: true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, ok, err := gradeOverridePatch(c.current, c.requested)
			if (err != nil) != c.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, c.wantErr)
			}
			if ok != c.wantOk || got != c.want {
				t.Fatalf("got (%v, %v), want (%v, %v)", got, ok, c.want, c.wantOk)
			}
		})
	}
}

func TestStripAdminOnlyFields(t *testing.T) {
	acc := &accountspb.Account{
		Title:         "client",
		Spending:      &accountspb.AccountSpending{Amount: 950},
		GradeOverride: ptr("A+"),
	}
	stripAdminOnlyFields(acc)
	if acc.Spending != nil || acc.GradeOverride != nil {
		t.Fatalf("admin-only fields left: %v", acc)
	}
	if acc.Title != "client" {
		t.Fatalf("title changed: %q", acc.Title)
	}
	stripAdminOnlyFields(nil)
}

func TestIsNocloudAdmin(t *testing.T) {
	cases := []struct {
		name string
		ctx  context.Context
		want bool
	}{
		{"no root claim", context.Background(), false},
		{"root", context.WithValue(context.Background(), nocloud.NoCloudRootAccess, 4), true},
		{"admin is not root", context.WithValue(context.Background(), nocloud.NoCloudRootAccess, 3), false},
		{"none", context.WithValue(context.Background(), nocloud.NoCloudRootAccess, 0), false},
	}
	for _, c := range cases {
		if got := isNocloudAdmin(c.ctx); got != c.want {
			t.Errorf("%s: got %v, want %v", c.name, got, c.want)
		}
	}
}
