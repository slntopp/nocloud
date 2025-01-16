package whmcs_gateway

import (
	"encoding/base64"
	"fmt"
	"github.com/google/go-querystring/query"
	pb "github.com/slntopp/nocloud-proto/billing"
	"net/url"
	"time"
)

func EncodeParam(s string) string {
	return url.QueryEscape(s)
}

func EncodeStringBase64(s string) string {
	return base64.StdEncoding.EncodeToString([]byte(s))
}

// TODO: review TaxRate, PaymentMethod, AutoApplyCredit and other fields
func (g *WhmcsGateway) buildCreateInvoiceQueryBase(inv *pb.Invoice, whmcsUserId int, _sendEmail bool, tax float64) (url.Values, error) {

	var sendEmail = "1"
	if !_sendEmail {
		sendEmail = "0"
	}

	var taxRate *floatAsString
	if tax > 0 {
		taxRate = ptr(floatAsString(tax))
	}

	res, err := query.Values(CreateInvoiceQuery{
		Action:          "CreateInvoice",
		Username:        g.apiUsername,
		Password:        g.apiPassword,
		UserId:          fmt.Sprintf("%d", whmcsUserId),
		Status:          statusToWhmcs(inv.Status),
		SendInvoice:     sendEmail,
		PaymentMethod:   nil,
		TaxRate:         taxRate,
		Notes:           inv.GetMeta()["note"].GetStringValue(),
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

func (g *WhmcsGateway) buildPaymentURIQueryBase(clientId int, invoiceId int) PaymentURIQuery {
	return PaymentURIQuery{
		Action:          "CreateSsoToken",
		Username:        g.apiUsername,
		Password:        g.apiPassword,
		ResponseType:    "json",
		ClientID:        clientId,
		Destination:     "sso:custom_redirect",
		SsoRedirectPath: fmt.Sprintf("viewinvoice.php?id=%d", invoiceId),
	}
}

func (g *WhmcsGateway) buildPaymentURI(_ int, data PaymentURIResponse) string {
	return data.RedirectUrl
}

func (g *WhmcsGateway) buildUpdateInvoiceQueryBase(whmcsInvoiceId int) UpdateInvoiceQuery {
	return UpdateInvoiceQuery{
		Action:       "UpdateInvoice",
		Username:     g.apiUsername,
		Password:     g.apiPassword,
		ResponseType: "json",
		InvoiceId:    whmcsInvoiceId,
	}
}

func (g *WhmcsGateway) buildGetInvoiceQueryBase(whmcsInvoiceId int) (url.Values, error) {
	res, err := query.Values(GetInvoiceQuery{
		Action:       "GetInvoice",
		InvoiceId:    whmcsInvoiceId,
		ResponseType: "json",
		Username:     g.apiUsername,
		Password:     g.apiPassword,
	})
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (g *WhmcsGateway) buildGetInvoicesQueryBase() GetInvoicesQuery {
	return GetInvoicesQuery{
		Action:       "GetInvoices",
		ResponseType: "json",
		Username:     g.apiUsername,
		Password:     g.apiPassword,
	}
}

func (g *WhmcsGateway) buildAddPaymentQueryBase(whmcsInvoiceId int) AddPaymentQuery {
	return AddPaymentQuery{
		Action:       "AddInvoicePayment",
		Username:     g.apiUsername,
		Password:     g.apiPassword,
		ResponseType: "json",
		InvoiceId:    whmcsInvoiceId,
	}
}

func (g *WhmcsGateway) buildUpdateClientQueryBase(clientId int, notes string) (url.Values, error) {
	res, err := query.Values(UpdateClientQuery{
		Action:       "GetInvoice",
		ClientID:     clientId,
		ResponseType: "json",
		Username:     g.apiUsername,
		Password:     g.apiPassword,
		Notes:        notes,
	})
	if err != nil {
		return nil, err
	}

	return res, nil
}
