package registry

import (
	"context"
	"fmt"
	"slices"

	"github.com/arangodb/go-driver"
	accountspb "github.com/slntopp/nocloud-proto/registry/accounts"
	"github.com/slntopp/nocloud/pkg/graph"
	"github.com/slntopp/nocloud/pkg/nocloud"
	"github.com/slntopp/nocloud/pkg/nocloud/schema"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
)

// The owner's list of members kept off its balance: they spend only its AI packages.
const balanceBlockedKey = "ai_balance_blocked"

// organizationCard is what a member sees of its organization: name, balance and currency, and
// whether the member itself is kept off the balance. Nothing else of the owner's account leaks.
func (s *AccountsServiceServer) organizationCard(ctx context.Context, owner, member string) (*accountspb.Account, error) {
	// Read as the owner itself, as GET /accounts/me does: members have no path to the owner.
	ownerId := driver.NewDocumentID(schema.ACCOUNTS_COL, owner)
	org, err := s.ctrl.GetWithAccess(ctx, ownerId, owner)
	if err != nil || org.Account == nil {
		return nil, status.Error(codes.NotFound, "Account not found")
	}

	blocked := []interface{}{}
	for _, id := range org.GetData().GetFields()[balanceBlockedKey].GetListValue().GetValues() {
		if id.GetStringValue() == member {
			blocked = append(blocked, member)
		}
	}
	data, err := structpb.NewStruct(map[string]interface{}{balanceBlockedKey: blocked})
	if err != nil {
		return nil, status.Error(codes.Internal, "Failed to build organization card")
	}

	return &accountspb.Account{
		Uuid:     owner,
		Title:    org.GetTitle(),
		Balance:  org.Balance,
		Currency: org.Currency,
		Data:     data,
	}, nil
}

// memberAccessPatch checks the member access in an account data patch: only members have it, and
// it may name only known access. It returns the access to store, nil when the patch has none.
func memberAccessPatch(data map[string]interface{}, isMember bool) ([]interface{}, error) {
	raw, ok := data[nocloud.MemberAccessKey]
	if !ok {
		return nil, nil
	}
	if !isMember {
		return nil, status.Error(codes.InvalidArgument, "Only subaccounts have member access")
	}
	list, ok := raw.([]interface{})
	if !ok {
		return nil, status.Error(codes.InvalidArgument, "Member access must be a list")
	}
	access := []string{}
	for _, a := range list {
		s, ok := a.(string)
		if !ok {
			return nil, status.Error(codes.InvalidArgument, "Member access must be a list of strings")
		}
		access = append(access, s)
	}
	access, err := nocloud.MemberAccess(access)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	result := []interface{}{}
	for _, a := range access {
		result = append(result, a)
	}
	return result, nil
}

// signOut ends every session of an account, so its next token carries what changed (access,
// suspension), and a removed member is gone at once.
func (s *AccountsServiceServer) signOut(ctx context.Context, account string) {
	keys, err := s.rdb.Keys(ctx, fmt.Sprintf("sessions:%s:*", account)).Result()
	if err == nil && len(keys) > 0 {
		err = s.rdb.Del(ctx, keys...).Err()
	}
	if err != nil {
		s.log.Error("Failed to sign the account out", zap.String("account", account), zap.Error(err))
	}
}

// dropMember takes a deleted member off its owner's subaccounts and signs it out.
func (s *AccountsServiceServer) dropMember(ctx context.Context, member graph.Account) {
	s.signOut(ctx, member.GetUuid())
	ownerId := driver.NewDocumentID(schema.ACCOUNTS_COL, member.GetAccountOwner())
	owner, err := s.ctrl.GetWithAccess(ctx, ownerId, member.GetAccountOwner())
	if err != nil || owner.Account == nil {
		s.log.Error("Failed to get the owner of a deleted member", zap.String("member", member.GetUuid()), zap.Error(err))
		return
	}
	subaccounts := slices.DeleteFunc(slices.Clone(owner.GetSubaccounts()), func(id string) bool {
		return id == member.GetUuid()
	})
	if err := s.ctrl.Update(ctx, owner, map[string]interface{}{"subaccounts": subaccounts}); err != nil {
		s.log.Error("Failed to take a deleted member off its owner", zap.String("member", member.GetUuid()), zap.Error(err))
	}
}
