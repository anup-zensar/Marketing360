package gsaplink

import (
	"cmp"
	"context"
	"time"

	"mapp/internal/measure"
	"mapp/internal/platform/actor"
)

type Deps struct {
	Scoped Scoped
	Store  Store
	Rates  Rates
}

// Service works out the GSAP link report: read-only, nothing on it changes a record (FR-002).
type Service struct {
	tx   Transactor
	deps Deps
	now  func() time.Time
}

func NewService(tx Transactor, deps Deps, now func() time.Time) *Service {
	return &Service{tx: tx, deps: deps, now: now}
}

// Report works out one tab of the records matching f, all its rows sorted, then one page of them, in one Read (R4).
func (s *Service) Report(ctx context.Context, a actor.Actor, f measure.Filter, t Table) (Report, error) {
	t = withDefaults(t)
	r := Report{Tab: t.Tab, Page: t.Page, PageSize: t.PageSize, Sort: t.Sort}
	err := s.tx.Read(ctx, func(ctx context.Context) error {
		y, err := s.year(ctx, a, f)
		if err != nil {
			return err
		}
		r.Currency = y.scope.Currency.Code
		return s.fill(ctx, y, t, &r)
	})
	if err != nil {
		return Report{}, err
	}
	return r, nil
}

func (s *Service) fill(ctx context.Context, y year, t Table, r *Report) error {
	switch t.Tab {
	case TabWaiting:
		tab, err := s.waiting(ctx, y, t.Sort)
		r.Total, r.Finance = len(tab.Rows), y.coverage()
		tab.Rows, r.Waiting = pageOf(tab.Rows, t.Page, t.PageSize), &tab
		return err
	case TabOrders:
		tab, err := s.orders(ctx, y, t.Sort)
		r.Total, r.Finance = len(tab.Rows), y.coverage()
		tab.Rows, r.Orders = pageOf(tab.Rows, t.Page, t.PageSize), &tab
		return err
	}
	tab, err := s.swbs(ctx, y, t.Sort)
	r.Total = len(tab.Rows)
	tab.Rows, r.Swbs = pageOf(tab.Rows, t.Page, t.PageSize), &tab
	return err
}

func (s *Service) year(ctx context.Context, a actor.Actor, f measure.Filter) (year, error) {
	sc, err := s.deps.Scoped.Scope(ctx, a, f, nil)
	return newYear(sc, s.now()), err
}

func (s *Service) swbs(ctx context.Context, y year, sort string) (SwbsTab, error) {
	ready := y.readyToSend()
	approved, err := s.deps.Store.ActivitiesApprovedAt(ctx, activitiesOf(ready))
	if err != nil {
		return SwbsTab{}, err
	}
	tab := y.swbsTab(ready, approved)
	sortRows(tab.Rows, swbsColumns(), sort, swbsID)
	return tab, nil
}

func (s *Service) waiting(ctx context.Context, y year, sort string) (WaitingTab, error) {
	requests, err := s.deps.Store.WaitingForGsap(ctx, y.scope.Matched.LineItems)
	if err != nil {
		return WaitingTab{}, err
	}
	rates, err := s.deps.Rates.Rates(ctx)
	if err != nil {
		return WaitingTab{}, err
	}
	tab := y.waitingTab(requests, rates)
	sortRows(tab.Rows, waitingColumns(), sort, waitingID)
	return tab, nil
}

// orders reads only the line items of plans whose financial data the person may see (044-C4).
func (s *Service) orders(ctx context.Context, y year, sort string) (OrdersTab, error) {
	var transactions []Transaction
	if ids := y.financeLineItems(); len(ids) > 0 {
		var err error
		if transactions, err = s.deps.Store.OrderTransactions(ctx, ids); err != nil {
			return OrdersTab{}, err
		}
	}
	rates, err := s.deps.Rates.Rates(ctx)
	if err != nil {
		return OrdersTab{}, err
	}
	tab := y.ordersTab(transactions, rates)
	sortRows(tab.Rows, orderColumns(), sort, orderID)
	return tab, nil
}

func withDefaults(t Table) Table {
	t.Tab = cmp.Or(t.Tab, TabSwbs)
	if !ValidSort(t.Tab, t.Sort) {
		t.Sort = DefaultSort(t.Tab)
	}
	t.Page, t.PageSize = max(t.Page, 1), cmp.Or(t.PageSize, DefaultPageSize)
	return t
}
