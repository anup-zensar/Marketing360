package measure

import (
	"context"
	"slices"

	"mapp/internal/allwork"
	"mapp/internal/platform/actor"
	"mapp/internal/platform/date"
)

const msgNoReport = "There is no such report."

// Frame works out what every report page shares for the report and the filter, in one Read (FR-006 – FR-019).
func (s *Service) Frame(ctx context.Context, a actor.Actor, report string, f Filter) (Frame, error) {
	if err := openable(ctx, report); err != nil {
		return Frame{}, err
	}
	var frame Frame
	err := s.tx.Read(ctx, func(ctx context.Context) error {
		sc, err := s.Scope(ctx, a, f, nil)
		if err != nil {
			return err
		}
		years, err := s.deps.Plans.YearsInReportsScope(ctx)
		if err != nil {
			return err
		}
		currencies, err := s.deps.Currencies.Active(ctx)
		if err != nil {
			return err
		}
		display, err := s.deps.Planning.Currency(ctx, a.ID)
		frame = Frame{Report: report, Filter: applied(f, sc), Choices: choicesOf(sc, f, withYear(years, sc.Year, date.Today(s.now()).Year()), currencies),
			InView: inViewOf(sc.Matched), Finance: financeOf(sc), LeftOut: len(sc.Matched.LeftOut),
			DisplayCurrency: display.Code, GeneratedAt: s.now().UTC(), FinanceAsOf: sc.FinanceAsOf}
		return err
	})
	if err != nil {
		return Frame{}, err
	}
	return frame, nil
}

// applied is the filter as the report uses it: this year by default, left-out values dropped, the resolved currency.
func applied(f Filter, sc Scope) AppliedFilter {
	q := sc.Query
	a := AppliedFilter{Year: sc.Year, Plans: orEmpty(q.Plans), PlanTypes: orEmpty(q.PlanTypes),
		Teams: orEmpty(q.Teams), Owners: orEmpty(q.Owners),
		ActivityTypes: orEmpty(q.ActivityTypes), ActivityStatuses: orEmpty(q.ActivityStatuses),
		Attributes: q.Attributes, Period: f.Period, Currency: sc.Currency.Code, Tables: orEmpty(f.Tables)}
	if a.Attributes == nil {
		a.Attributes = []allwork.AttrFilter{}
	}
	if f.Show != "" {
		show := f.Show
		a.Show = &show
	}
	return a
}

func orEmpty(values []string) []string {
	if values == nil {
		return []string{}
	}
	return values
}

// withYear is the years with plans in scope, the year shown and the current year, in order (FR-008, FR-011).
func withYear(years []int, shown, current int) []int {
	all := append(slices.Clone(years), shown, current)
	slices.Sort(all)
	return slices.Compact(all)
}

func inViewOf(m allwork.Matched) InView {
	return InView{Plans: len(m.Plans), Campaigns: len(m.Campaigns), Initiatives: len(m.Initiatives),
		Activities: len(m.Activities), LineItems: len(m.LineItems)}
}

// financeOf counts the plans in view whose financial data the person may see (FR-004, D-301).
func financeOf(sc Scope) FinanceCoverage {
	coverage := FinanceCoverage{Of: len(sc.Matched.Plans)}
	for _, id := range sc.Matched.Plans {
		if sc.Choices.SeesFinance && (sc.Choices.FinancePlans == nil || slices.Contains(sc.Choices.FinancePlans, id)) {
			coverage.Plans++
		}
	}
	return coverage
}
