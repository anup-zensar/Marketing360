package gsaplink_test

import (
	"context"
	"errors"
	"slices"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"mapp/internal/approval"
	"mapp/internal/measure/gsaplink"
	"mapp/internal/platform/actor"
	"mapp/internal/platform/authz"
	"mapp/internal/platform/database"
	"mapp/internal/platform/database/dbtest"
	"mapp/internal/platform/date"
	"mapp/internal/platform/money"
	"mapp/internal/purchaserequest"
	"mapp/internal/stores"
	"mapp/seed"
)

var (
	seededAt = time.Date(2026, 3, 12, 9, 0, 0, 0, time.UTC)
	routedAt = time.Date(2026, 3, 13, 10, 0, 0, 0, time.UTC)
)

const routedRequest = "PR-1001"

// seededDB is the demo data with PR-1001 approved and routed to the requisitioner, then sent again with Resend PR form.
func seededDB(tb testing.TB) *database.DB {
	tb.Helper()
	return dbtest.NewFilledDB(tb, "gsap044", func(db *database.DB) {
		ctx := context.Background()
		if err := seed.Load(ctx, stores.Postgres(db), func() time.Time { return seededAt }); err != nil {
			tb.Fatalf("seed: %v", err)
		}
		requests := purchaserequest.NewPostgresStore(db)
		if err := db.Write(ctx, func(ctx context.Context) error { return approveAndRoute(ctx, requests) }); err != nil {
			tb.Fatalf("route %s: %v", routedRequest, err)
		}
	})
}

func approveAndRoute(ctx context.Context, requests *purchaserequest.PostgresStore) error {
	pr, _, err := requests.Get(ctx, routedRequest)
	if err != nil {
		return err
	}
	lines := make([]purchaserequest.VersionLine, 0, len(pr.Lines))
	for _, l := range pr.Lines {
		lines = append(lines, purchaserequest.VersionLine{ID: l.ID, LineItemID: l.LineItemID, Currency: l.Currency,
			Amount: l.Amount, DeliveryDate: l.DeliveryDate, Attributes: l.Attributes})
	}
	pr.Status = purchaserequest.StatusApproved
	pr.Approved = &purchaserequest.Version{Name: pr.Name, OwnerID: pr.OwnerID, MoAApproverID: pr.MoAApproverID,
		Attributes: pr.Attributes, Vendor: pr.Vendor, Lines: lines, ApprovedAt: routedAt.Add(-time.Hour)}
	if err = requests.Put(ctx, pr); err != nil {
		return err
	}
	return errors.Join(requests.AddRouting(ctx, routing("PRR-9001", purchaserequest.RoutingOriginal, routedAt)),
		requests.AddRouting(ctx, routing("PRR-9002", purchaserequest.RoutingResend, routedAt.Add(50*time.Hour))))
}

func routing(id string, kind purchaserequest.RoutingType, at time.Time) purchaserequest.Routing {
	return purchaserequest.Routing{ID: id, PurchaseRequestID: routedRequest, Type: kind, RoutedAt: at,
		RoutedBy: actor.Actor{ID: "U-1002", Name: "Priya Raman"}, Recipients: []string{"requisitioner@example.com"},
		Form: []purchaserequest.FormRow{{PurchaseRequestID: routedRequest}}}
}

func allIDs(tb testing.TB, db *database.DB) (activities, lineItems []string) {
	tb.Helper()
	set := stores.Postgres(db)
	acts, err := set.Activities.List(tb.Context())
	if err != nil {
		tb.Fatal(err)
	}
	items, err := set.LineItems.List(tb.Context())
	if err != nil {
		tb.Fatal(err)
	}
	for _, a := range acts {
		activities = append(activities, a.ID)
	}
	for _, li := range items {
		lineItems = append(lineItems, li.ID)
	}
	return activities, lineItems
}

func requestIDs(list []gsaplink.Request) []string {
	ids := []string{}
	for _, r := range list {
		ids = append(ids, r.ID)
	}
	return ids
}

// TestStore_ScopeIsolation asks each read for one ID at a time: nothing of another record comes back.
func TestStore_ScopeIsolation(t *testing.T) {
	t.Parallel()
	db := seededDB(t)
	store := gsaplink.NewPostgresStore(db)
	activities, lineItems := allIDs(t, db)
	t.Run("approvals", func(t *testing.T) { t.Parallel(); assertApprovalsIsolated(t, store, activities) })
	t.Run("transactions", func(t *testing.T) { t.Parallel(); assertTransactionsIsolated(t, store, lineItems) })
	t.Run("waiting", func(t *testing.T) { t.Parallel(); assertWaitingIsolated(t, store, lineItems) })
}

func assertApprovalsIsolated(tb testing.TB, store *gsaplink.PostgresStore, activities []string) {
	tb.Helper()
	approved, err := store.ActivitiesApprovedAt(tb.Context(), activities)
	if err != nil || len(approved) == 0 {
		tb.Fatalf("approved activities: %v, %v", approved, err)
	}
	for id, at := range approved {
		one, oneErr := store.ActivitiesApprovedAt(tb.Context(), []string{id})
		if diff := cmp.Diff(map[string]time.Time{id: at}, one); oneErr != nil || diff != "" {
			tb.Errorf("approval of %s only (-want +got):\n%s %v", id, diff, oneErr)
		}
	}
}

func assertTransactionsIsolated(tb testing.TB, store *gsaplink.PostgresStore, lineItems []string) {
	tb.Helper()
	all, err := store.OrderTransactions(tb.Context(), lineItems)
	if err != nil || len(all) == 0 {
		tb.Fatalf("transactions: %v, %v", all, err)
	}
	for _, id := range lineItems {
		want := slices.DeleteFunc(slices.Clone(all), func(x gsaplink.Transaction) bool { return x.LineItemID != id })
		got, oneErr := store.OrderTransactions(tb.Context(), []string{id})
		if diff := cmp.Diff(want, got, exact()); oneErr != nil || diff != "" {
			tb.Errorf("transactions of %s only (-want +got):\n%s %v", id, diff, oneErr)
		}
	}
}

func assertWaitingIsolated(tb testing.TB, store *gsaplink.PostgresStore, lineItems []string) {
	tb.Helper()
	waiting, err := store.WaitingForGsap(tb.Context(), lineItems)
	if err != nil || len(waiting) != 1 {
		tb.Fatalf("waiting: %v, %v", waiting, err)
	}
	on := waiting[0].Lines[0].LineItemID
	others := slices.DeleteFunc(slices.Clone(lineItems), func(id string) bool { return id == on })
	if none, noneErr := store.WaitingForGsap(tb.Context(), others); noneErr != nil || len(none) != 0 {
		tb.Errorf("waiting on other line items = %v, %v; want none", requestIDs(none), noneErr)
	}
}

func TestWaiting_EqualsPurchaseRequestsView_SC002(t *testing.T) {
	t.Parallel()
	db := seededDB(t)
	_, lineItems := allIDs(t, db)
	page, err := purchaserequest.NewPostgresStore(db).Page(t.Context(), authz.Scope{Everything: true},
		purchaserequest.ListQuery{WaitingForGSAP: true, Sort: "name", Page: 1, PageSize: 500})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{}
	for _, pr := range page.Items {
		want = append(want, pr.ID)
	}
	slices.Sort(want)
	got, err := gsaplink.NewPostgresStore(db).WaitingForGsap(t.Context(), lineItems)
	if err != nil {
		t.Fatal(err)
	}
	ids := requestIDs(got)
	slices.Sort(ids)
	if diff := cmp.Diff(want, ids); diff != "" || len(want) == 0 {
		t.Errorf("Waiting for GSAP differs from 032's view (-view +report):\n%s", diff)
	}
}

func TestStore_LatestApprovalAndRouting(t *testing.T) {
	t.Parallel()
	db := seededDB(t)
	store := gsaplink.NewPostgresStore(db)
	activities, lineItems := allIDs(t, db)

	got, err := store.ActivitiesApprovedAt(t.Context(), activities)
	if diff := cmp.Diff(latestApprovals(t, db), got); err != nil || diff != "" {
		t.Errorf("latest approvals (-want +got):\n%s %v", diff, err)
	}

	assertRouting(t, store, lineItems, gsaplink.Request{Routing: gsaplink.RoutingOriginal, RoutedAt: routedAt})
	modified := routedAt.Add(72 * time.Hour)
	requests := purchaserequest.NewPostgresStore(db)
	if err = db.Write(t.Context(), func(ctx context.Context) error {
		return requests.AddRouting(ctx, routing("PRR-9003", purchaserequest.RoutingModification, modified))
	}); err != nil {
		t.Fatal(err)
	}
	assertRouting(t, store, lineItems, gsaplink.Request{Routing: gsaplink.RoutingModification, RoutedAt: modified})
}

// latestApprovals works out each activity's latest approval from 013's rounds.
func latestApprovals(tb testing.TB, db *database.DB) map[string]time.Time {
	tb.Helper()
	rounds, err := approval.NewPostgresStore(db).ListRounds(tb.Context())
	if err != nil {
		tb.Fatal(err)
	}
	latest := map[string]time.Time{}
	for _, r := range rounds {
		approved := r.SubjectKind == approval.KindActivity && r.Outcome == approval.OutcomeApproved && r.EndedAt != nil
		if approved && r.EndedAt.After(latest[r.SubjectID]) {
			latest[r.SubjectID] = *r.EndedAt
		}
	}
	return latest
}

// assertRouting checks the routing GSAP has not answered: Resend PR form never restarts it (FR-009).
func assertRouting(tb testing.TB, store *gsaplink.PostgresStore, lineItems []string, want gsaplink.Request) {
	tb.Helper()
	list, err := store.WaitingForGsap(tb.Context(), lineItems)
	if err != nil || len(list) != 1 {
		tb.Fatalf("waiting: %v, %v", list, err)
	}
	if got := list[0]; got.Routing != want.Routing || !got.RoutedAt.Equal(want.RoutedAt) {
		tb.Errorf("routing = %s at %v; want %s at %v", got.Routing, got.RoutedAt, want.Routing, want.RoutedAt)
	}
}

// exact compares amounts and dates by value: their fields are unexported.
func exact() cmp.Option {
	return cmp.Options{cmp.Comparer(func(x, y money.Amount) bool { return x.Cmp(y) == 0 }),
		cmp.Comparer(func(x, y date.Date) bool { return x.Compare(y) == 0 })}
}
