package handlers

import (
	"github.com/gorilla/mux"
	"github.com/slntopp/nocloud/pkg/oauth2/config"
)

type OAuthHandler interface {
	Setup(*mux.Router, config.OAuth2Config)
}

func GetOAuthHandler(handlerType string) OAuthHandler {
	switch handlerType {
	case "google":
		return &GoogleOauthHandler{}
	default:
		return nil
	}
}
