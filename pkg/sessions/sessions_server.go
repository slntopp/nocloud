package sessions

import (
	"context"

	"github.com/go-redis/redis/v8"
	sspb "github.com/slntopp/nocloud-proto/sessions"
	"github.com/slntopp/nocloud/pkg/nocloud"
	"go.uber.org/zap"
)

type SessionsServer struct {
	sspb.UnimplementedSessionsServiceServer

	log *zap.Logger
	rdb *redis.Client
}

func NewSessionsServer(log *zap.Logger, rdb *redis.Client) *SessionsServer {
	return &SessionsServer{
		log: log.Named("Sessions"),
		rdb: rdb,
	}
}

func (c *SessionsServer) Get(ctx context.Context, req *sspb.EmptyMessage) (*sspb.Sessions, error) {
	log := c.log.Named("Get")
	requestor := ctx.Value(nocloud.NoCloudAccount).(string)
	sid := ctx.Value(nocloud.NoCloudSession).(string)

	log.Debug("Invoked", zap.String("requestor", requestor), zap.String("sid", sid))

	result, err := Get(c.rdb, requestor)
	if err != nil {
		return nil, err
	}

	current := true
	for _, session := range result {
		if session.Id == sid {
			session.Current = &current
			break
		}
	}

	return &sspb.Sessions{
		Sessions: result,
	}, nil
}

func (c *SessionsServer) GetActivity(ctx context.Context, req *sspb.EmptyMessage) (*sspb.Activity, error) {
	log := c.log.Named("GetActivity")
	requestor := ctx.Value(nocloud.NoCloudAccount).(string)

	log.Debug("Invoked", zap.String("requestor", requestor))

	result, err := GetActivity(c.rdb, requestor)
	if err != nil {
		return nil, err
	}

	return &sspb.Activity{
		LastSeen: result,
	}, nil
}

func (c *SessionsServer) Revoke(ctx context.Context, req *sspb.Session) (*sspb.DeleteResponse, error) {
	log := c.log.Named("Revoke")
	requestor := ctx.Value(nocloud.NoCloudAccount).(string)

	log.Debug("Invoked", zap.String("requestor", requestor), zap.String("sid", req.Id))

	err := Revoke(c.rdb, requestor, req.Id)
	if err != nil {
		return nil, err
	}

	return &sspb.DeleteResponse{}, nil
}
