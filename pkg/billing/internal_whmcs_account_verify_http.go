package billing

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/slntopp/nocloud/pkg/nocloud"
	"github.com/slntopp/nocloud/pkg/nocloud/schema"
	"go.uber.org/zap"
)

const PathWhmcsAccountVerification = "/internal/whmcs/account-verification"

type whmcsAccountVerificationResponse struct {
	UUID                       string `json:"uuid"`
	IsPhoneVerified            bool   `json:"is_phone_verified"`
	RequirePhoneVerification   bool   `json:"require_phone_verification"`
	PhoneOKForWhmcsSSO         bool   `json:"phone_ok_for_whmcs_sso"`
}

func RegisterWhmcsModuleVerificationRoute(log *zap.Logger, router *mux.Router, srv *BillingServiceServer, secret string) {
	if strings.TrimSpace(secret) == "" {
		return
	}
	h := &whmcsAccountVerificationHandler{
		log:    log.Named("WhmcsAccountVerification"),
		srv:    srv,
		secret: strings.TrimSpace(secret),
	}
	router.Methods(http.MethodGet).Path(PathWhmcsAccountVerification).Handler(h)
}

type whmcsAccountVerificationHandler struct {
	log    *zap.Logger
	srv    *BillingServiceServer
	secret string
}

func (h *whmcsAccountVerificationHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	auth := strings.TrimSpace(r.Header.Get("Authorization"))
	const prefix = "Bearer "
	if !strings.HasPrefix(auth, prefix) || strings.TrimSpace(strings.TrimPrefix(auth, prefix)) != h.secret {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	uuid := strings.TrimSpace(r.URL.Query().Get("uuid"))
	if uuid == "" {
		http.Error(w, "uuid query parameter required", http.StatusBadRequest)
		return
	}

	ctx := context.WithValue(context.Background(), nocloud.NoCloudAccount, schema.ROOT_ACCOUNT_KEY)
	acc, err := h.srv.accounts.Get(ctx, uuid)
	if err != nil {
		h.log.Debug("account lookup failed", zap.String("uuid", uuid), zap.Error(err))
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	invConf := MakeInvoicesConf(h.srv.log, &h.srv.settingsClient)
	require := invConf.ForceRequirePhoneVerification
	verified := acc.GetIsPhoneVerified()
	ok := !require || verified

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(whmcsAccountVerificationResponse{
		UUID:                     uuid,
		IsPhoneVerified:          verified,
		RequirePhoneVerification: require,
		PhoneOKForWhmcsSSO:       ok,
	})
}
