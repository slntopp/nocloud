package whmcs_gateway

import (
	"context"
	"fmt"
	pb "github.com/slntopp/nocloud-proto/billing"
	rpb "github.com/slntopp/nocloud-proto/registry/accounts"
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
	return nil, fmt.Errorf("not found")
}

func (g *WhmcsGateway) GetAccountByWhmcsId(whmcsUserId int) (*rpb.Account, error) {
	return nil, fmt.Errorf("not implemented")
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
