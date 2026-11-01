package useraccess

import (
	"context"
	"testing"
	"time"

	"mapp/internal/platform/actor"
	"mapp/internal/platform/authz"
)

// now is Wednesday 2026-03-18, 12:00 UTC.
var now = time.Date(2026, 3, 18, 12, 0, 0, 0, time.UTC)

// office is five people: Priya (2 days ago, an inactive team too), Alex (45 days), Mei (no role, never), Carlos
// (exactly 30 days) and Daniel (inactive, 120 days).
type office struct {
	people []Person
	last   map[string]time.Time
	since  time.Time
}

func newOffice() *office {
	europe, china, archive := team("T-1001", "Lubricants Europe", true), team("T-1002", "Mobility China", true),
		team("T-1009", "Archive", false)
	return &office{
		people: []Person{
			person("U-1002", "Priya Raman", "Marketing planner", true, []authz.TeamRef{europe, archive}, role("Planner", "My teams")),
			person("U-1001", "Alex Morgan", "Platform administrator", true, nil, role("Administrator", "All markets")),
			person("U-1004", "Mei Lin", "Planner", true, []authz.TeamRef{china}),
			person("U-1005", "Carlos Mendes", "Finance lead", true, []authz.TeamRef{china}, role("Finance", "All markets")),
			person("U-1003", "Daniel Okoro", "Approver", false, []authz.TeamRef{europe}, role("Viewer", "My teams")),
		},
		last: map[string]time.Time{"U-1002": now.AddDate(0, 0, -2), "U-1001": now.AddDate(0, 0, -45),
			"U-1005": now.AddDate(0, 0, -30), "U-1003": now.AddDate(0, 0, -120)},
	}
}

func team(id, name string, active bool) authz.TeamRef {
	return authz.TeamRef{ID: id, Name: name, Active: active}
}

func role(name, scope string) authz.RoleAssignment {
	return authz.RoleAssignment{Role: authz.RoleRef{Name: name}, Scope: authz.AccessScope{Text: scope}}
}

func person(id, name, job string, active bool, teams []authz.TeamRef, roles ...authz.RoleAssignment) Person {
	email := map[string]string{"U-1001": "alex.morgan@shell.com", "U-1002": "priya.raman@shell.com",
		"U-1003": "daniel.okoro@shell.com", "U-1004": "mei.lin@shell.com", "U-1005": "carlos.mendes@shell.com"}[id]
	return Person{Ref: authz.PersonRef{ID: id, Name: name, JobTitle: job, Active: active}, Email: email,
		Roles: append([]authz.RoleAssignment{}, roles...), Teams: append([]authz.TeamRef{}, teams...)}
}

func (o *office) PeopleAccess(context.Context, actor.Actor) ([]Person, error) { return o.people, nil }

func (o *office) LastSignIns(context.Context) (map[string]time.Time, error) { return o.last, nil }

func (o *office) SignInsRecordedSince(context.Context) (time.Time, error) { return o.since, nil }

type fakeTx struct{}

func (fakeTx) Read(ctx context.Context, fn func(ctx context.Context) error) error { return fn(ctx) }

func (o *office) report(tb testing.TB, q Query) Report {
	tb.Helper()
	r, err := NewService(fakeTx{}, o, o, func() time.Time { return now }).Report(tb.Context(), actor.Actor{}, q)
	if err != nil {
		tb.Fatalf("Report() error = %v", err)
	}
	return r
}

func names(list []Row) []string {
	out := []string{}
	for _, r := range list {
		out = append(out, r.Person.Name)
	}
	return out
}
