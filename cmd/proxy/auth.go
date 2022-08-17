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
			w.Write([]byte("Error converting claims"))
			return
		}

		if _, ok := claims[nocloud.NOCLOUD_ACCOUNT_CLAIM]; !ok {
			log.Warn("Error reading Claims", zap.Any("claims", claims))
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Claims don't contain Account UUID"))
			return
		}

		ctx := context.WithValue(r.Context(), "bearer", bearer)
		r.Header.Set("Sec-WebSocket-Protocol", strings.Join(headers[1:], ", "))
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
