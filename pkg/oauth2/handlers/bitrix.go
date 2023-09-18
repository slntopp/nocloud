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
	"golang.org/x/oauth2"
	"google.golang.org/grpc/metadata"
	"io"
	"net/http"
	"time"
)

type BitrixHandler struct{}

type UserInfo struct {
	Result []map[string]interface{} `json:"result"`
}

func (g *BitrixHandler) Setup(router *mux.Router, cfg config.OAuth2Config, regClient registry.AccountsServiceClient) {
	oauth2Config := &oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		RedirectURL:  cfg.RedirectURL,
		Endpoint: oauth2.Endpoint{
			AuthURL:  cfg.AuthURL,
			TokenURL: cfg.TokenURL,
		},
	}

	stateString := cfg.StateString
	userInfoUrl := cfg.UserInfoURL
	field := cfg.AuthField

	router.Handle("/oauth/bitrix/login", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		url := oauth2Config.AuthCodeURL(stateString)
		http.Redirect(w, r, url, http.StatusTemporaryRedirect)
	}))
	router.Handle("/oauth/bitrix/checkout", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		state, code := r.FormValue("state"), r.FormValue("code")

		if state != stateString {
			return
		}

		token, err := oauth2Config.Exchange(context.Background(), code)
		if err != nil {
			return
		}

		response, err := http.Get(fmt.Sprintf("%s/rest/user.current.json?auth=%s", userInfoUrl, token.AccessToken))

		defer response.Body.Close()
		body, err := io.ReadAll(response.Body)

		var userInfo = UserInfo{}
		err = json.Unmarshal(body, &userInfo)
		if err != nil {
			return
		}

		if userInfo.Result == nil {
			return
		}

		if len(userInfo.Result) != 1 {
			return
		}

		user := userInfo.Result[0]
		value := user[field].(string)

		name := user["NAME"].(string)
		last_name := user["LAST_NAME"].(string)

		rootToken, err := auth.MakeToken(schema.ROOT_ACCOUNT_KEY)
		if err != nil {
			return
		}

		ctx := metadata.AppendToOutgoingContext(context.Background(), "authorization", "bearer "+rootToken)

		resp, err := regClient.Token(ctx, &accounts.TokenRequest{
			Auth: &accounts.Credentials{
				Type: "oauth2-github",
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
				return
			}
			resp, err = regClient.Token(ctx, &accounts.TokenRequest{
				Auth: &accounts.Credentials{
					Type: "oauth2-github",
					Data: []string{
						field,
						value,
					},
				},
				Exp: int32(time.Now().Unix() + int64(time.Hour.Seconds()*2160)),
			})
			if err != nil {
				return
			}
		}

		res := map[string]string{
			"token": resp.GetToken(),
		}
		marshal, _ := json.Marshal(res)
		w.Write(marshal)
	}))
}
