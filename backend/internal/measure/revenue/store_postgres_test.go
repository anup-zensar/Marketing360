package revenue_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"mapp/internal/activity"
	"mapp/internal/campaign"
	"mapp/internal/measure/revenue"
	"mapp/internal/people"
	"mapp/internal/performance"
	"mapp/internal/plan"
	"mapp/internal/platform/actor"
	"mapp/internal/platform/authz"
	"mapp/internal/platform/database"
	"mapp/internal/platform/database/dbtest"
	"mapp/internal/platform/money"
	"mapp/internal/platform/record"
	"mapp/internal/settings"
)

var pgMeta = record.NewMeta(actor.System, time.Date(2026, time.March, 12, 9, 0, 0, 0, time.UTC))

// team1Only holds plans.view in team T-1001 alone, so it reaches A-1001 and nothing of A-1002 — the same
// scope shape 046's own scope-isolation tests use (store_postgres_scope_test.go).
var team1Only = authz.Scope{PersonID: "U-2001", Holds: true, TeamIDs: []string{"T-1001"}}

// pgRevenueStore seeds A-1001 (team T-1001) and A-1002 (team T-1002), each with one 2026 revenue fact.
func pgRevenueStore(t *testing.T) *revenue.PostgresStore {
	t.Helper()
	db := dbtest.NewDB(t)
	store := revenue.NewPostgresStore(db)
	facts := performance.NewPostgresStore(db)
	directory, catalogue := people.NewPostgresStore(db), settings.NewPostgresStore(db)
	plans, containers := plan.NewPostgresStore(db), campaign.NewPostgresStore(db)
	activities := activity.NewPostgresStore(db)

	pgWrite(t, db, func(ctx context.Context) error {
		err := errors.Join(
			directory.Put(ctx, people.Person{ID: "U-2001", Name: "Sam Patel", Email: "sam@example.com", Meta: pgMeta}),
			catalogue.PutCurrency(ctx, settings.Currency{Code: "USD", Name: "US dollar", Active: true, Meta: pgMeta}),
			catalogue.PutActivityType(ctx, settings.ActivityType{ID: "AT-1001", Name: "Webinar", Meta: pgMeta}),
			catalogue.PutModel(ctx, settings.Model{ID: "MOD-1001", Name: "Leads", ActivityTypeIDs: []string{"AT-1001"},
				InqToMqlRate: money.PercentFromHundredths(200), MqlToDealRate: money.PercentFromHundredths(1000),
				AverageDealSize: money.USD(money.FromCents(100)), Meta: pgMeta}))
		for _, n := range []string{"1", "2"} {
			err = errors.Join(err,
				directory.PutTeam(ctx, people.Team{ID: "T-100" + n, Name: "Team " + n, Colour: "blue", Meta: pgMeta}),
				plans.Put(ctx, plan.Plan{ID: "P-100" + n, Name: "Plan " + n, Year: 2026, ModelID: "MOD-1001",
					TeamIDs: []string{"T-100" + n}, DefaultSpread: plan.SpreadEvenly, GoalStage: plan.StageINQ,
					Status: plan.StatusDraft, Meta: pgMeta}),
				containers.PutCampaign(ctx, campaign.Campaign{ID: "C-100" + n, PlanID: "P-100" + n, Name: "Campaign", Meta: pgMeta}),
				containers.PutInitiative(ctx, campaign.Initiative{ID: "I-100" + n, CampaignID: "C-100" + n, Name: "Initiative", Meta: pgMeta}),
				activities.Put(ctx, activity.Activity{ID: "A-100" + n, InitiativeID: "I-100" + n, Name: "Activity " + n,
					ExternalName: "activity-" + n, ActivityTypeID: "AT-1001", Status: activity.StatusApproved, Meta: pgMeta}))
		}
		return err
	})
	pgWrite(t, db, func(ctx context.Context) error {
		err := facts.PutSource(ctx, performance.Source{ID: "PS-1", Name: "Eloqua", OwnerID: "U-2001",
			Category: performance.CategoryMarketingAutomation, Status: performance.StatusActive,
			DeliveryRoute: performance.RouteUpload, Cadence: performance.CadenceMonthly, Meta: pgMeta})
		err = errors.Join(err, facts.PutImport(ctx, performance.Import{ID: "PIM-1", SourceID: "PS-1",
			FileName: "feed.csv", Origin: performance.OriginUpload, Status: performance.ImportCompleted,
			Counts: performance.Counts{Read: 2, Matched: 2}, Meta: pgMeta}))
		for _, n := range []string{"1", "2"} {
			amount := money.USD(money.FromCents(50_000))
			err = errors.Join(err, facts.PutFact(ctx, performance.Fact{ID: "PF-" + n, SourceID: "PS-1",
				ImportID: "PIM-1", SourceReference: "r" + n, ActivityID: "A-100" + n,
				MatchedOn: performance.MatchedActivityID, MatchedKey: "a-100" + n,
				FactDate: time.Date(2026, 3, 12, 0, 0, 0, 0, time.UTC), Stage: "revenue",
				Amount: &amount, Meta: pgMeta}))
		}
		return err
	})
	return store
}

func pgWrite(t *testing.T, db *database.DB, fn func(ctx context.Context) error) {
	t.Helper()
	if err := db.Write(t.Context(), fn); err != nil {
		t.Fatalf("Write() error: %v", err)
	}
}

func TestPostgresStore_ReportTotalsOutsideTheScopeAreInvisible_FR002(t *testing.T) {
	t.Parallel()
	store := pgRevenueStore(t)

	mine, err := store.ReportTotals(t.Context(), team1Only, []string{"A-1001", "A-1002"}, 2026)
	if err != nil {
		t.Fatalf("ReportTotals: %v", err)
	}
	if len(mine) != 1 || mine[0].ActivityID != "A-1001" {
		t.Fatalf("rows = %+v, want only A-1001's fact: a record outside the scope must look missing", mine)
	}
}

func TestPostgresStore_ReportTotalsStopsAtTheRequestedActivities(t *testing.T) {
	t.Parallel()
	store := pgRevenueStore(t)

	everything := authz.Scope{Everything: true}
	found, err := store.ReportTotals(t.Context(), everything, []string{"A-1001"}, 2026)
	if err != nil {
		t.Fatalf("ReportTotals: %v", err)
	}
	if len(found) != 1 || found[0].ActivityID != "A-1001" {
		t.Fatalf("rows = %+v, want only the activity asked for, even with an unrestricted scope", found)
	}
}
