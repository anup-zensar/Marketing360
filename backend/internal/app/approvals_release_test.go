package app

import (
	"context"
	"testing"
	"time"

	"mapp/internal/activity"
	"mapp/internal/platform/actor"
	"mapp/internal/platform/fake"
	"mapp/internal/platform/money"
	"mapp/internal/stores"
	"mapp/internal/stores/storestest"
	"mapp/seed"
)

// The release hook is one line of wiring; this proves an approval reaches it and a decline does not
// (046 FR-020, research R4).
func TestSetStatus_ApprovedReleasesHeldFacts_046FR020(t *testing.T) {
	t.Parallel()
	db := fake.NewDB()
	set := storestest.New(db)
	alex := actor.Actor{ID: "U-1001", Name: "Alex Morgan"}
	noon := time.Date(2026, 3, 12, 12, 0, 0, 0, time.UTC)
	if err := seed.Load(t.Context(), set, func() time.Time { return noon }); err != nil {
		t.Fatalf("seed: %v", err)
	}
	submitted := seededActivity(t, set, activity.StatusSubmitted)
	released := []string{}
	adapter := &approvalsActivities{
		activities: set.Activities, lineItems: set.LineItems, campaigns: set.Campaigns,
		plans: set.Plans, settings: set.Settings,
		rates: func(context.Context) (money.Rates, error) { return money.Rates{}, nil },
		now:   func() time.Time { return noon },
		approved: func(_ context.Context, _ actor.Actor, activityID string) error {
			released = append(released, activityID)
			return nil
		},
	}

	if _, err := adapter.SetStatus(t.Context(), alex, submitted, string(activity.StatusApproved)); err != nil {
		t.Fatalf("SetStatus(approved): %v", err)
	}
	if _, err := adapter.SetStatus(t.Context(), alex, submitted, string(activity.StatusDeclined)); err != nil {
		t.Fatalf("SetStatus(declined): %v", err)
	}

	if len(released) != 1 || released[0] != submitted {
		t.Fatalf("released = %v, want exactly the approval to release %s's held facts", released, submitted)
	}
}

func seededActivity(t *testing.T, set stores.Set, status activity.Status) string {
	t.Helper()
	all, err := set.Activities.List(t.Context())
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, act := range all {
		if act.Status == status {
			return act.ID
		}
	}
	t.Fatal("the demo data holds no activity in the wanted status")
	return ""
}
