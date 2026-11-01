package app

import (
	"context"
	"testing"
	"time"

	"mapp/internal/people"
	"mapp/internal/platform/actor"
	"mapp/internal/platform/fake"
	"mapp/internal/platform/money"
	"mapp/internal/purchaserequest"
	"mapp/internal/stores"
	"mapp/internal/stores/storestest"
)

// TestWaiting_ResendDoesNotRestart_FR009 and the Modification case of FR-008, on the in-memory reads: the wait
// counts from the latest original or modification routing; Resend PR form never restarts it.
func TestWaiting_ResendDoesNotRestart_FR009(t *testing.T) {
	t.Parallel()
	friday := time.Date(2026, 3, 13, 16, 0, 0, 0, time.UTC)
	tests := []struct {
		name     string
		routings []purchaserequest.RoutingType
		want     purchaserequest.RoutingType
		wantAt   time.Time
	}{
		{name: "Resend PR form does not restart the wait", want: purchaserequest.RoutingOriginal, wantAt: friday,
			routings: []purchaserequest.RoutingType{purchaserequest.RoutingOriginal, purchaserequest.RoutingResend}},
		{name: "an approved top-up waits from its modification", want: purchaserequest.RoutingModification,
			wantAt: friday.Add(48 * time.Hour), routings: []purchaserequest.RoutingType{purchaserequest.RoutingOriginal,
				purchaserequest.RoutingResend, purchaserequest.RoutingModification}},
	}
	for _, tt := range tests {
		set := storestest.New(fake.NewDB())
		if err := set.Tx.Write(t.Context(), func(ctx context.Context) error {
			return routed(ctx, set, tt.routings, friday)
		}); err != nil {
			t.Fatal(err)
		}
		list, err := gsapRecords{set: set}.WaitingForGsap(t.Context(), []string{"L-1006"})
		if err != nil || len(list) != 1 {
			t.Fatalf("%s: waiting = %v, %v", tt.name, list, err)
		}
		if got := list[0]; got.Routing != string(tt.want) || !got.RoutedAt.Equal(tt.wantAt) || got.Owner != "Priya Raman" {
			t.Errorf("%s: %s at %v by %s; want %s at %v", tt.name, got.Routing, got.RoutedAt, got.Owner, tt.want, tt.wantAt)
		}
	}
}

// routed saves an approved purchase request on L-1006 and its routings, a day apart from at.
func routed(ctx context.Context, set stores.Set, kinds []purchaserequest.RoutingType, at time.Time) error {
	priya := people.Person{ID: "U-1002", Name: "Priya Raman", Email: "priya.raman@shell.com"}
	if err := set.People.Put(ctx, priya); err != nil {
		return err
	}
	amount, err := money.ParseAmount("100.00")
	if err != nil {
		return err
	}
	pr := purchaserequest.PurchaseRequest{ID: "PR-1001", PlanID: "P-1001", Name: "PR for Stand build", OwnerID: "U-1002",
		Status: purchaserequest.StatusApproved, Vendor: "Brightside Media BV",
		Lines: []purchaserequest.Line{{ID: "PRL-1001", PurchaseRequestID: "PR-1001", Position: 1, LineItemID: "L-1006",
			Currency: "EUR", Amount: amount}}}
	if err = set.PurchaseRequests.Put(ctx, pr); err != nil {
		return err
	}
	for i, kind := range kinds {
		err = set.PurchaseRequests.AddRouting(ctx, purchaserequest.Routing{ID: "PRR-" + string(rune('1'+i)),
			PurchaseRequestID: pr.ID, Type: kind, RoutedAt: at.Add(time.Duration(i) * 24 * time.Hour),
			RoutedBy: actor.Actor{ID: "U-1002"}})
		if err != nil {
			return err
		}
	}
	return nil
}
