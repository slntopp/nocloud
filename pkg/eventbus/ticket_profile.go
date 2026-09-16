package eventbus

import (
	"strings"

	pb "github.com/slntopp/nocloud-proto/events"
)

const (
	autoTicketPrefix   = "auto_ticket"
	ipPoolTicketPrefix = "ip_pool_ticket"
)

type ccTicketProfile struct {
	Enabled     bool
	Department  string
	Admins      []string
	Responsible string
	SenderUUID  string
	Topic       string
	Message     string
}

func parseCCTicketProfile(values map[string]string, prefix string) ccTicketProfile {
	if values == nil {
		values = map[string]string{}
	}
	rawEnabled := strings.ToLower(strings.TrimSpace(values[prefix+".enabled"]))
	enabled := rawEnabled == "" || (rawEnabled != "false" && rawEnabled != "0" && rawEnabled != "no")
	return ccTicketProfile{
		Enabled:     enabled,
		Department:  strings.TrimSpace(values[prefix+".department"]),
		Admins:      splitCSV(values[prefix+".admins"]),
		Responsible: strings.TrimSpace(values[prefix+".responsible"]),
		SenderUUID:  strings.TrimSpace(values[prefix+".sender_uuid"]),
		Topic:       values[prefix+".topic"],
		Message:     values[prefix+".message"],
	}
}

func resolveTicketProfile(eventKey string, values map[string]string) ccTicketProfile {
	overdue := parseCCTicketProfile(values, autoTicketPrefix)
	if overdue.Department == "" {
		overdue.Department = overdueDepartmentKey
	}
	if overdue.SenderUUID == "" {
		overdue.SenderUUID = overdueWhmcsSenderUUID
	}
	if eventKey != "ip_pool_exhausted" {
		return overdue
	}

	ip := parseCCTicketProfile(values, ipPoolTicketPrefix)
	if ip.Department == "" {
		ip.Department = firstNonEmpty(overdue.Department, overdueDepartmentKey)
	}
	if ip.SenderUUID == "" {
		ip.SenderUUID = firstNonEmpty(overdue.SenderUUID, overdueWhmcsSenderUUID)
	}
	if ip.Responsible == "" {
		ip.Responsible = overdue.Responsible
	}
	if len(ip.Admins) == 0 {
		ip.Admins = overdue.Admins
	}
	return ip
}

func ticketText(event *pb.Event, info EventInfo, profile ccTicketProfile) (topic, content string) {
	fallbackTopic, fallbackMessage := supportTicketCopy(event, info)
	topic = strings.TrimSpace(profile.Topic)
	if topic == "" {
		topic = fallbackTopic
	} else {
		topic = applyTicketPlaceholders(topic, info, event)
	}
	content = strings.TrimSpace(profile.Message)
	if content == "" {
		content = fallbackMessage
	} else {
		content = applyTicketPlaceholders(content, info, event)
	}
	return topic, content
}

func applyTicketPlaceholders(s string, info EventInfo, event *pb.Event) string {
	clientName := info.AccountTitle
	if clientName == "" {
		clientName = info.Account
	}
	errMsg := eventError(event)
	if errMsg == "" {
		errMsg = "ip pool exhausted"
	}
	replacer := strings.NewReplacer(
		"{CLIENT_NAME}", clientName,
		"{INSTANCE}", stripOverdueBillingDecor(info.Instance),
		"{PRODUCT}", stripOverdueBillingDecor(info.Product),
		"{IPS}", formatOverdueIPs(info.Ips),
		"{SERVICE_DETAILS}", formatOverdueServiceDetails(info),
		"{SERVICE}", info.Service,
		"{ERROR}", errMsg,
		"{ACCOUNT}", info.Account,
	)
	return replacer.Replace(s)
}

func eventError(event *pb.Event) string {
	if event == nil {
		return ""
	}
	data := event.GetData()
	if data == nil {
		return ""
	}
	if v := data["error"]; v != nil {
		return v.GetStringValue()
	}
	return ""
}

func splitCSV(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	var out []string
	for _, part := range strings.Split(raw, ",") {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}
