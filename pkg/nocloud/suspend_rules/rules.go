package suspend_rules

import (
	"fmt"
	sppb "github.com/slntopp/nocloud-proto/services_providers"
	"strconv"
	"time"
)

func SuspendAllowed(rules *sppb.SuspendRules, now time.Time) bool {
	if rules == nil || !rules.Enabled || rules.Schedules == nil || len(rules.Schedules) == 0 {
		return true
	}
	dayOfWeek := int(now.Weekday())
	if dayOfWeek == 0 {
		dayOfWeek = 6
	} else {
		dayOfWeek--
	}
	nowTimeNum := timeToNumber(now.Hour(), now.Minute())
	for _, rule := range rules.GetSchedules() {
		if int(rule.Day) != dayOfWeek {
			continue
		}
		if len(rule.AllowedSuspendTime) == 0 {
			return true
		}
		for _, tme := range rule.AllowedSuspendTime {
			if tme == nil {
				continue
			}
			startHour, startMin, err := parseTime(tme.StartTime)
			if err != nil {
				fmt.Println("SuspendAllowed: error parsing start time: " + err.Error())
				continue
			}
			endHour, endMin, err := parseTime(tme.EndTime)
			if err != nil {
				fmt.Println("SuspendAllowed: error parsing end time: " + err.Error())
				continue
			}
			if nowTimeNum >= timeToNumber(startHour, startMin) && nowTimeNum <= timeToNumber(endHour, endMin) {
				return true
			}
		}
	}
	return false
}

func parseTime(t string) (int, int, error) {
	hour, err := parseHour(t)
	if err != nil {
		return 0, 0, err
	}
	minute, err := parseMinute(t)
	if err != nil {
		return 0, 0, err
	}
	return hour, minute, nil
}

func parseHour(t string) (int, error) {
	if len(t) != 5 {
		return 0, fmt.Errorf("invalid time format. Must be HH:MM")
	}
	hour, err := strconv.Atoi(t[0:2])
	if err != nil {
		return 0, fmt.Errorf("cannot parse integer")
	}
	if hour > 23 {
		return 23, nil
	}
	if hour < 0 {
		return 0, nil
	}
	return hour, nil
}

func parseMinute(t string) (int, error) {
	if len(t) != 5 {
		return 0, fmt.Errorf("invalid time format. Must be HH:MM")
	}
	minute, err := strconv.Atoi(t[2:5])
	if err != nil {
		return 0, fmt.Errorf("cannot parse integer")
	}
	if minute > 59 {
		return 59, nil
	}
	if minute < 0 {
		return 0, nil
	}
	return minute, nil
}

func timeToNumber(hour, minute int) int {
	return (hour * 60) + minute
}
