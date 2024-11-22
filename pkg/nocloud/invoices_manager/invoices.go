package invoices_manager

import (
	"connectrpc.com/connect"
	"context"
	"fmt"
	pb "github.com/slntopp/nocloud-proto/billing"
	"github.com/slntopp/nocloud-proto/billing/billingconnect"
	"github.com/slntopp/nocloud/pkg/graph"
	"github.com/slntopp/nocloud/pkg/nocloud"
	"github.com/slntopp/nocloud/pkg/nocloud/payments"
	"github.com/slntopp/nocloud/pkg/nocloud/schema"
)

type InvoicesManager interface {
	CreateInvoice(ctx context.Context, inv *pb.Invoice) error
	UpdateInvoice(ctx context.Context, inv *pb.Invoice) error
	UpdateInvoiceStatus(ctx context.Context, id string, newStatus pb.BillingStatus) (*pb.Invoice, error)
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

func (i *invoicesManager) CreateInvoice(ctx context.Context, inv *pb.Invoice) error {
	req := connect.NewRequest(&pb.CreateInvoiceRequest{
		Invoice:     inv,
		IsSendEmail: true,
	})
	token, err := i.tm.MakeToken(schema.ROOT_ACCOUNT_KEY)
	if err != nil {
		return err
	}
	req.Header().Set("Authorization", "Bearer "+token)
	if val := ctx.Value(payments.GatewayCallback); val != nil {
		fmt.Println("VALUE: ", val)
		fmt.Println("SETTING HEADER")
		req.Header().Set(string(payments.GatewayCallback), "true")
	}
	_, err = i.inv.CreateInvoice(context.WithValue(ctx, nocloud.NoCloudAccount, schema.ROOT_ACCOUNT_KEY), req)
	return err
}

func (i *invoicesManager) UpdateInvoice(ctx context.Context, inv *pb.Invoice) error {
	req := connect.NewRequest(&pb.UpdateInvoiceRequest{
		Invoice:     inv,
		IsSendEmail: true,
	})
	token, err := i.tm.MakeToken(schema.ROOT_ACCOUNT_KEY)
	if err != nil {
		return err
	}
	req.Header().Set("Authorization", "Bearer "+token)
	fmt.Println("SETTING HEADER")
	req.Header().Set(string(payments.GatewayCallback), "true")
	if val := ctx.Value(payments.GatewayCallback); val != nil {
		fmt.Println("VALUE: ", val)
		fmt.Println("SETTING HEADER INSIDE")
		req.Header().Set(string(payments.GatewayCallback), "true")
	}
	_, err = i.inv.UpdateInvoice(context.WithValue(ctx, nocloud.NoCloudAccount, schema.ROOT_ACCOUNT_KEY), req)
	return err
}

func (i *invoicesManager) UpdateInvoiceStatus(ctx context.Context, id string, newStatus pb.BillingStatus) (*pb.Invoice, error) {
	req := connect.NewRequest(&pb.UpdateInvoiceStatusRequest{
		Status: newStatus,
		Uuid:   id,
	})
	token, err := i.tm.MakeToken(schema.ROOT_ACCOUNT_KEY)
	if err != nil {
		return nil, err
	}
	req.Header().Set("Authorization", "Bearer "+token)
	if val := ctx.Value(payments.GatewayCallback); val != nil {
		fmt.Println("VALUE: ", val)
		fmt.Println("SETTING HEADER")
		req.Header().Set(string(payments.GatewayCallback), "true")
	}
	inv, err := i.inv.UpdateInvoiceStatus(context.WithValue(ctx, nocloud.NoCloudAccount, schema.ROOT_ACCOUNT_KEY), req)
	if err != nil {
		return nil, err
	}
	return inv.Msg, nil
}

func (i *invoicesManager) InvoicesController() graph.InvoicesController {
	return i.invCtrl
}
