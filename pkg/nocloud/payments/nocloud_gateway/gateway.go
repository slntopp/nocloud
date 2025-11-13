package nocloud_gateway

import (
	"context"
	"fmt"
	pb "github.com/slntopp/nocloud-proto/billing"
	"github.com/slntopp/nocloud/pkg/nocloud/auth"
	"net/url"
)

type NcGateway struct {
	baseHost string
}

func NewNoCloudGateway(baseHost string) *NcGateway {
	return &NcGateway{
		baseHost: baseHost,
	}
}

func (g *NcGateway) CreateInvoice(_ context.Context, _ *pb.Invoice, _ ...bool) error {

	return nil
}

func (g *NcGateway) UpdateInvoice(_ context.Context, _ *pb.Invoice, _ pb.BillingStatus, _ bool) error {
	return nil
}

func (g *NcGateway) PaymentURI(_ context.Context, inv *pb.Invoice) (string, error) {
	if inv == nil {
		return "", fmt.Errorf("invoice is nil")
	}
	token, _ := auth.MakeToken(inv.Account)
	return g.baseHost + "/billing/payments" + fmt.Sprintf("/%s/view?access_token=%s", inv.Uuid, url.QueryEscape(token)), nil
}
