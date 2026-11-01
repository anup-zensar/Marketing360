package app

import (
	"cmp"
	"context"
	"slices"

	"mapp/internal/allwork"
	"mapp/internal/measure"
	"mapp/internal/measure/plandetails"
	"mapp/internal/people"
	"mapp/internal/platform/actor"
	"mapp/internal/platform/authz"
	"mapp/internal/platform/clock"
	"mapp/internal/platform/date"
	"mapp/internal/settings"
	"mapp/internal/stores"
)

// registerPlanDetails registers and mounts the Plan details report (feature 043).
func registerPlanDetails(w *wiring, r *measure.Registry) error {
	service := plandetails.NewService(w.set.Tx, reportScope{source: w.allWork, plans: w.set.Plans, people: w.people,
		settings: w.settings, now: w.now})
	mount(w.signedIn, plandetails.NewHandler(service))
	return r.Register(service)
}

// reportScope is the report frame's step (039 research R1, R2) as 043 composes it: All work's year narrowed to the
// plans of the reports.view scope, matched by allwork.Match, and the filter's currency at its current rate.
type reportScope struct {
	source   allWorkSource
	plans    stores.PlanStore
	people   *people.Service
	settings *settings.Service
	now      clock.Clock
}

func (s reportScope) Matched(ctx context.Context, a actor.Actor, f measure.Filter, columns []string) (
	plandetails.Matching, error,
) {
	today := date.Today(s.now())
	year := cmp.Or(f.Query.Year, today.Year())
	data, err := s.source.Year(ctx, year, today, columns)
	if err != nil {
		return plandetails.Matching{}, err
	}
	nodes, err := s.inReportsScope(ctx, data.Nodes)
	if err != nil {
		return plandetails.Matching{}, err
	}
	choices, err := s.source.Choices(ctx, a.ID)
	if err != nil {
		return plandetails.Matching{}, err
	}
	attributes, err := s.source.Attributes(ctx)
	if err != nil {
		return plandetails.Matching{}, err
	}
	currency, err := s.currency(ctx, a, f.Currency)
	m := allwork.Match(nodes, f.Query, attributes, choices)
	matched := map[string]bool{}
	for _, id := range slices.Concat(m.Plans, m.Campaigns, m.Initiatives, m.Activities) {
		matched[id] = true
	}
	return plandetails.Matching{Year: year, Nodes: nodes, Matched: matched, ActivityTypes: choices.ActivityTypes,
		Currency: currency}, err
}

func (s reportScope) Attributes(ctx context.Context) ([]allwork.Attribute, error) {
	return s.source.Attributes(ctx)
}

// inReportsScope keeps the nodes of the plans the person may see reports of (039 FR-003).
func (s reportScope) inReportsScope(ctx context.Context, nodes []allwork.Node) ([]allwork.Node, error) {
	list, err := s.plans.ListInScope(ctx, authz.PermissionScopeOf(ctx, authz.ReportsView))
	if err != nil {
		return nil, err
	}
	ids := map[string]bool{}
	for _, p := range list {
		ids[p.ID] = true
	}
	return slices.DeleteFunc(slices.Clone(nodes), func(n allwork.Node) bool { return !ids[n.PlanID] }), nil
}

// currency is the chosen active currency with its current rate, or the person's display currency (039 FR-018).
func (s reportScope) currency(ctx context.Context, a actor.Actor, code string) (allwork.Currency, error) {
	if code != "" {
		options, err := s.people.DisplayCurrencies(ctx, a)
		if err != nil {
			return allwork.Currency{}, err
		}
		rates, err := s.settings.Rates(ctx)
		if err != nil {
			return allwork.Currency{}, err
		}
		active := slices.ContainsFunc(options, func(o people.DisplayCurrencyOption) bool { return o.Code == code && o.Active })
		if rate, err := rates.Of(code); active && err == nil {
			return allwork.Currency{Code: code, Rate: rate}, nil
		}
	}
	return s.source.Currency(ctx, a.ID)
}
