package whmcs_gateway

import (
	"encoding/json"
	"go.uber.org/zap"
)

func (g *WhmcsGateway) invoicePaidHandler(log *zap.Logger, data InvoicePaid) {

}

func unmarshal[T any](b []byte) (T, error) {
	var res T
	if err := json.Unmarshal(b, &res); err != nil {
		return res, err
	}
	return res, nil
}

func (g *WhmcsGateway) handleWhmcsEvent(log *zap.Logger, body []byte) {
	log = log.Named("WhmcsHandler")

	resp := struct {
		Event string `json:"event"`
	}{}
	if err := json.Unmarshal(body, &resp); err != nil {
		log.Error("Error decoding request", zap.Error(err))
		return
	}

	log.Info("Event received", zap.String("event", resp.Event))

	switch resp.Event {
	case "InvoicePaid":
		data, err := unmarshal[InvoicePaid](body)
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
