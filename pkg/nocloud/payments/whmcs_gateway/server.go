package whmcs_gateway

import (
	"go.uber.org/zap"
	"io"
	"net/http"
)

func (g *WhmcsGateway) BuildWhmcsHooksHandler(log *zap.Logger) func(http.ResponseWriter, *http.Request) {
	log = log.Named("WhmcsHooks")
	return func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		log.Info("Request received", zap.Any("method", r.Method), zap.Any("body", string(b)), zap.Any("url", r.URL))
		w.WriteHeader(http.StatusOK)
	}
}
