package whmcs_gateway

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/go-querystring/query"
	"github.com/google/uuid"
	pb "github.com/slntopp/nocloud-proto/billing"
	"github.com/slntopp/nocloud/pkg/graph"
	"github.com/slntopp/nocloud/pkg/nocloud/payments/types"
	ps "github.com/slntopp/nocloud/pkg/pubsub"
	"google.golang.org/protobuf/types/known/structpb"
	"io"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"
)

type NoCloudInvoicesManager interface {
	CreateInvoice(ctx context.Context, inv *pb.Invoice) error
	UpdateInvoice(ctx context.Context, inv *pb.Invoice, ignoreNulls bool) error
	UpdateInvoiceStatus(ctx context.Context, id string, newStatus pb.BillingStatus) (*pb.Invoice, error)
	InvoicesController() graph.InvoicesController
}

type WhmcsGateway struct {
	m *sync.Mutex

	apiUsername string
	apiPassword string
	baseUrl     string
	trustedIP   string

	accounts   graph.AccountsController
	currencies graph.CurrencyController
	invMan     NoCloudInvoicesManager

	taxExcluded bool
}

var ErrNotFound = fmt.Errorf("not found")

const invoiceNotFoundApiMsg = "Invoice ID Not Found"

const getInvoicesBatchSize = 10000 // WHMCS throws memory limit exception nearly on 50000 batch size

const BalancePayMethod = "balancepay"

var balancePayMethod = BalancePayMethod

func NewWhmcsGateway(data WhmcsData, acc graph.AccountsController, curr graph.CurrencyController, invMan NoCloudInvoicesManager, whmcsTaxExcluded bool) *WhmcsGateway {
	return &WhmcsGateway{
		m:           &sync.Mutex{},
		apiUsername: data.WhmcsUser,
		apiPassword: data.WhmcsPassHash,
		baseUrl:     data.WhmcsBaseUrl,
		trustedIP:   data.TrustedIP,
		accounts:    acc,
		invMan:      invMan,
		taxExcluded: whmcsTaxExcluded,
		currencies:  curr,
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
		err = fmt.Errorf("failed to perform action: %s. Response body: %+v", resp.Status, resultChecker)
		if resultChecker.Message == invoiceNotFoundApiMsg {
			err = fmt.Errorf("%w. Embedded error: %w", err, ErrNotFound)
		}
		return result, err
	}

	// Result
	err = json.Unmarshal(b, &result)
	if err != nil {
		return result, fmt.Errorf("failed to unmarshal response to result struct: %w", err)
	}

	return result, nil
}

func (g *WhmcsGateway) CreateInvoice(ctx context.Context, inv *pb.Invoice, noEmail ...bool) error {
	if inv.Status == pb.BillingStatus_DRAFT || inv.Status == pb.BillingStatus_TERMINATED {
		return nil
	}

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

	var sendEmail = (inv.Status != pb.BillingStatus_DRAFT) || !(len(noEmail) > 0 && noEmail[0])

	tax := inv.GetTaxOptions().GetTaxRate() * 100
	taxed := "0"
	if tax > 0 {
		taxed = "1"
	}

	cur, err := g.currencies.Get(ctx, inv.GetCurrency().GetId())
	if err != nil {
		return err
	}

	oldNote := inv.GetMeta()["note"].GetStringValue()
	newNote := oldNote
	inv.GetMeta()["note"] = structpb.NewStringValue(newNote)
	q, err := g.buildCreateInvoiceQueryBase(inv, userId, sendEmail, tax)
	if err != nil {
		return err
	}
	for i, item := range inv.GetItems() {
		var price float64
		if g.taxExcluded {
			price = item.GetPrice() * float64(item.GetAmount())
		} else {
			price = item.GetPrice()*float64(item.GetAmount()) + item.GetPrice()*float64(item.GetAmount())*(tax/100)
		}
		price = graph.Round(price, cur.Precision, cur.Rounding)

		q.Set(fmt.Sprintf("itemdescription%d", i+1), item.GetDescription())
		q.Set(fmt.Sprintf("itemamount%d", i+1), fmt.Sprintf("%.2f", price))
		q.Set(fmt.Sprintf("itemtaxed%d", i+1), taxed)
	}

	invResp, err := sendRequestToWhmcs[InvoiceResponse](http.MethodPost, reqUrl.String()+"?"+q.Encode(), nil)
	if err != nil {
		return err
	}

	// Set whmcs invoice id to invoice meta
	invoice, err := g.invMan.InvoicesController().Get(ctx, inv.GetUuid())
	if err != nil {
		return ps.NoNackErr(err)
	}
	meta := invoice.GetMeta()
	if meta == nil {
		meta = make(map[string]*structpb.Value)
	}

	// If invoice had whmcs invoice id, then cancel old invoice
	if invoiceId, ok := meta[invoiceIdField]; ok {
		updQuery := g.buildUpdateInvoiceQueryBase(int(invoiceId.GetNumberValue()))
		newSt := statusToWhmcs(pb.BillingStatus_TERMINATED)
		updQuery.Status = &newSt
		note := "CANCELLED BECAUSE IT WAS REPLACED BY OTHER INVOICE (CALLED BY NOCLOUD). NEW WHMCS INVOICE ID: " + strconv.Itoa(invResp.InvoiceId)
		updQuery.Notes = &note
		q, err = query.Values(updQuery)
		if err != nil {
			q = url.Values{}
		}
		_, _ = sendRequestToWhmcs[InvoiceResponse](http.MethodPost, reqUrl.String()+"?"+q.Encode(), nil)
	}

	meta[invoiceIdField] = structpb.NewNumberValue(float64(invResp.InvoiceId))
	meta["note"] = structpb.NewStringValue(newNote)
	invoice.Meta = meta
	// Update invoice using api but using gateway-callback flag to not trigger useless whmcs update
	if err := g.invMan.UpdateInvoice(context.WithValue(ctx, types.GatewayCallback, true), invoice.Invoice, false); err != nil {
		return ps.NoNackErr(err)
	}

	return nil
}

func (g *WhmcsGateway) UpdateInvoice(ctx context.Context, inv *pb.Invoice, oldStatus pb.BillingStatus) error {
	reqUrl, err := url.Parse(g.baseUrl)
	if err != nil {
		return err
	}
	newStatus := inv.Status

	id, ok := inv.GetMeta()[invoiceIdField]
	if !ok || int(id.GetNumberValue()) <= 0 {
		if (oldStatus == pb.BillingStatus_DRAFT || oldStatus == pb.BillingStatus_TERMINATED) &&
			(newStatus == pb.BillingStatus_UNPAID || newStatus == pb.BillingStatus_PAID || newStatus == pb.BillingStatus_RETURNED) {
			return g.CreateInvoice(ctx, inv)
		}
		return fmt.Errorf("no whmcs_invoice_id or zero value")
	}

	body := g.buildUpdateInvoiceQueryBase(int(id.GetNumberValue()))
	whmcsInv, err := g.GetInvoice(ctx, int(id.GetNumberValue()))
	if err != nil {
		return err
	}

	cur, err := g.currencies.Get(ctx, inv.GetCurrency().GetId())
	if err != nil {
		return err
	}

	if inv.Payment > 0 {
		body.DatePaid = ptr(time.Unix(inv.Payment, 0).Format("2006-01-02 15:04:05"))
	}
	if inv.Deadline > 0 {
		body.DueDate = ptr(time.Unix(inv.Deadline, 0).Format("2006-01-02"))
	}
	if inv.Created > 0 {
		body.Date = ptr(time.Unix(inv.Created, 0).Format("2006-01-02"))
	}

	body.Status = nil
	if inv.Status != statusToNoCloud(whmcsInv.Status) {
		if inv.Status == pb.BillingStatus_PAID && strings.ToLower(whmcsInv.Status) != strings.ToLower(statusToWhmcs(pb.BillingStatus_PAID)) {
			paidWithBalance, _ := ctx.Value("paid-with-balance").(bool)
			if err = g.PayInvoice(ctx, int(id.GetNumberValue()), paidWithBalance); err != nil {
				return fmt.Errorf("failed to pay invoice: %w", err)
			}
		}
	}
	body.Status = ptr(statusToWhmcs(inv.Status))

	body.Notes = ptr(inv.GetMeta()["note"].GetStringValue())
	tax := inv.GetTaxOptions().GetTaxRate() * 100
	_taxed := tax > 0
	isTaxed := "0"
	if _taxed {
		isTaxed = "1"
	}

	if tax == 0 {
		// If pass it as 0.0 then it will be omitted as default value or nil and
		// WHMCS will consider it like we don't want to change tax rate
		// If you pass 0.000000001 then it will not be omitted and WHMCS will round it to zero :(
		body.TaxRate = floatAsString(0.000001)
	} else {
		body.TaxRate = floatAsString(tax)
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
	taxed := make(map[int]string)
	for i, item := range inv.GetItems() {
		var price float64
		if g.taxExcluded {
			price = item.GetPrice() * float64(item.GetAmount())
		} else {
			price = item.GetPrice()*float64(item.GetAmount()) + item.GetPrice()*float64(item.GetAmount())*(tax/100)
		}
		price = graph.Round(price, cur.Precision, cur.Rounding)

		description[i] = item.GetDescription()
		amount[i] = floatAsString(price)
		taxed[i] = isTaxed
	}

	q, err := query.Values(body)
	if err != nil {
		return err
	}
	for i := range inv.GetItems() {
		q.Set(fmt.Sprintf("newitemdescription[%d]", i), description[i])
		q.Set(fmt.Sprintf("newitemamount[%d]", i), fmt.Sprintf("%.2f", amount[i]))
		q.Set(fmt.Sprintf("newitemtaxed[%d]", i), isTaxed)
	}
	_, err = sendRequestToWhmcs[InvoiceResponse](http.MethodPost, reqUrl.String()+"?"+q.Encode(), nil)
	if err != nil {
		return err
	}
	return nil
}

func (g *WhmcsGateway) PayInvoice(ctx context.Context, whmcsInvoiceId int, payWithBalance ...bool) error {
	reqUrl, err := url.Parse(g.baseUrl)
	if err != nil {
		return err
	}

	inv, err := g.GetInvoice(ctx, whmcsInvoiceId)
	if err != nil {
		return err
	}
	if inv.Status == statusToWhmcs(pb.BillingStatus_PAID) {
		return nil
	}

	isPayWithBalance := len(payWithBalance) > 0 && payWithBalance[0]

	if isPayWithBalance {
		// Update invoice payment method first, then pay it
		updBody := g.buildUpdateInvoiceQueryBase(whmcsInvoiceId)
		updBody.PaymentMethod = &balancePayMethod
		q, err := query.Values(updBody)
		if err != nil {
			return err
		}
		_, err = sendRequestToWhmcs[InvoiceResponse](http.MethodPost, reqUrl.String()+"?"+q.Encode(), nil)
		if err != nil {
			return err
		}
	}

	body := g.buildAddPaymentQueryBase(whmcsInvoiceId)
	body.TransId = uuid.New().String()
	body.Date = time.Now().Format("2006-01-02 15:04:05")
	body.Gateway = "system"
	if isPayWithBalance {
		body.Gateway = balancePayMethod
	}
	if inv.Balance <= 0 {
		body.Amount = ptr(inv.Total)
	}

	q, err := query.Values(body)
	if err != nil {
		return err
	}
	_, err = sendRequestToWhmcs[InvoiceResponse](http.MethodPost, reqUrl.String()+"?"+q.Encode(), nil)
	if err != nil {
		return err
	}

	return nil
}

func (g *WhmcsGateway) UpdateClient(_ context.Context, clientId int, notes string) error {
	reqUrl, err := url.Parse(g.baseUrl)
	if err != nil {
		return err
	}

	body, err := g.buildUpdateClientQueryBase(clientId, notes)
	if err != nil {
		return err
	}

	q, err := query.Values(body)
	if err != nil {
		return err
	}
	_, err = sendRequestToWhmcs[InvoiceResponse](http.MethodPost, reqUrl.String()+"?"+q.Encode(), nil)
	if err != nil {
		return err
	}

	return nil
}

func (g *WhmcsGateway) AddNote(_ context.Context, clientId int, notes string, sticky bool) error {
	reqUrl, err := url.Parse(g.baseUrl)
	if err != nil {
		return err
	}

	body, err := g.buildAddNoteQueryBase(clientId, notes, sticky)
	if err != nil {
		return err
	}

	q, err := query.Values(body)
	if err != nil {
		return err
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

	body := g.buildPaymentURIQueryBase(userId, invId)
	q, err := query.Values(body)
	if err != nil {
		return "", err
	}
	resp, err := sendRequestToWhmcs[PaymentURIResponse](http.MethodPost, reqUrl.String()+"?"+q.Encode(), nil)
	if err != nil {
		return "", err
	}

	return g.buildPaymentURI(invId, resp), nil
}

func (g *WhmcsGateway) GetClients(_ context.Context) ([]ListClient, error) {
	reqUrl, err := url.Parse(g.baseUrl)
	if err != nil {
		return nil, err
	}

	query, err := g.buildGetClientsQueryBase()
	if err != nil {
		return nil, err
	}

	resp, err := sendRequestToWhmcs[GetClientsResponse](http.MethodPost, reqUrl.String()+"?"+query.Encode(), nil)
	if err != nil {
		return nil, err
	}

	return resp.Clients.Client, nil
}

func (g *WhmcsGateway) GetClientsProducts(_ context.Context, clientId int) ([]ListProduct, error) {
	reqUrl, err := url.Parse(g.baseUrl)
	if err != nil {
		return nil, err
	}

	query, err := g.buildGetClientsProductsQueryBase(clientId)
	if err != nil {
		return nil, err
	}

	resp, err := sendRequestToWhmcs[GetClientsProductsResponse](http.MethodPost, reqUrl.String()+"?"+query.Encode(), nil)
	if err != nil {
		return nil, err
	}

	return resp.Products.Product, nil
}

func (g *WhmcsGateway) GetClientsDetails(_ context.Context, clientId int) (Client, error) {
	reqUrl, err := url.Parse(g.baseUrl)
	if err != nil {
		return Client{}, err
	}

	query, err := g.buildGetClientsDetailsQueryBase(clientId)
	if err != nil {
		return Client{}, err
	}

	resp, err := sendRequestToWhmcs[GetClientsDetailsResponse](http.MethodPost, reqUrl.String()+"?"+query.Encode(), nil)
	if err != nil {
		return Client{}, err
	}

	return resp.Client, nil
}

func (g *WhmcsGateway) GetPaymentMethods(_ context.Context) ([]PaymentMethod, error) {
	reqUrl, err := url.Parse(g.baseUrl)
	if err != nil {
		return nil, err
	}

	query, err := g.buildGetPaymentMethodsQueryBase()
	if err != nil {
		return nil, err
	}

	resp, err := sendRequestToWhmcs[GetPaymentMethodsResponse](http.MethodPost, reqUrl.String()+"?"+query.Encode(), nil)
	if err != nil {
		return nil, err
	}

	return resp.Methods.Method, nil
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

func (g *WhmcsGateway) GetInvoices(_ context.Context) ([]InvoiceInList, error) {
	res := make([]InvoiceInList, 0)

	reqUrl, err := url.Parse(g.baseUrl)
	if err != nil {
		return res, err
	}
	body := g.buildGetInvoicesQueryBase()

	// Do not use async here
	body.LimitNum = ptr(getInvoicesBatchSize)
	offset := 0
	for {
		body.LimitStart = ptr(offset)
		q, err := query.Values(body)
		if err != nil {
			return res, err
		}
		invResp, err := sendRequestToWhmcs[GetInvoicesResponse](http.MethodPost, reqUrl.String()+"?"+q.Encode(), nil)
		if err != nil {
			return res, err
		}
		if invResp.Invoices.Invoice == nil {
			return res, fmt.Errorf("error getting invoices. Got nil")
		}
		res = append(res, invResp.Invoices.Invoice...)
		offset += getInvoicesBatchSize
		if offset > invResp.TotalResults {
			break
		}
	}

	return res, nil
}

func (g *WhmcsGateway) SendEmail(_ context.Context, template string, relatedID int, customType *string) error {
	reqUrl, err := url.Parse(g.baseUrl)
	if err != nil {
		return err
	}

	body, err := g.buildSendEmailQueryBase(template, relatedID, customType)
	if err != nil {
		return err
	}

	q, err := query.Values(body)
	if err != nil {
		return err
	}
	_, err = sendRequestToWhmcs[SendEmailResponse](http.MethodPost, reqUrl.String()+"?"+q.Encode(), nil)
	if err != nil {
		return err
	}

	return nil
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
		if errors.Is(err, ErrNotFound) {
			return ps.NoNackErr(fmt.Errorf("error syncWhmcsInvoice: %w", err))
		}
		return fmt.Errorf("error syncWhmcsInvoice: %w", err)
	}

	cur, err := g.currencies.Get(ctx, inv.GetCurrency().GetId())
	if err != nil {
		return err
	}

	if inv.Status == pb.BillingStatus_TERMINATED && whmcsInv.Status == statusToWhmcs(pb.BillingStatus_CANCELED) {
		goto skipStatus
	}
	if inv.Status != statusToNoCloud(whmcsInv.Status) {
		inv.Status = statusToNoCloud(whmcsInv.Status)
		if inv, err = g.invMan.UpdateInvoiceStatus(ctx, inv.GetUuid(), inv.Status); err != nil {
			return fmt.Errorf("error syncWhmcsInvoice: %w", err)
		}
	}

skipStatus:
	if !strings.Contains(whmcsInv.DatePaid, "0000-00-00") {
		t, err := time.Parse("2006-01-02 15:04:05", whmcsInv.DatePaid)
		if err != nil {
			return fmt.Errorf("error syncWhmcsInvoice: %w", err)
		}
		if !isDateEqualWithoutTime(time.Unix(inv.Payment, 0), t) {
			inv.Payment = t.Unix()
		}
	}
	if !strings.Contains(whmcsInv.DueDate, "0000-00-00") {
		t, err := time.Parse("2006-01-02", whmcsInv.DueDate)
		if err != nil {
			return fmt.Errorf("error syncWhmcsInvoice: %w", err)
		}
		if !isDateEqualWithoutTime(time.Unix(inv.Deadline, 0), t) {
			inv.Deadline = t.Unix()
		}
	}
	if !strings.Contains(whmcsInv.Date, "0000-00-00") {
		t, err := time.Parse("2006-01-02", whmcsInv.Date)
		if err != nil {
			return fmt.Errorf("error syncWhmcsInvoice: %w", err)
		}
		if !isDateEqualWithoutTime(time.Unix(inv.Created, 0), t) {
			inv.Created = t.Unix()
		}
	}

	tax := float64(whmcsInv.TaxRate / 100)
	inv.TaxOptions.TaxRate = tax

	whmcsItems := whmcsInv.Items.Items
	ncItems := slices.Clone(inv.GetItems())
	synced := len(whmcsItems) == len(ncItems)
	newItems := make([]*pb.Item, 0)
	var total float64
	for _, item := range whmcsItems {
		var price float64
		whmcsAmount := float64(item.Amount)
		if g.taxExcluded {
			price = whmcsAmount
		} else {
			price = whmcsAmount / (1 + tax)
		}
		price = graph.Round(price, cur.Precision, cur.Rounding)

		found := false
		index := 0
		for i, ncItem := range ncItems {
			ncPrice := ncItem.Price * float64(ncItem.Amount)
			ncPrice = graph.Round(ncPrice, cur.Precision, cur.Rounding)
			if item.Description == ncItem.Description && compareFloat(price, ncPrice, int(cur.Precision)) {
				found = true
				index = i
				break
			}
		}
		if found {
			ncItems = slices.Delete(ncItems, index, index+1)
		} else {
			synced = false
		}

		total += price
		newItems = append(newItems, &pb.Item{
			Description: item.Description,
			Amount:      1,
			Price:       price,
			Unit:        "Pcs",
			ApplyTax:    item.Taxed > 0,
		})
	}
	total = graph.Round(total, cur.Precision, cur.Rounding)

	var warning string
	if !synced {
		if inv.Type != pb.ActionType_WHMCS_INVOICE {
			warning = "[WARNING]: THIS INVOICE ITEMS WERE UPDATED DIRECTLY FROM WHMCS.\n"
		}
		inv.Items = newItems
		inv.Total = total
	}

	meta := inv.GetMeta()
	meta["note"] = structpb.NewStringValue(warning + whmcsInv.Notes)
	inv.Meta = meta

	inv.Transactions = nil
	inv.Instances = nil
	if synced {
		inv.Items = nil
	}
	if err = g.invMan.UpdateInvoice(ctx, inv, true); err != nil {
		return fmt.Errorf("error syncWhmcsInvoice: failed to update invoice: %w", err)
	}
	return nil
}
