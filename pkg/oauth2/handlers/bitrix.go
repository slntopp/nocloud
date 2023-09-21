package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/slntopp/nocloud-proto/registry"
	"github.com/slntopp/nocloud-proto/registry/accounts"
	"github.com/slntopp/nocloud/pkg/nocloud/auth"
	"github.com/slntopp/nocloud/pkg/nocloud/schema"
	"github.com/slntopp/nocloud/pkg/oauth2/config"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
	"google.golang.org/grpc/metadata"
	"io"
	"net/http"
	"sync"
	"time"
)

type BitrixOauthHandler struct {
	states map[string]string
	m      *sync.Mutex
}

type UserInfo struct {
	Result map[string]interface{} `json:"result"`
}

func (g *BitrixOauthHandler) Setup(log *zap.Logger, router *mux.Router, cfg config.OAuth2Config, regClient registry.AccountsServiceClient) {
	g.states = map[string]string{}
	g.m = &sync.Mutex{}

	oauth2Config := &oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		RedirectURL:  cfg.RedirectURL,
		Endpoint: oauth2.Endpoint{
			AuthURL:  cfg.AuthURL,
			TokenURL: cfg.TokenURL,
		},
	}

	userInfoUrl := cfg.UserInfoURL
	field := cfg.AuthField

	router.Handle("/oauth/bitrix/sign_in", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		state, redirect := r.FormValue("state"), r.FormValue("redirect")

		g.m.Lock()
		g.states[state] = redirect
		g.m.Unlock()

		url := oauth2Config.AuthCodeURL(state)
		http.Redirect(w, r, url, http.StatusTemporaryRedirect)
	}))
	router.Handle("/oauth/bitrix/checkout", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		state, code := r.FormValue("state"), r.FormValue("code")
		var redirect string

		g.m.Lock()

		if _, ok := g.states[state]; !ok {
			log.Debug("State string not equal to state", zap.String("state", state), zap.String("stateString", state))
			return
		}

		redirect = g.states[state]
		delete(g.states, state)
		g.m.Unlock()

		token, err := oauth2Config.Exchange(context.Background(), code)
		if err != nil {
			log.Error("Failed to get token from exchange", zap.Error(err))
			return
		}

		response, err := http.Get(fmt.Sprintf("%s?auth=%s", userInfoUrl, token.AccessToken))

		if err != nil {
			log.Error("Failed to make request", zap.Error(err))
			return
		}

		defer response.Body.Close()
		body, err := io.ReadAll(response.Body)

		if err != nil {
			log.Error("Failed to read body", zap.Error(err))
			return
		}

		var userInfo = UserInfo{}
		err = json.Unmarshal(body, &userInfo)
		if err != nil {
			log.Error("Failed unmarshal body", zap.Error(err))
			return
		}

		if userInfo.Result == nil {
			return
		}

		if len(userInfo.Result) != 1 {
			return
		}

		user := userInfo.Result
		value := user[field].(string)

		name := user["NAME"].(string)
		last_name := user["LAST_NAME"].(string)

		rootToken, err := auth.MakeToken(schema.ROOT_ACCOUNT_KEY)
		if err != nil {
			log.Error("Failed create token", zap.Error(err))
			return
		}

		ctx := metadata.AppendToOutgoingContext(context.Background(), "authorization", "bearer "+rootToken)

		resp, err := regClient.Token(ctx, &accounts.TokenRequest{
			Auth: &accounts.Credentials{
				Type: "oauth2-bitrix",
				Data: []string{
					field,
					value,
				},
			},
			Exp: int32(time.Now().Unix() + int64(time.Hour.Seconds()*2160)),
		})
		if err != nil {
			_, err = regClient.Create(ctx, &accounts.CreateRequest{
				Title:     fmt.Sprintf("%s %s", name, last_name),
				Namespace: schema.ROOT_NAMESPACE_KEY,
				Auth: &accounts.Credentials{
					Type: "oauth2-bitrix",
					Data: []string{
						field,
						value,
					},
				},
			})
			if err != nil {
				log.Error("Failed create account", zap.Error(err))
				return
			}
			resp, err = regClient.Token(ctx, &accounts.TokenRequest{
				Auth: &accounts.Credentials{
					Type: "oauth2-bitrix",
					Data: []string{
						field,
						value,
					},
				},
				Exp: int32(time.Now().Unix() + int64(time.Hour.Seconds()*2160)),
			})
			if err != nil {
				log.Error("Failed get token", zap.Error(err))
				return
			}
		}

		http.Redirect(w, r, fmt.Sprintf("%s?token=?%s", redirect, resp.GetToken()), http.StatusSeeOther)

	}))
}
