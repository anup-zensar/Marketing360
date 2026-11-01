package useraccess

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"mapp/internal/platform/date"
)

func TestBands_Boundaries_FR007(t *testing.T) {
	t.Parallel()
	ago := func(days int) *time.Time {
		at := now.AddDate(0, 0, -days)
		return &at
	}
	tests := []struct {
		at   *time.Time
		want string
	}{
		{at: ago(0), want: BandWithin30Days}, {at: ago(30), want: BandWithin30Days}, {at: ago(31), want: BandDays31to90},
		{at: ago(90), want: BandDays31to90}, {at: ago(91), want: BandOver90Days}, {at: nil, want: BandNever},
	}
	for _, tt := range tests {
		if got := bandOf(tt.at, now); got != tt.want {
			t.Errorf("bandOf(%v) = %s, want %s", tt.at, got, tt.want)
		}
	}
}

func TestCards_FR007(t *testing.T) {
	t.Parallel()
	r := newOffice().report(t, Query{})

	if diff := cmp.Diff(Cards{People: 4, SignedInWithin30Days: 2, NeverSignedIn: 1}, r.Cards); diff != "" {
		t.Errorf("cards (-want +got):\n%s", diff)
	}
	want := []Band{{Key: BandWithin30Days, Label: "Within 30 days", Count: 2},
		{Key: BandDays31to90, Label: "31 – 90 days", Count: 1}, {Key: BandOver90Days, Label: "Over 90 days"},
		{Key: BandNever, Label: "Never", Count: 1}}
	if diff := cmp.Diff(want, r.Bands); diff != "" {
		t.Errorf("bands, adding up to People (-want +got):\n%s", diff)
	}
	if r.RecordedSince != date.New(2026, 3, 18) {
		t.Errorf("recorded since = %v, want today when unknown", r.RecordedSince)
	}
}
