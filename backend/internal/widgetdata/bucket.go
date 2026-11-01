package widgetdata

import (
	"time"

	"mapp/internal/platform/date"
)

const (
	bucketDay     = "day"
	bucketWeek    = "week"
	bucketMonth   = "month"
	bucketQuarter = "quarter"
	bucketYear    = "year"
)

var buckets = []string{bucketDay, bucketWeek, bucketMonth, bucketQuarter, bucketYear} //nolint:gochecknoglobals // allowlist

// bucketOf is the first day of the bucket holding d; weeks are ISO weeks, from Monday.
func bucketOf(d date.Date, bucket string) date.Date {
	switch bucket {
	case bucketWeek:
		sinceMonday := (int(d.Time().Weekday()) + 6) % 7
		return d.AddDays(-sinceMonday)
	case bucketMonth:
		return date.New(d.Year(), d.Month(), 1)
	case bucketQuarter:
		first := time.Month((int(d.Month())-1)/3*3 + 1)
		return date.New(d.Year(), first, 1)
	case bucketYear:
		return date.YearStart(d.Year())
	}
	return d
}
