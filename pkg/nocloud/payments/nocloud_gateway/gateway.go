package nocloud_gateway

import (
	"context"
	pb "github.com/slntopp/nocloud-proto/billing"
	"github.com/slntopp/nocloud/pkg/nocloud/payments/types"
)

type NcGateway struct {
}

func NewNoCloudGateway() *NcGateway {
	return &NcGateway{}
}

func (g *NcGateway) CreateInvoice(_ context.Context, _ *pb.Invoice) error {

	return nil
}

func (g *NcGateway) UpdateInvoice(_ context.Context, _ *pb.Invoice, _ *pb.Invoice) error {
	return nil
}

func (g *NcGateway) PaymentURI(_ context.Context, _ *pb.Invoice) (string, error) {
	return "", nil
}

func (g *NcGateway) AddClient(_ types.CreateUserParams) (int, error) {
	return 0, nil
}
