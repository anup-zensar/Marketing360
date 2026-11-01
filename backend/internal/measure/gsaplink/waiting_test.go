package gsaplink

import (
	"testing"
	"time"
)

func TestWaitingHours_FR006(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		since time.Time
		want  int
	}{
		{name: "approved three days ago", since: now.Add(-72 * time.Hour), want: 72},
		{name: "part of an hour is not counted", since: now.Add(-90 * time.Minute), want: 1},
		{name: "a clock ahead is never negative", since: now.Add(time.Hour), want: 0},
	}
	for _, tt := range tests {
		if got := waitingHours(tt.since, now); got != tt.want {
			t.Errorf("%s: waitingHours = %d, want %d", tt.name, got, tt.want)
		}
	}
}

func TestBand_FR007(t *testing.T) {
	t.Parallel()
	tests := map[int]string{0: BandUnder48h, 48: BandUnder48h, 49: BandDays2to5, 120: BandDays2to5, 121: BandDays5to10,
		240: BandDays5to10, 241: BandOver10Days}
	for hours, want := range tests {
		if got := band(hours); got != want {
			t.Errorf("band(%d) = %s, want %s", hours, got, want)
		}
	}
}

func TestWorkingDays_FR009(t *testing.T) {
	t.Parallel()
	friday := time.Date(2026, 3, 13, 16, 0, 0, 0, time.UTC)
	tests := []struct {
		name       string
		since, now time.Time
		want       int
	}{
		{name: "the same day", since: friday, now: friday.Add(4 * time.Hour), want: 0},
		{name: "Friday to Monday", since: friday, now: friday.AddDate(0, 0, 3), want: 1},
		{name: "Friday to Wednesday", since: friday, now: now, want: 3},
		{name: "over a weekend only", since: friday, now: friday.AddDate(0, 0, 2), want: 0},
		{name: "two weeks", since: friday, now: friday.AddDate(0, 0, 14), want: 10},
		{name: "a clock ahead is never negative", since: now, now: friday, want: 0},
	}
	for _, tt := range tests {
		if got := workingDays(tt.since, tt.now); got != tt.want {
			t.Errorf("%s: workingDays = %d, want %d", tt.name, got, tt.want)
		}
	}
}
