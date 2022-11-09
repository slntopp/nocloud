/*
Copyright © 2021-2022 Nikita Ivanovski info@slnt-opp.xyz

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
	billpb "github.com/slntopp/nocloud/pkg/billing/proto"
	healthpb "github.com/slntopp/nocloud/pkg/health/proto"
	"github.com/slntopp/nocloud/pkg/nocloud"
	sppb "github.com/slntopp/nocloud/pkg/services_providers/proto"
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
	SIGNING_KEY []byte
)

func SetContext(logger *zap.Logger, key []byte) {
	log = logger.Named("JWT")
	SIGNING_KEY = key
	log.Debug("Context set", zap.ByteString("signing_key", key))
}

func MakeToken(account string) (string, error) {
	claims := jwt.MapClaims{}
	claims[nocloud.NOCLOUD_ACCOUNT_CLAIM] = account
	claims[nocloud.NOCLOUD_INSTANCE_CLAIM] = "placeholder"
	claims[nocloud.NOCLOUD_ROOT_CLAIM] = 4
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
	case "/nocloud.registry.AccountsService/Token", "/nocloud.health.InternalProbeService/Service":
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
	case "/nocloud.billing.BillingService/ListPlans":
		probe := req.(*billpb.ListRequest)
		if probe.Anonymously {
			return handler(ctx, req)
		}
	}
	ctx, err := JWT_AUTH_MIDDLEWARE(ctx)
	if err != nil {
		return nil, err
	}

	return handler(ctx, req)
}

func JWT_AUTH_MIDDLEWARE(ctx context.Context) (context.Context, error) {
	l := log.Named("Middleware")
	tokenString, err := grpc_auth.AuthFromMD(ctx, "bearer")
	if err != nil {
		l.Debug("Error extracting token", zap.Any("error", err))
		return nil, err
	}

	token, err := validateToken(tokenString)
	if err != nil {
		return nil, err
	}
	log.Debug("Validated token", zap.Any("claims", token))

	acc := token[nocloud.NOCLOUD_ACCOUNT_CLAIM]
	if acc == nil {
		return nil, status.Error(codes.Unauthenticated, "Invalid token format: no requestor ID")
	}
	ctx = context.WithValue(ctx, nocloud.NoCloudAccount, acc.(string))
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
