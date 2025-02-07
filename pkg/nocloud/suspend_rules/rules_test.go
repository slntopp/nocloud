package suspend_rules

import (
	sppb "github.com/slntopp/nocloud-proto/services_providers"
	"testing"
	"time"
)

func TestSuspendAllowed(t *testing.T) {
	timeNow := time.Date(2024, 2, 7, 10, 30, 0, 0, time.UTC) // Среда, 10:30 UTC
	rules := &sppb.SuspendRules{
		Enabled: true,
		Schedules: []*sppb.DaySchedule{
			{
				Day: sppb.DayOfWeek_WEDNESDAY,
				AllowedSuspendTime: []*sppb.TimeRange{
					{StartTime: "09:00", EndTime: "11:00"},
				},
			},
		},
	}
	if !SuspendAllowed(rules, timeNow) {
		t.Errorf("Expected suspend to be allowed, but it was not")
	}
}

func TestSuspendNotAllowedOutsideRange(t *testing.T) {
	timeNow := time.Date(2024, 2, 7, 12, 0, 0, 0, time.UTC) // Среда, 12:00 UTC
	rules := &sppb.SuspendRules{
		Enabled: true,
		Schedules: []*sppb.DaySchedule{
			{
				Day: sppb.DayOfWeek_WEDNESDAY,
				AllowedSuspendTime: []*sppb.TimeRange{
					{StartTime: "09:00", EndTime: "11:00"},
				},
			},
		},
	}
	if SuspendAllowed(rules, timeNow) {
		t.Errorf("Expected suspend to not be allowed, but it was")
	}
}

func TestSuspendAllowedWithNoSchedules(t *testing.T) {
	rules := &sppb.SuspendRules{
		Enabled:   true,
		Schedules: []*sppb.DaySchedule{},
	}
	if !SuspendAllowed(rules, time.Now()) {
		t.Errorf("Expected suspend to be allowed, but it was not")
	}
}

func TestSuspendAllowedWhenDisabled(t *testing.T) {
	rules := &sppb.SuspendRules{
		Enabled: false,
	}
	if !SuspendAllowed(rules, time.Now()) {
		t.Errorf("Expected suspend to be allowed, but it was not")
	}
}

func TestSuspendAllowedNilRules(t *testing.T) {
	if !SuspendAllowed(nil, time.Now()) {
		t.Errorf("Expected suspend to be allowed with nil rules, but it was not")
	}
}
