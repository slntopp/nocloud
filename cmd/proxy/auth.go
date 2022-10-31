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
package main

import (
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v4"
	"github.com/slntopp/nocloud/pkg/nocloud"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ContextKey string

func AuthMiddleware(next http.Handler) http.Handler {
	log.Info("Using Auth Middleware")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		protocol := r.Header.Get("Sec-WebSocket-Protocol")

		headers := strings.Split(protocol, ", ")
		bearer := headers[0]
		if bearer == "" {
			log.Warn("Bearer is empty", zap.String("bearer", bearer))
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		token, err := jwt.Parse(bearer, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, status.Errorf(codes.Unauthenticated, "Unexpected signing method: %v", t.Header["alg"])
			}
			return SIGNING_KEY, nil
		})
		if err != nil {
			log.Warn("Invalid token", zap.Error(err))
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		var claims jwt.MapClaims
		var ok bool
		if claims, ok = token.Claims.(jwt.MapClaims); !ok {
			log.Warn("Error converting Claims", zap.Any("claims", token.Claims))
			w.WriteHeader(http.StatusBadRequest)
			if _, err = w.Write([]byte("Error converting claims")); err != nil {
				log.Warn("Couldn't Write response bytes", zap.Error(err))
			}
			return
		}

		if _, ok := claims[nocloud.NOCLOUD_ACCOUNT_CLAIM]; !ok {
			log.Warn("Error reading Claims", zap.Any("claims", claims))
			w.WriteHeader(http.StatusBadRequest)
			if _, err = w.Write([]byte("Claims don't contain Account UUID")); err != nil {
				log.Warn("Couldn't Write response bytes", zap.Error(err))
			}
			return
		}

		ctx := context.WithValue(r.Context(), ContextKey("bearer"), bearer)
		r.Header.Set("Sec-WebSocket-Protocol", strings.Join(headers[1:], ", "))
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
