package gsaplink

import (
	"time"

	"mapp/internal/platform/date"
)

// The turnaround limits of 044-C5: 48 hours for an SWBS code, 2 working days for GSAP's answer to a routing.
const (
	swbsTurnaroundHours = 48
	gsapTurnaroundDays  = 2
	hoursPerDay         = 24
	bandDays5           = 5
	bandDays10          = 10
)

func waitingHours(since, now time.Time) int {
	return max(0, int(now.Sub(since)/time.Hour))
}

// band places a wait in the Waiting by time chart; Under 48 hours holds every row not flagged (FR-007).
func band(hours int) string {
	switch {
	case hours <= swbsTurnaroundHours:
		return BandUnder48h
	case hours <= bandDays5*hoursPerDay:
		return BandDays2to5
	case hours <= bandDays10*hoursPerDay:
		return BandDays5to10
	}
	return BandOver10Days
}

func bands() []Band {
	return []Band{{Key: BandUnder48h, Label: "Under 48 hours"}, {Key: BandDays2to5, Label: "2 – 5 days"},
		{Key: BandDays5to10, Label: "5 – 10 days"}, {Key: BandOver10Days, Label: "Over 10 days"}}
}

// workingDays counts the days Monday to Friday after since's day up to now's day, in UTC (FR-009, R5).
func workingDays(since, now time.Time) int {
	from, to := date.Today(since.UTC()), date.Today(now.UTC())
	days := 0
	for d := from.AddDays(1); !d.After(to); d = d.AddDays(1) {
		if weekday := d.Time().Weekday(); weekday != time.Saturday && weekday != time.Sunday {
			days++
		}
	}
	return days
}
