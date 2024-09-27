package whmcs_gateway

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/go-querystring/query"
	pb "github.com/slntopp/nocloud-proto/billing"
	"github.com/slntopp/nocloud/pkg/graph"
	"google.golang.org/protobuf/types/known/structpb"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const invoiceIdField = "whmcs_invoice_id"

type WhmcsGateway struct {
	apiUsername string
	apiPassword string
	baseUrl     string
	trustedIP   string

	accounts *graph.AccountsController
	invoices *graph.InvoicesController
}

func NewWhmcsGateway(data WhmcsData, acc *graph.AccountsController, inv *graph.InvoicesController) *WhmcsGateway {
	return &WhmcsGateway{
		apiUsername: data.WhmcsUser,
		apiPassword: data.WhmcsPassHash,
		baseUrl:     data.WhmcsBaseUrl,
		trustedIP:   data.TrustedIP,
		accounts:    acc,
		invoices:    inv,
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

	// TODO: review taxed field
	query, err := g.buildCreateInvoiceQueryBase(inv, int(account.GetData().AsMap()["whmcs_id"].(float64)))
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
	invoice, err := g.invoices.Get(ctx, inv.GetUuid())
	if err != nil {
		return err
	}
	meta := invoice.GetMeta()
	meta[invoiceIdField] = structpb.NewNumberValue(float64(invResp.InvoiceId))
	invoice.Meta = meta
	if _, err := g.invoices.Update(ctx, invoice); err != nil {
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

func (g *WhmcsGateway) syncWhmcsInvoice(ctx context.Context, invoiceId int) error {
	whmcsInv, err := g.GetInvoice(ctx, invoiceId)
	if err != nil {
		return err
	}
	inv, err := g.getInvoiceByWhmcsId(invoiceId)
	if err != nil {
		return err
	}

	inv.Status = statusToNoCloud(whmcsInv.Status)
	if !strings.Contains(whmcsInv.DatePaid, "0000-00-00") {
		t, err := time.Parse("2006-01-02 15:04:05", whmcsInv.DatePaid)
		if err != nil {
			return err
		}
		inv.Payment = t.Unix()
		inv.Processed = inv.Payment
	}
	if !strings.Contains(whmcsInv.DueDate, "0000-00-00") {
		t, err := time.Parse("2006-01-02", whmcsInv.DueDate)
		if err != nil {
			return err
		}
		inv.Deadline = t.Unix()
	}
	if !strings.Contains(whmcsInv.Date, "0000-00-00") {
		t, err := time.Parse("2006-01-02", whmcsInv.Date)
		if err != nil {
			return err
		}
		inv.Created = t.Unix()
	}
	inv.Total = float64(whmcsInv.Total)
	// TODO: sync invoice number too. Have to refactor number logic in invoice controller
	inv.Items = []*pb.Item{}
	for _, item := range whmcsInv.Items.Items {
		inv.Items = append(inv.Items, &pb.Item{
			Description: item.Description,
			Amount:      1,
			Price:       float64(item.Amount),
			Unit:        "Pcs",
		})
	}
	meta := inv.GetMeta()
	meta["note"] = structpb.NewStringValue(whmcsInv.Notes)
	inv.Meta = meta

	if _, err = g.invoices.Update(ctx, &graph.Invoice{
		Invoice: inv,
	}); err != nil {
		return fmt.Errorf("failed to update invoice: %w", err)
	}
	return nil
}
