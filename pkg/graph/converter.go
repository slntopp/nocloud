package graph

import (
	"connectrpc.com/connect"
	"context"
	"fmt"
	bpb "github.com/slntopp/nocloud-proto/billing"
	pb "github.com/slntopp/nocloud-proto/billing/addons"
	ipb "github.com/slntopp/nocloud-proto/instances"
	spb "github.com/slntopp/nocloud-proto/services"
	"github.com/slntopp/nocloud/pkg/nocloud/schema"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"net/http"
)

const CurrencyHeader = "NoCloud-Primary-Currency-Code"
const PrimaryCurrencyCtxKey = ContextKey("nocloud-primary-currency")

func CtxWithPrimaryCurrency(ctx context.Context, currCode string) context.Context {
	if currCode != "" {
		return context.WithValue(ctx, PrimaryCurrencyCtxKey, currCode)
	}
	return ctx
}

func GetPrimaryCurrencyCode(ctx context.Context) (string, bool) {
	if val, ok := ctx.Value(PrimaryCurrencyCtxKey).(string); ok && val != "" {
		return val, true
	}
	return "", false
}

type PricesConverter struct {
	currencies CurrencyController
	target     Currency
	rate       float64
	failed     bool
}

func ExplicitSetPrimaryCurrencyHeader(header http.Header, code string) {
	header.Set(CurrencyHeader, code)
}

func HandleConvertionError[T any](resp *connect.Response[T], conv PricesConverter) (*connect.Response[T], error) {
	if conv.failed {
		return nil, status.Error(codes.Internal, "Internal error. Couldn't convert result prices")
	}
	return resp, nil
}

func NewConverter(header http.Header, curr CurrencyController) PricesConverter {
	code := header.Get(CurrencyHeader)
	if code == "" {
		fmt.Println("skipping: no desired currency")
		return PricesConverter{currencies: curr, failed: true}
	}
	ctx := context.Background()
	c, err := curr.GetByCode(ctx, code)
	if err != nil {
		fmt.Println("error: failed to get currency by code: " + err.Error())
		return PricesConverter{currencies: curr, failed: true}
	}
	rate, _, err := curr.GetExchangeRate(ctx,
		&bpb.Currency{Id: schema.DEFAULT_CURRENCY_ID, Title: schema.DEFAULT_CURRENCY_NAME},
		&bpb.Currency{Id: c.Id, Title: c.Title})
	if err != nil {
		fmt.Println("error: failed to get exchange rate: " + err.Error())
		return PricesConverter{currencies: curr, failed: true}
	}
	return PricesConverter{currencies: curr, target: c, rate: rate}
}

func (conv *PricesConverter) SetResponseHeader(header http.Header) {
	if conv.failed {
		header.Set(CurrencyHeader, schema.DEFAULT_CURRENCY_NAME)
		return
	}
	header.Set(CurrencyHeader, conv.target.Code)
}

func (conv *PricesConverter) ConvertObjectPrices(obj interface{}) {
	if conv.failed {
		return
	}
	if obj == nil {
		return
	}
	switch val := obj.(type) {
	case *pb.Addon:
		ConvertAddon(val, conv.rate, conv.target.Precision, conv.target.Rounding)
	case *bpb.Plan:
		ConvertPlan(val, conv.rate, conv.target.Precision, conv.target.Rounding)
	case *spb.Service:
		ConvertService(val, conv.rate, conv.target.Precision, conv.target.Rounding)
	case *ipb.Instance:
		ConvertInstance(val, conv.rate, conv.target.Precision, conv.target.Rounding)
	default:
		fmt.Println("error: provided invalid object to convert")
		return
	}
}

func ConvertService(s *spb.Service, rate float64, precision int32, round bpb.Rounding) {
	if s == nil {
		return
	}
	for _, ig := range s.GetInstancesGroups() {
		for _, i := range ig.GetInstances() {
			ConvertInstance(i, rate, precision, round)
		}
	}
}

func ConvertInstance(i *ipb.Instance, rate float64, precision int32, round bpb.Rounding) {
	if i == nil {
		return
	}
	i.Estimate = Round(i.Estimate*rate, precision, round)
	if i.BillingPlan != nil {
		ConvertPlan(i.BillingPlan, rate, precision, round)
	}
}

func ConvertPlan(p *bpb.Plan, rate float64, precision int32, round bpb.Rounding) {
	if p == nil {
		return
	}
	for _, res := range p.GetResources() {
		if res == nil {
			continue
		}
		res.Price = Round(res.Price*rate, precision, round)
	}
	for _, prod := range p.GetProducts() {
		if prod == nil {
			continue
		}
		prod.Price = Round(prod.Price*rate, precision, round)
	}
}

func ConvertAddon(a *pb.Addon, rate float64, precision int32, round bpb.Rounding) {
	if a == nil {
		return
	}
	if a.Periods == nil {
		return
	}
	for key, price := range a.Periods {
		a.Periods[key] = Round(price*rate, precision, round)
	}
}
