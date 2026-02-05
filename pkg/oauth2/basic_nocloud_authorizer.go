package oauth2

import (
	"context"
	"fmt"
	"github.com/golang-jwt/jwt/v4"
	"github.com/slntopp/nocloud/pkg/nocloud"
	"github.com/slntopp/nocloud/pkg/nocloud/connect_auth"
	"net/http"
	"strings"
	"time"
)

type BasicAuthorizer struct {
	Ic  *connect_auth.Interceptor
	Key []byte
}

func (b *BasicAuthorizer) Subject(ctx context.Context, r *http.Request) (subject string, ok bool, err error) {
	var token string
	tokenCookie, err := r.Cookie("nocloud_token")
	if err == nil {
		token = tokenCookie.Value
	} else {
		header := r.Header.Get("Authorization")
		segments := strings.Split(header, " ")
		if len(segments) != 2 {
			segments = []string{"", ""}
		}
		token = segments[1]
	}
	ctx, err = b.Ic.JwtAuthMiddleware(ctx, token)
	if err != nil {
		return "", false, nil
	}
	account, _ := ctx.Value(nocloud.NoCloudAccount).(string)
	if account == "" {
		return "", false, nil
	}
	return account, true, nil
}

func (b *BasicAuthorizer) ConsentedScopes(_ context.Context, subject, clientID string) ([]string, error) {
	return make([]string, 0), nil // Always request consent window
}

func (b *BasicAuthorizer) SaveConsent(_ context.Context, subject, clientID string, scopes []string) error {
	// Must be implemented if you want ConsentedScopes method to work
	return nil
}

func (b *BasicAuthorizer) IssueAccessToken(ttl time.Duration, clientID, subject string, scopes []string) (AccessToken, error) {
	now := time.Now().UTC()
	expirationTime := now.Add(ttl)
	claims := jwt.MapClaims{}
	claims[nocloud.NOCLOUD_ACCOUNT_CLAIM] = subject
	claims["expires"] = expirationTime.UTC().Unix()
	claims["exp"] = expirationTime.UTC().Unix()
	claims["scopes"] = scopes
	claims["client_id"] = clientID
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(b.Key)
	if err != nil {
		return AccessToken{}, fmt.Errorf("failed to sign token: %v", err)
	}
	return AccessToken{
		Token:     tokenString,
		ClientID:  clientID,
		Subject:   subject,
		Scopes:    scopes,
		IssuedAt:  now,
		ExpiresAt: now.Add(ttl),
		Revoked:   false,
	}, nil
}
