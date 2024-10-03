package billing

import (
	"context"
	pb "github.com/slntopp/nocloud-proto/billing"
)

type PaymentGateway interface {
	CreateInvoice(context.Context, *pb.Invoice) error
	// UpdateInvoice 1. Context; 2. New invoice; 3. Old invoice
	UpdateInvoice(context.Context, *pb.Invoice, *pb.Invoice) error
	PaymentURI(context.Context, *pb.Invoice) (string, error)
}
