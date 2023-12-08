package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/golang-jwt/jwt/v4"
	"github.com/gorilla/mux"
	"github.com/slntopp/nocloud-proto/registry"
	"github.com/slntopp/nocloud-proto/registry/accounts"
	"github.com/slntopp/nocloud/pkg/nocloud"
	"github.com/slntopp/nocloud/pkg/nocloud/auth"
	"github.com/slntopp/nocloud/pkg/nocloud/schema"
	"github.com/slntopp/nocloud/pkg/oauth2/config"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

type GithubOauthHandler struct {
	states map[string]*StateInfo
	m      *sync.Mutex
}

func (g *GithubOauthHandler) Setup(
	log *zap.Logger,
	router *mux.Router,
	cfg config.OAuth2Config,
	regClient registry.AccountsServiceClient,
	signingKey []byte,
) {
	g.states = map[string]*StateInfo{}
	g.m = &sync.Mutex{}

	oauth2Config := &oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		RedirectURL:  cfg.RedirectURL,
		Endpoint:     github.Endpoint,
	}

	userInfoUrl := cfg.UserInfoURL

	router.Handle("/oauth/github/sign_in", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		state, redirect := r.FormValue("state"), r.FormValue("redirect")

		g.m.Lock()
		g.states[state] = &StateInfo{
			RedirectUrl: redirect,
			Method:      "sign_in",
		}
		log.Debug("Put state", zap.Any("state", g.states[state]))
		g.m.Unlock()

		url := oauth2Config.AuthCodeURL(state)

		result := map[string]string{
			"url": url,
		}

		marshal, _ := json.Marshal(result)
		w.Write(marshal)
	}))

	router.Handle("/oauth/github/link", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		state, redirect := r.FormValue("state"), r.FormValue("redirect")

		authHeader := r.Header.Get("Authorization")
		authHeaderSplit := strings.Split(authHeader, " ")

		if len(authHeaderSplit) != 2 {
			log.Error("Len is not 2")
			return
		}

		if strings.ToLower(authHeaderSplit[0]) != "bearer" {
			log.Error("No bearer")
			return
		}

		g.m.Lock()
		g.states[state] = &StateInfo{
			RedirectUrl: redirect,
			Token:       authHeaderSplit[1],
			Method:      "link",
		}
		g.m.Unlock()

		url := oauth2Config.AuthCodeURL(state)

		result := map[string]string{
			"url": url,
		}

		marshal, _ := json.Marshal(result)
		w.Write(marshal)
	}))

	router.Handle("/oauth/github/checkout", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		state, code := r.FormValue("state"), r.FormValue("code")

		g.m.Lock()

		if _, ok := g.states[state]; !ok {
			log.Debug("State string not equal to state", zap.String("state", state), zap.String("stateString", state))
			return
		}

		stateInfo := g.states[state]

		log.Debug("Get state", zap.Any("state", stateInfo))
		delete(g.states, state)
		g.m.Unlock()

		token, err := oauth2Config.Exchange(context.Background(), code)
		if err != nil {
			log.Error("Failed to get token from exchange", zap.Error(err))
			return
		}

		request, err := http.NewRequest("GET", userInfoUrl, nil)
		if err != nil {
			log.Error("Failed to create user info request", zap.Error(err))
			return
		}
		request.Header.Set("Accept", "application/vnd.github+json")
		request.Header.Set("Authorization", "Bearer "+token.AccessToken)
		request.Header.Set("X-GitHub-Api-Version", "2022-11-28")

		client := http.Client{}
		response, err := client.Do(request)

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

		var userInfo = map[string]interface{}{}
		err = json.Unmarshal(body, &userInfo)
		if err != nil {
			log.Error("Failed unmarshal body", zap.Error(err))
			return
		}

		log.Debug("Info", zap.Any("info", userInfo))

		name := userInfo["name"].(string)

		rootToken, err := auth.MakeToken(schema.ROOT_ACCOUNT_KEY)
		if err != nil {
			log.Error("Failed create token", zap.Error(err))
			return
		}

		ctx := metadata.AppendToOutgoingContext(context.Background(), "authorization", "bearer "+rootToken)

		if stateInfo.Method == "link" {
			ncToken, err := jwt.Parse(stateInfo.Token, func(t *jwt.Token) (interface{}, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, status.Errorf(codes.Unauthenticated, "Unexpected signing method: %v", t.Header["alg"])
				}
				return signingKey, nil
			})

			if err != nil {
				log.Error("Failed to get token", zap.Error(err))
				return
			}

			if !ncToken.Valid {
				log.Error("Token invalid", zap.Error(err))
				return
			}

			claims := ncToken.Claims.(jwt.MapClaims)
			acc := claims[nocloud.NOCLOUD_ACCOUNT_CLAIM].(string)

			resp, err := regClient.SetCredentials(ctx, &accounts.SetCredentialsRequest{
				Account: acc,
				Auth: &accounts.Credentials{
					Type: "oauth2-github",
					Data: []string{
						token.AccessToken,
					},
				},
			})

			if err != nil {
				log.Error("Failed set creds", zap.Error(err))
				return
			}

			if !resp.GetResult() {
				log.Error("False result")
				return
			}

			get, err := regClient.Get(ctx, &accounts.GetRequest{
				Uuid: acc,
			})

			if err != nil {
				log.Error("Failed get acc", zap.Error(err))
				return
			}

			if get.GetData() == nil {
				get.Data = &structpb.Struct{
					Fields: make(map[string]*structpb.Value),
				}
			}

			if get.GetData().GetFields() == nil {
				get.Data.Fields = map[string]*structpb.Value{}
			}

			_, ok := get.GetData().GetFields()["oauth_types"]

			if !ok {
				list, _ := structpb.NewList([]interface{}{
					"oauth2-github",
				})
				get.Data.Fields["oauth_types"] = structpb.NewListValue(list)
			} else {
				get.Data.GetFields()["oauth_types"].GetListValue().Values = append(get.Data.GetFields()["oauth_types"].GetListValue().GetValues(), structpb.NewStringValue("oauth2-github"))
			}

			_, err = regClient.Update(ctx, get)
			if err != nil {
				log.Error("Failed to update")
				return
			}

			http.Redirect(w, r, fmt.Sprintf("%s?token=%s", stateInfo.RedirectUrl, stateInfo.Token), http.StatusSeeOther)
		} else {
			resp, err := regClient.Token(ctx, &accounts.TokenRequest{
				Auth: &accounts.Credentials{
					Type: "oauth2-github",
					Data: []string{
						token.AccessToken,
					},
				},
				Exp: int32(time.Now().Unix() + int64(time.Hour.Seconds()*2160)),
			})
			if err != nil {
				list, _ := structpb.NewList([]interface{}{
					"oauth2-github",
				})
				listValue := structpb.NewListValue(list)

				data := &structpb.Struct{
					Fields: map[string]*structpb.Value{
						"oauth_types": listValue,
					},
				}
				_, err := regClient.Create(ctx, &accounts.CreateRequest{
					Title:     name,
					Namespace: schema.ROOT_NAMESPACE_KEY,
					Data:      data,
					Auth: &accounts.Credentials{
						Type: "oauth2-github",
						Data: []string{
							token.AccessToken,
						},
					},
				})
				if err != nil {
					log.Error("Failed create account", zap.Error(err))
					return
				}
				resp, err = regClient.Token(ctx, &accounts.TokenRequest{
					Auth: &accounts.Credentials{
						Type: "oauth2-github",
						Data: []string{
							token.AccessToken,
						},
					},
					Exp: int32(time.Now().Unix() + int64(time.Hour.Seconds()*2160)),
				})
				if err != nil {
					log.Error("Failed get token", zap.Error(err))
					return
				}
			}
			http.Redirect(w, r, fmt.Sprintf("%s?token=%s", stateInfo.RedirectUrl, resp.GetToken()), http.StatusSeeOther)
		}
	}))
}
