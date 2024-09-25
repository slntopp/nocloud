package whmcs_gateway

import (
	"context"
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
}

func NewWhmcsGateway(username string, passwordHash string, host string, acc *graph.AccountsController) *WhmcsGateway {
	return &WhmcsGateway{
		apiUsername: username,
		apiPassword: passwordHash,
		baseUrl:     host,
		accounts:    acc,
	}
}

func (g *WhmcsGateway) CreateInvoice(_ context.Context, inv *pb.Invoice) error {
	reqUrl, err := url.Parse(g.baseUrl)
	if err != nil {
		return err
	}

	account, err := g.accounts.Get(context.Background(), inv.GetAccount())
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
	fmt.Println("Create Invoice Body: ", string(b))

	if resp.StatusCode != 200 {
		return fmt.Errorf("failed to create invoice: %s", resp.Status)
	}

	return nil
}

func (g *WhmcsGateway) UpdateInvoice(_ context.Context, inv *pb.Invoice) error {
	return nil
}
