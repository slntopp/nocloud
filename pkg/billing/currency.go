package billing

import (
	"connectrpc.com/connect"
	"context"
	"errors"
	"github.com/slntopp/nocloud-proto/access"
	"github.com/slntopp/nocloud/pkg/nocloud"
	"github.com/slntopp/nocloud/pkg/nocloud/schema"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"strings"

	"github.com/arangodb/go-driver"
	pb "github.com/slntopp/nocloud-proto/billing"
	"github.com/slntopp/nocloud/pkg/graph"
	"go.uber.org/zap"
)

type CurrencyServiceServer struct {
	log *zap.Logger

	ctrl     graph.CurrencyController
	accounts graph.AccountsController
	ca       graph.CommonActionsController

	db driver.Database
}

func NewCurrencyServiceServer(log *zap.Logger, db driver.Database, currencies graph.CurrencyController, accounts graph.AccountsController, ca graph.CommonActionsController) *CurrencyServiceServer {
	return &CurrencyServiceServer{
		log:      log.Named("CurrencyServer"),
		db:       db,
		ctrl:     currencies,
		ca:       ca,
		accounts: accounts,
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
	currencies, err := s.ctrl.GetCurrencies(ctx, true)
	if err != nil {
		log.Error("Failed to get currencies", zap.Error(err))
		return nil, err
	}
	if req.Currency.Id == 0 {
		var idMax int32 = 0
		for _, currency := range currencies {
			if currency.GetId() > idMax {
				idMax = currency.GetId()
			}
		}
		req.Currency.Id = idMax + 1
	}
	if req.Currency.Precision == 0 {
		req.Currency.Precision = 2
	}
	req.Currency.Title = strings.TrimSpace(req.Currency.Title)
	req.Currency.Code = strings.TrimSpace(req.Currency.Code)
	if req.Currency.Title == "" || req.Currency.Code == "" {
		return nil, status.Error(codes.InvalidArgument, "title and code must be provided")
	}
	if req.Currency.Default {
		for _, currency := range currencies {
			if currency.Default {
				return nil, status.Error(codes.AlreadyExists, "default currency already exists")
			}
		}
	}
	for _, currency := range currencies {
		if currency.GetCode() == req.Currency.Code {
			return nil, status.Error(codes.AlreadyExists, "currency with this code already exists")
		}
	}

	err = s.ctrl.CreateCurrency(ctx, req.Currency)
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
	currencies, err := s.ctrl.GetCurrencies(ctx, true)
	if err != nil {
		log.Error("Failed to get currencies", zap.Error(err))
		return nil, err
	}
	req.Currency.Title = strings.TrimSpace(req.Currency.Title)
	req.Currency.Code = strings.TrimSpace(req.Currency.Code)
	if req.Currency.Title == "" || req.Currency.Code == "" {
		return nil, status.Error(codes.InvalidArgument, "title and code must be provided")
	}
	if req.Currency.Default {
		for _, currency := range currencies {
			if currency.Default && req.GetCurrency().GetId() != currency.GetId() {
				return nil, status.Error(codes.AlreadyExists, "default currency already exists")
			}
		}
	}
	for _, currency := range currencies {
		if currency.GetCode() == req.Currency.Code && currency.GetId() != req.GetCurrency().GetId() {
			return nil, status.Error(codes.AlreadyExists, "currency with this code already exists")
		}
	}

	err = s.ctrl.UpdateCurrency(ctx, req.Currency)
	if err != nil {
		log.Error("Error updating Currency", zap.Error(err))
		return nil, err
	}
	return connect.NewResponse(&pb.UpdateCurrencyResponse{}), nil
}

func (s *CurrencyServiceServer) ChangeDefaultCurrency(ctx context.Context, r *connect.Request[pb.ChangeDefaultCurrencyRequest]) (*connect.Response[pb.ChangeDefaultCurrencyResponse], error) {
	log := s.log.Named("ChangeDefaultCurrency")
	req := r.Msg
	log.Debug("Request received", zap.Any("request", req))
	requester := ctx.Value(nocloud.NoCloudAccount).(string)
	if !s.ca.HasAccess(ctx, requester, driver.NewDocumentID(schema.NAMESPACES_COL, schema.ROOT_NAMESPACE_KEY), access.Level_ROOT) {
		return nil, status.Error(codes.PermissionDenied, "Not enough Access rights to manage Currencies")
	}
	def := &pb.Currency{Id: schema.DEFAULT_CURRENCY_ID, Title: schema.DEFAULT_CURRENCY_NAME}

	if req.GetId() == def.GetId() {
		return nil, status.Error(codes.InvalidArgument, "Can't use root currency(NCU) as default currency")
	}

	trID, err := s.db.BeginTransaction(ctx, driver.TransactionCollections{
		Exclusive: []string{schema.CUR_COL, schema.CUR2CUR},
	}, &driver.BeginTransactionOptions{})
	if err != nil {
		log.Error("Failed to start transaction", zap.Error(err))
		return nil, status.Error(codes.Internal, "Failed to begin operation")
	}
	ctx = driver.WithTransactionID(ctx, trID)
	abort := func() {
		if err := s.db.AbortTransaction(ctx, trID, &driver.AbortTransactionOptions{}); err != nil {
			log.Error("Failed to abort transaction")
		}
	}

	currencies, err := s.ctrl.GetCurrencies(ctx, true)
	if err != nil {
		abort()
		log.Error("Failed to get currencies", zap.Error(err))
		return nil, status.Error(codes.Internal, "Failed to get existing currencies")
	}
	var prev *pb.Currency
	var next *pb.Currency
	var defaultsCount int
	for _, currency := range currencies {
		if currency.Default {
			prev = currency
			defaultsCount++
		}
		if req.GetId() == currency.GetId() {
			next = currency
		}
	}
	if defaultsCount > 1 {
		abort()
		log.Error("FATAL: More than 1 default currency. Fix it")
		return nil, status.Error(codes.Internal, "Fatal error: DB contains more than 1 default currency")
	}
	if next == nil {
		abort()
		return nil, status.Error(codes.InvalidArgument, "Currency you want to make default not found")
	}
	if next.Default {
		abort()
		return nil, status.Error(codes.InvalidArgument, "This currency is already default currency")
	}
	if prev != nil {
		prev.Default = false
		if err = s.ctrl.UpdateCurrency(ctx, prev); err != nil {
			abort()
			log.Error("Failed to remove old default currency mark", zap.Error(err))
			return nil, status.Error(codes.Internal, "Failed to remove old default currency")
		}
	}
	next.Default = true
	if err = s.ctrl.UpdateCurrency(ctx, next); err != nil {
		abort()
		log.Error("Failed to mark default", zap.Error(err))
		return nil, status.Error(codes.Internal, "Failed to mark default")
	}
	zeroRate := func(from, to *pb.Currency) error {
		if _, _, err = s.ctrl.GetExchangeRateDirect(ctx, from, to); err != nil {
			if driver.IsNotFoundGeneral(err) {
				if err = s.ctrl.CreateExchangeRate(ctx, from, to, 1, 0); err != nil {
					log.Error("Failed to create new exchange rate", zap.Error(err))
					return status.Error(codes.Internal, "Failed to create new exchange rate. Error: "+err.Error())
				}
			} else {
				log.Error("Failed to get direct rate", zap.Error(err))
				return status.Error(codes.Internal, "Failed to get direct rate. Error: "+err.Error())
			}
		} else {
			if err = s.ctrl.UpdateExchangeRate(ctx, from, to, 1, 0); err != nil {
				log.Error("Failed to update rate", zap.Error(err))
				return status.Error(codes.Internal, "Failed to update new rate. Error: "+err.Error())
			}
		}
		return nil
	}
	if errors.Join(zeroRate(next, def), zeroRate(def, next)); err != nil {
		abort()
		return nil, err
	}

	if err = s.db.CommitTransaction(ctx, trID, &driver.CommitTransactionOptions{}); err != nil {
		log.Error("Failed to commit transaction", zap.Error(err))
		return nil, status.Error(codes.Internal, "Failed to commit operation result")
	}

	return connect.NewResponse(&pb.ChangeDefaultCurrencyResponse{}), nil
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

	if req.GetTo().GetId() != schema.DEFAULT_CURRENCY_ID &&
		req.GetFrom().GetId() != schema.DEFAULT_CURRENCY_ID {
		return nil, status.Error(codes.InvalidArgument, "Can't create rate with 2 platform currencies. One edge must be root currency")
	}

	err := s.ctrl.CreateExchangeRate(ctx, req.From, req.To, req.Rate, req.Commission)
	if err != nil {
		log.Error("Error creating Exchange rate", zap.Error(err))
		return connect.NewResponse(&pb.CreateExchangeRateResponse{}), err
	}

	_ = s.ctrl.DeleteExchangeRate(ctx, req.To, req.From)
	if err = s.ctrl.CreateExchangeRate(ctx, req.To, req.From, 1/req.Rate, req.Commission); err != nil {
		log.Error("Error creating reverse rate", zap.Error(err))
		return nil, err
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

	if err := s.ctrl.UpdateExchangeRate(ctx, req.From, req.To, req.Rate, req.Commission); err != nil {
		log.Error("Error updating Exchange rate", zap.Error(err))
		return nil, err
	}
	_ = s.ctrl.DeleteExchangeRate(ctx, req.To, req.From)
	if err := s.ctrl.CreateExchangeRate(ctx, req.To, req.From, 1/req.Rate, req.Commission); err != nil {
		log.Error("Error updating reverse rate", zap.Error(err))
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
		log.Error("Error deleting exchange rate", zap.Error(err))
		return nil, err
	}
	if err := s.ctrl.DeleteExchangeRate(ctx, req.To, req.From); err != nil {
		log.Error("Error deleting reverse rate", zap.Error(err))
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

	mustFetch := make([]int32, 0)
	if !isAdmin && ok {
		acc, err := s.accounts.Get(ctx, requester)
		if err == nil {
			if acc.Currency == nil {
				mustFetch = append(mustFetch, schema.DEFAULT_CURRENCY_ID)
			} else {
				mustFetch = append(mustFetch, acc.Currency.GetId())
			}
		}
	}

	currencies, err := s.ctrl.GetCurrencies(ctx, isAdmin, mustFetch...)
	if err != nil {
		log.Error("Error getting Currencies", zap.Error(err))
		return nil, err
	}

	return connect.NewResponse(&pb.GetCurrenciesResponse{Currencies: currencies}), nil
}

func (s *CurrencyServiceServer) GetExchangeRates(ctx context.Context, r *connect.Request[pb.GetExchangeRatesRequest]) (*connect.Response[pb.GetExchangeRatesResponse], error) {
	var currencies map[int32]*pb.Currency
	curr, err := s.ctrl.GetCurrencies(ctx, true)
	if err != nil {
		return nil, err
	}
	for _, cur := range curr {
		currencies[cur.Id] = cur
	}
	rates, err := s.ctrl.GetExchangeRates(ctx)
	if err != nil {
		return nil, err
	}
	for _, rate := range rates {
		rate.From = currencies[rate.From.GetId()]
		rate.To = currencies[rate.To.GetId()]
	}

	return connect.NewResponse(&pb.GetExchangeRatesResponse{Rates: rates}), nil
}
