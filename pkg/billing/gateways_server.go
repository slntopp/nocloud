package billing

import (
	"github.com/arangodb/go-driver"
	"github.com/gorilla/mux"
	accesspb "github.com/slntopp/nocloud-proto/access"
	"github.com/slntopp/nocloud/pkg/graph"
	"github.com/slntopp/nocloud/pkg/nocloud"
	redisdb "github.com/slntopp/nocloud/pkg/nocloud/redis"
	"github.com/slntopp/nocloud/pkg/nocloud/rest_auth"
	"github.com/slntopp/nocloud/pkg/nocloud/schema"
	"go.uber.org/zap"
	"net/http"
)

type PaymentGatewayType string

var (
	PaymentGatewayFacture PaymentGatewayType = "facture"
)

const gatewaysBase = "/billing/payments"

type PaymentGatewayServer struct {
	log          *zap.Logger
	db           driver.Database
	ctrl         graph.PaymentGatewaysController
	invoicesCtrl graph.InvoicesController
	caCtrl       graph.CommonActionsController

	rdb redisdb.Client
	sk  []byte
}

func NewPaymentGatewayServer(_log *zap.Logger, db driver.Database, rdb redisdb.Client, sighingKey []byte) *PaymentGatewayServer {
	log := _log.Named("PaymentGatewayServer")
	ctrl := graph.NewPaymentGatewaysController(log, db)
	invoicesCtrl := graph.NewInvoicesController(log, db)
	caCtrl := graph.NewCommonActionsController(log, db)
	return &PaymentGatewayServer{
		log:          log,
		db:           db,
		ctrl:         ctrl,
		invoicesCtrl: invoicesCtrl,
		caCtrl:       caCtrl,
		rdb:          rdb,
		sk:           sighingKey,
	}
}

func (s *PaymentGatewayServer) RegisterRoutes(router *mux.Router) {
	interceptor := rest_auth.NewInterceptor(s.log, s.rdb, s.sk)
	subRouter := router.PathPrefix(gatewaysBase).Subrouter()
	subRouter.Handle("/{key}/{invoice_uuid}/action", interceptor.JwtMiddleWare(http.HandlerFunc(s.HandlePaymentAction))).Methods("POST")
	subRouter.Handle("/{invoice_uuid}/view", interceptor.JwtMiddleWare(http.HandlerFunc(s.HandlePaymentAction))).Methods("GET")
}

func (s *PaymentGatewayServer) HandleViewInvoice(writer http.ResponseWriter, request *http.Request) {
	invoiceUuid := mux.Vars(request)["invoice_uuid"]
	if invoiceUuid == "" {
		http.Error(writer, "Invoice UUID is required", http.StatusBadRequest)
		return
	}
	invoice, err := s.invoicesCtrl.Get(request.Context(), invoiceUuid)
	if err != nil {
		http.Error(writer, "Failed to get invoice: "+err.Error(), http.StatusInternalServerError)
		return
	}
	requester, _ := request.Context().Value(nocloud.NoCloudAccount).(string)
	if !s.caCtrl.HasAccess(request.Context(), requester, driver.NewDocumentID(schema.NAMESPACES_COL, schema.ROOT_NAMESPACE_KEY), accesspb.Level_ADMIN) &&
		requester != invoice.Account {
		http.Error(writer, "Access denied to view this invoice", http.StatusForbidden)
		return
	}

	// Generate and send invoice html

	writer.WriteHeader(http.StatusOK)
}

func (s *PaymentGatewayServer) HandlePaymentAction(writer http.ResponseWriter, request *http.Request) {
	pgKey := mux.Vars(request)["key"]
	if pgKey == "" {
		http.Error(writer, "Payment Gateway Key is required", http.StatusBadRequest)
		return
	}
	invoiceUuid := mux.Vars(request)["invoice_uuid"]
	if invoiceUuid == "" {
		http.Error(writer, "Invoice UUID is required", http.StatusBadRequest)
		return
	}

	switch PaymentGatewayType(pgKey) {
	case PaymentGatewayFacture:
		s.log.Debug("Handling Facture Payment Action", zap.String("invoice_uuid", invoiceUuid))
		_, err := s.invoicesCtrl.Get(request.Context(), invoiceUuid)
		if err != nil {
			http.Error(writer, "Failed to get invoice: "+err.Error(), http.StatusInternalServerError)
			return
		}
		// Generate PDF facture and send back to client to download it

	default:
		http.Error(writer, "Unsupported Payment Gateway Key", http.StatusBadRequest)
		return
	}

	writer.WriteHeader(http.StatusOK)
}
