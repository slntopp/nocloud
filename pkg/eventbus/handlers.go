package eventbus

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/arangodb/go-driver"
	"github.com/golang-jwt/jwt/v4"
	pb "github.com/slntopp/nocloud-proto/events"
	elpb "github.com/slntopp/nocloud-proto/events_logging"
	"github.com/slntopp/nocloud/pkg/nocloud"
	"github.com/slntopp/nocloud/pkg/nocloud/schema"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"
)

type EventHandler func(context.Context, *zap.Logger, *pb.Event, driver.Database) (*pb.Event, error)

var (
	overdueCCHost     string
	overdueSigningKey []byte
)

func SetupOverdueTicketHandler(ccHost string, signingKey []byte) {
	overdueCCHost = ccHost
	overdueSigningKey = signingKey
}

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
	"inactive_chat_closed":        nil,
	"logging":                     EventLoggingHandler,
	"invoice_published":           nil,
	"invoice_paid":                nil,
	"overdue_ticket":              OverdueTicketHandler,
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
    FILTER (i.to.id == 0 || i.from.id == 0) && i.rate == 1
        RETURN i
)

LET default_cur = rate_one.to.id == 0 ? rate_one.from : rate_one.to

LET currency = account.currency != null ? account.currency : default_cur
LET rate = PRODUCT(
	FOR vertex, edge IN OUTBOUND
	SHORTEST_PATH DOCUMENT(CONCAT(@currencies, "/", default_cur.id))
	TO DOCUMENT(CONCAT(@currencies, "/", currency.id))
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

func GetInstAccountHandler(ctx context.Context, _ *zap.Logger, event *pb.Event, db driver.Database) (*pb.Event, error) {
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

func OverdueTicketHandler(ctx context.Context, log *zap.Logger, event *pb.Event, db driver.Database) (*pb.Event, error) {
	if overdueCCHost == "" {
		log.Warn("CC_HOST not set, skipping overdue ticket creation")
		event.Type = "noop"
		return event, nil
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
		"inner_price": 0,
	})
	if err != nil {
		return nil, fmt.Errorf("overdue ticket: query account: %w", err)
	}
	defer cursor.Close()

	var info EventInfo
	for cursor.HasMore() {
		if _, err := cursor.ReadDocument(ctx, &info); err != nil {
			return nil, fmt.Errorf("overdue ticket: read account: %w", err)
		}
	}

	if info.Account == "" {
		log.Warn("overdue ticket: account not found", zap.String("instance", event.GetUuid()))
		event.Type = "noop"
		return event, nil
	}

	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		nocloud.NOCLOUD_ACCOUNT_CLAIM:   schema.ROOT_ACCOUNT_KEY,
		nocloud.NOCLOUD_INSTANCE_CLAIM:  "placeholder",
		nocloud.NOCLOUD_ROOT_CLAIM:      4,
		nocloud.NOCLOUD_NOSESSION_CLAIM: true,
	}).SignedString(overdueSigningKey)
	if err != nil {
		return nil, fmt.Errorf("overdue ticket: sign token: %w", err)
	}

	body, _ := json.Marshal(map[string]any{
		"owner":  info.Account,
		"users":  []string{info.Account},
		"topic":  fmt.Sprintf("Overdue payment: %s (%s)", info.Instance, event.GetUuid()),
		"status": 0,
	})

	req, _ := http.NewRequestWithContext(ctx, http.MethodPost,
		overdueCCHost+"/cc.ChatsAPI/Create", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("overdue ticket: http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		log.Error("overdue ticket: unexpected status", zap.Int("status", resp.StatusCode))
	} else {
		log.Info("overdue ticket created", zap.String("account", info.Account), zap.String("instance", event.GetUuid()))
	}

	event.Type = "noop"
	return event, nil
}

func EventLoggingHandler(_ context.Context, log *zap.Logger, event *pb.Event, _ driver.Database) (*pb.Event, error) {
	data := event.GetData()
	scope := data["scope"].GetStringValue()
	action := data["action"].GetStringValue()
	diff := data["diff"].GetStringValue()
	if scope == "" || action == "" {
		log.Warn("Invalid event for logging. Scope or action missing. skip logging", zap.Any("event", event))
		return event, nil
	}

	logEvent := &elpb.Event{
		Scope:     scope,
		Action:    action,
		Rc:        0,
		Requestor: schema.ROOT_ACCOUNT_KEY,
		Ts:        time.Now().Unix(),
		Snapshot: &elpb.Snapshot{
			Diff: diff,
		},
		Priority: event.Priority,
		Entity:   event.Type,
		Uuid:     event.Uuid,
	}

	nocloud.Log(log, logEvent)
	log.Debug("Logged event", zap.Any("event", logEvent))

	event.Type = "log"
	event.Uuid = scope
	return event, nil
}
