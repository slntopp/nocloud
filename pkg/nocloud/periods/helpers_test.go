package periods

import (
	"testing"
	"time"
)

func TestGetNextDate_BillingMonth(t *testing.T) {
	tests := []struct {
		name         string
		cycleStart   string
		current      string
		expectedNext string
	}{
		{
			name:         "End of Jan to Feb",
			cycleStart:   "2021-01-31T10:00:00Z",
			current:      "2021-01-31T10:00:00Z",
			expectedNext: "2021-02-28T10:00:00Z",
		},
		{
			name:         "Feb to Mar",
			cycleStart:   "2021-01-31T10:00:00Z",
			current:      "2021-02-28T10:00:00Z",
			expectedNext: "2021-03-31T10:00:00Z",
		},
		{
			name:         "March to April",
			cycleStart:   "2021-01-31T10:00:00Z",
			current:      "2021-03-31T10:00:00Z",
			expectedNext: "2021-04-30T10:00:00Z",
		},
		{
			name:         "30th start",
			cycleStart:   "2021-01-30T15:30:00Z",
			current:      "2021-01-30T15:30:00Z",
			expectedNext: "2021-02-28T15:30:00Z",
		},
		{
			name:         "30th start to March 30th",
			cycleStart:   "2021-01-30T15:30:00Z",
			current:      "2021-02-28T15:30:00Z",
			expectedNext: "2021-03-30T15:30:00Z",
		},
		{
			name:         "Leap year feb",
			cycleStart:   "2020-01-31T00:00:00Z",
			current:      "2020-01-31T00:00:00Z",
			expectedNext: "2020-02-29T00:00:00Z",
		},
		{
			name:         "Year rollover",
			cycleStart:   "2021-12-31T23:59:59Z",
			current:      "2021-12-31T23:59:59Z",
			expectedNext: "2022-01-31T23:59:59Z",
		},
		{
			name:         "Default",
			cycleStart:   "2021-05-15T08:00:00Z",
			current:      "2021-05-15T08:00:00Z",
			expectedNext: "2021-06-15T08:00:00Z",
		},
		{
			name:         "Other case 1",
			cycleStart:   "2020-04-23T08:00:00Z",
			current:      "2021-02-28T08:00:00Z",
			expectedNext: "2021-03-28T08:00:00Z",
		},
		{
			name:         "Other case 2",
			cycleStart:   "2025-04-23T08:00:00Z",
			current:      "2025-02-28T08:00:00Z",
			expectedNext: "2025-03-28T08:00:00Z",
		},
		{
			name:         ">28 but mid-cycle current day",
			cycleStart:   "2021-01-31T10:00:00Z",
			current:      "2021-02-15T10:00:00Z",
			expectedNext: "2021-03-15T10:00:00Z",
		},
		{
			name:         ">28 and end-of-month current day",
			cycleStart:   "2021-01-29T09:00:00Z",
			current:      "2021-02-28T09:00:00Z",
			expectedNext: "2021-03-29T09:00:00Z",
		},
		{
			name:         "<=28 but >27",
			cycleStart:   "2021-03-27T12:00:00Z",
			current:      "2021-04-28T12:00:00Z",
			expectedNext: "2021-05-28T12:00:00Z",
		},
		{
			name:         "February for 29th start",
			cycleStart:   "2021-01-29T00:00:00Z",
			current:      "2021-02-28T00:00:00Z",
			expectedNext: "2021-03-29T00:00:00Z",
		},
		{
			name:         "February leap for 30th start",
			cycleStart:   "2020-01-30T05:00:00Z",
			current:      "2020-02-29T05:00:00Z",
			expectedNext: "2020-03-30T05:00:00Z",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cs, _ := time.Parse(time.RFC3339, tt.cycleStart)
			cur, _ := time.Parse(time.RFC3339, tt.current)
			exp, _ := time.Parse(time.RFC3339, tt.expectedNext)
			got := time.Unix(GetNextDate(cur.Unix(), BillingMonth, cs.Unix()), 0).UTC()
			if !got.Equal(exp) {
				t.Errorf("%s: expected %s, got %s", tt.name, exp, got)
			}
		})
	}
}
