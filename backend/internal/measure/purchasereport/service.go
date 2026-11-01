package purchasereport

import (
	"context"
	"slices"

	"mapp/internal/measure"
	"mapp/internal/platform/actor"
)

type Service struct {
	tx   Transactor
	deps Deps
}

func NewService(tx Transactor, deps Deps) *Service {
	return &Service{tx: tx, deps: deps}
}

// Report works the report out in one Read: the hub's scope, then one read of the purchase requests in view.
func (s *Service) Report(ctx context.Context, a actor.Actor, f measure.Filter, q Query) (View, error) {
	d, err := s.snapshot(ctx, a, f)
	if err != nil {
		return View{}, err
	}
	return d.view(q), nil
}

func (s *Service) snapshot(ctx context.Context, a actor.Actor, f measure.Filter) (data, error) {
	var d data
	err := s.tx.Read(ctx, func(ctx context.Context) error {
		scope, err := s.deps.Scope.Scope(ctx, a, f, nil)
		if err != nil {
			return err
		}
		requests, err := s.deps.Purchases.OnLineItems(ctx, scope.Matched.LineItems)
		if err != nil {
			return err
		}
		d = newData(scope, requests)
		return nil
	})
	return d, err
}

func (d data) view(q Query) View {
	rs := filtered(d.requests, q)
	return View{PurchaseSupport: d.switchedOn, Cards: d.cards(rs), StatusMix: d.statusMix(rs), ByVendor: d.chart(rs, byVendor), ByOwner: d.chart(rs, byOwner),
		Grid: d.grid(rs, d.requests, q), FinanceCovers: d.covers(rs)}
}

func (d data) cards(rs []Request) Cards {
	var committed, invoiced tally
	vendors := map[string]bool{}
	cards := Cards{PurchaseRequests: len(rs)}
	for _, r := range rs {
		vendors[r.Vendor] = true
		cards.Lines += len(r.Lines)
		for _, l := range r.Lines {
			if isCommitted(r.Status) {
				committed.add(l.Amount, d.sees(r.PlanID))
			}
			if l.Invoiced != nil {
				invoiced.add(*l.Invoiced, d.sees(r.PlanID))
			}
		}
	}
	cards.Vendors = len(vendors)
	cards.Committed, cards.Invoiced = committed.money(d.shown), invoiced.money(d.shown)
	return cards
}

// covers counts the plans that have purchase requests, and those whose amounts the person may see.
func (d data) covers(rs []Request) FinanceCovers {
	plans := []string{}
	for _, r := range rs {
		if !slices.Contains(plans, r.PlanID) {
			plans = append(plans, r.PlanID)
		}
	}
	covers := FinanceCovers{Of: len(plans)}
	for _, p := range plans {
		if d.sees(p) {
			covers.Plans++
		}
	}
	return covers
}
