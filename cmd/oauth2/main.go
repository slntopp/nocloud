package main

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/golang-jwt/jwt/v4"
	"github.com/slntopp/nocloud-proto/registry"
	"github.com/slntopp/nocloud/pkg/graph"
	"github.com/slntopp/nocloud/pkg/nocloud"
	"github.com/slntopp/nocloud/pkg/nocloud/auth"
	"github.com/slntopp/nocloud/pkg/nocloud/connect_auth"
	"github.com/slntopp/nocloud/pkg/nocloud/connectdb"
	"github.com/slntopp/nocloud/pkg/oauth2"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"net/http"
	"strings"
	"time"
)

var (
	port string
	log  *zap.Logger

	corsAllowed  []string
	registryHost string
	redisHost    string
	SIGNING_KEY  []byte

	oauthIssuer string

	arangodbHost string
	arangodbCred string
	arangodbName string
)

func init() {
	viper.AutomaticEnv()
	log = nocloud.NewLogger()

	viper.SetDefault("PORT", "8000")

	viper.SetDefault("REGISTRY_HOST", "registry:8000")
	viper.SetDefault("REDIS_HOST", "redis:6379")

	viper.SetDefault("CORS_ALLOWED", []string{"*"})

	viper.SetDefault("SIGNING_KEY", "seeeecreet")

	port = viper.GetString("PORT")

	registryHost = viper.GetString("REGISTRY_HOST")
	redisHost = viper.GetString("REDIS_HOST")

	corsAllowed = strings.Split(viper.GetString("CORS_ALLOWED"), ",")

	SIGNING_KEY = []byte(viper.GetString("SIGNING_KEY"))

	oauthIssuer = viper.GetString("OAUTH_ISSUER")

	arangodbHost = viper.GetString("DB_HOST")
	arangodbCred = viper.GetString("DB_CRED")
	arangodbName = viper.GetString("DB_NAME")
}

type basicAuthorizer struct {
	ic  *connect_auth.Interceptor
	key []byte
}

func (b *basicAuthorizer) Authorize(ctx context.Context, r *http.Request, request oauth2.AuthorizeRequest) (oauth2.AuthorizeResult, error) {
	header := r.Header.Get("Authorization")
	segments := strings.Split(header, " ")
	if len(segments) != 2 {
		segments = []string{"", ""}
	}
	ctx, err := b.ic.JwtAuthMiddleware(ctx, segments[1])
	if err != nil {
		return oauth2.AuthorizeResult{}, err
	}
	account, _ := ctx.Value(nocloud.NoCloudAccount).(string)
	if account == "" {
		return oauth2.AuthorizeResult{}, fmt.Errorf("no account in token")
	}
	return oauth2.AuthorizeResult{
		Subject: account,
		Scopes:  request.Scopes,
	}, nil
}

func (b *basicAuthorizer) IssueAccessToken(ttl time.Duration, clientID, subject string, scopes []string) (oauth2.AccessToken, error) {
	now := time.Now().UTC()
	expirationTime := now.Add(ttl)
	claims := jwt.MapClaims{}
	claims[nocloud.NOCLOUD_ACCOUNT_CLAIM] = subject
	claims["expires"] = expirationTime.UTC().Unix()
	claims["exp"] = expirationTime.UTC().Unix()
	claims["scopes"] = scopes
	claims["client_id"] = clientID
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(b.key)
	if err != nil {
		return oauth2.AccessToken{}, fmt.Errorf("failed to sign token: %v", err)
	}
	return oauth2.AccessToken{
		Token:     tokenString,
		ClientID:  clientID,
		Subject:   subject,
		Scopes:    scopes,
		IssuedAt:  now,
		ExpiresAt: now.Add(ttl),
		Revoked:   false,
	}, nil
}

func main() {
	log.Info("Setting up DB Connection")
	db := connectdb.MakeDBConnection(log, arangodbHost, arangodbCred, arangodbName)
	log.Info("DB connection established")

	rdb := redis.NewClient(&redis.Options{
		Addr: redisHost,
		DB:   0,
	})

	auth.SetContext(log, rdb, SIGNING_KEY)

	registryConn, err := grpc.Dial(registryHost, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	defer registryConn.Close()

	registryClient := registry.NewAccountsServiceClient(registryConn)

	oauthRepository := graph.NewOAuthController(log, db, nil)
	authorizer := &basicAuthorizer{key: SIGNING_KEY, ic: connect_auth.NewInterceptor(log, rdb, SIGNING_KEY)}

	server := oauth2.NewOAuth2Server(log, SIGNING_KEY)
	server.SetupRegistryClient(registryClient)
	server.Start(port, corsAllowed, oauth2.Dependencies{
		Clients:       oauthRepository,
		Codes:         oauthRepository,
		Tokens:        oauthRepository,
		Authorization: authorizer,
	}, oauth2.Config{Issuer: oauthIssuer})
}
