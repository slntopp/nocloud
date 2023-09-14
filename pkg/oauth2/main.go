package oauth2

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"github.com/slntopp/nocloud-proto/registry"
	config2 "github.com/slntopp/nocloud/pkg/oauth2/config"
	"github.com/slntopp/nocloud/pkg/oauth2/handlers"
	"go.uber.org/zap"
	"net/http"
)

type OAuth2Server struct {
	router         *mux.Router
	registryClient registry.AccountsServiceClient

	log *zap.Logger
}

func NewOAuth2Server(log *zap.Logger, host string) *OAuth2Server {
	return &OAuth2Server{
		log:    log,
		router: mux.NewRouter(),
	}
}

func (s *OAuth2Server) SetupRegistryClient(registryClient registry.AccountsServiceClient) {
	s.registryClient = registryClient
}

func (s *OAuth2Server) registerOAuthHandlers() {
	config, err := config2.Config()
	if err != nil {
		s.log.Fatal("Failed to read config", zap.Error(err))
	}

	for key, conf := range config {
		handler := handlers.GetOAuthHandler(key)
		if handler == nil {
			continue
		}

		handler.Setup(s.router, conf, s.registryClient)
	}

}

func (s *OAuth2Server) Start(port string, corsAllowed []string) {
	s.registerOAuthHandlers()

	handler := cors.New(cors.Options{
		AllowedOrigins:   corsAllowed,
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowedMethods:   []string{"GET", "POST", "DELETE", "PUT", "PATCH", "OPTIONS", "HEAD"},
		AllowCredentials: true,
	}).Handler(s.router)

	err := http.ListenAndServe(fmt.Sprintf("localhost:%s", port), handler)
	if err != nil {
		s.log.Fatal("Failed to start server", zap.Error(err))
	}
}
