package eventbus

import (
	"context"
	"errors"
	"github.com/arangodb/go-driver"
	pb "github.com/slntopp/nocloud-proto/events"
	"github.com/slntopp/nocloud/pkg/nocloud/schema"
	"google.golang.org/protobuf/types/known/structpb"
)

type EventHandler func(context.Context, *pb.Event, driver.Database) (*pb.Event, error)

var handlers = map[string]EventHandler{
	"instance_suspended":   GetInstAccountHandler,
	"instance_unsuspended": GetInstAccountHandler,
	"instance_created":     GetInstAccountHandler,
	"instance_deleted":     GetInstAccountHandler,
	"expiry_notification":  ExpiryHandler,
}

var getInstanceAccount = `
LET doc = DOCUMENT(@inst)


LET srv = LAST(
FOR node, edge, path IN 2
    INBOUND doc
    GRAPH @permissions
    FILTER path.edges[*].role == ["owner","owner"]
    FILTER IS_SAME_COLLECTION(node, @@services)
        RETURN node
    )

LET account = LAST(
    FOR node, edge, path IN 2
    INBOUND srv
    GRAPH @permissions
    FILTER path.edges[*].role == ["owner","owner"]
    FILTER IS_SAME_COLLECTION(node, @@accounts)
        RETURN node
    )
    
RETURN {account: account._key, service: srv._key}
`

type AccountWithService struct {
	Account string `json:"account"`
	Service string `json:"service"`
}

func GetInstAccountHandler(ctx context.Context, event *pb.Event, db driver.Database) (*pb.Event, error) {
	inst := driver.NewDocumentID(schema.INSTANCES_COL, event.GetUuid())

	cursor, err := db.Query(ctx, getInstanceAccount, map[string]interface{}{
		"inst":        inst,
		"permissions": schema.PERMISSIONS_GRAPH.Name,
		"@services":   schema.SERVICES_COL,
		"@accounts":   schema.ACCOUNTS_COL,
	})
	if err != nil {
		return nil, err
	}

	defer cursor.Close()

	var accountWithService AccountWithService
	for cursor.HasMore() {
		_, err := cursor.ReadDocument(ctx, &accountWithService)
		if err != nil {
			return nil, err
		}
	}

	event.Uuid = accountWithService.Account
	event.Type = "email"
	return event, nil
}

func ExpiryHandler(ctx context.Context, event *pb.Event, db driver.Database) (*pb.Event, error) {
	if event.GetData() == nil {
		return nil, errors.New("event don't have data")
	}

	inst := driver.NewDocumentID(schema.INSTANCES_COL, event.GetUuid())

	cursor, err := db.Query(ctx, getInstanceAccount, map[string]interface{}{
		"inst":        inst,
		"permissions": schema.PERMISSIONS_GRAPH.Name,
		"@services":   schema.SERVICES_COL,
		"@accounts":   schema.ACCOUNTS_COL,
	})
	if err != nil {
		return nil, err
	}

	defer cursor.Close()

	var accountWithService AccountWithService
	for cursor.HasMore() {
		_, err := cursor.ReadDocument(ctx, &accountWithService)
		if err != nil {
			return nil, err
		}
	}

	event.Data["instance"] = structpb.NewStringValue(event.GetUuid())
	event.Data["service"] = structpb.NewStringValue(accountWithService.Service)
	event.Uuid = accountWithService.Account
	event.Type = "email"

	return event, nil
}
