// Package revenuetest fakes feature 049's one store read, for tests that build the service or the whole
// API without a database.
package revenuetest

import (
	"context"
	"slices"

	"mapp/internal/measure/revenue"
	"mapp/internal/platform/authz"
)

var _ revenue.Facts = (*Store)(nil)

// Store holds every fact total a test seeds directly; it applies no scope filter itself, because the
// activity IDs a caller passes are already the ones its own scope resolved (as 046's fake does).
type Store struct {
	Totals []revenue.FactTotal
}

func NewStore() *Store {
	return &Store{}
}

func (s *Store) ReportTotals(_ context.Context, _ authz.Scope, activityIDs []string, year int) ([]revenue.FactTotal, error) {
	found := []revenue.FactTotal{}
	for _, t := range s.Totals {
		if slices.Contains(activityIDs, t.ActivityID) && yearOf(t.Month) == year {
			found = append(found, t)
		}
	}
	return found, nil
}

func yearOf(month string) int {
	if len(month) < 4 {
		return 0
	}
	year := 0
	for _, r := range month[:4] {
		year = year*10 + int(r-'0')
	}
	return year
}
