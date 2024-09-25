package billing

import (
	"context"
	pb "github.com/slntopp/nocloud-proto/billing"
)

type PaymentGateway interface {
	CreateInvoice(context.Context, *pb.Invoice) error
	UpdateInvoice(context.Context, *pb.Invoice) error
}
