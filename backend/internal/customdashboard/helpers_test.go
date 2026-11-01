package customdashboard

import (
	"context"
	"errors"
	"testing"
	"time"

	"mapp/internal/history"
	"mapp/internal/history/historytest"
	"mapp/internal/people"
	"mapp/internal/platform/actor"
	"mapp/internal/platform/errs"
	"mapp/internal/platform/fake"
	"mapp/internal/platform/ids"
	"mapp/internal/platform/record"
)

var (
	priya  = actor.Actor{ID: "U-1002", Name: "Priya Raman"}
	daniel = actor.Actor{ID: "U-1003", Name: "Daniel Okoro"}
	mei    = actor.Actor{ID: "U-1004", Name: "Mei Lin"}
	start  = time.Date(2026, time.March, 12, 9, 0, 0, 0, time.UTC)
)

type testClock struct{ now time.Time }

func (c *testClock) Now() time.Time { return c.now }

type fixture struct {
	service *Service
	db      *fake.DB
	clock   *testClock
	history *history.Service
	teams   *fake.Table[people.Team]
	members *fake.Table[people.Membership]
}

// newFixture has Priya and Daniel in Lubricants Europe (T-1001) and Mei in Mobility China (T-1002).
func newFixture(t *testing.T) *fixture {
	t.Helper()
	db := fake.NewDB()
	clock := &testClock{now: start}
	generator := ids.NewGenerator(fake.NewCounters(db))
	hist := history.NewService(db, historytest.NewStore(db), generator, clock.Now)
	f := &fixture{db: db, clock: clock, history: hist,
		teams:   fake.TableOf(db, "teams", func(t people.Team) string { return t.ID }),
		members: fake.TableOf(db, "team_memberships", func(m people.Membership) string { return m.ID })}
	f.service = NewService(history.Tools{Tx: db, IDs: generator, Log: hist, Now: clock.Now}, newFakeStore(db),
		fakeDirectory{teams: f.teams, members: f.members})
	if err := hist.Register(f.service.Kinds()[0]); err != nil {
		t.Fatalf("Register: %v", err)
	}
	f.write(t, func(ctx context.Context) error {
		return errors.Join(
			f.teams.Put(ctx, people.Team{ID: "T-1001", Name: "Lubricants Europe", Colour: "blue", Active: true}),
			f.teams.Put(ctx, people.Team{ID: "T-1002", Name: "Mobility China", Colour: "teal", Active: true}),
			f.join(ctx, "TM-1001", "T-1001", priya.ID), f.join(ctx, "TM-1002", "T-1001", daniel.ID),
			f.join(ctx, "TM-1003", "T-1002", mei.ID))
	})
	return f
}

func (f *fixture) join(ctx context.Context, id, teamID, personID string) error {
	return f.members.Put(ctx, people.Membership{ID: id, TeamID: teamID, PersonID: personID, Meta: record.NewMeta(actor.System, start)})
}

func (f *fixture) write(t *testing.T, fn func(ctx context.Context) error) {
	t.Helper()
	if err := f.db.Write(t.Context(), fn); err != nil {
		t.Fatalf("Write: %v", err)
	}
}

func (f *fixture) create(t *testing.T, a actor.Actor, name string, widgets ...Widget) View {
	t.Helper()
	view, err := f.service.Create(t.Context(), a, CreateInput{Name: name, Widgets: widgets})
	if err != nil {
		t.Fatalf("Create(%s): %v", name, err)
	}
	return view
}

func (f *fixture) share(t *testing.T, view View, teamIDs ...string) View {
	t.Helper()
	shared, err := f.service.Share(t.Context(), priya, view.ID, SharingInput{Version: &view.Version, TeamIDs: teamIDs})
	if err != nil {
		t.Fatalf("Share(%s): %v", view.ID, err)
	}
	return shared
}

func (f *fixture) log(t *testing.T) []history.Entry {
	t.Helper()
	entries, err := f.history.Entries(t.Context())
	if err != nil {
		t.Fatalf("Entries: %v", err)
	}
	return entries
}

func assertCode(t *testing.T, err error, status int, code string) {
	t.Helper()
	apiErr, ok := errors.AsType[*errs.Error](err)
	if !ok || apiErr.Status != status || apiErr.Code != code {
		t.Fatalf("error = %v, want %d %s", err, status, code)
	}
}

type fakeDirectory struct {
	teams   *fake.Table[people.Team]
	members *fake.Table[people.Membership]
}

func (d fakeDirectory) PersonRefs(_ context.Context, wanted []string) (map[string]people.PersonRef, error) {
	names := map[string]string{priya.ID: priya.Name, daniel.ID: daniel.Name, mei.ID: mei.Name}
	refs := map[string]people.PersonRef{}
	for _, id := range wanted {
		refs[id] = people.PersonRef{ID: id, Name: names[id], Active: true}
	}
	return refs, nil
}

func (d fakeDirectory) TeamRefs(_ context.Context, wanted []string) (map[string]people.TeamRef, error) {
	refs := map[string]people.TeamRef{}
	for _, id := range wanted {
		team, found, err := d.teams.Get(id)
		if err != nil {
			return nil, err
		}
		if found {
			refs[id] = team.Ref()
		}
	}
	return refs, nil
}

func (d fakeDirectory) TeamsOf(_ context.Context, personID string) ([]string, error) {
	of, err := d.members.Where(func(m people.Membership) bool { return m.PersonID == personID })
	teams := []string{}
	for _, m := range of {
		teams = append(teams, m.TeamID)
	}
	return teams, err
}
