package billing

import (
	"context"
	"math/rand"

	"connectrpc.com/connect"
	"github.com/slntopp/nocloud-proto/access"
	"github.com/slntopp/nocloud/pkg/nocloud"
	"github.com/slntopp/nocloud/pkg/nocloud/schema"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/arangodb/go-driver"
	pb "github.com/slntopp/nocloud-proto/billing"
	"github.com/slntopp/nocloud/pkg/graph"
	"go.uber.org/zap"
)

type CurrencyServiceServer struct {
	log *zap.Logger

	ctrl graph.CurrencyController
	ca   graph.CommonActionsController

	db driver.Database
}

func NewCurrencyServiceServer(log *zap.Logger, db driver.Database, currencies graph.CurrencyController, ca graph.CommonActionsController) *CurrencyServiceServer {
	return &CurrencyServiceServer{
		log:  log.Named("CurrencyServer"),
		db:   db,
		ctrl: currencies,
		ca:   ca,
	}
}

func (s *CurrencyServiceServer) CreateCurrency(ctx context.Context, r *connect.Request[pb.CreateCurrencyRequest]) (*connect.Response[pb.CreateCurrencyResponse], error) {
	log := s.log.Named("CreateCurrency")
	req := r.Msg
	log.Debug("Request received", zap.Any("request", req))
	requester := ctx.Value(nocloud.NoCloudAccount).(string)
	if !s.ca.HasAccess(ctx, requester, driver.NewDocumentID(schema.NAMESPACES_COL, schema.ROOT_NAMESPACE_KEY), access.Level_ROOT) {
		return nil, status.Error(codes.PermissionDenied, "Not enough Access rights to manage Currencies")
	}

	if req.Currency == nil {
		return nil, status.Error(codes.InvalidArgument, "no currency provided")
	}
	if req.Currency.Id == 0 {
		req.Currency.Id = int32(rand.Int())
	}
	err := s.ctrl.CreateCurrency(ctx, req.Currency)
	if err != nil {
		log.Error("Error creating Currency", zap.Error(err))
		return nil, err
	}

	return connect.NewResponse(&pb.CreateCurrencyResponse{}), nil
}

func (s *CurrencyServiceServer) UpdateCurrency(ctx context.Context, r *connect.Request[pb.UpdateCurrencyRequest]) (*connect.Response[pb.UpdateCurrencyResponse], error) {
	log := s.log.Named("UpdateCurrency")
	req := r.Msg
	log.Debug("Request received", zap.Any("request", req))
	requester := ctx.Value(nocloud.NoCloudAccount).(string)
	if !s.ca.HasAccess(ctx, requester, driver.NewDocumentID(schema.NAMESPACES_COL, schema.ROOT_NAMESPACE_KEY), access.Level_ROOT) {
		return nil, status.Error(codes.PermissionDenied, "Not enough Access rights to manage Currencies")
	}
	if req.Currency == nil {
		return nil, status.Error(codes.InvalidArgument, "no currency provided")
	}

	err := s.ctrl.UpdateCurrency(ctx, req.Currency)
	if err != nil {
		log.Error("Error updating Currency", zap.Error(err))
		return nil, err
	}

	return connect.NewResponse(&pb.UpdateCurrencyResponse{}), nil
}

func (s *CurrencyServiceServer) GetExchangeRate(ctx context.Context, req *connect.Request[pb.GetExchangeRateRequest]) (*connect.Response[pb.GetExchangeRateResponse], error) {
	rate, commission, err := s.ctrl.GetExchangeRate(ctx, req.Msg.GetFrom(), req.Msg.GetTo())
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&pb.GetExchangeRateResponse{Rate: rate, Commission: commission}), nil
}

func (s *CurrencyServiceServer) CreateExchangeRate(ctx context.Context, r *connect.Request[pb.CreateExchangeRateRequest]) (*connect.Response[pb.CreateExchangeRateResponse], error) {
	log := s.log.Named("CreateExchangeRate")
	req := r.Msg
	log.Debug("Request received", zap.Any("request", req))
	requester := ctx.Value(nocloud.NoCloudAccount).(string)
	if !s.ca.HasAccess(ctx, requester, driver.NewDocumentID(schema.NAMESPACES_COL, schema.ROOT_NAMESPACE_KEY), access.Level_ROOT) {
		return nil, status.Error(codes.PermissionDenied, "Not enough Access rights to manage Currencies")
	}

	err := s.ctrl.CreateExchangeRate(ctx, *req.From, *req.To, req.Rate, req.Commission)
	if err != nil {
		log.Error("Error creating Exchange rate", zap.Error(err))
		return connect.NewResponse(&pb.CreateExchangeRateResponse{}), err
	}

	_, _, err = s.ctrl.GetExchangeRateDirect(ctx, *req.To, *req.From)
	if err == nil {
		return connect.NewResponse(&pb.CreateExchangeRateResponse{}), nil
	}

	s.log.Info("Reverse rate is not set yet, setting automatically", zap.String("from", req.To.String()), zap.String("to", req.From.String()))
	err = s.ctrl.CreateExchangeRate(ctx, *req.To, *req.From, 1/req.Rate, req.Commission)
	if err != nil {
		log.Error("Couldn't automatically create reverse Exchange rate", zap.Error(err))
	}

	return connect.NewResponse(&pb.CreateExchangeRateResponse{}), nil
}

func (s *CurrencyServiceServer) UpdateExchangeRate(ctx context.Context, r *connect.Request[pb.UpdateExchangeRateRequest]) (*connect.Response[pb.UpdateExchangeRateResponse], error) {
	log := s.log.Named("UpdateExchangeRate")
	req := r.Msg
	log.Debug("Request received", zap.Any("request", req))
	requester := ctx.Value(nocloud.NoCloudAccount).(string)
	if !s.ca.HasAccess(ctx, requester, driver.NewDocumentID(schema.NAMESPACES_COL, schema.ROOT_NAMESPACE_KEY), access.Level_ROOT) {
		return nil, status.Error(codes.PermissionDenied, "Not enough Access rights to manage Currencies")
	}

	if err := s.ctrl.UpdateExchangeRate(ctx, *req.From, *req.To, req.Rate, req.Commission); err != nil {
		log.Error("Error updating Exchange rate", zap.Error(err))
		return nil, err
	}
	return connect.NewResponse(&pb.UpdateExchangeRateResponse{}), nil
}

func (s *CurrencyServiceServer) DeleteExchangeRate(ctx context.Context, r *connect.Request[pb.DeleteExchangeRateRequest]) (*connect.Response[pb.DeleteExchangeRateResponse], error) {
	log := s.log.Named("DeleteExchangeRate")
	req := r.Msg
	log.Debug("Request received", zap.Any("request", req))
	requester := ctx.Value(nocloud.NoCloudAccount).(string)
	if !s.ca.HasAccess(ctx, requester, driver.NewDocumentID(schema.NAMESPACES_COL, schema.ROOT_NAMESPACE_KEY), access.Level_ROOT) {
		return nil, status.Error(codes.PermissionDenied, "Not enough Access rights to manage Currencies")
	}

	if err := s.ctrl.DeleteExchangeRate(ctx, req.From, req.To); err != nil {
		log.Error("Error deleting Exchange rate", zap.Error(err))
		return nil, err
	}
	return connect.NewResponse(&pb.DeleteExchangeRateResponse{}), nil
}

func (s *CurrencyServiceServer) Convert(ctx context.Context, r *connect.Request[pb.ConversionRequest]) (*connect.Response[pb.ConversionResponse], error) {
	req := r.Msg
	amount, err := s.ctrl.Convert(ctx, req.From, req.To, req.Amount)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&pb.ConversionResponse{Amount: amount}), nil
}

func (s *CurrencyServiceServer) GetCurrencies(ctx context.Context, r *connect.Request[pb.GetCurrenciesRequest]) (*connect.Response[pb.GetCurrenciesResponse], error) {
	log := s.log.Named("GetCurrencies")
	log.Debug("Request received")

	var isAdmin bool
	requester, ok := ctx.Value(nocloud.NoCloudAccount).(string)
	if ok {
		isAdmin = s.ca.HasAccess(ctx, requester, driver.NewDocumentID(schema.NAMESPACES_COL, schema.ROOT_NAMESPACE_KEY), access.Level_ADMIN)
	}

	currencies, err := s.ctrl.GetCurrencies(ctx, isAdmin)
	if err != nil {
		log.Error("Error getting Currencies", zap.Error(err))
		return nil, err
	}

	return connect.NewResponse(&pb.GetCurrenciesResponse{Currencies: currencies}), nil
}

func (s *CurrencyServiceServer) GetExchangeRates(ctx context.Context, r *connect.Request[pb.GetExchangeRatesRequest]) (*connect.Response[pb.GetExchangeRatesResponse], error) {
	rates, err := s.ctrl.GetExchangeRates(ctx)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&pb.GetExchangeRatesResponse{Rates: rates}), nil
}
