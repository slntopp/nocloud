package registry

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/arangodb/go-driver"
	jwt "github.com/golang-jwt/jwt/v4"
	"github.com/slntopp/nocloud-proto/access"
	accountspb "github.com/slntopp/nocloud-proto/registry/accounts"
	"github.com/slntopp/nocloud/pkg/graph"
	"github.com/slntopp/nocloud/pkg/nocloud"
	"github.com/slntopp/nocloud/pkg/nocloud/schema"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	mcpPublicURL = "https://mcp.support.by/mcp"

	mcpOrdersKey     = "public_mcp:orders:%s:%s"
	mcpRebootsKey    = "public_mcp:reboots:%s:%s"
	mcpSpendDayKey   = "public_mcp:spend:day:%s:%s"
	mcpSpendMonthKey = "public_mcp:spend:month:%s:%s"
)

func (s *AccountsServiceServer) GetMcp(ctx context.Context, req *accountspb.GetMcpRequest) (*accountspb.GetMcpResponse, error) {
	acc, err := s.mcpAccount(ctx, req.GetUuid())
	if err != nil {
		return nil, err
	}
	if acc.GetMcp() == nil || acc.GetMcp().GetGen() == 0 {
		acc.Mcp = defaultMcp()
		if err := s.ctrl.Update(ctx, *acc, map[string]any{"mcp": mcpDocument(acc.Mcp)}); err != nil {
			return nil, status.Error(codes.Internal, "failed to store MCP settings")
		}
	}
	return s.mcpResponse(ctx, acc)
}

func (s *AccountsServiceServer) SetMcp(ctx context.Context, req *accountspb.SetMcpRequest) (*accountspb.GetMcpResponse, error) {
	acc, err := s.mcpAccount(ctx, req.GetUuid())
	if err != nil {
		return nil, err
	}
	if acc.GetMcp() == nil || acc.GetMcp().GetGen() == 0 {
		acc.Mcp = defaultMcp()
	}
	if err := applyMcp(acc.Mcp, req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if err := s.ctrl.Update(ctx, *acc, map[string]any{"mcp": mcpDocument(acc.Mcp)}); err != nil {
		return nil, status.Error(codes.Internal, "failed to store MCP settings")
	}
	return s.mcpResponse(ctx, acc)
}

func (s *AccountsServiceServer) RotateMcp(ctx context.Context, req *accountspb.RotateMcpRequest) (*accountspb.GetMcpResponse, error) {
	acc, err := s.mcpAccount(ctx, req.GetUuid())
	if err != nil {
		return nil, err
	}
	if acc.GetMcp() == nil || acc.GetMcp().GetGen() == 0 {
		acc.Mcp = defaultMcp()
	} else {
		acc.Mcp.Gen++
	}
	if err := s.ctrl.Update(ctx, *acc, map[string]any{"mcp": mcpDocument(acc.Mcp)}); err != nil {
		return nil, status.Error(codes.Internal, "failed to store MCP settings")
	}
	return s.mcpResponse(ctx, acc)
}

func (s *AccountsServiceServer) mcpAccount(ctx context.Context, uuid string) (*graph.Account, error) {
	requestor, _ := ctx.Value(nocloud.NoCloudAccount).(string)
	if requestor == "" {
		return nil, status.Error(codes.Unauthenticated, "no account")
	}
	if uuid == "" || uuid == "me" {
		uuid = requestor
	}
	acc, err := s.ctrl.Get(ctx, uuid)
	if err != nil {
		return nil, status.Error(codes.NotFound, "Account not found")
	}
	if acc.GetAccountOwner() != "" {
		return nil, status.Error(codes.PermissionDenied, "MCP settings belong to the main account")
	}
	if requestor != acc.Key && !s.ca.HasAccess(ctx, requestor, driver.NewDocumentID(schema.ACCOUNTS_COL, acc.Key), access.Level_ADMIN) {
		return nil, status.Error(codes.PermissionDenied, "Not enough access rights")
	}
	return &acc, nil
}

func (s *AccountsServiceServer) mcpResponse(ctx context.Context, acc *graph.Account) (*accountspb.GetMcpResponse, error) {
	if len(s.SIGNING_KEY) == 0 {
		return nil, status.Error(codes.Internal, "signing key is not configured")
	}
	token, err := mcpToken(s.SIGNING_KEY, acc.Key, acc.GetMcp().GetGen())
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to build MCP token")
	}
	prec := currencyPrecision(acc)
	now := time.Now()
	return &accountspb.GetMcpResponse{
		Mcp:          acc.Mcp,
		Token:        token,
		Config:       mcpConfig(mcpPublicURL, token),
		SpentDay:     s.mcpMoney(ctx, fmt.Sprintf(mcpSpendDayKey, acc.Key, now.UTC().Format("2006-01-02")), prec),
		SpentMonth:   s.mcpMoney(ctx, fmt.Sprintf(mcpSpendMonthKey, acc.Key, now.UTC().Format("2006-01")), prec),
		OrdersToday:  int32(s.mcpInt(ctx, fmt.Sprintf(mcpOrdersKey, acc.Key, now.UTC().Format("2006-01-02")))),
		RebootsToday: int32(s.mcpInt(ctx, fmt.Sprintf(mcpRebootsKey, acc.Key, now.UTC().Format("2006-01-02")))),
	}, nil
}

func (s *AccountsServiceServer) mcpInt(ctx context.Context, key string) int64 {
	if s.rdb == nil {
		return 0
	}
	raw, err := s.rdb.Get(ctx, key).Result()
	if err != nil {
		return 0
	}
	n, _ := strconv.ParseInt(raw, 10, 64)
	return n
}

func (s *AccountsServiceServer) mcpMoney(ctx context.Context, key string, precision int) float64 {
	minor := s.mcpInt(ctx, key)
	if precision < 0 {
		precision = 0
	}
	return float64(minor) / math.Pow10(precision)
}

func defaultMcp() *accountspb.McpSettings {
	return &accountspb.McpSettings{Enabled: false, Pay: true, Reboot: true, Tickets: true, OrdersDay: 5, RebootsDay: 20, Gen: 1}
}

func currencyPrecision(acc *graph.Account) int {
	if acc.GetCurrency() != nil && acc.GetCurrency().GetPrecision() > 0 {
		return int(acc.GetCurrency().GetPrecision())
	}
	return 2
}

func mcpDocument(m *accountspb.McpSettings) map[string]any {
	doc := map[string]any{
		"enabled":     m.GetEnabled(),
		"pay":         m.GetPay(),
		"reboot":      m.GetReboot(),
		"tickets":     m.GetTickets(),
		"orders_day":  m.GetOrdersDay(),
		"reboots_day": m.GetRebootsDay(),
		"gen":         m.GetGen(),
	}
	if m.SpendDay != nil {
		doc["spend_day"] = *m.SpendDay
	}
	if m.SpendMonth != nil {
		doc["spend_month"] = *m.SpendMonth
	}
	return doc
}

func mcpToken(key []byte, account string, gen int32) (string, error) {
	claims := jwt.MapClaims{"account": account, "mcp": true, "gen": gen}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(key)
}

func mcpConfig(publicURL, token string) string {
	raw, _ := json.Marshal(map[string]any{
		"mcpServers": map[string]any{
			"support.by": map[string]any{
				"url":     publicURL,
				"headers": map[string]any{"Authorization": "Bearer " + token},
			},
		},
	})
	return string(raw)
}

func applyMcp(cur *accountspb.McpSettings, req *accountspb.SetMcpRequest) error {
	if req.GetUpdateMask() == nil || len(req.GetUpdateMask().GetPaths()) == 0 {
		return fmt.Errorf("update_mask is required")
	}
	for _, raw := range req.GetUpdateMask().GetPaths() {
		switch normMcpPath(raw) {
		case "enabled":
			if req.Enabled == nil {
				return fmt.Errorf("enabled is in the mask but has no value")
			}
			cur.Enabled = *req.Enabled
		case "pay":
			if req.Pay == nil {
				return fmt.Errorf("pay is in the mask but has no value")
			}
			cur.Pay = *req.Pay
		case "reboot":
			if req.Reboot == nil {
				return fmt.Errorf("reboot is in the mask but has no value")
			}
			cur.Reboot = *req.Reboot
		case "tickets":
			if req.Tickets == nil {
				return fmt.Errorf("tickets is in the mask but has no value")
			}
			cur.Tickets = *req.Tickets
		case "spend_day":
			if req.SpendDay == nil {
				cur.SpendDay = nil
				continue
			}
			if *req.SpendDay < 0 {
				return fmt.Errorf("spend_day cannot be negative")
			}
			v := *req.SpendDay
			cur.SpendDay = &v
		case "spend_month":
			if req.SpendMonth == nil {
				cur.SpendMonth = nil
				continue
			}
			if *req.SpendMonth < 0 {
				return fmt.Errorf("spend_month cannot be negative")
			}
			v := *req.SpendMonth
			cur.SpendMonth = &v
		case "orders_day":
			if req.OrdersDay == nil {
				return fmt.Errorf("orders_day is in the mask but has no value")
			}
			if *req.OrdersDay < 0 {
				return fmt.Errorf("orders_day cannot be negative")
			}
			cur.OrdersDay = *req.OrdersDay
		case "reboots_day":
			if req.RebootsDay == nil {
				return fmt.Errorf("reboots_day is in the mask but has no value")
			}
			if *req.RebootsDay < 0 {
				return fmt.Errorf("reboots_day cannot be negative")
			}
			cur.RebootsDay = *req.RebootsDay
		default:
			return fmt.Errorf("unknown MCP field %q", raw)
		}
	}
	return nil
}

func normMcpPath(path string) string {
	path = strings.TrimSpace(path)
	var b strings.Builder
	for i, r := range path {
		if r >= 'A' && r <= 'Z' {
			if i > 0 {
				b.WriteByte('_')
			}
			b.WriteRune(r + ('a' - 'A'))
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}
