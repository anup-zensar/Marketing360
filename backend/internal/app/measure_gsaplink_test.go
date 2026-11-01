package app

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"mapp/internal/measure/gsaplink"
	"mapp/internal/platform/actor"
	"mapp/internal/platform/database"
	"mapp/internal/platform/database/dbtest"
	"mapp/internal/platform/date"
	"mapp/internal/platform/money"
	"mapp/internal/purchaserequest"
	"mapp/internal/stores"
	"mapp/seed"
)

// TestGsapRecords_EqualPostgresStore pins the in-memory reads to measure_gsaplink.sql on the demo data, with
// PR-1001 approved and routed, then sent again with Resend PR form.
func TestGsapRecords_EqualPostgresStore(t *testing.T) {
	t.Parallel()
	seededAt := time.Date(2026, 3, 12, 9, 0, 0, 0, time.UTC)
	db := dbtest.NewFilledDB(t, "gsap044app", func(db *database.DB) {
		ctx := context.Background()
		if err := seed.Load(ctx, stores.Postgres(db), func() time.Time { return seededAt }); err != nil {
			t.Fatalf("seed: %v", err)
		}
		if err := db.Write(ctx, func(ctx context.Context) error { return routePR(ctx, stores.Postgres(db), seededAt) }); err != nil {
			t.Fatalf("route: %v", err)
		}
	})
	set := stores.Postgres(db)
	activities, lineItems := idsOf(t, set)
	both := [2]gsaplink.Store{gsaplink.NewPostgresStore(db), gsapRecords{set: set}}
	ctx := t.Context()

	compare(t, "approvals", func(s gsaplink.Store) (any, error) { return s.ActivitiesApprovedAt(ctx, activities) }, both)
	compare(t, "waiting", func(s gsaplink.Store) (any, error) { return s.WaitingForGsap(ctx, lineItems) }, both)
	compare(t, "orders", func(s gsaplink.Store) (any, error) { return s.OrderTransactions(ctx, lineItems) }, both)
}

// compare reads with both stores, PostgreSQL first.
func compare(t *testing.T, name string, read func(gsaplink.Store) (any, error), both [2]gsaplink.Store) {
	t.Helper()
	want, err := read(both[0])
	if err != nil {
		t.Fatalf("%s on PostgreSQL: %v", name, err)
	}
	got, err := read(both[1])
	if diff := cmp.Diff(want, got, exactValues()); err != nil || diff != "" {
		t.Errorf("%s (-sql +records):\n%s %v", name, diff, err)
	}
}

func routePR(ctx context.Context, set stores.Set, at time.Time) error {
	pr, _, err := set.PurchaseRequests.Get(ctx, "PR-1001")
	if err != nil {
		return err
	}
	pr.Status = purchaserequest.StatusApproved
	pr.Approved = &purchaserequest.Version{Name: pr.Name, OwnerID: pr.OwnerID, Vendor: pr.Vendor, ApprovedAt: at}
	if err = set.PurchaseRequests.Put(ctx, pr); err != nil {
		return err
	}
	for i, kind := range []purchaserequest.RoutingType{purchaserequest.RoutingOriginal, purchaserequest.RoutingResend} {
		r := purchaserequest.Routing{ID: []string{"PRR-9001", "PRR-9002"}[i], PurchaseRequestID: pr.ID, Type: kind,
			RoutedAt: at.Add(time.Duration(i+1) * time.Hour), RoutedBy: actor.Actor{ID: "U-1002", Name: "Priya Raman"},
			Recipients: []string{"requisitioner@example.com"}, Form: []purchaserequest.FormRow{{PurchaseRequestID: pr.ID}}}
		if err = set.PurchaseRequests.AddRouting(ctx, r); err != nil {
			return err
		}
	}
	return nil
}

func idsOf(t *testing.T, set stores.Set) (activities, lineItems []string) {
	t.Helper()
	acts, err := set.Activities.List(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	items, err := set.LineItems.List(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	for _, a := range acts {
		activities = append(activities, a.ID)
	}
	for _, li := range items {
		lineItems = append(lineItems, li.ID)
	}
	return activities, lineItems
}

// exactValues compares amounts and dates by value: their fields are unexported.
func exactValues() cmp.Option {
	return cmp.Options{cmp.Comparer(func(x, y money.Amount) bool { return x.Cmp(y) == 0 }),
		cmp.Comparer(func(x, y date.Date) bool { return x.Compare(y) == 0 })}
}
