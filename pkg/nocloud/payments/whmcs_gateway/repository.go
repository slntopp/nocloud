package whmcs_gateway

import (
	"context"
	"fmt"
	pb "github.com/slntopp/nocloud-proto/billing"
)

func (g *WhmcsGateway) getInvoiceByWhmcsId(whmcsInvoiceId int) (*pb.Invoice, error) {
	invoices, err := g.invoices.List(context.Background())
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
