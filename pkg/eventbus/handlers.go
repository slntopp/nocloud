package eventbus

import (
	"context"
	"github.com/arangodb/go-driver"
	pb "github.com/slntopp/nocloud-proto/events"
	"github.com/slntopp/nocloud/pkg/nocloud/schema"
)

type EventHandler func(context.Context, *pb.Event, driver.Database) (*pb.Event, error)

var handlers = map[string]EventHandler{
	"instance_suspended":   GetInstAccountHandler,
	"instance_unsuspended": GetInstAccountHandler,
	"instance_created":     GetInstAccountHandler,
	"instance_deleted":     GetInstAccountHandler,
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
    
RETURN account._key
`

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

	var accountUuid string
	for cursor.HasMore() {
		_, err := cursor.ReadDocument(ctx, &accountUuid)
		if err != nil {
			return nil, err
		}
	}

	event.Uuid = accountUuid
	event.Type = "email"
	return event, nil
}
