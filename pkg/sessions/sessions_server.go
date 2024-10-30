package sessions

import (
	"context"
	"github.com/arangodb/go-driver"
	"github.com/slntopp/nocloud-proto/access"
	"github.com/slntopp/nocloud/pkg/graph"
	redisdb "github.com/slntopp/nocloud/pkg/nocloud/redis"
	"github.com/slntopp/nocloud/pkg/nocloud/schema"
	"github.com/slntopp/nocloud/pkg/nocloud/sessions"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	sspb "github.com/slntopp/nocloud-proto/sessions"
	"github.com/slntopp/nocloud/pkg/nocloud"
	"go.uber.org/zap"
)

type SessionsServer struct {
	sspb.UnimplementedSessionsServiceServer

	log *zap.Logger
	rdb redisdb.Client
	db  driver.Database
}

func NewSessionsServer(log *zap.Logger, rdb redisdb.Client, db driver.Database) *SessionsServer {
	return &SessionsServer{
		log: log.Named("Sessions"),
		rdb: rdb,
		db:  db,
	}
}

func (c *SessionsServer) Get(ctx context.Context, req *sspb.GetSessions) (*sspb.Sessions, error) {
	log := c.log.Named("Get")
	requestor := ctx.Value(nocloud.NoCloudAccount).(string)
	sid := ctx.Value(nocloud.NoCloudSession).(string)

	var user_id = "*"

	if req.UserId == nil {
		if ok := graph.HasAccess(ctx, c.db, requestor, driver.NewDocumentID(schema.NAMESPACES_COL, schema.ROOT_NAMESPACE_KEY), access.Level_ROOT); !ok {
			user_id = requestor
		}
	} else {
		uuid := req.GetUserId()
		if ok := graph.HasAccess(ctx, c.db, requestor, driver.NewDocumentID(schema.ACCOUNTS_COL, uuid), access.Level_ROOT); !ok {
			return nil, status.Error(codes.PermissionDenied, "No access to account")
		}
		user_id = uuid
	}

	log.Debug("Invoked", zap.String("requestor", requestor), zap.String("sid", sid))

	result, err := sessions.Get(c.rdb, user_id)
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

func (c *SessionsServer) GetActivity(ctx context.Context, req *sspb.GetActivityRequest) (*sspb.Activity, error) {
	log := c.log.Named("GetActivity")
	requestor := ctx.Value(nocloud.NoCloudAccount).(string)

	uuid := req.GetUuid()
	if ok := graph.HasAccess(ctx, c.db, requestor, driver.NewDocumentID(schema.ACCOUNTS_COL, uuid), access.Level_ROOT); !ok {
		return nil, status.Error(codes.PermissionDenied, "No access to account")
	}

	log.Debug("Invoked", zap.String("requestor", uuid))

	result, err := sessions.GetActivity(c.rdb, uuid)
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

	err := sessions.Revoke(c.rdb, requestor, req.Id)
	if err != nil {
		return nil, err
	}

	return &sspb.DeleteResponse{}, nil
}
