package sc

import (
	"context"
	"encoding/json"
	"errors"

	pb "github.com/slntopp/nocloud-proto/settings"
	"go.uber.org/zap"
)

var log *zap.Logger
var ctx context.Context
var client *pb.SettingsServiceClient

func Setup(logger *zap.Logger, _ctx context.Context, c *pb.SettingsServiceClient) {
	log = logger
	ctx = _ctx
	client = c
}

type Setting[T any] struct {
	Value       T
	Description string
	Public      bool
}

func Fetch[T any](key string, _conf *T, _default *Setting[T]) error {
	if client == nil {
		return errors.New("error: Settings Client unset")
	}
	c := *client

	r, err := c.Get(ctx, &pb.GetRequest{Keys: []string{key}})
	if err != nil {
		log.Warn("Failed to Get setting", zap.Error(err))
		goto set_default
	}

	if _, ok := r.GetFields()[key]; !ok {
		log.Warn("Failed to Get setting in response", zap.Any("result", r), zap.Error(err))
		goto set_default
	}

	err = json.Unmarshal([]byte(r.GetFields()[key].GetStringValue()), _conf)
	if err != nil {
		log.Warn("Failed to Unmarshal setting", zap.Any("result", r), zap.Error(err))
		goto set_default
	}
	return nil

set_default:
	log.Info("Setting default conf")
	if _default == nil {
		log.Error("No default conf")
		return errors.New("error: default setting is nil")
	}
	payload, err := json.Marshal(_default.Value)
	if err == nil {
		_, err := c.Put(ctx, &pb.PutRequest{
			Key:         key,
			Value:       string(payload),
			Description: &_default.Description,
			Public:      &_default.Public,
		})
		if err != nil {
			log.Error("Error Putting Monitoring Configuration", zap.Error(err))
		}
	}

	*_conf = (*_default).Value
	return nil
}
