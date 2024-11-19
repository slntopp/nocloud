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

type InvoicesManager interface {
	CreateInvoice(inv *pb.Invoice) error
	UpdateInvoiceStatus(id string, newStatus pb.BillingStatus) (*pb.Invoice, error)
	InvoicesController() graph.InvoicesController
}

type tokenMaker interface {
	MakeToken(string) (string, error)
}

type invoicesManager struct {
	inv     billingconnect.BillingServiceClient
	invCtrl graph.InvoicesController
	tm      tokenMaker
}

func NewInvoicesManager(inv billingconnect.BillingServiceClient, invCtrl graph.InvoicesController, tm tokenMaker) InvoicesManager {
	return &invoicesManager{inv: inv, invCtrl: invCtrl, tm: tm}
}

func (i *invoicesManager) CreateInvoice(inv *pb.Invoice) error {
	req := connect.NewRequest(&pb.CreateInvoiceRequest{
		Invoice:     inv,
		IsSendEmail: true,
	})
	token, err := i.tm.MakeToken(schema.ROOT_ACCOUNT_KEY)
	if err != nil {
		return err
	}
	req.Header().Set("Authorization", "Bearer "+token)
	_, err = i.inv.CreateInvoice(context.WithValue(context.Background(), nocloud.NoCloudAccount, schema.ROOT_ACCOUNT_KEY), req)
	return err
}

func (i *invoicesManager) UpdateInvoiceStatus(id string, newStatus pb.BillingStatus) (*pb.Invoice, error) {
	req := connect.NewRequest(&pb.UpdateInvoiceStatusRequest{
		Status: newStatus,
		Uuid:   id,
	})
	token, err := i.tm.MakeToken(schema.ROOT_ACCOUNT_KEY)
	if err != nil {
		return nil, err
	}
	req.Header().Set("Authorization", "Bearer "+token)
	inv, err := i.inv.UpdateInvoiceStatus(context.WithValue(context.Background(), nocloud.NoCloudAccount, schema.ROOT_ACCOUNT_KEY), req)
	if err != nil {
		return nil, err
	}
	return inv.Msg, nil
}

func (i *invoicesManager) InvoicesController() graph.InvoicesController {
	return i.invCtrl
}
