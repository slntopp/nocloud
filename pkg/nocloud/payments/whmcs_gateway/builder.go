package whmcs_gateway

import (
	"fmt"
	"github.com/google/go-querystring/query"
	pb "github.com/slntopp/nocloud-proto/billing"
	"net/url"
	"time"
)

// TODO: review TaxRate, PaymentMethod, AutoApplyCredit and other fields
func (g *WhmcsGateway) buildCreateInvoiceQueryBase(inv *pb.Invoice, whmcsUserId int) (url.Values, error) {
	res, err := query.Values(CreateInvoiceQuery{
		Action:          "CreateInvoice",
		Username:        g.apiUsername,
		Password:        g.apiPassword,
		UserId:          fmt.Sprintf("%d", whmcsUserId),
		Status:          statusToWhmcs(inv.Status),
		SendInvoice:     "1",
		PaymentMethod:   "mailin",
		TaxRate:         "10",
		Date:            time.Unix(inv.Created, 0).Format("2006-01-02"),
		DueDate:         time.Unix(inv.Deadline, 0).Format("2006-01-02"),
		AutoApplyCredit: "0",
		ResponseType:    "json",
	})
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (g *WhmcsGateway) buildGetInvoiceQueryBase(whmcsInvoiceId int) (url.Values, error) {
	res, err := query.Values(GetInvoiceQuery{
		Action:    "GetInvoice",
		InvoiceId: whmcsInvoiceId,
	})
	if err != nil {
		return nil, err
	}

	return res, nil
}
