package handlers

import (
	"github.com/gorilla/mux"
	"github.com/slntopp/nocloud-proto/registry"
	"github.com/slntopp/nocloud/pkg/oauth2/config"
)

type OAuthHandler interface {
	Setup(*mux.Router, config.OAuth2Config, registry.AccountsServiceClient)
}

func GetOAuthHandler(handlerType string) OAuthHandler {
	switch handlerType {
	case "google":
		return &GoogleOauthHandler{}
	case "github":
		return &GithubOauthHandler{}
	default:
		return nil
	}
}
