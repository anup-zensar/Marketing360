// Package gsaplinktest is the in-memory fake of 044's reads, over the other areas' fake tables, for tests that
// build the whole API without a database. It answers as measure_gsaplink.sql does.
package gsaplinktest

import (
	"cmp"
	"context"
	"slices"
	"strings"
	"time"

	"mapp/internal/approval"
	"mapp/internal/financials"
	"mapp/internal/measure/gsaplink"
	"mapp/internal/people"
	"mapp/internal/platform/fake"
	"mapp/internal/platform/money"
	"mapp/internal/purchaserequest"
)

type Store struct {
	rounds       *fake.Table[approval.Round]
	requests     *fake.Table[purchaserequest.PurchaseRequest]
	lines        *fake.Table[purchaserequest.Line]
	routings     *fake.Table[purchaserequest.Routing]
	transactions *fake.Table[financials.Transaction]
	people       *fake.Table[people.Person]
}

func NewStore(db *fake.DB) Store {
	return Store{
		rounds:   fake.TableOf(db, "approval_rounds", func(r approval.Round) string { return r.ID }),
		requests: fake.TableOf(db, "purchase_requests", func(pr purchaserequest.PurchaseRequest) string { return pr.ID }),
		lines:    fake.TableOf(db, "purchase_request_lines", func(l purchaserequest.Line) string { return l.ID }),
		routings: fake.TableOf(db, "purchase_request_routings", func(r purchaserequest.Routing) string { return r.ID }),
		transactions: fake.TableOf(db, "financial_transactions",
			func(t financials.Transaction) string { return t.ID }),
		people: fake.TableOf(db, "people", func(p people.Person) string { return p.ID }),
	}
}

func (s Store) ActivitiesApprovedAt(_ context.Context, activityIDs []string) (map[string]time.Time, error) {
	rounds, err := s.rounds.Where(func(r approval.Round) bool {
		return r.SubjectKind == approval.KindActivity && r.Outcome == approval.OutcomeApproved && r.EndedAt != nil &&
			slices.Contains(activityIDs, r.SubjectID)
	})
	latest := map[string]time.Time{}
	for _, r := range rounds {
		if r.EndedAt.After(latest[r.SubjectID]) {
			latest[r.SubjectID] = *r.EndedAt
		}
	}
	return latest, err
}

func (s Store) WaitingForGsap(_ context.Context, lineItemIDs []string) ([]gsaplink.Request, error) {
	all, err := s.requests.Where(func(pr purchaserequest.PurchaseRequest) bool {
		return pr.Status == purchaserequest.StatusApproved
	})
	if err != nil {
		return nil, err
	}
	slices.SortFunc(all, func(a, b purchaserequest.PurchaseRequest) int { return compareIDs(a.ID, b.ID) })
	list := []gsaplink.Request{}
	for _, pr := range all {
		request, waiting, err := s.request(pr, lineItemIDs)
		if err != nil {
			return nil, err
		}
		if waiting {
			list = append(list, request)
		}
	}
	return list, nil
}

// request is pr with its lines and routings when it waits for GSAP on one of the line items (032 FR-032).
func (s Store) request(pr purchaserequest.PurchaseRequest, lineItemIDs []string) (gsaplink.Request, bool, error) {
	lines, err := s.linesOf(pr.ID)
	if err != nil {
		return gsaplink.Request{}, false, err
	}
	routings, err := s.routingsOf(pr.ID)
	if err != nil || len(routings) == 0 || !slices.ContainsFunc(lines, unmatched) ||
		!slices.ContainsFunc(lines, func(l purchaserequest.Line) bool { return slices.Contains(lineItemIDs, l.LineItemID) }) {
		return gsaplink.Request{}, false, err
	}
	owner, _, err := s.people.Get(pr.OwnerID)
	request := gsaplink.Request{ID: pr.ID, Name: pr.Name, Vendor: pr.Vendor, Owner: owner.Name, PlanID: pr.PlanID,
		GsapPrNumber: pr.GSAP.PRNumber, Routings: routings}
	for _, l := range lines {
		request.Lines = append(request.Lines, gsaplink.RequestLine{LineItemID: l.LineItemID,
			Amount: money.Money{Amount: l.Amount, Currency: l.Currency}})
	}
	return request, true, err
}

// unmatched is a line GSAP has not answered with a purchase order line of its value and currency.
func unmatched(l purchaserequest.Line) bool {
	return l.PO.Line == "" || l.PO.Value == nil || l.PO.Value.Currency != l.Currency || l.PO.Value.Amount.Cmp(l.Amount) != 0
}

func (s Store) linesOf(prID string) ([]purchaserequest.Line, error) {
	lines, err := s.lines.Where(func(l purchaserequest.Line) bool { return l.PurchaseRequestID == prID })
	slices.SortFunc(lines, func(a, b purchaserequest.Line) int { return cmp.Compare(a.Position, b.Position) })
	return lines, err
}

func (s Store) routingsOf(prID string) ([]gsaplink.Routing, error) {
	found, err := s.routings.Where(func(r purchaserequest.Routing) bool { return r.PurchaseRequestID == prID })
	slices.SortFunc(found, func(a, b purchaserequest.Routing) int {
		return cmp.Or(a.RoutedAt.Compare(b.RoutedAt), compareIDs(a.ID, b.ID))
	})
	list := make([]gsaplink.Routing, 0, len(found))
	for _, r := range found {
		list = append(list, gsaplink.Routing{Type: string(r.Type), At: r.RoutedAt})
	}
	return list, err
}

func (s Store) OrderTransactions(_ context.Context, lineItemIDs []string) ([]gsaplink.Transaction, error) {
	kinds := []financials.Type{financials.TypePurchaseOrder, financials.TypeInvoice, financials.TypePayment}
	found, err := s.transactions.Where(func(t financials.Transaction) bool {
		return t.PONumber != "" && slices.Contains(kinds, t.Type) && slices.Contains(lineItemIDs, t.LineItemID)
	})
	if err != nil {
		return nil, err
	}
	slices.SortFunc(found, func(a, b financials.Transaction) int { return compareIDs(a.ID, b.ID) })
	list := make([]gsaplink.Transaction, 0, len(found))
	for _, t := range found {
		pr, gsapPR, err := s.raisedFrom(t)
		if err != nil {
			return nil, err
		}
		list = append(list, gsaplink.Transaction{ID: t.ID, Type: string(t.Type), LineItemID: t.LineItemID,
			PoNumber: t.PONumber, LineNumber: t.LineNumber, PostingDate: t.PostingDate, Vendor: t.VendorName,
			Amount: money.Money{Amount: t.Amount, Currency: t.Currency}, PurchaseRequestID: pr, GsapPrNumber: gsapPR})
	}
	return list, nil
}

// raisedFrom is the first purchase request with a line of the transaction's PO number on its line item.
func (s Store) raisedFrom(t financials.Transaction) (id, gsapPR string, err error) {
	lines, err := s.lines.Where(func(l purchaserequest.Line) bool {
		return l.PO.Number == t.PONumber && l.LineItemID == t.LineItemID
	})
	if err != nil || len(lines) == 0 {
		return "", "", err
	}
	ids := make([]string, 0, len(lines))
	for _, l := range lines {
		ids = append(ids, l.PurchaseRequestID)
	}
	first := slices.MinFunc(ids, compareIDs)
	pr, _, err := s.requests.Get(first)
	return first, pr.GSAP.PRNumber, err
}

// compareIDs puts PR-999 before PR-1000, as the queries order by length, then ID.
func compareIDs(a, b string) int {
	return cmp.Or(cmp.Compare(len(a), len(b)), strings.Compare(a, b))
}
