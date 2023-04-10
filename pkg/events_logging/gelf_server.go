package events_logging

import (
	"context"
	"encoding/json"
	"github.com/Graylog2/go-gelf/gelf"
	"github.com/slntopp/nocloud/pkg/nocloud"
	"go.uber.org/zap"
)

type GelfServer struct {
	*gelf.Reader
	rep *SqliteRepository

	log *zap.Logger
}

type ShortLogMessage struct {
	Level string `json:"level"`
	Msg   string `json:"msg"`

	Entity    string `json:"entity"`
	Uuid      string `json:"uuid"`
	Scope     string `json:"scope"`
	Action    string `json:"action"`
	Rc        int32  `json:"rc"`
	Requestor string `json:"requestor"`
	Ts        int64  `json:"ts"`

	Diff string `json:"diff,omitempty"`
}

func NewGelfServer(_log *zap.Logger, host string, rep *SqliteRepository) *GelfServer {
	log := _log.Named("GelfServer")

	log.Debug("Creating Gelf Server")

	reader, err := gelf.NewReader(host)
	if err != nil {
		log.Fatal("Failed to create GelfServer", zap.Error(err))
		return nil
	}
	return &GelfServer{Reader: reader, rep: rep, log: log}
}

func (s *GelfServer) Run() {
	log := s.log.Named("Run")

	log.Info("Start accepting messages")
	nocloudLevelVal := nocloud.NOCLOUD_LOG_LEVEL.String()

	for {
		message, err := s.ReadMessage()
		if err != nil {
			log.Error("Failed to read message", zap.Error(err))
			continue
		}
		var shortMessage ShortLogMessage

		err = json.Unmarshal([]byte(message.Short), &shortMessage)
		if err != nil {
			log.Error("Failed to parse short message", zap.Error(err))
			continue
		}

		if shortMessage.Level != nocloudLevelVal {
			continue
		}

		err = s.rep.CreateEvent(context.Background(), &shortMessage)
		if err != nil {
			log.Error("Failed to create event", zap.Error(err))
		}
	}
}
