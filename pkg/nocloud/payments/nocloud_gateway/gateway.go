package nocloud_gateway

import (
	"context"
	pb "github.com/slntopp/nocloud-proto/billing"
)

type NcGateway struct {
}

func NewNoCloudGateway() *NcGateway {
	return &NcGateway{}
}

func (g *NcGateway) CreateInvoice(_ context.Context, _ *pb.Invoice, _ ...bool) error {

	return nil
}

func (g *NcGateway) UpdateInvoice(_ context.Context, _ *pb.Invoice, _ pb.BillingStatus) error {
	return nil
}

func (g *NcGateway) PaymentURI(_ context.Context, _ *pb.Invoice) (string, error) {
	return "", nil
}
