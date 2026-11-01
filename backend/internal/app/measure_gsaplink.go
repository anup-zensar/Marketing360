package app

import (
	"cmp"
	"context"
	"slices"
	"strings"
	"time"

	"mapp/internal/approval"
	"mapp/internal/financials"
	"mapp/internal/measure"
	"mapp/internal/measure/gsaplink"
	"mapp/internal/platform/database"
	"mapp/internal/platform/money"
	"mapp/internal/purchaserequest"
	"mapp/internal/stores"
)

// registerGsapLink registers and mounts the GSAP link report (feature 044) on 039's shared Scope step.
func registerGsapLink(w *wiring, r *measure.Registry) error {
	service := gsaplink.NewService(w.set.Tx, gsaplink.Deps{Scoped: w.reports, Store: gsapStore(w.set), Rates: w.settings},
		w.now)
	mount(w.signedIn, gsaplink.NewHandler(service))
	return r.Register(service)
}

// gsapStore reads with 044's queries on PostgreSQL; in memory (API tests) through the other areas' stores.
func gsapStore(set stores.Set) gsaplink.Store {
	if db, ok := set.Tx.(*database.DB); ok {
		return gsaplink.NewPostgresStore(db)
	}
	return gsapRecords{set: set}
}

// gsapRecords answers gsaplink.Store as measure_gsaplink.sql does, from the stores' own reads.
type gsapRecords struct {
	set stores.Set
}

func (g gsapRecords) ActivitiesApprovedAt(ctx context.Context, activityIDs []string) (map[string]time.Time, error) {
	rounds, err := g.set.Approvals.RoundsOfSubjects(ctx, activityIDs)
	latest := map[string]time.Time{}
	for _, r := range rounds {
		approved := r.SubjectKind == approval.KindActivity && r.Outcome == approval.OutcomeApproved && r.EndedAt != nil
		if approved && r.EndedAt.After(latest[r.SubjectID]) {
			latest[r.SubjectID] = *r.EndedAt
		}
	}
	return latest, err
}

func (g gsapRecords) WaitingForGsap(ctx context.Context, lineItemIDs []string) ([]gsaplink.Request, error) {
	all, err := g.set.PurchaseRequests.All(ctx)
	if err != nil {
		return nil, err
	}
	slices.SortFunc(all, func(a, b purchaserequest.PurchaseRequest) int { return compareIDs(a.ID, b.ID) })
	list := []gsaplink.Request{}
	for _, pr := range all {
		if !waitingForGsap(pr, lineItemIDs) {
			continue
		}
		request, found, err := g.request(ctx, pr)
		if err != nil {
			return nil, err
		}
		if found {
			list = append(list, request)
		}
	}
	return list, nil
}

// waitingForGsap is 032 FR-032 on lines of the given line items: approved, with a line GSAP has not matched.
func waitingForGsap(pr purchaserequest.PurchaseRequest, lineItemIDs []string) bool {
	unmatched := slices.ContainsFunc(pr.Lines, func(l purchaserequest.Line) bool {
		return l.PO.Line == "" || l.PO.Value == nil || l.PO.Value.Currency != l.Currency || l.PO.Value.Amount.Cmp(l.Amount) != 0
	})
	inView := slices.ContainsFunc(pr.Lines, func(l purchaserequest.Line) bool { return slices.Contains(lineItemIDs, l.LineItemID) })
	return pr.Status == purchaserequest.StatusApproved && unmatched && inView
}

// request is the purchase request with the latest original or modification routing; Resend PR form never counts.
func (g gsapRecords) request(ctx context.Context, pr purchaserequest.PurchaseRequest) (gsaplink.Request, bool, error) {
	routings, err := g.set.PurchaseRequests.Routings(ctx, pr.ID)
	if err != nil {
		return gsaplink.Request{}, false, err
	}
	routings = slices.DeleteFunc(routings, func(r purchaserequest.Routing) bool {
		return r.Type != purchaserequest.RoutingOriginal && r.Type != purchaserequest.RoutingModification
	})
	if len(routings) == 0 {
		return gsaplink.Request{}, false, nil
	}
	latest := slices.MaxFunc(routings, func(a, b purchaserequest.Routing) int {
		return cmp.Or(a.RoutedAt.Compare(b.RoutedAt), compareIDs(a.ID, b.ID))
	})
	owner, _, err := g.set.People.Get(ctx, pr.OwnerID)
	request := gsaplink.Request{ID: pr.ID, Name: pr.Name, Vendor: pr.Vendor, Owner: owner.Name, PlanID: pr.PlanID,
		GsapPrNumber: pr.GSAP.PRNumber, Routing: string(latest.Type), RoutedAt: latest.RoutedAt}
	lines := slices.Clone(pr.Lines)
	slices.SortFunc(lines, func(a, b purchaserequest.Line) int { return cmp.Compare(a.Position, b.Position) })
	for _, l := range lines {
		request.Lines = append(request.Lines, gsaplink.RequestLine{LineItemID: l.LineItemID,
			Amount: money.Money{Amount: l.Amount, Currency: l.Currency}})
	}
	return request, true, err
}

func (g gsapRecords) OrderTransactions(ctx context.Context, lineItemIDs []string) ([]gsaplink.Transaction, error) {
	all, err := g.set.Financials.ListTransactions(ctx)
	if err != nil {
		return nil, err
	}
	requests, err := g.set.PurchaseRequests.All(ctx)
	if err != nil {
		return nil, err
	}
	slices.SortFunc(requests, func(a, b purchaserequest.PurchaseRequest) int { return compareIDs(a.ID, b.ID) })
	slices.SortFunc(all, func(a, b financials.Transaction) int { return compareIDs(a.ID, b.ID) })
	list := []gsaplink.Transaction{}
	for _, t := range all {
		if !onOrder(t) || !slices.Contains(lineItemIDs, t.LineItemID) {
			continue
		}
		pr, gsapPR := raisedFrom(requests, t)
		list = append(list, gsaplink.Transaction{ID: t.ID, Type: string(t.Type), LineItemID: t.LineItemID,
			PoNumber: t.PONumber, LineNumber: t.LineNumber, PostingDate: t.PostingDate, Vendor: t.VendorName,
			Amount: money.Money{Amount: t.Amount, Currency: t.Currency}, PurchaseRequestID: pr, GsapPrNumber: gsapPR})
	}
	return list, nil
}

func onOrder(t financials.Transaction) bool {
	kinds := []financials.Type{financials.TypePurchaseOrder, financials.TypeInvoice, financials.TypePayment}
	return t.PONumber != "" && slices.Contains(kinds, t.Type)
}

// raisedFrom is the first purchase request with a line of the transaction's PO number on its line item.
func raisedFrom(requests []purchaserequest.PurchaseRequest, t financials.Transaction) (id, gsapPR string) {
	for _, pr := range requests {
		if slices.ContainsFunc(pr.Lines, func(l purchaserequest.Line) bool {
			return l.PO.Number == t.PONumber && l.LineItemID == t.LineItemID
		}) {
			return pr.ID, pr.GSAP.PRNumber
		}
	}
	return "", ""
}

// compareIDs puts PR-999 before PR-1000, as the queries order by length, then ID.
func compareIDs(a, b string) int {
	return cmp.Or(cmp.Compare(len(a), len(b)), strings.Compare(a, b))
}
