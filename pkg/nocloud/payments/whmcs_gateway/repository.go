package whmcs_gateway

import (
	"context"
	"fmt"
	pb "github.com/slntopp/nocloud-proto/billing"
	rpb "github.com/slntopp/nocloud-proto/registry/accounts"
	"github.com/slntopp/nocloud/pkg/nocloud/schema"
	ps "github.com/slntopp/nocloud/pkg/pubsub"
	"google.golang.org/protobuf/types/known/structpb"
)

const invoiceIdField = "whmcs_invoice_id"
const userIdField = "whmcs_id"

func (g *WhmcsGateway) getInvoiceByWhmcsId(whmcsInvoiceId int) (*pb.Invoice, error) {
	invoices, err := g.invMan.InvoicesController().List(context.Background(), "")
	if err != nil {
		return nil, err
	}
	for _, inv := range invoices {
		id, ok := inv.GetMeta()[invoiceIdField]
		if !ok {
			continue
		}
		if int(id.GetNumberValue()) == whmcsInvoiceId {
			return inv.Invoice, nil
		}
	}
	return nil, ps.NoNackErr(ErrNotFound)
}

func (g *WhmcsGateway) GetInvoiceByWhmcsId(whmcsInvoiceId int) (*pb.Invoice, error) {
	return g.getInvoiceByWhmcsId(whmcsInvoiceId)
}

func (g *WhmcsGateway) GetAccountByWhmcsId(whmcsUserId int) (*rpb.Account, error) {
	acc, _, _, err := g.accounts.ListImproved(context.Background(), schema.ROOT_ACCOUNT_KEY, 100, 0, 0, "", "", map[string]*structpb.Value{
		fmt.Sprintf("data.%s", userIdField): structpb.NewNumberValue(float64(whmcsUserId)),
	})
	if err != nil {
		return nil, err
	}
	if len(acc) == 0 {
		return nil, ErrNotFound
	}
	if len(acc) > 1 {
		return nil, fmt.Errorf("multiple accounts found")
	}
	return acc[0].Account, nil
}

func (g *WhmcsGateway) getWhmcsUser(acc *rpb.Account) (int, bool) {
	data, ok := acc.GetData().AsMap()[userIdField]
	if !ok {
		return 0, false
	}
	value, ok := data.(float64)
	if !ok {
		return 0, false
	}
	casted := int(value)
	if casted == 0 {
		return 0, false
	}
	return casted, true
}

func (g *WhmcsGateway) getWhmcsInvoice(inv *pb.Invoice) (int, bool) {
	id, ok := inv.GetMeta()[invoiceIdField]
	if !ok {
		return 0, false
	}
	value := id.GetNumberValue()
	casted := int(value)
	if casted == 0 {
		return 0, false
	}
	return casted, true
}

const getAccountByWhmcsId = `

`
