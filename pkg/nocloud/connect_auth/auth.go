package connect_auth

import (
	"connectrpc.com/connect"
	"context"
	"errors"
	"github.com/golang-jwt/jwt/v4"
	billpb "github.com/slntopp/nocloud-proto/billing"
	healthpb "github.com/slntopp/nocloud-proto/health"
	sppb "github.com/slntopp/nocloud-proto/services_providers"
	"github.com/slntopp/nocloud/pkg/nocloud"
	redisdb "github.com/slntopp/nocloud/pkg/nocloud/redis"
	"github.com/slntopp/nocloud/pkg/nocloud/sessions"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"strings"
)

type interceptor struct {
	log         *zap.Logger
	rdb         redisdb.Client
	signing_key []byte
}

func NewInterceptor(logger *zap.Logger, rdb redisdb.Client, key []byte) *interceptor {
	return &interceptor{
		log:         logger,
		rdb:         rdb,
		signing_key: key,
	}
}

func (i *interceptor) MakeToken(account string) (string, error) {
	claims := jwt.MapClaims{}
	claims[nocloud.NOCLOUD_ACCOUNT_CLAIM] = account
	claims[nocloud.NOCLOUD_INSTANCE_CLAIM] = "placeholder"
	claims[nocloud.NOCLOUD_ROOT_CLAIM] = 4
	claims[nocloud.NOCLOUD_NOSESSION_CLAIM] = true
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(i.signing_key)
}

func (i *interceptor) MakeTokenInstance(instance string) (string, error) {
	claims := jwt.MapClaims{}
	claims[nocloud.NOCLOUD_ACCOUNT_CLAIM] = "placeholder"
	claims[nocloud.NOCLOUD_INSTANCE_CLAIM] = instance
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(i.signing_key)
}

func (i *interceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return connect.UnaryFunc(func(
		ctx context.Context,
		req connect.AnyRequest,
	) (connect.AnyResponse, error) {
		i.log.Debug("Invoked", zap.String("method", req.Spec().Procedure))

		header := req.Header().Get("Authorization")

		switch req.Spec().Procedure {
		case "/nocloud.registry.AccountsService/SignUp":
			return next(ctx, req)
		case "/nocloud.health.InternalProbeService/Service":
			return next(ctx, req)
		case "/nocloud.health.HealthService/Probe":
			probe := req.Any().(*healthpb.ProbeRequest)
			if probe.GetProbeType() == "PING" {
				return next(ctx, req)
			}
		case "/nocloud.services_providers.ServicesProvidersService/List":
			probe := req.Any().(*sppb.ListRequest)
			if probe.GetAnonymously() {
				return next(ctx, req)
			}
		case "/nocloud.services_providers.ShowcasesService/List":
			probe := req.Any().(*sppb.ListRequest)
			if probe.GetAnonymously() {
				return next(ctx, req)
			}
		case "/nocloud.billing.BillingService/ListPlans":
			probe := req.Any().(*billpb.ListRequest)
			if probe.GetAnonymously() {
				return next(ctx, req)
			}
		case "/nocloud.billing.CurrencyService/GetExchangeRates":
			return next(ctx, req)
		case "/nocloud.billing.PromocodesService/ApplySale":
			return next(ctx, req)
		}

		segments := strings.Split(header, " ")
		if len(segments) != 2 {
			segments = []string{"", ""}
		}

		ctx, err := i.jwtAuthMiddleware(ctx, segments[1])
		if req.Spec().Procedure != "/nocloud.registry.AccountsService/Token" &&
			req.Spec().Procedure != "/nocloud.billing.CurrencyService/GetCurrencies" &&
			req.Spec().Procedure != "/nocloud.billing.AddonsService/List" &&
			req.Spec().Procedure != "/nocloud.billing.PromocodesService/GetByCode" &&
			req.Spec().Procedure != "/nocloud.services_providers.ShowcasesService/Get" &&
			req.Spec().Procedure != "/nocloud.billing.CurrencyService/Convert" &&
			req.Spec().Procedure != "/nocloud.billing.CurrencyService/ConvertMany" &&
			req.Spec().Procedure != "/nocloud.billing.BillingService/GetPlan" &&
			req.Spec().Procedure != "/nocloud.billing.DescriptionsService/Get" {
			if err != nil {
				return nil, err
			}
		}

		go i.handleLogActivity(ctx)

		return next(ctx, req)
	})
}

func (i *interceptor) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	i.log.Debug("WrapStreamingClient")
	return next
}

func (i *interceptor) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	i.log.Debug("Setup Wrap Streaming Handler")
	return func(ctx context.Context, shc connect.StreamingHandlerConn) error {
		l := i.log.Named("StreamInterceptor")
		l.Debug("Invoked", zap.String("method", shc.Spec().Procedure))

		header := shc.RequestHeader().Get("Authorization")

		segments := strings.Split(header, " ")
		if len(segments) != 2 {
			segments = []string{"", ""}
		}

		if strings.ToLower(segments[0]) != "bearer" {
			return errors.New("wrong auth type")
		}

		ctx, err := i.jwtAuthMiddleware(ctx, segments[1])
		if err != nil {
			return err
		}

		return next(ctx, shc)
	}
}

func (i *interceptor) jwtAuthMiddleware(ctx context.Context, tokenString string) (context.Context, error) {
	l := i.log.Named("Middleware")

	token, err := i.validateToken(tokenString)
	if err != nil {
		return ctx, err
	}
	l.Debug("Validated token", zap.Any("claims", token))

	acc := token[nocloud.NOCLOUD_ACCOUNT_CLAIM]
	if acc == nil {
		return ctx, status.Error(codes.Unauthenticated, "Invalid token format: no requestor ID")
	}
	uuid, ok := acc.(string)
	if !ok {
		return ctx, status.Error(codes.Unauthenticated, "Invalid token format: requestor ID isn't string")
	}

	if token[nocloud.NOCLOUD_NOSESSION_CLAIM] == nil {
		session := token[nocloud.NOCLOUD_SESSION_CLAIM]
		if session == nil {
			return ctx, status.Error(codes.Unauthenticated, "Invalid token format: no session ID")
		}
		sid, ok := session.(string)
		if !ok {
			return ctx, status.Error(codes.Unauthenticated, "Invalid token format: session ID isn't string")
		}

		// Check if session is valid
		if err := sessions.Check(i.rdb, uuid, sid); err != nil {
			i.log.Debug("Session check failed", zap.Any("error", err))
			return ctx, status.Error(codes.Unauthenticated, "Session is expired, revoked or invalid")
		}

		ctx = context.WithValue(ctx, nocloud.NoCloudSession, sid)
	}

	var exp int64
	if token["expires"] != nil {
		exp = int64(token["expires"].(float64))
	}

	ctx = context.WithValue(ctx, nocloud.NoCloudAccount, acc.(string))
	ctx = context.WithValue(ctx, nocloud.Expiration, exp)
	ctx = metadata.AppendToOutgoingContext(ctx, nocloud.NOCLOUD_ACCOUNT_CLAIM, acc.(string))

	ctx, err = func(ctx context.Context) (context.Context, error) {
		sp := token[nocloud.NOCLOUD_SP_CLAIM]
		if sp == nil {
			return ctx, nil
		}

		s, ok := sp.(string)

		if !ok {
			return ctx, errors.New("wrong type of sp")
		}

		ctx = context.WithValue(ctx, nocloud.NoCloudSp, s)
		return metadata.AppendToOutgoingContext(ctx, nocloud.NOCLOUD_SP_CLAIM, sp.(string)), nil
	}(ctx)

	if err != nil {
		return ctx, err
	}

	ctx, err = func(ctx context.Context) (context.Context, error) {
		inst := token[nocloud.NOCLOUD_INSTANCE_CLAIM]
		if inst == nil {
			return ctx, nil
		}

		s, ok := inst.(string)

		if !ok {
			return ctx, errors.New("wrong type of inst")
		}

		ctx = context.WithValue(ctx, nocloud.NoCloudInstance, s)
		return metadata.AppendToOutgoingContext(ctx, nocloud.NOCLOUD_INSTANCE_CLAIM, inst.(string)), nil
	}(ctx)

	if err != nil {
		return ctx, err
	}

	ctx = func(ctx context.Context) context.Context {
		rootAccessClaim := token[nocloud.NOCLOUD_ROOT_CLAIM]
		lvlF, ok := rootAccessClaim.(float64)
		if !ok {
			return ctx
		}

		return context.WithValue(ctx, nocloud.NoCloudRootAccess, int(lvlF))
	}(ctx)

	ctx = context.WithValue(ctx, nocloud.NoCloudToken, tokenString)
	ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "bearer "+tokenString)

	return ctx, nil
}

func (i *interceptor) validateToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, status.Errorf(codes.Unauthenticated, "Unexpected signing method: %v", t.Header["alg"])
		}
		return i.signing_key, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		return claims, nil
	}

	return nil, status.Error(codes.Unauthenticated, "Cannot Validate Token")
}

func (i *interceptor) handleLogActivity(ctx context.Context) {
	sid_ctx := ctx.Value(nocloud.NoCloudSession)
	if sid_ctx == nil {
		return
	}

	sid := sid_ctx.(string)
	req := ctx.Value(nocloud.NoCloudAccount).(string)
	exp := ctx.Value(nocloud.Expiration).(int64)

	if err := sessions.LogActivity(i.rdb, req, sid, exp); err != nil {
		i.log.Warn("Error logging activity", zap.Any("error", err))
	}
}
