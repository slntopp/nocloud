package payments

import (
	"context"
	pb "github.com/slntopp/nocloud-proto/billing"
	"github.com/slntopp/nocloud/pkg/graph"
	"github.com/slntopp/nocloud/pkg/nocloud/invoices_manager"
	"github.com/slntopp/nocloud/pkg/nocloud/payments/nocloud_gateway"
	"github.com/slntopp/nocloud/pkg/nocloud/payments/whmcs_gateway"
)

type PaymentGateway interface {
	CreateInvoice(context.Context, *pb.Invoice) error
	UpdateInvoice(context.Context, *pb.Invoice, *pb.Invoice) error
	PaymentURI(context.Context, *pb.Invoice) (string, error)
	//AddClient(types.CreateUserParams) (int, error)
}

type ContextKey string

const GatewayCallback = ContextKey("payment-gateway-callback")

var (
	_registered bool

	whmcsData whmcs_gateway.WhmcsData

	accountController graph.AccountsController
	invoicesManager   invoices_manager.InvoicesManager
)

func RegisterGateways(whmcs whmcs_gateway.WhmcsData,
	accountCtrl graph.AccountsController,
	invoicesMan invoices_manager.InvoicesManager) {
	if _registered {
		panic("payment gateways are already registered")
	}
	whmcsData = whmcs
	accountController = accountCtrl
	invoicesManager = invoicesMan
	_registered = true
}

func GetPaymentGateway(t string) PaymentGateway {
	if !_registered {
		panic("payment gateways are not registered")
	}
	switch t {
	case "nocloud":
		return nocloud_gateway.NewNoCloudGateway()
	case "whmcs":
		return whmcs_gateway.NewWhmcsGateway(whmcsData, accountController, &invoicesManager)
	default:
		return whmcs_gateway.NewWhmcsGateway(whmcsData, accountController, &invoicesManager)
	}
}
