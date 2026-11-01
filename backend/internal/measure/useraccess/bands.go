package useraccess

import (
	"slices"
	"time"
)

const (
	recentDays = 30
	longDays   = 90
	hoursInDay = 24
)

// bandOf places a last sign-in by whole days before now: within 30, 31 – 90, over 90, or never (FR-007, R4).
func bandOf(at *time.Time, now time.Time) string {
	if at == nil {
		return BandNever
	}
	switch days := int(now.Sub(*at) / (hoursInDay * time.Hour)); {
	case days <= recentDays:
		return BandWithin30Days
	case days <= longDays:
		return BandDays31to90
	}
	return BandOver90Days
}

// bandsAndCards count the matching people, adding up to the People card.
func bandsAndCards(list []Row, now time.Time) ([]Band, Cards) {
	bands := []Band{{Key: BandWithin30Days, Label: "Within 30 days"}, {Key: BandDays31to90, Label: "31 – 90 days"},
		{Key: BandOver90Days, Label: "Over 90 days"}, {Key: BandNever, Label: "Never"}}
	for _, r := range list {
		key := bandOf(r.LastSignedInAt, now)
		bands[slices.IndexFunc(bands, func(b Band) bool { return b.Key == key })].Count++
	}
	return bands, Cards{People: len(list), SignedInWithin30Days: bands[0].Count, NeverSignedIn: bands[3].Count}
}
