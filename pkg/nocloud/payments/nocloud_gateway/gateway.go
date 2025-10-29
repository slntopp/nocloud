package nocloud_gateway

import (
	"context"
	"fmt"
	pb "github.com/slntopp/nocloud-proto/billing"
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
	return g.baseHost + "/billing/payments" + fmt.Sprintf("/%s/view", inv.Uuid), nil
}
