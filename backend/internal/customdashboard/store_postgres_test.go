package customdashboard_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"mapp/internal/customdashboard"
	"mapp/internal/people"
	"mapp/internal/platform/actor"
	"mapp/internal/platform/database"
	"mapp/internal/platform/database/dbtest"
	"mapp/internal/platform/date"
	"mapp/internal/platform/record"
)

var (
	pgMeta  = record.NewMeta(actor.System, time.Date(2026, time.March, 12, 9, 0, 0, 0, time.UTC))
	pgDates = cmpopts.EquateComparable(date.Date{})
)

// pgStore gives a store with the owner U-1001, Daniel U-1002 in T-1001, Sara U-1003 who left T-1001 and Mei
// U-1004; DB-1001 is shared with T-1001 and reads DS-1001, DB-1002 is private, DB-1003 is Mei's.
func pgStore(t *testing.T) (*customdashboard.PostgresStore, []customdashboard.Dashboard) {
	t.Helper()
	db := dbtest.NewDB(t)
	s, directory := customdashboard.NewPostgresStore(db), people.NewPostgresStore(db)
	dashboards := []customdashboard.Dashboard{
		pgDashboard("DB-1001", "U-1001", []string{"T-1001"}, "DS-1001"),
		pgDashboard("DB-1002", "U-1001", []string{}, "activities"),
		pgDashboard("DB-1003", "U-1004", []string{"T-1002"}, "DS-1001"),
	}
	pgWrite(t, db, func(ctx context.Context) error {
		err := errors.Join(
			directory.PutTeam(ctx, people.Team{ID: "T-1001", Name: "Lubricants Europe", Colour: "blue", Meta: pgMeta}),
			directory.PutTeam(ctx, people.Team{ID: "T-1002", Name: "Mobility China", Colour: "teal", Meta: pgMeta}),
			directory.PutMembership(ctx, people.Membership{ID: "TM-1001", TeamID: "T-1001", PersonID: "U-1002", Meta: pgMeta}),
			directory.PutMembership(ctx, people.Membership{ID: "TM-1002", TeamID: "T-1001", PersonID: "U-1003", Meta: pgMeta}))
		for _, id := range []string{"U-1001", "U-1002", "U-1003", "U-1004"} {
			err = errors.Join(err, directory.Put(ctx, people.Person{ID: id, Name: id, Email: id + "@example.com", Meta: pgMeta}))
		}
		for _, d := range dashboards {
			err = errors.Join(err, s.Insert(ctx, d))
		}
		return err
	})
	pgWrite(t, db, func(ctx context.Context) error { return directory.DeleteMembership(ctx, "TM-1002") })
	return s, dashboards
}

func pgWrite(t *testing.T, db *database.DB, fn func(ctx context.Context) error) {
	t.Helper()
	if err := db.Write(t.Context(), fn); err != nil {
		t.Fatalf("Write() error: %v", err)
	}
}

func pgDashboard(id, ownerID string, teams []string, source string) customdashboard.Dashboard {
	from := date.MustParse("2026-07-01")
	return customdashboard.Dashboard{
		ID: id, Name: "Dashboard " + id, Description: "About " + id, OwnerID: ownerID, SharedWith: teams,
		Filters: customdashboard.Filters{DateFrom: &from, PlanIDs: []string{"P-1001"}, TeamIDs: []string{}},
		Widgets: []customdashboard.Widget{
			{ID: "w1", Type: "kpi", Layout: customdashboard.Layout{W: 3, H: 2}, Query: &customdashboard.Query{
				Source: source, Measures: []customdashboard.Measure{{Aggregation: "count"}}, Filters: []customdashboard.Filter{},
				Sort: customdashboard.Sort{By: "measure", Direction: "desc"}, Limit: 50,
			}},
			{ID: "w2", Type: "text", Layout: customdashboard.Layout{X: 3, W: 4, H: 2}, Note: "Figures in USD."},
		},
		Meta: pgMeta,
	}
}

func TestPostgresStoreVisibleDashboards_ScopeIsolation_FR005(t *testing.T) {
	t.Parallel()
	s, all := pgStore(t)
	tests := []struct {
		name     string
		personID string
		want     []customdashboard.Dashboard
		using    []customdashboard.Ref
	}{
		{name: "the owner sees their own", personID: "U-1001", want: all[:2],
			using: []customdashboard.Ref{{ID: "DB-1001", Name: "Dashboard DB-1001"}}},
		{name: "a member of the shared team", personID: "U-1002", want: all[:1],
			using: []customdashboard.Ref{{ID: "DB-1001", Name: "Dashboard DB-1001"}}},
		{name: "a former member", personID: "U-1003", want: []customdashboard.Dashboard{}, using: []customdashboard.Ref{}},
		{name: "another person", personID: "U-1004", want: all[2:],
			using: []customdashboard.Ref{{ID: "DB-1003", Name: "Dashboard DB-1003"}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			listed, err := s.ListVisible(t.Context(), tt.personID)
			if err != nil {
				t.Fatalf("ListVisible() error: %v", err)
			}
			if diff := cmp.Diff(tt.want, listed, pgDates); diff != "" {
				t.Errorf("ListVisible(%s) mismatch (-want +got):\n%s", tt.personID, diff)
			}
			for _, d := range all {
				assertGetVisible(t, s, visibility{personID: tt.personID, dashboard: d, want: containsID(tt.want, d.ID)})
			}
			using, err := s.UsingDataset(t.Context(), tt.personID, "DS-1001")
			if err != nil || !cmp.Equal(tt.using, using) {
				t.Errorf("UsingDataset(%s) = %v, %v; want %v", tt.personID, using, err, tt.using)
			}
		})
	}
}

type visibility struct {
	personID  string
	dashboard customdashboard.Dashboard
	want      bool
}

func assertGetVisible(t *testing.T, s *customdashboard.PostgresStore, v visibility) {
	t.Helper()
	got, found, err := s.GetVisible(t.Context(), v.personID, v.dashboard.ID)
	if err != nil || found != v.want || (found && !cmp.Equal(v.dashboard, got, pgDates)) {
		t.Errorf("GetVisible(%s, %s) = found %v, error %v; want found %v", v.personID, v.dashboard.ID, found, err, v.want)
	}
}

func containsID(list []customdashboard.Dashboard, id string) bool {
	for _, d := range list {
		if d.ID == id {
			return true
		}
	}
	return false
}
