package nocloud

import (
	"context"
	"fmt"
	"os"
	"slices"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
)

// A subaccount (member) acts for its owner (organization) and only as far as the owner allows.
// Its tokens carry the owner's uuid and the member's access; the auth interceptors let such a token
// reach the base methods as the member itself, and the methods of its access as the owner.
const NOCLOUD_MEMBER_CLAIM = "member"
const NOCLOUD_MEMBER_ACCESS_CLAIM = "member_access"

const NoCloudMember = ContextKey("member")
const NoCloudMemberAccess = ContextKey("member_access")

// NoCloudActingMember is the member behind a call it makes as its owner.
const NoCloudActingMember = ContextKey("acting_member")

// The member's access, kept in its account data under this key; only the owner (or an admin) sets it.
const MemberAccessKey = "member_access"

// Spending on AI (the owner's AI balance and packages, through the AI driver) every member may;
// it is not an access.
const (
	MemberOrder    = "order"    // buy subscriptions and services, top the owner's balance up
	MemberInvoices = "invoices" // see and pay the owner's invoices
	MemberSupport  = "support"  // write to support (the chats service checks it, NoCloud has nothing)
	MemberServices = "services" // see the owner's services
)

// A member with no access set (one from LibreChat, or from before access existed) only spends on AI.
var MemberAccessDefault = []string{}

var memberImplies = map[string][]string{
	MemberOrder:    {MemberServices},
	MemberInvoices: nil,
	MemberSupport:  nil,
	MemberServices: nil,
}

// MemberAccess checks a member's access and adds what it implies: ordering needs seeing what was
// ordered.
func MemberAccess(access []string) ([]string, error) {
	result := []string{}
	for _, a := range access {
		implied, ok := memberImplies[a]
		if !ok {
			return nil, fmt.Errorf("unknown member access %q", a)
		}
		result = append(append(result, a), implied...)
	}
	slices.Sort(result)
	return slices.Compact(result), nil
}

// MemberAccessOf reads a member's access from its account data. No key means MemberAccessDefault,
// an empty list means base methods only. Unknown entries are dropped.
func MemberAccessOf(data *structpb.Struct) []string {
	value, ok := data.GetFields()[MemberAccessKey]
	if !ok {
		access, _ := MemberAccess(MemberAccessDefault)
		return access
	}
	access := []string{}
	for _, v := range value.GetListValue().GetValues() {
		if _, ok := memberImplies[v.GetStringValue()]; ok {
			access = append(access, v.GetStringValue())
		}
	}
	access, _ = MemberAccess(access)
	return access
}

// MEMBER_GATE: "log" lets a member's call outside its methods through and logs it, "off" lets it
// through silently; anything else (the default) refuses it.
var memberGate = os.Getenv("MEMBER_GATE")

func methods(names ...string) map[string]bool {
	m := map[string]bool{}
	for _, n := range names {
		m[n] = true
	}
	return m
}

// What any member may call, as itself. The handlers narrow Get to self and the owner's card, and
// GetTransactions to self and the member's own charges on the owner.
var memberBase = methods(
	"/nocloud.registry.AccountsService/Token",
	"/nocloud.registry.AccountsService/Logout",
	"/nocloud.registry.AccountsService/Get",
	"/nocloud.registry.AccountsService/SetCredentials",
	"/nocloud.registry.AccountsService/ChangeLanguageCode",
	"/nocloud.sessions.SessionsService/Get",
	"/nocloud.sessions.SessionsService/Revoke",
	"/nocloud.sessions.SessionsService/GetActivity",
	"/nocloud.billing.BillingService/GetTransactions",
	"/nocloud.billing.BillingService/GetPlan",
	"/nocloud.billing.BillingService/ListPlans",
	"/nocloud.billing.CurrencyService/GetCurrencies",
	"/nocloud.billing.CurrencyService/GetExchangeRate",
	"/nocloud.billing.CurrencyService/GetExchangeRates",
	"/nocloud.billing.CurrencyService/Convert",
	"/nocloud.billing.CurrencyService/ConvertMany",
	"/nocloud.billing.DescriptionsService/Get",
	"/nocloud.billing.DescriptionsService/List",
	"/nocloud.billing.AddonsService/Get",
	"/nocloud.billing.AddonsService/List",
	"/nocloud.billing.PromocodesService/GetByCode",
	"/nocloud.services_providers.ServicesProvidersService/List",
	"/nocloud.services_providers.ShowcasesService/List",
	"/nocloud.services_providers.ShowcasesService/Get",
	"/nocloud.services_providers.ShowcaseCategoriesService/List",
	"/nocloud.settings.SettingsService/Get",
)

// What each access lets a member call, as its owner. Nothing here may touch accounts themselves:
// as the owner, a member would have ROOT on the owner's account.
var memberScopes = map[string]map[string]bool{
	MemberInvoices: methods(
		"/nocloud.billing.BillingService/GetTransactions",
		"/nocloud.billing.BillingService/GetTransactionsCount",
		"/nocloud.billing.BillingService/GetRecords",
		"/nocloud.billing.BillingService/GetInstancesReports",
		"/nocloud.billing.BillingService/GetInstancesReportsCount",
		"/nocloud.billing.BillingService/GetRecordsReports",
		"/nocloud.billing.BillingService/GetRecordsReportsCount",
		"/nocloud.billing.BillingService/GetInvoice",
		"/nocloud.billing.BillingService/GetInvoices",
		"/nocloud.billing.BillingService/GetInvoicesCount",
		"/nocloud.billing.BillingService/Pay",
		"/nocloud.billing.BillingService/PayWithBalance",
		"/nocloud.billing.BillingService/CreateRenewalInvoice",
	),
	// Buying ends in an invoice to pay, so ordering pays the invoices it makes.
	MemberOrder: methods(
		"/nocloud.services.ServicesService/TestConfig",
		"/nocloud.services.ServicesService/Create",
		"/nocloud.services.ServicesService/Update",
		"/nocloud.services.ServicesService/Up",
		"/nocloud.instances.InstancesService/Create",
		"/nocloud.billing.PromocodesService/Apply",
		"/nocloud.billing.PromocodesService/Detach",
		"/nocloud.billing.BillingService/CreateTopUpBalanceInvoice",
		"/nocloud.billing.BillingService/GetInvoice",
		"/nocloud.billing.BillingService/Pay",
		"/nocloud.billing.BillingService/PayWithBalance",
	),
	MemberServices: methods(
		"/nocloud.services.ServicesService/Get",
		"/nocloud.services.ServicesService/List",
		"/nocloud.services.ServicesService/Stream",
		"/nocloud.instances.InstancesService/Get",
		"/nocloud.instances.InstancesService/List",
		"/nocloud.instances.InstancesService/GetUnique",
		"/nocloud.registry.NamespacesService/List",
		"/nocloud.registry.NamespacesService/Get",
		"/nocloud.billing.BillingService/ListPlansInstances",
		"/nocloud.billing.RecordsService/GetActive",
	),
	MemberSupport: methods(),
}

// MemberOwner is the owner of the member who makes this request as itself, "" for anyone else,
// including a member acting as its owner.
func MemberOwner(ctx context.Context) string {
	owner, _ := ctx.Value(NoCloudMember).(string)
	return owner
}

// MemberClaims puts a member's claims from its token into ctx.
func MemberClaims(ctx context.Context, owner, access interface{}) context.Context {
	o, _ := owner.(string)
	if o == "" {
		return ctx
	}
	list := []string{}
	if a, ok := access.([]interface{}); ok {
		for _, v := range a {
			if s, ok := v.(string); ok {
				list = append(list, s)
			}
		}
	}
	ctx = context.WithValue(ctx, NoCloudMember, o)
	return context.WithValue(ctx, NoCloudMemberAccess, list)
}

// CheckMember lets a member's call through as the member (base methods) or as its owner (methods of
// its access), and refuses anything else as MEMBER_GATE says. Anyone else passes untouched.
func CheckMember(ctx context.Context, log *zap.Logger, method string) (context.Context, error) {
	owner := MemberOwner(ctx)
	if owner == "" {
		return ctx, nil
	}
	access, _ := ctx.Value(NoCloudMemberAccess).([]string)
	for _, a := range access {
		if memberScopes[a][method] {
			return asOwner(ctx, owner), nil
		}
	}
	if memberBase[method] || memberGate == "off" {
		return ctx, nil
	}
	member, _ := ctx.Value(NoCloudAccount).(string)
	if memberGate == "log" {
		log.Warn("Member call outside its access", zap.String("method", method),
			zap.String("member", member), zap.String("owner", owner))
		return ctx, nil
	}
	return ctx, status.Error(codes.PermissionDenied, "Your organization did not allow you "+method)
}

func asOwner(ctx context.Context, owner string) context.Context {
	member, _ := ctx.Value(NoCloudAccount).(string)
	ctx = context.WithValue(ctx, NoCloudActingMember, member)
	ctx = context.WithValue(ctx, NoCloudMember, nil)
	ctx = context.WithValue(ctx, NoCloudAccount, owner)
	md, _ := metadata.FromOutgoingContext(ctx)
	md = md.Copy()
	md.Set(NOCLOUD_ACCOUNT_CLAIM, owner)
	return metadata.NewOutgoingContext(ctx, md)
}
