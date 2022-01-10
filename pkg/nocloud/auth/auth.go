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
	"fmt"

	"github.com/golang-jwt/jwt/v4"
	"github.com/slntopp/nocloud/pkg/health/healthpb"
	"github.com/slntopp/nocloud/pkg/nocloud"
	"go.uber.org/zap"

	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

var (
	log *zap.Logger
	SIGNING_KEY		[]byte
)

func SetContext(logger *zap.Logger, key []byte) {
	log = logger.Named("JWT")
	SIGNING_KEY = key
	log.Debug("Context set", zap.ByteString("signing_key", key))
}

func JWT_AUTH_INTERCEPTOR(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	l := log.Named("Interceptor")
	l.Debug("Invoked", zap.String("method", info.FullMethod))

	switch info.FullMethod {
	case "/nocloud.registry.AccountsService/Token":
		return handler(ctx, req)
	case "/nocloud.health.HealthService/Probe":
		probe := req.(*healthpb.ProbeRequest)
		if probe.ProbeType == "PING" {
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

	account := token[nocloud.NOCLOUD_ACCOUNT_CLAIM]
	if account == nil {
		return nil, status.Error(codes.Unauthenticated, "Invalid token format: no requestor ID")
	}
	ctx = context.WithValue(ctx, nocloud.NoCloudAccount, account.(string))
	ctx = metadata.AppendToOutgoingContext(ctx, nocloud.NOCLOUD_ACCOUNT_CLAIM, account.(string))

	return ctx, nil
}

func validateToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, status.Error(codes.Unauthenticated, fmt.Sprintf("Unexpected signing method: %v", t.Header["alg"]))
		}
		return SIGNING_KEY, nil
	})
	
	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("Invalid token")
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		return claims, nil
	}

	return nil, status.Error(codes.Unauthenticated, "Cannot Validate Token")
}