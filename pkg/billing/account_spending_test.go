package billing

import (
	"testing"
	"time"
)

func TestAccountSpendingMonths(t *testing.T) {
	now := time.Date(2026, 11, 12, 12, 0, 0, 0, time.UTC)
	day := int64(24 * 60 * 60)
	cases := []struct {
		name  string
		first int64
		want  float64
	}{
		{"paid today", now.Unix(), 1},
		{"paid yesterday", now.Unix() - day, 1},
		{"first payment 40 days ago", now.Unix() - 40*day, 2},
		{"first payment 60 days ago", now.Unix() - 60*day, 2},
		{"first payment 62 days ago", now.Unix() - 62*day, 3},
		{"first payment half a year ago", now.Unix() - 182*day, 6},
		{"first payment at the window start", now.Unix() - AccountSpendingPeriodDays*day, 13},
		{"older than the window", now.Unix() - 1000*day, 13},
	}
	for _, c := range cases {
		if got := accountSpendingMonths(c.first, now); got != c.want {
			t.Errorf("%s: got %v, want %v", c.name, got, c.want)
		}
	}
}
