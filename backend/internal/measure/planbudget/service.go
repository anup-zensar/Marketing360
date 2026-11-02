package planbudget

import (
	"context"

	"mapp/internal/allwork"
	"mapp/internal/measure"
	"mapp/internal/platform/actor"
	"mapp/internal/platform/clock"
	"mapp/internal/platform/date"
	"mapp/internal/platform/money"
)

type Deps struct {
	Scoper   Scoper
	Spending Spending
}

type Service struct {
	tx   Transactor
	deps Deps
	now  clock.Clock
}

func NewService(tx Transactor, deps Deps, now clock.Clock) *Service {
	return &Service{tx: tx, deps: deps, now: now}
}

// Report is the cards and the periods of the records the filter keeps (US1).
func (s *Service) Report(ctx context.Context, a actor.Actor, f measure.Filter) (Report, error) {
	var out Report
	err := s.tx.Read(ctx, func(ctx context.Context) error {
		r, err := s.records(ctx, a, f)
		if err != nil {
			return err
		}
		out.Cards, out.Period = r.cards(), periodOf(f)
		out.Periods, out.Notes = r.periods(out.Period, date.Today(s.now()))
		return nil
	})
	return out, err
}

// records reads the scope once and the spend of the line items whose finance the person may see (research R1, R2).
func (s *Service) records(ctx context.Context, a actor.Actor, f measure.Filter) (*records, error) {
	sc, err := s.deps.Scoper.Scope(ctx, a, f, nil)
	if err != nil {
		return nil, err
	}
	r := newRecords(sc)
	r.display = money.DisplayIn(sc.Currency.Code, sc.Currency.Rate.String())
	r.planned = allwork.PlannedByActivity(sc.Nodes, sc.Year)
	if ids := r.financedLineItems(); len(ids) > 0 {
		if r.spend, err = s.deps.Spending.SpendByMonth(ctx, a, ids); err != nil {
			return nil, err
		}
	}
	return r, nil
}

// Ranked is one page of the ranked list for the Show level, default campaigns (US2, US3).
func (s *Service) Ranked(ctx context.Context, a actor.Actor, f measure.Filter, q RankedQuery) (Ranked, error) {
	var out Ranked
	err := s.tx.Read(ctx, func(ctx context.Context) error {
		r, err := s.records(ctx, a, f)
		if err != nil {
			return err
		}
		out = r.ranked(levelOf(f), q)
		return nil
	})
	return out, err
}

func levelOf(f measure.Filter) string {
	if f.Show == "" {
		return measure.ShowCampaign
	}
	return f.Show
}

func periodOf(f measure.Filter) string {
	if f.Period == measure.PeriodQuarter {
		return measure.PeriodQuarter
	}
	return measure.PeriodMonth
}
