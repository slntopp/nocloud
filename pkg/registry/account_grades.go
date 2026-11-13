package registry

import (
	"context"
	"strings"

	"github.com/slntopp/nocloud-proto/access"
	accountspb "github.com/slntopp/nocloud-proto/registry/accounts"
	"github.com/slntopp/nocloud/pkg/nocloud"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const maxGradeOverrideLength = 32

// isNocloudAdmin reports root on the root namespace. Access to the account
// itself is not enough: clients get ROOT on their own account and resellers
// on their clients.
//
// The level comes from the token's root claim, which Token computes from the
// permissions graph when asked (the admin UI always asks; service tokens carry
// ROOT), and the auth interceptor puts it into the context. This is the check
// the settings service uses, and it saves a graph query on every Get and List.
// A token without the claim is treated as not an admin.
func isNocloudAdmin(ctx context.Context) bool {
	lvl, ok := ctx.Value(nocloud.NoCloudRootAccess).(int)
	return ok && lvl >= int(access.Level_ROOT)
}

// stripAdminOnlyFields drops what only nocloud admins may see.
func stripAdminOnlyFields(acc *accountspb.Account) {
	if acc == nil {
		return
	}
	acc.Spending = nil
	acc.GradeOverride = nil
}

// gradeOverridePatch returns the value to store when the request changes the
// manual grade; ok is false when there is nothing to write. An empty grade
// clears the override.
func gradeOverridePatch(current string, requested *string) (value any, ok bool, err error) {
	if requested == nil {
		return nil, false, nil
	}
	grade := strings.TrimSpace(*requested)
	if len(grade) > maxGradeOverrideLength {
		return nil, false, status.Error(codes.InvalidArgument, "Grade is too long")
	}
	if grade == current {
		return nil, false, nil
	}
	if grade == "" {
		return nil, true, nil
	}
	return grade, true, nil
}

// isAdminOnlySortField reports sort fields that would reveal admin-only data.
func isAdminOnlySortField(field string) bool {
	return strings.HasPrefix(field, "spending") || strings.HasPrefix(field, "grade_override")
}
