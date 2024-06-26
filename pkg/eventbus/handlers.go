package eventbus

import (
	"context"

	"github.com/arangodb/go-driver"
	pb "github.com/slntopp/nocloud-proto/events"
	"github.com/slntopp/nocloud/pkg/nocloud/schema"
	"google.golang.org/protobuf/types/known/structpb"
)

type EventHandler func(context.Context, *pb.Event, driver.Database) (*pb.Event, error)

var handlers = map[string]EventHandler{
	"instance_suspended":          GetInstAccountHandler,
	"instance_unsuspended":        GetInstAccountHandler,
	"instance_created":            GetInstAccountHandler,
	"instance_deleted":            GetInstAccountHandler,
	"expiry_notification":         GetInstAccountHandler,
	"suspend_expiry_notification": GetInstAccountHandler,
	"suspend_delete_instance":     GetInstAccountHandler,
	"instance_renew":              GetInstAccountHandler,
	"pending_notification":        GetInstAccountHandler,
	"instance_credentials":        GetInstAccountHandler,
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

LET rate_one = LAST(
	FOR i IN @@c2c
    FILTER (i.to == 0 || i.from == 0) && i.rate == 1
        RETURN i
)

LET default_cur = rate_one.to == 0 ? rate_one.from : rate_one.to

LET currency = account.currency != null ? account.currency : default_cur
LET rate = PRODUCT(
	FOR vertex, edge IN OUTBOUND
	SHORTEST_PATH DOCUMENT(CONCAT(@currencies, "/", default_cur))
	TO DOCUMENT(CONCAT(@currencies, "/", currency))
	GRAPH @graph
	FILTER edge
		RETURN edge.rate
)

LET price = doc.billing_plan.products[doc.product] == null ? 0 : doc.billing_plan.products[doc.product].price

LET total = @inner_price == 0 ? price : @inner_price

RETURN {
	account: account._key, 
	service: srv.title, 
	instance: doc.title, 
	product: doc.product, 
	next_payment_date: doc.data.next_payment_date,
	ips: doc.state.meta.networking.public,
	price: total * rate
}
`

type EventInfo struct {
	Account         string  `json:"account"`
	Service         string  `json:"service"`
	Instance        string  `json:"instance"`
	Product         string  `json:"product,omitempty"`
	Ips             []any   `json:"ips,omitempty"`
	NextPaymentDate float64 `json:"next_payment_date,omitempty"`
	Price           float64 `json:"price,omitempty"`
}

func GetInstAccountHandler(ctx context.Context, event *pb.Event, db driver.Database) (*pb.Event, error) {
	if event.GetData() == nil {
		event.Data = make(map[string]*structpb.Value)
	}

	var innerPrice float64
	price, ok := event.GetData()["price"]
	if ok {
		innerPrice = price.GetNumberValue()
	}

	inst := driver.NewDocumentID(schema.INSTANCES_COL, event.GetUuid())

	cursor, err := db.Query(ctx, getInstanceAccount, map[string]interface{}{
		"inst":        inst,
		"permissions": schema.PERMISSIONS_GRAPH.Name,
		"@services":   schema.SERVICES_COL,
		"@accounts":   schema.ACCOUNTS_COL,
		"currencies":  schema.CUR_COL,
		"graph":       schema.BILLING_GRAPH.Name,
		"@c2c":        schema.CUR2CUR,
		"inner_price": innerPrice,
	})
	if err != nil {
		return nil, err
	}

	defer cursor.Close()

	var eventInfo EventInfo
	for cursor.HasMore() {
		_, err := cursor.ReadDocument(ctx, &eventInfo)
		if err != nil {
			return nil, err
		}
	}

	event.Data["service"] = structpb.NewStringValue(eventInfo.Service)
	event.Data["instance"] = structpb.NewStringValue(eventInfo.Instance)
	if eventInfo.Product != "" {
		event.Data["product"] = structpb.NewStringValue(eventInfo.Product)
	}
	if eventInfo.Ips != nil {
		listValue, _ := structpb.NewList(eventInfo.Ips)
		event.Data["ips"] = structpb.NewListValue(listValue)
	}
	if eventInfo.NextPaymentDate != 0 {
		event.Data["next_payment_date"] = structpb.NewNumberValue(eventInfo.NextPaymentDate)
	}
	event.Data["instance_uuid"] = structpb.NewStringValue(event.GetUuid())
	event.Data["price"] = structpb.NewNumberValue(eventInfo.Price)
	event.Uuid = eventInfo.Account
	event.Type = "email"

	return event, nil
}
