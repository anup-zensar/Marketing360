package dataset

import (
	"context"
	"encoding/base64"
	"errors"
	"slices"
	"strings"
	"testing"
	"time"

	"mapp/internal/history"
	"mapp/internal/history/historytest"
	"mapp/internal/people"
	"mapp/internal/platform/actor"
	"mapp/internal/platform/errs"
	"mapp/internal/platform/fake"
	"mapp/internal/platform/ids"
	"mapp/internal/settings"
)

const (
	europe = "T-1001"
	china  = "T-1002"
)

var (
	start = time.Date(2026, 3, 12, 9, 0, 0, 0, time.UTC)
	priya = actor.Actor{ID: "U-1002", Name: "Priya Raman"}
	sara  = actor.Actor{ID: "U-1006", Name: "Sara Novak"}
	mei   = actor.Actor{ID: "U-1004", Name: "Mei Lin"}
)

const leadsCSV = "Event,Date,Leads,Cost EUR\nExpo Hamburg,2026-01-22,140,\"18,500.00\"\nFleet Day,12/02/2026,85,9200\n"

type fixture struct {
	service *Service
	store   fakeStore
	history *history.Service
	db      *fake.DB
}

func newFixture(t *testing.T) *fixture {
	t.Helper()
	db := fake.NewDB()
	generator := ids.NewGenerator(fake.NewCounters(db))
	hist := history.NewService(db, historytest.NewStore(db), generator, func() time.Time { return start })
	store := newFakeStore(db)
	f := &fixture{store: store, history: hist, db: db}
	f.service = NewService(history.Tools{Tx: db, IDs: generator, Log: hist, Now: func() time.Time { return start }}, store,
		Deps{Directory: fakeDirectory{store: store}, Currencies: fakeCurrencies{}})
	for _, k := range f.service.Kinds() {
		if err := hist.Register(k); err != nil {
			t.Fatalf("Register: %v", err)
		}
	}
	f.member(t, priya.ID, europe)
	f.member(t, sara.ID, europe)
	f.member(t, mei.ID, china)
	return f
}

func (f *fixture) member(t *testing.T, personID, teamID string) {
	t.Helper()
	m := people.Membership{ID: "TM-" + personID + teamID, PersonID: personID, TeamID: teamID}
	if err := f.db.Write(t.Context(), func(ctx context.Context) error { return f.store.memberships.Put(ctx, m) }); err != nil {
		t.Fatalf("member: %v", err)
	}
}

func upload(name, content string) Upload {
	return Upload{FileName: name, Content: base64.StdEncoding.EncodeToString([]byte(content))}
}

func leadsInput(name string) CreateInput {
	up := upload("leads.csv", leadsCSV)
	eur := "EUR"
	return CreateInput{FileName: up.FileName, Content: up.Content, Name: name, Columns: []ColumnInput{
		{Position: 0, Label: "Event", Kind: KindText}, {Position: 1, Label: "Date", Kind: KindDate},
		{Position: 2, Label: "Leads", Kind: KindNumber}, {Position: 3, Label: "Agency cost", Kind: KindMoney, Currency: &eur},
	}}
}

func (f *fixture) create(t *testing.T, a actor.Actor, name string) View {
	t.Helper()
	view, err := f.service.Create(t.Context(), a, leadsInput(name))
	if err != nil {
		t.Fatalf("Create(%s): %v", name, err)
	}
	return view
}

func fieldErrors(t *testing.T, err error) []errs.FieldError {
	t.Helper()
	apiErr, ok := errors.AsType[*errs.Error](err)
	if !ok || apiErr.Code != errs.CodeValidationFailed {
		t.Fatalf("error = %v, want validation_failed", err)
	}
	return apiErr.Fields
}

func codeOf(err error) string {
	if apiErr, ok := errors.AsType[*errs.Error](err); ok {
		return apiErr.Code
	}
	return ""
}

func ptr[T any](v T) *T {
	return &v
}

type fakeDirectory struct {
	store fakeStore
}

func (fakeDirectory) PersonRefs(_ context.Context, personIDs []string) (map[string]people.PersonRef, error) {
	refs := map[string]people.PersonRef{}
	for _, a := range []actor.Actor{priya, sara, mei} {
		if slices.Contains(personIDs, a.ID) {
			refs[a.ID] = people.PersonRef{ID: a.ID, Name: a.Name, Initials: a.Name[:1], Active: true}
		}
	}
	return refs, nil
}

func (fakeDirectory) TeamRefs(_ context.Context, teamIDs []string) (map[string]people.TeamRef, error) {
	names := map[string]string{europe: "Lubricants Europe", china: "Mobility China"}
	refs := map[string]people.TeamRef{}
	for _, id := range teamIDs {
		if name, found := names[id]; found {
			refs[id] = people.TeamRef{ID: id, Name: name, Colour: "blue", Active: true}
		}
	}
	return refs, nil
}

func (d fakeDirectory) TeamIDsOf(_ context.Context, personID string) ([]string, error) {
	memberships, err := d.store.memberships.Where(func(m people.Membership) bool { return m.PersonID == personID })
	teamIDs := []string{}
	for _, m := range memberships {
		teamIDs = append(teamIDs, m.TeamID)
	}
	return teamIDs, err
}

type fakeCurrencies struct{}

func (fakeCurrencies) Currency(_ context.Context, code string) (settings.Currency, bool, error) {
	active := map[string]bool{"USD": true, "EUR": true, "GBP": false}
	isActive, found := active[strings.ToUpper(code)]
	return settings.Currency{Code: code, Active: isActive}, found, nil
}
