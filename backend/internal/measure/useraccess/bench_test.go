package useraccess

import (
	"fmt"
	"testing"
	"time"

	"mapp/internal/platform/actor"
	"mapp/internal/platform/authz"
)

// The client's scale (SC-001, research R5).
const (
	benchPeople = 2_000
	benchTeams  = 200
	benchRoles  = 5
)

// largeOffice is 2,000 people in 200 teams with two roles each; three in four have signed in.
func largeOffice() *office {
	o := &office{last: map[string]time.Time{}}
	for i := range benchPeople {
		id := fmt.Sprintf("U-%d", 1000+i)
		teams := []authz.TeamRef{team(fmt.Sprintf("T-%d", 1000+i%benchTeams), fmt.Sprintf("Team %d", i%benchTeams), true),
			team(fmt.Sprintf("T-%d", 1000+(i+7)%benchTeams), fmt.Sprintf("Team %d", (i+7)%benchTeams), i%13 != 0)}
		p := person(authz.PersonRef{ID: id, Name: fmt.Sprintf("Person %d", i), JobTitle: "Planner", Active: i%10 != 0},
			teams, role(fmt.Sprintf("Role %d", i%benchRoles), "My teams"), role("Viewer", "All markets"))
		o.people = append(o.people, p)
		if i%4 != 0 {
			o.last[id] = now.Add(-time.Duration(i) * time.Hour)
		}
	}
	return o
}

// BenchmarkUserAccess_SC001 times the first page, a search, and a sort by last sign-in; under 100 ms each (R5).
func BenchmarkUserAccess_SC001(b *testing.B) {
	o := largeOffice()
	s := NewService(fakeTx{}, o, o, func() time.Time { return now })
	for name, q := range map[string]Query{"first page": {}, "search": {Search: "team 12"},
		"last sign-in": {Status: StatusAll, Sort: "-lastSignedIn"}} {
		b.Run(name, func(b *testing.B) {
			for b.Loop() {
				if _, err := s.Report(b.Context(), actor.Actor{}, q); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
