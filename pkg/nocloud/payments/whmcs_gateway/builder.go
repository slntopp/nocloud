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
func (g *WhmcsGateway) buildCreateInvoiceQueryBase(inv *pb.Invoice, whmcsUserId int, _sendEmail bool) (url.Values, error) {

	var sendEmail string = "1"
	if !_sendEmail {
		sendEmail = "0"
	}

	res, err := query.Values(CreateInvoiceQuery{
		Action:          "CreateInvoice",
		Username:        g.apiUsername,
		Password:        g.apiPassword,
		UserId:          fmt.Sprintf("%d", whmcsUserId),
		Status:          statusToWhmcs(inv.Status),
		SendInvoice:     sendEmail,
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

func (g *WhmcsGateway) buildPaymentURIQueryBase(clientId int) PaymentURIQuery {
	return PaymentURIQuery{
		Action:       "CreateSsoToken",
		Username:     g.apiUsername,
		Password:     g.apiPassword,
		ResponseType: "json",
		ClientID:     clientId,
	}
}

func (g *WhmcsGateway) buildPaymentURI(invoiceId int, token string) string {
	u, err := url.Parse(g.baseUrl)
	if err != nil {
		return ""
	}
	return fmt.Sprintf("%s?access_token=%s&id=%d", u.Scheme+"://"+u.Host+"/viewinvoice.php", token, invoiceId)
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
