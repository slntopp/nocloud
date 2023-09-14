package handlers

import (
	"github.com/dghubble/gologin/v2"
	github_gologin "github.com/dghubble/gologin/v2/github"
	"github.com/gorilla/mux"
	"github.com/slntopp/nocloud/pkg/oauth2/config"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"net/http"
)

type GithubOauthHandler struct{}

func (g *GithubOauthHandler) Setup(router *mux.Router, cfg config.OAuth2Config) {
	oauth2Config := &oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		RedirectURL:  cfg.RedirectURL,
		Endpoint:     github.Endpoint,
	}
	stateConfig := gologin.DefaultCookieConfig

	router.Handle("/oauth/github/login", github_gologin.StateHandler(stateConfig, github_gologin.LoginHandler(oauth2Config, nil)))
	router.Handle("/oauth/github/checkout", github_gologin.StateHandler(stateConfig, github_gologin.CallbackHandler(oauth2Config, g.successHandler(), nil)))
}

func (g *GithubOauthHandler) successHandler() http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		github_gologin.UserFromContext(request.Context())
	})
}
