package dataset_test

import (
	"context"
	"slices"
	"testing"
	"time"

	"mapp/internal/dataset"
	"mapp/internal/people"
	"mapp/internal/platform/actor"
	"mapp/internal/platform/database"
	"mapp/internal/platform/database/dbtest"
	"mapp/internal/platform/record"
)

const (
	owner   = "U-1"
	member  = "U-2"
	other   = "U-3"
	shared  = "DS-1"
	private = "DS-2"
)

// scopeDB holds two datasets of the owner, one shared with the member's team, and a third person in no team.
func scopeDB(t *testing.T) (*dataset.PostgresStore, *database.DB) {
	t.Helper()
	db := dbtest.NewDB(t)
	meta := record.NewMeta(actor.System, time.Date(2026, 3, 12, 9, 0, 0, 0, time.UTC))
	persons := people.NewPostgresStore(db)
	store := dataset.NewPostgresStore(db)
	eur := "EUR"
	columns := []dataset.Column{{Key: "c1", Label: "Event", Kind: dataset.KindText},
		{Key: "c2", Position: 1, Label: "Cost", Kind: dataset.KindMoney, Currency: &eur}}
	rows := []dataset.Row{{dataset.TextCell("Expo"), dataset.MinorCell(1850000)}, {dataset.TextCell("Fair"), {Empty: true}}}
	err := db.Write(t.Context(), func(ctx context.Context) error {
		return firstError(
			func() error { return putPeople(ctx, persons, meta) },
			func() error {
				return store.Insert(ctx, dataset.Dataset{ID: shared, Name: "Shared", OwnerID: owner, FileName: "a.csv",
					RowCount: 2, Columns: columns, SharedWith: []string{"T-1"}, Meta: meta}, rows)
			},
			func() error {
				return store.Insert(ctx, dataset.Dataset{ID: private, Name: "Private", OwnerID: owner, FileName: "b.csv",
					Columns: columns, SharedWith: []string{}, Meta: meta}, nil)
			},
		)
	})
	if err != nil {
		t.Fatalf("fill: %v", err)
	}
	return store, db
}

func putPeople(ctx context.Context, persons *people.PostgresStore, meta record.Meta) error {
	return firstError(
		func() error {
			return persons.Put(ctx, people.Person{ID: owner, Name: "Owner", Email: "o@shell.com", Active: true, Meta: meta})
		},
		func() error {
			return persons.Put(ctx, people.Person{ID: member, Name: "Member", Email: "m@shell.com", Active: true, Meta: meta})
		},
		func() error {
			return persons.Put(ctx, people.Person{ID: other, Name: "Other", Email: "x@shell.com", Active: true, Meta: meta})
		},
		func() error {
			return persons.PutTeam(ctx, people.Team{ID: "T-1", Name: "One", Colour: "blue", Active: true, Meta: meta})
		},
		func() error {
			return persons.PutTeam(ctx, people.Team{ID: "T-2", Name: "Two", Colour: "teal", Active: true, Meta: meta})
		},
		func() error {
			return persons.PutMembership(ctx, people.Membership{ID: "TM-1", TeamID: "T-1", PersonID: member, Meta: meta})
		},
		func() error {
			return persons.PutMembership(ctx, people.Membership{ID: "TM-2", TeamID: "T-2", PersonID: other, Meta: meta})
		},
	)
}

func firstError(steps ...func() error) error {
	for _, step := range steps {
		if err := step(); err != nil {
			return err
		}
	}
	return nil
}

func TestPostgresStoreVisibleDatasets_ScopeIsolation_FR044(t *testing.T) {
	t.Parallel()
	store, _ := scopeDB(t)
	tests := []struct {
		person string
		want   []string
	}{
		{person: owner, want: []string{private, shared}},
		{person: member, want: []string{shared}},
		{person: other, want: []string{}},
	}
	for _, tt := range tests {
		t.Run(tt.person, func(t *testing.T) {
			t.Parallel()
			listed, err := store.ListVisible(t.Context(), tt.person)
			if err != nil {
				t.Fatalf("ListVisible: %v", err)
			}
			ids := []string{}
			for _, d := range listed {
				ids = append(ids, d.ID)
			}
			if !slices.Equal(ids, tt.want) {
				t.Errorf("ListVisible(%s) = %v, want %v", tt.person, ids, tt.want)
			}
			for _, id := range []string{shared, private} {
				_, found, err := store.GetVisible(t.Context(), tt.person, id)
				if err != nil || found != slices.Contains(tt.want, id) {
					t.Errorf("GetVisible(%s, %s) = %v, %v", tt.person, id, found, err)
				}
			}
		})
	}
}

func TestPostgresStoreRowsAndColumns_AsStored_FR042(t *testing.T) {
	t.Parallel()
	store, _ := scopeDB(t)
	d, found, err := store.GetVisible(t.Context(), member, shared)
	if err != nil || !found {
		t.Fatalf("GetVisible = %v, %v", found, err)
	}
	page, err := store.RowsPage(t.Context(), shared, 1, 5)
	if err != nil {
		t.Fatalf("RowsPage: %v", err)
	}
	swapped := []dataset.Column{d.Columns[0], d.Columns[1]}
	swapped[0].Label, swapped[1].Label = "Cost", "Event"
	swapErr := store.UpdateLabels(t.Context(), shared, swapped)
	ownerID, _, _ := store.OwnerOf(t.Context(), shared)

	same(t, "dataset", []any{2, "EUR", []string{"T-1"}}, []any{len(d.Columns), *d.Columns[1].Currency, d.SharedWith})
	same(t, "page", []dataset.Row{{dataset.TextCell("Fair"), {Empty: true}}}, page)
	same(t, "swap and owner", []any{true, owner}, []any{swapErr == nil, ownerID})
}

var _ dataset.Store = (*dataset.PostgresStore)(nil)

func TestPostgresStore_SharesLeaveWithADeletedTeam_D379(t *testing.T) {
	t.Parallel()
	store, db := scopeDB(t)
	persons := people.NewPostgresStore(db)

	err := db.Write(t.Context(), func(ctx context.Context) error {
		return firstError(func() error { return persons.DeleteMembership(ctx, "TM-1") },
			func() error { return persons.DeleteTeam(ctx, "T-1") })
	})
	if err != nil {
		t.Fatalf("deleting the shared team: %v", err)
	}

	d, found, err := store.Get(t.Context(), shared)
	if err != nil || !found || len(d.SharedWith) != 0 {
		t.Errorf("Get(%s) = shared with %q, found %v, error %v; want shared with nobody", shared, d.SharedWith, found, err)
	}
}
