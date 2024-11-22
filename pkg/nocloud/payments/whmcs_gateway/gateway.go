package whmcs_gateway

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/go-querystring/query"
	"github.com/google/uuid"
	pb "github.com/slntopp/nocloud-proto/billing"
	"github.com/slntopp/nocloud/pkg/graph"
	"google.golang.org/protobuf/types/known/structpb"
	"io"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"time"
)

type NoCloudInvoicesManager interface {
	CreateInvoice(ctx context.Context, inv *pb.Invoice) error
	UpdateInvoice(ctx context.Context, inv *pb.Invoice) error
	UpdateInvoiceStatus(ctx context.Context, id string, newStatus pb.BillingStatus) (*pb.Invoice, error)
	InvoicesController() graph.InvoicesController
}

type WhmcsGateway struct {
	apiUsername string
	apiPassword string
	baseUrl     string
	trustedIP   string

	accounts graph.AccountsController
	invMan   NoCloudInvoicesManager
}

func NewWhmcsGateway(data WhmcsData, acc graph.AccountsController, invMan NoCloudInvoicesManager) *WhmcsGateway {
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

	var sendEmail = inv.Status != pb.BillingStatus_DRAFT

	oldNote := inv.GetMeta()["note"].GetStringValue()
	newNote := "NOCLOUD INVOICE ID: " + inv.GetUuid() + "\n" + oldNote
	inv.GetMeta()["note"] = structpb.NewStringValue(newNote)
	q, err := g.buildCreateInvoiceQueryBase(inv, userId, sendEmail)
	if err != nil {
		return err
	}
	for i, item := range inv.GetItems() {
		q.Set(fmt.Sprintf("itemdescription%d", i+1), item.GetDescription())
		q.Set(fmt.Sprintf("itemamount%d", i+1), fmt.Sprintf("%.2f", item.GetPrice()*float64(item.GetAmount())))
		q.Set(fmt.Sprintf("itemtaxed%d", i+1), "0")
	}

	invResp, err := sendRequestToWhmcs[InvoiceResponse](http.MethodPost, reqUrl.String()+"?"+q.Encode(), nil)
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

	// If invoice had whmcs invoice id, then cancel old invoice
	if invoiceId, ok := meta[invoiceIdField]; ok {
		updQuery := g.buildUpdateInvoiceQueryBase(int(invoiceId.GetNumberValue()))
		newSt := statusToWhmcs(pb.BillingStatus_TERMINATED)
		updQuery.Status = &newSt
		note := "CANCELLED BECAUSE IT WAS REPLACED BY OTHER INVOICE (CALLED BY NOCLOUD). NEW WHMCS INVOICE ID: " + strconv.Itoa(invResp.InvoiceId)
		updQuery.Notes = &note
		q, err = query.Values(updQuery)
		if err != nil {
			fmt.Println("WHMCS Gateway: failed to build update invoice query: ", err.Error())
			q = url.Values{}
		}
		_, err := sendRequestToWhmcs[InvoiceResponse](http.MethodPost, reqUrl.String()+"?"+q.Encode(), nil)
		if err != nil {
			fmt.Println("WHMCS Gateway: error cancelling old whmcs invoice: ", err.Error())
		}
	}

	meta[invoiceIdField] = structpb.NewNumberValue(float64(invResp.InvoiceId))
	meta["note"] = structpb.NewStringValue(newNote)
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
		return g.CreateInvoice(ctx, inv)
	}

	body := g.buildUpdateInvoiceQueryBase(int(id.GetNumberValue()))
	whmcsInv, err := g.GetInvoice(ctx, int(id.GetNumberValue()))
	if err != nil {
		return err
	}

	body.DatePaid = ptr(time.Unix(inv.Payment, 0).Format("2006-01-02 15:04:05"))
	body.DueDate = ptr(time.Unix(inv.Deadline, 0).Format("2006-01-02"))
	body.Date = ptr(time.Unix(inv.Created, 0).Format("2006-01-02"))

	if inv.Status != old.Status {
		body.Status = ptr(statusToWhmcs(inv.Status))
		if inv.Status == pb.BillingStatus_PAID {
			if err = g.PayInvoice(ctx, int(id.GetNumberValue())); err != nil {
				return fmt.Errorf("failed to pay invoice: %w", err)
			}
			body.Status = nil
		}
	}

	body.Notes = ptr(inv.GetMeta()["note"].GetStringValue())

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

func (g *WhmcsGateway) PayInvoice(ctx context.Context, whmcsInvoiceId int) error {
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

	body := g.buildAddPaymentQueryBase(whmcsInvoiceId)
	body.TransId = uuid.New().String()
	body.Date = time.Now().Format("2006-01-02 15:04:05")
	body.Gateway = "system"

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
		if inv, err = g.invMan.UpdateInvoiceStatus(ctx, inv.GetUuid(), inv.Status); err != nil {
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

	whmcsItems := whmcsInv.Items.Items
	ncItems := slices.Clone(inv.GetItems())
	newItems := make([]*pb.Item, 0)
	synced := true
	for _, item := range whmcsItems {
		found := false
		index := 0
		for i, ncItem := range ncItems {
			if item.Description == ncItem.Description && equalFloats(float64(item.Amount), ncItem.Price*float64(ncItem.Amount)) {
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

		newItems = append(newItems, &pb.Item{
			Description: item.Description,
			Amount:      1,
			Price:       float64(item.Amount),
			Unit:        "Pcs",
		})
	}
	var warning string
	if !synced && inv.Type != pb.ActionType_WHMCS_INVOICE {
		warning = "[WARNING]: THIS INVOICE ITEMS WERE UPDATED DIRECTLY FROM WHMCS. THIS MEANS THAT NOW THIS INVOICE DO NOT PROVIDE FUNCTIONALITY IN NOCLOUD! (DONT UPDATE AUTOGENERATED AND FUNCTIONAL INVOICES THROUGH WHMCS)\n"
		inv.Items = newItems
		inv.Total = float64(whmcsInv.Total)
	}

	meta := inv.GetMeta()
	meta["note"] = structpb.NewStringValue(warning + whmcsInv.Notes)
	inv.Meta = meta

	if err = g.invMan.UpdateInvoice(ctx, inv); err != nil {
		return fmt.Errorf("error syncWhmcsInvoice: failed to update invoice: %w", err)
	}
	return nil
}
