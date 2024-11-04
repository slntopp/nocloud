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
	"github.com/slntopp/nocloud/pkg/nocloud"
	redisdb "github.com/slntopp/nocloud/pkg/nocloud/redis"
	"github.com/slntopp/nocloud/pkg/nocloud/sessions"
	"go.uber.org/zap"

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

func JWT_AUTH_INTERCEPTOR(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	l := log.Named("Interceptor")
	l.Debug("Invoked", zap.String("method", info.FullMethod))

	if info.FullMethod == "/nocloud.health.InternalProbeService/Service" {
		return handler(ctx, req)
	}

	ctx, err := JWT_AUTH_MIDDLEWARE(ctx)
	if err != nil {
		return nil, err
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
	rootAccessClaim := token[nocloud.NOCLOUD_ROOT_CLAIM]
	lvlF, ok := rootAccessClaim.(float64)
	if !ok {
		return ctx, status.Error(codes.Unauthenticated, "Root Access Claim not provided")
	}
	lvl := int(lvlF)
	if lvl == 0 {
		return ctx, status.Error(codes.PermissionDenied, "Account has no root access")
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
	ctx = context.WithValue(ctx, nocloud.NoCloudRootAccess, lvl)

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
