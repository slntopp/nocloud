package handlers

import (
	"context"
	"encoding/json"
	"github.com/dghubble/gologin/v2"
	google_gologin "github.com/dghubble/gologin/v2/google"
	"github.com/gorilla/mux"
	"github.com/slntopp/nocloud-proto/registry"
	"github.com/slntopp/nocloud-proto/registry/accounts"
	"github.com/slntopp/nocloud/pkg/nocloud/auth"
	"github.com/slntopp/nocloud/pkg/nocloud/schema"
	"github.com/slntopp/nocloud/pkg/oauth2/config"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/grpc/metadata"
	"net/http"
	"time"
)

type GoogleOauthHandler struct{}

func (g *GoogleOauthHandler) Setup(router *mux.Router, cfg config.OAuth2Config, regClient registry.AccountsServiceClient) {
	oauth2Config := &oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		RedirectURL:  cfg.RedirectURL,
		Endpoint:     google.Endpoint,
	}
	stateConfig := gologin.DefaultCookieConfig

	router.Handle("/oauth/google/login", google_gologin.StateHandler(stateConfig, google_gologin.LoginHandler(oauth2Config, nil)))
	router.Handle("/oauth/google/checkout", google_gologin.StateHandler(stateConfig, google_gologin.CallbackHandler(oauth2Config, g.successHandler(regClient), nil)))
}

func (g *GoogleOauthHandler) successHandler(regClient registry.AccountsServiceClient) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		user, err := google_gologin.UserFromContext(request.Context())
		if err != nil {
			return
		}

		token, err := auth.MakeToken(schema.ROOT_ACCOUNT_KEY)
		if err != nil {
			return
		}

		ctx := metadata.AppendToOutgoingContext(context.Background(), "authorization", "bearer "+token)

		response, err := regClient.Token(ctx, &accounts.TokenRequest{
			Auth: &accounts.Credentials{
				Type: "oauth2-google",
				Data: []string{
					"email",
					user.Email,
				},
			},
			Exp: int32(time.Now().Unix() + int64(time.Hour.Seconds()*2160)),
		})
		if err != nil {
			_, err = regClient.Create(ctx, &accounts.CreateRequest{
				Title:     user.Name,
				Namespace: schema.ROOT_NAMESPACE_KEY,
				Auth: &accounts.Credentials{
					Type: "oauth2-google",
					Data: []string{
						"email",
						user.Email,
					},
				},
			})
			if err != nil {
				return
			}
			response, err = regClient.Token(ctx, &accounts.TokenRequest{
				Auth: &accounts.Credentials{
					Type: "oauth2-google",
					Data: []string{
						"email",
						user.Email,
					},
				},
				Exp: int32(time.Now().Unix() + int64(time.Hour.Seconds()*2160)),
			})
			if err != nil {
				return
			}
		}

		res := map[string]string{
			"token": response.GetToken(),
		}
		marshal, _ := json.Marshal(res)
		writer.Write(marshal)
	})
}
