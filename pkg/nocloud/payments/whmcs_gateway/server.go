package whmcs_gateway

import (
	"github.com/slntopp/nocloud/pkg/nocloud"
	"github.com/slntopp/nocloud/pkg/nocloud/schema"
	"go.uber.org/zap"
	"io"
	"net/http"
)

func (g *WhmcsGateway) BuildWhmcsHooksHandler(log *zap.Logger) func(http.ResponseWriter, *http.Request) {
	log = log.Named("WhmcsHooks")
	return func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		log.Info("Request received", zap.Any("method", r.Method), zap.Any("body", string(b)), zap.Any("url", r.URL))
		requester := r.Context().Value(nocloud.NoCloudAccount)
		requesterStr, ok := requester.(string)
		if !ok {
			log.Error("Error converting requester to string")
			w.WriteHeader(http.StatusForbidden)
			_, _ = w.Write([]byte("unauthorized access"))
			return
		}
		if requesterStr != schema.ROOT_ACCOUNT_KEY {
			log.Error("Not root account")
			w.WriteHeader(http.StatusForbidden)
			_, _ = w.Write([]byte("only root access allowed"))
			return
		}
		go g.handleWhmcsEvent(log, r)
		w.WriteHeader(http.StatusOK)
	}
}
