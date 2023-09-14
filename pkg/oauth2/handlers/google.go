package handlers

import (
	"github.com/slntopp/nocloud/pkg/oauth2/config"
	"net/http"

	"github.com/dghubble/gologin/v2"
	"github.com/gorilla/mux"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	google_gologin "github.com/dghubble/gologin/v2/google"
)

type GoogleOauthHandler struct{}

func (g *GoogleOauthHandler) Setup(router *mux.Router, cfg config.OAuth2Config) {
	oauth2Config := &oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		RedirectURL:  cfg.RedirectURL,
		Endpoint:     google.Endpoint,
	}
	stateConfig := gologin.DefaultCookieConfig

	router.Handle("/oauth/google/login", google_gologin.StateHandler(stateConfig, google_gologin.LoginHandler(oauth2Config, nil)))
	router.Handle("/oauth/google/checkout", google_gologin.StateHandler(stateConfig, google_gologin.CallbackHandler(oauth2Config, g.successHandler(), nil)))
}

func (g *GoogleOauthHandler) successHandler() http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {

	})
}
