package invoices_manager

import (
	"connectrpc.com/connect"
	"context"
	pb "github.com/slntopp/nocloud-proto/billing"
	"github.com/slntopp/nocloud-proto/billing/billingconnect"
	"github.com/slntopp/nocloud/pkg/graph"
	"github.com/slntopp/nocloud/pkg/nocloud"
	"github.com/slntopp/nocloud/pkg/nocloud/schema"
)

type InvoicesManager struct {
	inv     billingconnect.BillingServiceClient
	invCtrl *graph.InvoicesController
}

func NewInvoicesManager(inv billingconnect.BillingServiceClient, invCtrl *graph.InvoicesController) *InvoicesManager {
	return &InvoicesManager{inv: inv, invCtrl: invCtrl}
}

func (i *InvoicesManager) CreateInvoice(inv *pb.Invoice) error {
	req := connect.NewRequest(&pb.CreateInvoiceRequest{
		Invoice:     inv,
		IsSendEmail: true,
	})
	_, err := i.inv.CreateInvoice(context.WithValue(context.Background(), nocloud.NoCloudAccount, schema.ROOT_ACCOUNT_KEY), req)
	return err
}

func (i *InvoicesManager) UpdateInvoiceStatus(id string, newStatus pb.BillingStatus) error {
	req := connect.NewRequest(&pb.UpdateInvoiceStatusRequest{
		Status: newStatus,
		Uuid:   id,
	})
	_, err := i.inv.UpdateInvoiceStatus(context.WithValue(context.Background(), nocloud.NoCloudAccount, schema.ROOT_ACCOUNT_KEY), req)
	return err
}

func (i *InvoicesManager) InvoicesController() *graph.InvoicesController {
	return i.invCtrl
}
