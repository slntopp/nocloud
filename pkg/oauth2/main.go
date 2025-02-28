package oauth2

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"github.com/slntopp/nocloud-proto/registry"
	config "github.com/slntopp/nocloud/pkg/oauth2/config"
	"github.com/slntopp/nocloud/pkg/oauth2/handlers"
	"go.uber.org/zap"
	"net/http"
)

type OAuth2Server struct {
	router         *mux.Router
	registryClient registry.AccountsServiceClient
	signingKey     []byte

	log *zap.Logger
}

func NewOAuth2Server(log *zap.Logger, signingKey []byte) *OAuth2Server {
	log.Debug("New Server")
	return &OAuth2Server{
		log:        log.Named("OauthServer"),
		router:     mux.NewRouter(),
		signingKey: signingKey,
	}
}

func (s *OAuth2Server) SetupRegistryClient(registryClient registry.AccountsServiceClient) {
	s.log.Debug("Reg registry client")
	s.registryClient = registryClient
}

func (s *OAuth2Server) registerOAuthHandlers() {
	cfg, err := config.Config()
	s.log.Debug("Read config", zap.Any("cfg", cfg))
	if err != nil {
		s.log.Fatal("Failed to read config", zap.Error(err))
	}

	for key, conf := range cfg {
		s.log.Debug("Setting handler", zap.String("key", key))
		handler := handlers.GetOAuthHandler(key)
		if handler == nil {
			continue
		}

		handler.Setup(s.log.Named(key), s.router, conf, s.registryClient, s.signingKey)
	}

}

func (s *OAuth2Server) Start(port string, corsAllowed []string) {
	s.registerOAuthHandlers()

	s.router.HandleFunc("/oauth", func(writer http.ResponseWriter, request *http.Request) {
		cfg, err := config.Config()
		if err != nil {
			s.log.Error("Failed to read config", zap.Error(err))
			return
		}

		var resp []string

		for key := range cfg {
			resp = append(resp, key)
		}

		marshal, err := json.Marshal(resp)
		if err != nil {
			s.log.Error("Failed to marshal config", zap.Error(err))
			return
		}

		writer.Write(marshal)
	})

	handler := cors.New(cors.Options{
		AllowedOrigins:   corsAllowed,
		AllowedHeaders:   []string{"Content-Type", "Authorization", "*", "Connect-Protocol-Version", "grpc-metadata-nocloud-primary-currency-code", "NoCloud-Primary-Currency-Code"},
		AllowedMethods:   []string{"GET", "POST", "DELETE", "PUT", "PATCH", "OPTIONS", "HEAD"},
		AllowCredentials: true,
	}).Handler(s.router)

	s.log.Debug("listen", zap.String("port", port))
	err := http.ListenAndServe(fmt.Sprintf("0.0.0.0:%s", port), handler)
	if err != nil {
		s.log.Fatal("Failed to start server", zap.Error(err))
	}
}
