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
	overdueCCHost          string
	overdueSigningKey      []byte
	overdueDepartmentKey   string
	overdueWhmcsSenderUUID string
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
	"ip_pool_exhausted":           OverdueTicketHandler,
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

type ccDefaultsDoc struct {
	Departments []struct {
		Key          string   `json:"key"`
		Admins       []string `json:"admins"`
		WhmcsID      string   `json:"whmcsId"`
		WhmcsIDSnake string   `json:"whmcs_id"`
	} `json:"departments"`
	Bot struct {
		Values map[string]string `json:"values"`
	} `json:"bot"`
}

func fetchCCDefaults(ctx context.Context, token string) (*ccDefaultsDoc, error) {
	status, body, err := overdueCCPost(ctx, "/cc.UsersAPI/FetchDefaults", map[string]any{
		"fetchTemplates": false,
	}, token)
	if err != nil {
		return nil, err
	}
	if status >= 300 {
		return nil, fmt.Errorf("fetch defaults: status %d: %s", status, string(body))
	}
	var defaults ccDefaultsDoc
	if err := json.Unmarshal(body, &defaults); err != nil {
		return nil, fmt.Errorf("parse defaults: %w", err)
	}
	return &defaults, nil
}

func (d *ccDefaultsDoc) department(departmentKey string) (*overdueCCDepartmentInfo, error) {
	if d == nil {
		return nil, fmt.Errorf("department %q not found in CC config", departmentKey)
	}
	for _, dep := range d.Departments {
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

func (d *ccDefaultsDoc) values() map[string]string {
	if d == nil {
		return nil
	}
	return d.Bot.Values
}

func overdueCCDepartmentInfoFetch(ctx context.Context, token, departmentKey string) (*overdueCCDepartmentInfo, error) {
	defaults, err := fetchCCDefaults(ctx, token)
	if err != nil {
		return nil, err
	}
	return defaults.department(departmentKey)
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

func supportTicketCopy(event *pb.Event, info EventInfo) (topic, content string) {
	if event.GetKey() == "ip_pool_exhausted" {
		return formatIPPoolTicketTopic(info), formatIPPoolTicketMessage(info, event)
	}
	return formatOverdueTicketTopic(info), formatOverdueTicketMessage(info)
}

func formatIPPoolTicketTopic(info EventInfo) string {
	name := stripOverdueBillingDecor(info.Instance)
	if name == "" {
		name = stripOverdueBillingDecor(info.Product)
	}
	if name == "" {
		name = info.AccountTitle
	}
	return fmt.Sprintf("Нет свободных IP: %s", name)
}

func formatIPPoolTicketMessage(info EventInfo, event *pb.Event) string {
	clientName := info.AccountTitle
	if clientName == "" {
		clientName = info.Account
	}
	errMsg := ""
	if event != nil {
		if data := event.GetData(); data != nil {
			if v := data["error"]; v != nil {
				errMsg = v.GetStringValue()
			}
		}
	}
	if errMsg == "" {
		errMsg = "ip pool exhausted"
	}
	return fmt.Sprintf(`Здравствуйте.

Услуга "%s" (клиент %s) не создана: в пуле закончились свободные IP-адреса.
Ошибка: %s

После пополнения пула создание возобновится автоматически.

С уважением, служба поддержки.`,
		formatOverdueServiceDetails(info), clientName, errMsg)
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

	defaults, err := fetchCCDefaults(ctx, token)
	if err != nil {
		log.Warn("overdue ticket: CC defaults not loaded", zap.Error(err))
	}
	var botValues map[string]string
	if defaults != nil {
		botValues = defaults.values()
	}
	profile := resolveTicketProfile(event.GetKey(), botValues)
	if !profile.Enabled {
		log.Info("auto ticket disabled", zap.String("key", event.GetKey()))
		event.Type = "noop"
		return event, nil
	}

	topic, content := ticketText(event, info, profile)
	createPayload := map[string]any{
		"owner":  info.Account,
		"users":  []string{info.Account},
		"topic":  topic,
		"status": 0,
	}
	if profile.Responsible != "" {
		createPayload["responsible"] = profile.Responsible
	}

	var deptWhmcsID string
	var admins []string
	if profile.Department != "" {
		createPayload["department"] = profile.Department
		if defaults != nil {
			deptInfo, err := defaults.department(profile.Department)
			if err != nil {
				log.Warn("overdue ticket: department config not loaded", zap.Error(err))
			} else {
				admins = deptInfo.Admins
				deptWhmcsID = deptInfo.WhmcsID
			}
		}
	}
	for _, uuid := range profile.Admins {
		admins = overdueAppendUniqueAdminUUID(admins, uuid)
	}
	admins = overdueAppendUniqueAdminUUID(admins, profile.Responsible)
	admins = overdueAppendUniqueAdminUUID(admins, profile.SenderUUID)
	if len(admins) > 0 {
		createPayload["admins"] = admins
	} else if profile.Department != "" {
		log.Warn("overdue ticket: department has no admins", zap.String("department", profile.Department))
	}

	if deptWhmcsID != "" {
		createPayload["meta"] = map[string]any{
			"data": map[string]any{
				"dept_id": deptWhmcsID,
			},
		}
	} else if profile.Department != "" {
		log.Warn("overdue ticket: CC department has no whmcsId; WHMCS OpenTicket may fail",
			zap.String("department", profile.Department))
	}

	if profile.Department != "" && deptWhmcsID != "" && profile.SenderUUID == "" {
		log.Warn("overdue ticket: set sender (staff NoCloud UUID with whmcs_admin_id) so the first message opens WHMCS as admin")
	}

	// ChatsAPI/Create sets chat owner from JWT (not from payload). Root → owner "0" (nocloud);
	// WHMCS OpenTicket treats the first message as admin opener only if sender ∈ chat.admins.
	// Use the same staff JWT for Create+Send when configured so CC owner matches the opener.
	createToken := token
	sendToken := token
	if profile.SenderUUID != "" {
		staffTok, err := overdueCCJWT(profile.SenderUUID)
		if err != nil {
			return nil, fmt.Errorf("overdue ticket: sign staff token: %w", err)
		}
		createToken = staffTok
		sendToken = staffTok
	}

	createStatus, createBody, err := overdueCCPost(ctx, "/cc.ChatsAPI/Create", createPayload, createToken)
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

	sendStatus, sendBody, err := overdueCCPost(ctx, "/cc.MessagesAPI/Send", map[string]any{
		"chat":    chatUUID,
		"content": content,
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
