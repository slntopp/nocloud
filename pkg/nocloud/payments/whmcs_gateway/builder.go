package whmcs_gateway

import (
	"encoding/base64"
	"encoding/json"
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

type CreateUserParams struct {
	Firstname       string       `json:"firstname" validate:"required"`
	Lastname        string       `json:"lastname" validate:"required"`
	Email           string       `json:"email" validate:"required,email"`
	Language        string       `json:"language"`
	Country         string       `json:"country" validate:"omitempty,iso3166_1_alpha2"`
	PhoneNumber     string       `json:"phonenumber" validate:"required,e164"`
	Password        string       `json:"password2" validate:"required"`
	Currency        *pb.Currency `json:"currency"`
	CompanyName     string       `json:"companyname"`
	City            string       `json:"city" validate:"required"`
	Postcode        string       `json:"postcode" validate:"required"`
	Address         string       `json:"address1" validate:"required"`
	NoEmail         bool         `json:"noemail"`
	AccountNumber   string       `json:"account_number"`
	BankName        string       `json:"bankname"`
	BIC             string       `json:"bic"`
	CheckingAccount string       `json:"checking_account"`
	TaxID           string       `json:"tax_id"`
}

func (g *WhmcsGateway) buildCreateUserQueryBase(params CreateUserParams, whmcsCurrencyCode int) (url.Values, error) {
	type CustomFields struct {
		AccountNumber   string `json:"3212,omitempty"`
		CheckingAccount string `json:"3409,omitempty"`
		BankName        string `json:"3213,omitempty"`
		BIC             string `json:"3408,omitempty"`
		CustomerType    string `json:"3629,omitempty"`
		TaxID           string `json:"11,omitempty"`
	}
	cf := CustomFields{
		AccountNumber:   params.AccountNumber,
		CheckingAccount: params.CheckingAccount,
		BankName:        params.BankName,
		BIC:             params.BIC,
	}

	if params.CompanyName != "" {
		cf.CustomerType = "Юридическое лицо"
	} else {
		cf.CustomerType = "Физическое лицо"
	}

	const Poland = "PL"
	if params.Country == Poland {
		cf.TaxID = params.TaxID
	}

	jsonData, err := json.Marshal(cf)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal custom fields: %w", err)
	}
	base64CustomFields := base64.StdEncoding.EncodeToString(jsonData)

	res, err := query.Values(AddClientQuery{
		Action:       "AddClient",
		ResponseType: "json",
		Username:     g.apiUsername,
		Password:     g.apiPassword,

		Firstname:    params.Firstname,
		Lastname:     params.Lastname,
		Email:        params.Email,
		Language:     params.Language,
		Country:      params.Country,
		PhoneNumber:  params.PhoneNumber,
		Password2:    params.Password,
		Currency:     whmcsCurrencyCode,
		CompanyName:  params.CompanyName,
		City:         params.City,
		Postcode:     params.Postcode,
		Address:      params.Address,
		CustomFields: base64CustomFields,
		NoEmail:      params.NoEmail,
		TaxID:        params.TaxID,
	})
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (g *WhmcsGateway) buildUpdateClientPaymentMethodQueryBase(clientId int, newMethod string) (url.Values, error) {
	res, err := query.Values(UpdateClientPaymentMethodQuery{
		Action:        "UpdateClient",
		ResponseType:  "json",
		Username:      g.apiUsername,
		Password:      g.apiPassword,
		ClientId:      clientId,
		PaymentMethod: newMethod,
	})
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (g *WhmcsGateway) buildGetCurrenciesQueryBase() (url.Values, error) {
	res, err := query.Values(GetCurrenciesRequest{
		Action:       "GetCurrencies",
		ResponseType: "json",
		Username:     g.apiUsername,
		Password:     g.apiPassword,
	})
	if err != nil {
		return nil, err
	}
	return res, nil
}
