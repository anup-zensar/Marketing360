package purchasereport

import (
	"fmt"
	"slices"
	"strings"
	"time"

	"mapp/internal/platform/date"
)

const noMonth = "none"

func monthOf(d date.Date) string {
	if d.IsZero() {
		return noMonth
	}
	return fmt.Sprintf("%04d-%02d", d.Year(), int(d.Month()))
}

// monthRange lists every month from the earliest to the latest of keys, then none when a key has no delivery date
// (042-C2).
func monthRange(keys []string) []string {
	real := slices.DeleteFunc(slices.Clone(keys), func(k string) bool { return k == noMonth })
	slices.Sort(real)
	months := []string{}
	if len(real) > 0 {
		first, errFirst := time.Parse("2006-01", real[0])
		last, errLast := time.Parse("2006-01", real[len(real)-1])
		if errFirst == nil && errLast == nil {
			for m := first; !m.After(last); m = m.AddDate(0, 1, 0) {
				months = append(months, m.Format("2006-01"))
			}
		}
	}
	if slices.Contains(keys, noMonth) {
		months = append(months, noMonth)
	}
	return months
}

func monthLabel(key string) string {
	if key == noMonth {
		return "No delivery date"
	}
	return strings.Replace(key, "-", " ", 1)
}
