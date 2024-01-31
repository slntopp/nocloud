package billing

import (
	"connectrpc.com/connect"
	"context"

	"github.com/arangodb/go-driver"
	pb "github.com/slntopp/nocloud-proto/billing"
	"github.com/slntopp/nocloud/pkg/graph"
	"go.uber.org/zap"
)

type CurrencyServiceServer struct {
	log *zap.Logger

	ctrl graph.CurrencyController

	db driver.Database
}

func NewCurrencyServiceServer(log *zap.Logger, db driver.Database) *CurrencyServiceServer {
	return &CurrencyServiceServer{
		log:  log.Named("CurrencyServer"),
		db:   db,
		ctrl: graph.NewCurrencyController(log, db),
	}
}

func (s *CurrencyServiceServer) GetExchangeRate(ctx context.Context, r *connect.Request[pb.GetExchangeRateRequest]) (*connect.Response[pb.GetExchangeRateResponse], error) {
	req := r.Msg
	rate, err := s.ctrl.GetExchangeRate(ctx, req.From, req.To)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&pb.GetExchangeRateResponse{Rate: rate}), nil
}

func (s *CurrencyServiceServer) CreateExchangeRate(ctx context.Context, r *connect.Request[pb.CreateExchangeRateRequest]) (*connect.Response[pb.CreateExchangeRateResponse], error) {
	req := r.Msg
	err := s.ctrl.CreateExchangeRate(ctx, req.From, req.To, req.Rate)
	if err != nil {
		return connect.NewResponse(&pb.CreateExchangeRateResponse{}), err
	}

	_, err = s.ctrl.GetExchangeRateDirect(ctx, req.To, req.From)
	if err == nil {
		return connect.NewResponse(&pb.CreateExchangeRateResponse{}), nil
	}

	s.log.Info("Reverse rate is not set yet, setting automatically", zap.String("from", req.To.String()), zap.String("to", req.From.String()))
	err = s.ctrl.CreateExchangeRate(ctx, req.To, req.From, 1/req.Rate)
	if err != nil {
		s.log.Warn("Couldn't automatically create reverse Exchange rate", zap.Error(err))
	}

	return connect.NewResponse(&pb.CreateExchangeRateResponse{}), nil
}

func (s *CurrencyServiceServer) UpdateExchangeRate(ctx context.Context, r *connect.Request[pb.UpdateExchangeRateRequest]) (*connect.Response[pb.UpdateExchangeRateResponse], error) {
	req := r.Msg
	err := s.ctrl.UpdateExchangeRate(ctx, req.From, req.To, req.Rate)
	return connect.NewResponse(&pb.UpdateExchangeRateResponse{}), err
}

func (s *CurrencyServiceServer) DeleteExchangeRate(ctx context.Context, r *connect.Request[pb.DeleteExchangeRateRequest]) (*connect.Response[pb.DeleteExchangeRateResponse], error) {
	req := r.Msg
	err := s.ctrl.DeleteExchangeRate(ctx, req.From, req.To)
	return connect.NewResponse(&pb.DeleteExchangeRateResponse{}), err
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
	currencies, err := s.ctrl.GetCurrencies(ctx)
	if err != nil {
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
