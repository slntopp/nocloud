package whmcs_gateway

import (
	"context"
	"encoding/json"
	"fmt"
	validator "github.com/go-playground/validator/v10"
	"github.com/google/go-querystring/query"
	pb "github.com/slntopp/nocloud-proto/billing"
	"github.com/slntopp/nocloud/pkg/graph"
	"github.com/slntopp/nocloud/pkg/nocloud/payments/types"
	"google.golang.org/protobuf/types/known/structpb"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type NoCloudInvoicesManager interface {
	CreateInvoice(inv *pb.Invoice) error
	UpdateInvoiceStatus(id string, newStatus pb.BillingStatus) error
	InvoicesController() *graph.InvoicesController
}

type WhmcsGateway struct {
	apiUsername string
	apiPassword string
	baseUrl     string
	trustedIP   string

	accounts *graph.AccountsController
	invMan   NoCloudInvoicesManager
}

func NewWhmcsGateway(data WhmcsData, acc *graph.AccountsController, invMan NoCloudInvoicesManager) *WhmcsGateway {
	return &WhmcsGateway{
		apiUsername: data.WhmcsUser,
		apiPassword: data.WhmcsPassHash,
		baseUrl:     data.WhmcsBaseUrl,
		trustedIP:   data.TrustedIP,
		accounts:    acc,
		invMan:      invMan,
	}
}

func sendRequestToWhmcs[T any](method string, url string, body io.Reader) (T, error) {
	var result T

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return result, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return result, err
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return result, err
	}

	resultChecker := struct {
		Result  string `json:"result"`
		Message string `json:"message"`
	}{}
	err = json.Unmarshal(b, &resultChecker)
	if err != nil {
		return result, fmt.Errorf("failed to unmarshal response to result checker struct: %w. Body: %s", err, string(b))
	}
	if resp.StatusCode != 200 || resultChecker.Result != "success" {
		return result, fmt.Errorf("failed to perform action: %s. Response body: %+v", resp.Status, resultChecker)
	}

	// Result
	err = json.Unmarshal(b, &result)
	if err != nil {
		return result, fmt.Errorf("failed to unmarshal response to result struct: %w", err)
	}

	return result, nil
}

func (g *WhmcsGateway) CreateInvoice(ctx context.Context, inv *pb.Invoice) error {
	reqUrl, err := url.Parse(g.baseUrl)
	if err != nil {
		return err
	}

	account, err := g.accounts.Get(ctx, inv.GetAccount())
	if err != nil {
		return err
	}

	userId, ok := g.getWhmcsUser(account.Account)
	if !ok {
		return fmt.Errorf("whmcs user not found")
	}

	// TODO: review taxed field
	query, err := g.buildCreateInvoiceQueryBase(inv, userId)
	if err != nil {
		return err
	}
	for i, item := range inv.GetItems() {
		query.Set(fmt.Sprintf("itemdescription%d", i+1), item.GetDescription())
		query.Set(fmt.Sprintf("itemamount%d", i+1), fmt.Sprintf("%.2f", item.GetPrice()*float64(item.GetAmount())))
		query.Set(fmt.Sprintf("itemtaxed%d", i+1), "0")
	}

	invResp, err := sendRequestToWhmcs[InvoiceResponse](http.MethodPost, reqUrl.String()+"?"+query.Encode(), nil)
	if err != nil {
		return err
	}

	// Set whmcs invoice id to invoice meta
	invoice, err := g.invMan.InvoicesController().Get(ctx, inv.GetUuid())
	if err != nil {
		return err
	}
	meta := invoice.GetMeta()
	if meta == nil {
		meta = make(map[string]*structpb.Value)
	}
	meta[invoiceIdField] = structpb.NewNumberValue(float64(invResp.InvoiceId))
	invoice.Meta = meta
	if _, err := g.invMan.InvoicesController().Update(ctx, invoice); err != nil {
		return err
	}

	return nil
}

func (g *WhmcsGateway) UpdateInvoice(ctx context.Context, inv *pb.Invoice, old *pb.Invoice) error {
	reqUrl, err := url.Parse(g.baseUrl)
	if err != nil {
		return err
	}

	id, ok := inv.GetMeta()[invoiceIdField]
	if !ok {
		return fmt.Errorf("failed to get invoice id from meta")
	}
	body := g.buildUpdateInvoiceQueryBase(int(id.GetNumberValue()))

	whmcsInv, err := g.GetInvoice(ctx, int(id.GetNumberValue()))
	if err != nil {
		return err
	}

	// Process params
	if inv.Payment != old.Payment {
		body.DatePaid = ptr(time.Unix(inv.Payment, 0).Format("2006-01-02 15:04:05"))
	}
	if inv.Deadline != old.Deadline {
		body.DueDate = ptr(time.Unix(inv.Deadline, 0).Format("2006-01-02"))
	}
	if inv.Status != old.Status {
		body.Status = ptr(statusToWhmcs(inv.Status))
	}
	if inv.Created != old.Created {
		body.Date = ptr(time.Unix(inv.Created, 0).Format("2006-01-02"))
	}
	if inv.GetMeta()["note"].GetStringValue() != old.GetMeta()["note"].GetStringValue() {
		body.Notes = ptr(inv.GetMeta()["note"].GetStringValue())
	}

	// Delete all existing invoice items
	toDelete := make([]int, 0)
	for _, item := range whmcsInv.Items.Items {
		toDelete = append(toDelete, item.Id)
	}
	body.DeleteLineIds = toDelete

	// From new list of items
	description := make(map[int]string)
	amount := make(map[int]floatAsString)
	taxed := make(map[int]bool)
	for i, item := range inv.GetItems() {
		description[i] = item.GetDescription()
		amount[i] = floatAsString(item.GetPrice() * float64(item.GetAmount()))
		taxed[i] = false
	}
	body.NewItemDescription = description
	body.NewItemAmount = amount
	body.NewItemTaxed = taxed

	q, err := query.Values(body)
	if err != nil {
		return err
	}
	for i := range inv.GetItems() {
		q.Set(fmt.Sprintf("newitemdescription[%d]", i), description[i])
		q.Set(fmt.Sprintf("newitemamount[%d]", i), fmt.Sprintf("%f", amount[i]))
		q.Set(fmt.Sprintf("newitemtaxed[%d]", i), strconv.FormatBool(taxed[i]))
	}
	_, err = sendRequestToWhmcs[InvoiceResponse](http.MethodPost, reqUrl.String()+"?"+q.Encode(), nil)
	if err != nil {
		return err
	}
	return nil
}

func (g *WhmcsGateway) PaymentURI(ctx context.Context, inv *pb.Invoice) (string, error) {
	reqUrl, err := url.Parse(g.baseUrl)
	if err != nil {
		return "", err
	}

	acc, err := g.accounts.Get(ctx, inv.GetAccount())
	if err != nil {
		return "", err
	}

	userId, ok := g.getWhmcsUser(acc.Account)
	if !ok {
		return "", fmt.Errorf("failed to get whmcs user")
	}
	invId, ok := g.getWhmcsInvoice(inv)
	if !ok {
		return "", fmt.Errorf("failed to get whmcs invoice")
	}

	body := g.buildPaymentURIQueryBase(userId)
	q, err := query.Values(body)
	if err != nil {
		return "", err
	}
	resp, err := sendRequestToWhmcs[PaymentURIResponse](http.MethodPost, reqUrl.String()+"?"+q.Encode(), nil)
	if err != nil {
		return "", err
	}

	return g.buildPaymentURI(invId, resp.AccessToken), nil
}

func (g *WhmcsGateway) GetInvoice(ctx context.Context, whmcsInvoiceId int) (Invoice, error) {
	reqUrl, err := url.Parse(g.baseUrl)
	if err != nil {
		return Invoice{}, err
	}

	query, err := g.buildGetInvoiceQueryBase(whmcsInvoiceId)
	if err != nil {
		return Invoice{}, err
	}

	invResp, err := sendRequestToWhmcs[Invoice](http.MethodPost, reqUrl.String()+"?"+query.Encode(), nil)
	if err != nil {
		return Invoice{}, err
	}

	return invResp, nil
}

func (g *WhmcsGateway) AddClient(params types.CreateUserParams) (int, error) {
	reqUrl, err := url.Parse(g.baseUrl)
	if err != nil {
		return 0, err
	}

	// Validate with validator
	if err = validator.New().Struct(params); err != nil {
		return 0, fmt.Errorf("validation error: %w", err)
	}
	// Other validation
	if params.Currency == nil {
		return 0, fmt.Errorf("failed to validate currency: currency is required")
	}
	var currencyCode int
	currencies, err := g.fetchCurrencies()
	if err != nil {
		return 0, fmt.Errorf("failed to validate currency: %w", err)
	}
	var found bool
	for _, c := range currencies {
		if c.Code == params.Currency.Title {
			found = true
			currencyCode = c.Id
			break
		}
	}
	if !found {
		return 0, fmt.Errorf("failed to validate currency: can't syncronize currency. Not found")
	}

	// User creation
	q, err := g.buildCreateUserQueryBase(params, currencyCode)
	if err != nil {
		return 0, err
	}

	clientResp, err := sendRequestToWhmcs[AddClientResponse](http.MethodPost, reqUrl.String()+"?"+q.Encode(), nil)
	if err != nil {
		return 0, fmt.Errorf("failed to create client: %w", err)
	}

	id, err := strconv.Atoi(clientResp.ClientID)
	if err != nil {
		return 0, fmt.Errorf("client is created but failed to convert client id to int: %w", err)
	}

	// Updating payment method for company client and individual clients. See whmcs module
	var vals url.Values
	if params.CompanyName != "" {
		vals, err = g.buildUpdateClientPaymentMethodQueryBase(id, "schetfacturagate")
	} else {
		vals, err = g.buildUpdateClientPaymentMethodQueryBase(id, "begateway")
	}
	if err != nil {
		return id, err
	}
	_, err = sendRequestToWhmcs[struct{}](http.MethodPost, reqUrl.String()+"?"+vals.Encode(), nil)
	if err != nil {
		return id, fmt.Errorf("failed to update client payment method: %w", err)
	}

	return id, nil
}

func (g *WhmcsGateway) _SyncWhmcsInvoice(ctx context.Context, invoiceId int) error {
	return g.syncWhmcsInvoice(ctx, invoiceId)
}

func (g *WhmcsGateway) syncWhmcsInvoice(ctx context.Context, invoiceId int) error {
	whmcsInv, err := g.GetInvoice(ctx, invoiceId)
	if err != nil {
		return fmt.Errorf("error syncWhmcsInvoice: %w", err)
	}
	inv, err := g.getInvoiceByWhmcsId(invoiceId)
	if err != nil {
		return fmt.Errorf("error syncWhmcsInvoice: %w", err)
	}

	if inv.Status != statusToNoCloud(whmcsInv.Status) {
		inv.Status = statusToNoCloud(whmcsInv.Status)
		if err := g.invMan.UpdateInvoiceStatus(inv.GetUuid(), inv.Status); err != nil {
			return fmt.Errorf("error syncWhmcsInvoice: %w", err)
		}
	}
	if !strings.Contains(whmcsInv.DatePaid, "0000-00-00") {
		t, err := time.Parse("2006-01-02 15:04:05", whmcsInv.DatePaid)
		if err != nil {
			return fmt.Errorf("error syncWhmcsInvoice: %w", err)
		}
		inv.Payment = t.Unix()
		inv.Processed = inv.Payment
	}
	if !strings.Contains(whmcsInv.DueDate, "0000-00-00") && inv.Deadline == 0 {
		t, err := time.Parse("2006-01-02", whmcsInv.DueDate)
		if err != nil {
			return fmt.Errorf("error syncWhmcsInvoice: %w", err)
		}
		inv.Deadline = t.Unix()
	}
	if !strings.Contains(whmcsInv.Date, "0000-00-00") && inv.Created == 0 {
		t, err := time.Parse("2006-01-02", whmcsInv.Date)
		if err != nil {
			return fmt.Errorf("error syncWhmcsInvoice: %w", err)
		}
		inv.Created = t.Unix()
	}
	inv.Total = float64(whmcsInv.Total)
	// TODO: sync invoice number too. Have to refactor number logic in invoice controller
	// TODO: Update items too or refactor
	//inv.Items = []*pb.Item{}
	//for _, item := range whmcsInv.Items.Items {
	//	inv.Items = append(inv.Items, &pb.Item{
	//		Description: item.Description,
	//		Amount:      1,
	//		Price:       float64(item.Amount),
	//		Unit:        "Pcs",
	//	})
	//}
	meta := inv.GetMeta()
	meta["note"] = structpb.NewStringValue(whmcsInv.Notes)
	inv.Meta = meta

	if _, err = g.invMan.InvoicesController().Update(ctx, &graph.Invoice{
		Invoice: inv,
	}); err != nil {
		return fmt.Errorf("error syncWhmcsInvoice: failed to update invoice: %w", err)
	}
	return nil
}

func (g *WhmcsGateway) fetchCurrencies() ([]Currency, error) {
	reqUrl, err := url.Parse(g.baseUrl)
	if err != nil {
		return nil, err
	}

	q, err := g.buildGetCurrenciesQueryBase()
	if err != nil {
		return nil, err
	}
	resp, err := sendRequestToWhmcs[GetCurrenciesResponse](http.MethodPost, reqUrl.String()+"?"+q.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch currencies: %w", err)
	}

	if resp.Currencies == nil {
		return nil, fmt.Errorf("currencies is empty or nil")
	}

	return resp.Currencies, nil
}
