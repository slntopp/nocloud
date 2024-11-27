/*
Copyright © 2021-2023 Nikita Ivanovski info@slnt-opp.xyz

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package auth

import (
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

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

var (
	log         *zap.Logger
	rdb         redisdb.Client
	SIGNING_KEY []byte
)

func SetContext(logger *zap.Logger, _rdb redisdb.Client, key []byte) {
	log = logger.Named("JWT")
	rdb = _rdb
	SIGNING_KEY = key
	log.Debug("Context set", zap.ByteString("signing_key", key))
}

func MakeToken(account string) (string, error) {
	claims := jwt.MapClaims{}
	claims[nocloud.NOCLOUD_ACCOUNT_CLAIM] = account
	claims[nocloud.NOCLOUD_INSTANCE_CLAIM] = "placeholder"
	claims[nocloud.NOCLOUD_ROOT_CLAIM] = 4
	claims[nocloud.NOCLOUD_NOSESSION_CLAIM] = true
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(SIGNING_KEY)
}

func MakeTokenInstance(instance string) (string, error) {
	claims := jwt.MapClaims{}
	claims[nocloud.NOCLOUD_ACCOUNT_CLAIM] = "placeholder"
	claims[nocloud.NOCLOUD_INSTANCE_CLAIM] = instance
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(SIGNING_KEY)
}

func JWT_STREAM_INTERCEPTOR(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	l := log.Named("StreamInterceptor")
	l.Debug("Invoked", zap.String("method", info.FullMethod))

	ctx, err := JWT_AUTH_MIDDLEWARE(stream.Context())
	if err != nil {
		return err
	}

	return handler(srv, &grpc_middleware.WrappedServerStream{
		ServerStream:   stream,
		WrappedContext: ctx,
	})
}

func JWT_AUTH_INTERCEPTOR(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	l := log.Named("Interceptor")
	l.Debug("Invoked", zap.String("method", info.FullMethod))

	switch info.FullMethod {
	case "/nocloud.registry.AccountsService/SignUp":
		return handler(ctx, req)
	case "/nocloud.health.InternalProbeService/Service":
		return handler(ctx, req)
	case "/nocloud.health.HealthService/Probe":
		probe := req.(*healthpb.ProbeRequest)
		if probe.ProbeType == "PING" {
			return handler(ctx, req)
		}
	case "/nocloud.services_providers.ServicesProvidersService/List":
		probe := req.(*sppb.ListRequest)
		if probe.Anonymously {
			return handler(ctx, req)
		}
	case "/nocloud.services_providers.ShowcasesService/List":
		probe := req.(*sppb.ListRequest)
		if probe.Anonymously {
			return handler(ctx, req)
		}
	case "/nocloud.billing.BillingService/ListPlans":
		probe := req.(*billpb.ListRequest)
		if probe.Anonymously {
			return handler(ctx, req)
		}
	case "/nocloud.billing.CurrencyService/GetExchangeRates":
		return handler(ctx, req)
	}
	ctx, err := JWT_AUTH_MIDDLEWARE(ctx)
	if info.FullMethod != "/nocloud.registry.AccountsService/Token" &&
		info.FullMethod != "/nocloud.billing.CurrencyService/GetCurrencies" &&
		info.FullMethod != "/nocloud.billing.AddonsService/List" {
		if err != nil {
			return nil, err
		}
	}

	go handleLogActivity(ctx)

	return handler(ctx, req)
}

func JWT_AUTH_MIDDLEWARE(ctx context.Context) (context.Context, error) {
	l := log.Named("Middleware")
	tokenString, err := grpc_auth.AuthFromMD(ctx, "bearer")
	if err != nil {
		l.Debug("Error extracting token", zap.Any("error", err))
		return ctx, err
	}

	token, err := validateToken(tokenString)
	if err != nil {
		return ctx, err
	}
	log.Debug("Validated token", zap.Any("claims", token))

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
		if err := sessions.Check(rdb, uuid, sid); err != nil {
			log.Debug("Session check failed", zap.Any("error", err))
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

func validateToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, status.Errorf(codes.Unauthenticated, "Unexpected signing method: %v", t.Header["alg"])
		}
		return SIGNING_KEY, nil
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

func handleLogActivity(ctx context.Context) {
	sid_ctx := ctx.Value(nocloud.NoCloudSession)
	if sid_ctx == nil {
		return
	}

	sid := sid_ctx.(string)
	req := ctx.Value(nocloud.NoCloudAccount).(string)
	exp := ctx.Value(nocloud.Expiration).(int64)

	if err := sessions.LogActivity(rdb, req, sid, exp); err != nil {
		log.Warn("Error logging activity", zap.Any("error", err))
	}
}
