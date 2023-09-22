package handlers

import (
	"github.com/gorilla/mux"
	"github.com/slntopp/nocloud-proto/registry"
	"github.com/slntopp/nocloud/pkg/oauth2/config"
	"go.uber.org/zap"
)

type OAuthHandler interface {
	Setup(*zap.Logger, *mux.Router, config.OAuth2Config, registry.AccountsServiceClient, []byte)
}

func GetOAuthHandler(handlerType string) OAuthHandler {
	switch handlerType {
	case "google":
		return &GoogleOauthHandler{}
	case "github":
		return &GithubOauthHandler{}
	case "bitrix":
		return &BitrixOauthHandler{}
	default:
		return nil
	}
}
