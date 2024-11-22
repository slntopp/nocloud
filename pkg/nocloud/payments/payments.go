package payments

import (
	"context"
	"fmt"
	pb "github.com/slntopp/nocloud-proto/billing"
	"github.com/slntopp/nocloud/pkg/graph"
	"github.com/slntopp/nocloud/pkg/nocloud/payments/nocloud_gateway"
	"github.com/slntopp/nocloud/pkg/nocloud/payments/types"
	"github.com/slntopp/nocloud/pkg/nocloud/payments/whmcs_gateway"
	"google.golang.org/grpc/metadata"
	"net/http"
)

type PaymentGateway interface {
	CreateInvoice(context.Context, *pb.Invoice) error
	UpdateInvoice(context.Context, *pb.Invoice, *pb.Invoice) error
	PaymentURI(context.Context, *pb.Invoice) (string, error)
	//AddClient(types.CreateUserParams) (int, error)
}

func GetGatewayCallbackValue(ctx context.Context, h ...http.Header) bool {
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		if v, ok := md[string(types.GatewayCallback)]; ok {
			fmt.Println("FOUND IN INCOMING METADATA")
			return v[0] == "true"
		}
	}
	md, ok = metadata.FromOutgoingContext(ctx)
	if ok {
		if v, ok := md[string(types.GatewayCallback)]; ok {
			fmt.Println("FOUND IN OUTGOING METADATA")
			return v[0] == "true"
		}
	}
	if len(h) > 0 {
		header := h[0].Get(string(types.GatewayCallback))
		if header != "" {
			fmt.Println("FOUND IN HEADER METADATA")
			return header == "true"
		}
	}
	fmt.Println("RETURNING CTX VALUE")
	val, _ := ctx.Value(types.GatewayCallback).(bool)
	return val
}

var (
	_registered bool

	whmcsData whmcs_gateway.WhmcsData

	accountController graph.AccountsController
	invoicesManager   whmcs_gateway.NoCloudInvoicesManager
)

func RegisterGateways(whmcs whmcs_gateway.WhmcsData,
	accountCtrl graph.AccountsController,
	invoicesMan whmcs_gateway.NoCloudInvoicesManager) {
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
		return whmcs_gateway.NewWhmcsGateway(whmcsData, accountController, invoicesManager)
	default:
		return whmcs_gateway.NewWhmcsGateway(whmcsData, accountController, invoicesManager)
	}
}
