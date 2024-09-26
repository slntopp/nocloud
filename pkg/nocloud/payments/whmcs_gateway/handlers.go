package whmcs_gateway

import (
	"encoding/json"
	"go.uber.org/zap"
	"io"
	"net/http"
)

func (g *WhmcsGateway) invoicePaidHandler(log *zap.Logger, data InvoicePaid) {

}

func unmarshal[T any](body io.ReadCloser) (T, error) {
	var res T
	if err := json.NewDecoder(body).Decode(&res); err != nil {
		return res, err
	}
	return res, nil
}

func (g *WhmcsGateway) handleWhmcsEvent(log *zap.Logger, r *http.Request) {
	log = log.Named("WhmcsHandler")

	resp := struct {
		Event string `json:"event"`
	}{}
	if err := json.NewDecoder(r.Body).Decode(&resp); err != nil {
		log.Error("Error decoding request", zap.Error(err))
		return
	}

	log.Info("Event received", zap.String("event", resp.Event))

	switch resp.Event {
	case "InvoicePaid":
		data, err := unmarshal[InvoicePaid](r.Body)
		if err != nil {
			log.Error("Error decoding request", zap.Error(err))
			return
		}
		g.invoicePaidHandler(log, data)
	default:
		log.Error("Unknown event", zap.String("event", resp.Event))
		return
	}
}
