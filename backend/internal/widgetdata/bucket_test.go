package widgetdata

import (
	"testing"

	"mapp/internal/platform/date"
)

func TestBuckets_DayWeekMonthQuarterYear_FR017(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name, day, bucket, want string
	}{
		{"a day is itself", "2026-08-15", bucketDay, "2026-08-15"},
		{"a week starts on Monday", "2026-08-15", bucketWeek, "2026-08-10"},
		{"a Monday starts its own week", "2026-08-10", bucketWeek, "2026-08-10"},
		{"a Sunday belongs to the week before", "2026-08-16", bucketWeek, "2026-08-10"},
		{"an ISO week across years starts in December", "2027-01-01", bucketWeek, "2026-12-28"},
		{"New Year's Day 2026 is in the week of 29 December 2025", "2026-01-01", bucketWeek, "2025-12-29"},
		{"a month starts on the first", "2026-08-15", bucketMonth, "2026-08-01"},
		{"a quarter starts in its first month", "2026-08-15", bucketQuarter, "2026-07-01"},
		{"the first quarter", "2026-03-31", bucketQuarter, "2026-01-01"},
		{"a year starts on 1 January", "2026-08-15", bucketYear, "2026-01-01"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := bucketOf(date.MustParse(tt.day), tt.bucket).String(); got != tt.want {
				t.Errorf("bucketOf(%s, %s) = %s, want %s", tt.day, tt.bucket, got, tt.want)
			}
		})
	}
}
