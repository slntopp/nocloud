package eventbus

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/arangodb/go-driver"
	"github.com/slntopp/nocloud/pkg/nocloud/schema"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"
)

const (
	autoTicketEnabledKey     = "auto_ticket.enabled"
	autoTicketDepartmentKey  = "auto_ticket.department"
	autoTicketTopicKey       = "auto_ticket.topic"
	autoTicketMessageKey     = "auto_ticket.message"
	autoTicketSenderKey      = "auto_ticket.sender_uuid"
	autoTicketAdminsKey      = "auto_ticket.admins"
	autoTicketResponsibleKey = "auto_ticket.responsible"
	autoTicketDelayHoursKey  = "auto_ticket.delay_hours"
	autoTicketDelaysKey      = "auto_ticket.delays"
	defaultOverdueDelayHours = int64(24)
	overduePeriodMatchWindow = int64(2 * 24 * 3600)
)

type overdueCCDefaults struct {
	Departments []overdueCCDepartment `json:"departments"`
	Bot         *overdueCCBot         `json:"bot"`
}

type overdueCCDepartment struct {
	Key          string   `json:"key"`
	Admins       []string `json:"admins"`
	WhmcsID      string   `json:"whmcsId"`
	WhmcsIDSnake string   `json:"whmcs_id"`
}

type overdueCCBot struct {
	Values map[string]string `json:"values"`
}

type overdueDelayRule struct {
	Period int64 `json:"period"`
	Hours  int64 `json:"hours"`
}

type overdueAutoTicketConf struct {
	Enabled           bool
	Department        string
	Admins            []string
	Responsible       string
	Topic             string
	Message           string
	SenderUUID        string
	DefaultDelayHours int64
	Delays            []overdueDelayRule
}

type pbEventData map[string]*structpb.Value

func overdueCCFetchDefaults(ctx context.Context, token string) (*overdueCCDefaults, error) {
	status, body, err := overdueCCPost(ctx, "/cc.UsersAPI/FetchDefaults", map[string]any{
		"fetchTemplates": false,
	}, token)
	if err != nil {
		return nil, err
	}
	if status >= 300 {
		return nil, fmt.Errorf("fetch defaults: status %d: %s", status, string(body))
	}

	var defaults overdueCCDefaults
	if err := json.Unmarshal(body, &defaults); err != nil {
		return nil, fmt.Errorf("parse defaults: %w", err)
	}
	return &defaults, nil
}

func parseOverdueAutoTicket(values map[string]string) overdueAutoTicketConf {
	conf := overdueAutoTicketConf{Enabled: true, DefaultDelayHours: defaultOverdueDelayHours}
	if len(values) == 0 {
		return conf
	}

	switch strings.ToLower(strings.TrimSpace(values[autoTicketEnabledKey])) {
	case "false", "0", "no":
		conf.Enabled = false
	}

	conf.Department = strings.TrimSpace(values[autoTicketDepartmentKey])
	conf.Topic = values[autoTicketTopicKey]
	conf.Message = values[autoTicketMessageKey]
	conf.SenderUUID = strings.TrimSpace(values[autoTicketSenderKey])
	conf.Responsible = strings.TrimSpace(values[autoTicketResponsibleKey])

	if raw := strings.TrimSpace(values[autoTicketDelayHoursKey]); raw != "" {
		var hours int64
		if _, err := fmt.Sscanf(raw, "%d", &hours); err == nil && hours >= 0 {
			conf.DefaultDelayHours = hours
		}
	}

	if raw := strings.TrimSpace(values[autoTicketDelaysKey]); raw != "" {
		var rules []overdueDelayRule
		if err := json.Unmarshal([]byte(raw), &rules); err == nil {
			conf.Delays = rules
		}
	}

	for _, part := range strings.Split(values[autoTicketAdminsKey], ",") {
		if uuid := strings.TrimSpace(part); uuid != "" {
			conf.Admins = append(conf.Admins, uuid)
		}
	}
	return conf
}

func (c overdueAutoTicketConf) delaySeconds(period int64) int64 {
	hours := c.DefaultDelayHours
	if hours < 0 {
		hours = defaultOverdueDelayHours
	}
	if period > 0 {
		var bestHours int64
		bestDiff := int64(-1)
		for _, rule := range c.Delays {
			if rule.Period <= 0 || rule.Hours < 0 {
				continue
			}
			if rule.Period == period {
				return rule.Hours * 3600
			}
			diff := int64(math.Abs(float64(rule.Period - period)))
			if diff <= overduePeriodMatchWindow && (bestDiff < 0 || diff < bestDiff) {
				bestDiff = diff
				bestHours = rule.Hours
			}
		}
		if bestDiff >= 0 {
			return bestHours * 3600
		}
	}
	return hours * 3600
}

func resolveOverdueDueAndPeriod(eventData pbEventData, info EventInfo) (due, period int64) {
	due = int64(info.Due)
	period = int64(info.Period)
	if eventData != nil {
		if v, ok := eventData["due"]; ok && v.GetNumberValue() != 0 {
			due = int64(v.GetNumberValue())
		}
		if v, ok := eventData["period"]; ok && v.GetNumberValue() != 0 {
			period = int64(v.GetNumberValue())
		}
	}
	return due, period
}

func (d *overdueCCDefaults) departmentInfo(key string) (*overdueCCDepartmentInfo, error) {
	if d == nil || key == "" {
		return nil, fmt.Errorf("department %q not found in CC config", key)
	}
	for _, dep := range d.Departments {
		if dep.Key != key {
			continue
		}
		wid := strings.TrimSpace(dep.WhmcsID)
		if wid == "" {
			wid = strings.TrimSpace(dep.WhmcsIDSnake)
		}
		return &overdueCCDepartmentInfo{Admins: dep.Admins, WhmcsID: wid}, nil
	}
	return nil, fmt.Errorf("department %q not found in CC config", key)
}

func (d *overdueCCDefaults) autoTicketConf() overdueAutoTicketConf {
	if d == nil || d.Bot == nil {
		return overdueAutoTicketConf{Enabled: true, DefaultDelayHours: defaultOverdueDelayHours}
	}
	return parseOverdueAutoTicket(d.Bot.Values)
}

func renderOverdueTemplate(tpl string, info EventInfo) string {
	instance := stripOverdueBillingDecor(info.Instance)
	product := stripOverdueBillingDecor(info.Product)
	if instance == "" {
		instance = product
	}
	clientName := info.AccountTitle
	if clientName == "" {
		clientName = info.Account
	}
	return strings.NewReplacer(
		"{CLIENT_NAME}", clientName,
		"{INSTANCE}", instance,
		"{PRODUCT}", product,
		"{IPS}", formatOverdueIPs(info.Ips),
		"{SERVICE_DETAILS}", formatOverdueServiceDetails(info),
		"{SERVICE}", info.Service,
	).Replace(tpl)
}

func resolveOverdueTicketTexts(conf overdueAutoTicketConf, info EventInfo) (topic, message string) {
	topic = strings.TrimSpace(conf.Topic)
	if topic == "" {
		topic = formatOverdueTicketTopic(info)
	} else {
		topic = renderOverdueTemplate(topic, info)
	}

	message = strings.TrimSpace(conf.Message)
	if message == "" {
		message = formatOverdueTicketMessage(info)
	} else {
		message = renderOverdueTemplate(message, info)
	}
	return topic, message
}

const markOverdueTicketCreatedQuery = `
LET doc = DOCUMENT(@inst)
FILTER doc != null
UPDATE doc WITH {
	data: MERGE(
		doc.data == null ? {} : doc.data,
		{ overdue_ticket_created: true }
	)
} IN @@instances
`

func markOverdueTicketCreated(ctx context.Context, db driver.Database, instanceUUID string) error {
	cursor, err := db.Query(ctx, markOverdueTicketCreatedQuery, map[string]interface{}{
		"inst":       driver.NewDocumentID(schema.INSTANCES_COL, instanceUUID),
		"@instances": schema.INSTANCES_COL,
	})
	if err != nil {
		return err
	}
	defer cursor.Close()
	return nil
}

func skipOverdueTicketUntilDelay(log *zap.Logger, conf overdueAutoTicketConf, eventData pbEventData, info EventInfo) bool {
	due, period := resolveOverdueDueAndPeriod(eventData, info)
	if due <= 0 {
		return false
	}
	delay := conf.delaySeconds(period)
	if time.Now().Unix() >= due+delay {
		return false
	}
	log.Info("overdue ticket: delay not elapsed",
		zap.Int64("due", due),
		zap.Int64("period", period),
		zap.Int64("delay_seconds", delay))
	return true
}
