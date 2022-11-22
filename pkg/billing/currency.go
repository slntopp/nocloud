package billing

import (
	"context"

	"github.com/arangodb/go-driver"
	pb "github.com/slntopp/nocloud-proto/billing"
	"github.com/slntopp/nocloud/pkg/graph"
	"go.uber.org/zap"
)

type CurrencyServiceServer struct {
	pb.UnimplementedCurrencyServiceServer
	log *zap.Logger

	ctrl *graph.CurrencyController

	db driver.Database
}

func NewCurrencyServiceServer(log *zap.Logger, db driver.Database) *CurrencyServiceServer {
	return &CurrencyServiceServer{
		log:  log.Named("CurrencyServer"),
		db:   db,
		ctrl: graph.NewCurrencyController(log, db),
	}
}

func (s *CurrencyServiceServer) GetExchangeRate(ctx context.Context, req *pb.GetExchangeRateRequest) (*pb.GetExchangeRateResponse, error) {
	rate, err := s.ctrl.GetExchangeRate(ctx, req.From, req.To)
	if err != nil {
		return nil, err
	}

	return &pb.GetExchangeRateResponse{Rate: rate}, nil
}

func (s *CurrencyServiceServer) CreateExchangeRate(ctx context.Context, req *pb.CreateExchangeRateRequest) (*pb.CreateExchangeRateResponse, error) {
	err := s.ctrl.CreateExchangeRate(ctx, req.From, req.To, req.Rate)
	return &pb.CreateExchangeRateResponse{}, err
}

func (s *CurrencyServiceServer) UpdateExchangeRate(ctx context.Context, req *pb.UpdateExchangeRateRequest) (*pb.UpdateExchangeRateResponse, error) {
	err := s.ctrl.UpdateExchangeRate(ctx, req.From, req.To, req.Rate)
	return &pb.UpdateExchangeRateResponse{}, err
}

func (s *CurrencyServiceServer) DeleteExchangeRate(ctx context.Context, req *pb.DeleteExchangeRateRequest) (*pb.DeleteExchangeRateResponse, error) {
	err := s.ctrl.DeleteExchangeRate(ctx, req.From, req.To)
	return &pb.DeleteExchangeRateResponse{}, err
}

func (s *CurrencyServiceServer) Convert(ctx context.Context, req *pb.ConversionRequest) (*pb.ConversionResponse, error) {
	amount, err := s.ctrl.Convert(ctx, req.From, req.To, req.Amount)
	if err != nil {
		return nil, err
	}

	return &pb.ConversionResponse{Amount: amount}, nil
}

func (s *CurrencyServiceServer) GetCurrencies(ctx context.Context, req *pb.GetCurrenciesRequest) (*pb.GetCurrenciesResponse, error) {
	currencies, err := s.ctrl.Get(ctx)
	if err != nil {
		return nil, err
	}

	return &pb.GetCurrenciesResponse{Currencies: currencies}, nil
}
