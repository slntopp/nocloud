package whmcs_gateway

import (
	"context"
	"encoding/json"
	"fmt"
	pb "github.com/slntopp/nocloud-proto/billing"
	"github.com/slntopp/nocloud/pkg/graph"
	"io"
	"net/http"
	"net/url"
)

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
		return result, fmt.Errorf("failed to unmarshal response to result checker struct: %w", err)
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

	newInv, err := g.GetInvoice(ctx, invResp.InvoiceId)
	if err != nil {
		return err
	}
	fmt.Printf("WhmcsGetInvoice: %+v\n", newInv)

	// Update NoCloud invoice
	patch := map[string]interface{}{
		"meta.whmcs_invoice_id": invResp.InvoiceId,
	}
	if err := g.invoices.Patch(ctx, inv.GetUuid(), patch); err != nil {
		return err
	}

	return nil
}

func (g *WhmcsGateway) UpdateInvoice(ctx context.Context, inv *pb.Invoice) error {
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
