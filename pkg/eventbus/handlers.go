package eventbus

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strings"
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
	overdueCCHost                string
	overdueSigningKey            []byte
	overdueDepartmentKey         string
	overdueWhmcsSenderUUID       string
)

func SetupOverdueTicketHandler(ccHost string, signingKey []byte, departmentKey, whmcsSenderUUID string) {
	overdueCCHost = ccHost
	overdueSigningKey = signingKey
	overdueDepartmentKey = departmentKey
	overdueWhmcsSenderUUID = strings.TrimSpace(whmcsSenderUUID)
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
	account_title: account.title,
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
	AccountTitle    string  `json:"account_title"`
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

func overdueServiceToken() (string, error) {
	return overdueCCJWT(schema.ROOT_ACCOUNT_KEY)
}

func overdueCCJWT(account string) (string, error) {
	return jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		nocloud.NOCLOUD_ACCOUNT_CLAIM:   account,
		nocloud.NOCLOUD_INSTANCE_CLAIM:  "placeholder",
		nocloud.NOCLOUD_ROOT_CLAIM:      4,
		nocloud.NOCLOUD_NOSESSION_CLAIM: true,
	}).SignedString(overdueSigningKey)
}

func overdueCCPost(ctx context.Context, path string, payload any, token string) (int, []byte, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return 0, nil, fmt.Errorf("marshal body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, overdueCCHost+path, bytes.NewReader(body))
	if err != nil {
		return 0, nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, nil, err
	}
	return resp.StatusCode, respBody, nil
}

type overdueCCDepartmentInfo struct {
	Admins  []string
	WhmcsID string
}

func overdueCCDepartmentInfoFetch(ctx context.Context, token, departmentKey string) (*overdueCCDepartmentInfo, error) {
	status, body, err := overdueCCPost(ctx, "/cc.UsersAPI/FetchDefaults", map[string]any{
		"fetchTemplates": false,
	}, token)
	if err != nil {
		return nil, err
	}
	if status >= 300 {
		return nil, fmt.Errorf("fetch defaults: status %d: %s", status, string(body))
	}

	var defaults struct {
		Departments []struct {
			Key          string   `json:"key"`
			Admins       []string `json:"admins"`
			WhmcsID      string   `json:"whmcsId"`
			WhmcsIDSnake string   `json:"whmcs_id"`
		} `json:"departments"`
	}
	if err := json.Unmarshal(body, &defaults); err != nil {
		return nil, fmt.Errorf("parse defaults: %w", err)
	}

	for _, dep := range defaults.Departments {
		if dep.Key != departmentKey {
			continue
		}
		wid := strings.TrimSpace(dep.WhmcsID)
		if wid == "" {
			wid = strings.TrimSpace(dep.WhmcsIDSnake)
		}
		return &overdueCCDepartmentInfo{Admins: dep.Admins, WhmcsID: wid}, nil
	}
	return nil, fmt.Errorf("department %q not found in CC config", departmentKey)
}

func overdueAppendUniqueAdminUUID(admins []string, uuid string) []string {
	uuid = strings.TrimSpace(uuid)
	if uuid == "" {
		return admins
	}
	if slices.Contains(admins, uuid) {
		return admins
	}
	return append(slices.Clone(admins), uuid)
}

func stripOverdueBillingDecor(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, "&monthly", "")
	s = strings.ReplaceAll(s, "&Monthly", "")
	return strings.TrimSpace(s)
}

func parseChatUUIDFromCreate(respBody []byte) (string, error) {
	var chat struct {
		UUID string `json:"uuid"`
	}
	if err := json.Unmarshal(respBody, &chat); err != nil {
		return "", err
	}
	if chat.UUID == "" {
		return "", fmt.Errorf("empty chat uuid in create response")
	}
	return chat.UUID, nil
}

func formatOverdueTicketTopic(info EventInfo) string {
	name := stripOverdueBillingDecor(info.Instance)
	if name == "" {
		name = stripOverdueBillingDecor(info.Product)
	}
	return fmt.Sprintf("Уведомление об удалении услуги: %s", name)
}

func formatOverdueServiceDetails(info EventInfo) string {
	var parts []string
	if inst := stripOverdueBillingDecor(info.Instance); inst != "" {
		parts = append(parts, fmt.Sprintf("название: %s", inst))
	}
	if prod := stripOverdueBillingDecor(info.Product); prod != "" {
		parts = append(parts, fmt.Sprintf("тариф: %s", prod))
	}
	if ips := formatOverdueIPs(info.Ips); ips != "" {
		parts = append(parts, fmt.Sprintf("IP: %s", ips))
	}
	if len(parts) == 0 {
		return "не указано"
	}
	return strings.Join(parts, ", ")
}

func formatOverdueIPs(ips []any) string {
	var out []string
	for _, ip := range ips {
		switch v := ip.(type) {
		case string:
			if v != "" {
				out = append(out, v)
			}
		}
	}
	return strings.Join(out, ", ")
}

func formatOverdueTicketMessage(info EventInfo) string {
	clientName := info.AccountTitle
	if clientName == "" {
		clientName = info.Account
	}
	return fmt.Sprintf(`Здравствуйте.

Уважаемый %s, сообщаем, что оказание услуги: "%s" приостановлено в связи с истечением срока оплаты.

Информация о выставленных счетах доступна в личном кабинете.
Обращаем внимание, что в случае неоплаты счета, размещенные данные будут удалены без возможности восстановления.
Если Вам нужна помощь, пожалуйста, свяжитесь с нами.

С уважением, служба поддержки.`,
		clientName, formatOverdueServiceDetails(info))
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

	token, err := overdueServiceToken()
	if err != nil {
		return nil, fmt.Errorf("overdue ticket: sign token: %w", err)
	}

	createPayload := map[string]any{
		"owner":  info.Account,
		"users":  []string{info.Account},
		"topic":  formatOverdueTicketTopic(info),
		"status": 0,
	}
	var deptWhmcsID string
	if overdueDepartmentKey != "" {
		createPayload["department"] = overdueDepartmentKey
		deptInfo, err := overdueCCDepartmentInfoFetch(ctx, token, overdueDepartmentKey)
		if err != nil {
			log.Warn("overdue ticket: department config not loaded", zap.Error(err))
		} else {
			admins := deptInfo.Admins
			deptWhmcsID = deptInfo.WhmcsID
			if overdueWhmcsSenderUUID != "" {
				admins = overdueAppendUniqueAdminUUID(admins, overdueWhmcsSenderUUID)
			}
			if len(admins) == 0 {
				log.Warn("overdue ticket: department has no admins", zap.String("department", overdueDepartmentKey))
			} else {
				createPayload["admins"] = admins
				log.Debug("overdue ticket: assigned department admins",
					zap.String("department", overdueDepartmentKey),
					zap.Int("admins", len(admins)))
			}
			if deptWhmcsID != "" {
				createPayload["meta"] = map[string]any{
					"data": map[string]any{
						"dept_id": deptWhmcsID,
					},
				}
			} else {
				log.Warn("overdue ticket: CC department has no whmcsId; WHMCS OpenTicket may fail",
					zap.String("department", overdueDepartmentKey))
			}
		}
	} else if overdueWhmcsSenderUUID != "" {
		createPayload["admins"] = []string{overdueWhmcsSenderUUID}
		log.Warn("overdue ticket: OVERDUE_TICKET_WHMCS_SENDER_UUID set but OVERDUE_TICKET_DEPARTMENT empty; need department for WHMCS dept_id")
	}

	if overdueDepartmentKey != "" && deptWhmcsID != "" && overdueWhmcsSenderUUID == "" {
		log.Warn("overdue ticket: set OVERDUE_TICKET_WHMCS_SENDER_UUID (staff NoCloud UUID with whmcs_admin_id) so the first message opens WHMCS as admin")
	}

	createStatus, createBody, err := overdueCCPost(ctx, "/cc.ChatsAPI/Create", createPayload, token)
	if err != nil {
		return nil, fmt.Errorf("overdue ticket: create chat: %w", err)
	}
	if createStatus >= 300 {
		log.Error("overdue ticket: create chat failed",
			zap.Int("status", createStatus),
			zap.String("body", string(createBody)))
		event.Type = "noop"
		return event, nil
	}

	chatUUID, err := parseChatUUIDFromCreate(createBody)
	if err != nil {
		log.Error("overdue ticket: parse chat uuid", zap.Error(err), zap.String("body", string(createBody)))
		event.Type = "noop"
		return event, nil
	}

	sendToken := token
	if overdueWhmcsSenderUUID != "" {
		st, err := overdueCCJWT(overdueWhmcsSenderUUID)
		if err != nil {
			return nil, fmt.Errorf("overdue ticket: sign opener token: %w", err)
		}
		sendToken = st
	}

	sendStatus, sendBody, err := overdueCCPost(ctx, "/cc.MessagesAPI/Send", map[string]any{
		"chat":    chatUUID,
		"content": formatOverdueTicketMessage(info),
		"kind":    0,
	}, sendToken)
	if err != nil {
		return nil, fmt.Errorf("overdue ticket: send message: %w", err)
	}
	if sendStatus >= 300 {
		log.Error("overdue ticket: send message failed",
			zap.Int("status", sendStatus),
			zap.String("chat", chatUUID),
			zap.String("body", string(sendBody)))
	} else {
		log.Info("overdue ticket created with message",
			zap.String("account", info.Account),
			zap.String("instance", event.GetUuid()),
			zap.String("chat", chatUUID))
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
