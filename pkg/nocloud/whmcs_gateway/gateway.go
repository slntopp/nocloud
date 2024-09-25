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

	accounts *graph.AccountsController
	invoices *graph.InvoicesController
}

func NewWhmcsGateway(username string, passwordHash string, host string, acc *graph.AccountsController, inv *graph.InvoicesController) *WhmcsGateway {
	return &WhmcsGateway{
		apiUsername: username,
		apiPassword: passwordHash,
		baseUrl:     host,
		accounts:    acc,
		invoices:    inv,
	}
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
	query := g.buildCreateInvoiceQueryBase(inv, account.GetData().AsMap()["whmcs_id"].(int))
	for i, item := range inv.GetItems() {
		query.Set(fmt.Sprintf("itemdescription%d", i+1), item.GetDescription())
		query.Set(fmt.Sprintf("itemamount%d", i+1), fmt.Sprintf("%.2f", item.GetPrice()*float64(item.GetAmount())))
		query.Set(fmt.Sprintf("itemtaxed%d", i+1), "0")
	}

	req, err := http.NewRequest(http.MethodPost, reqUrl.String()+"?"+query.Encode(), nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var invResp InvoiceResponse
	err = json.Unmarshal(b, &invResp)
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 || invResp.Status != "success" {
		return fmt.Errorf("failed to create invoice: %s", resp.Status)
	}

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
