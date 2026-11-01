package app

import (
	"context"
	"slices"

	"mapp/internal/allwork"
	"mapp/internal/measure"
	"mapp/internal/measure/plandetails"
	"mapp/internal/platform/actor"
)

// registerPlanDetails registers and mounts the Plan details report (feature 043).
func registerPlanDetails(w *wiring, r *measure.Registry) error {
	service := plandetails.NewService(w.set.Tx, planDetailsScope{reports: w.reports, source: w.allWork})
	mount(w.signedIn, plandetails.NewHandler(service))
	return r.Register(service)
}

// planDetailsScope is the report frame's step (measure.Service.Scope, 039 research R1, R2) as 043 reads it.
type planDetailsScope struct {
	reports *measure.Service
	source  allWorkSource
}

func (s planDetailsScope) Matched(ctx context.Context, a actor.Actor, f measure.Filter, columns []string) (
	plandetails.Matching, error,
) {
	sc, err := s.reports.Scope(ctx, a, f, columns)
	matched := map[string]bool{}
	m := sc.Matched
	for _, id := range slices.Concat(m.Plans, m.Campaigns, m.Initiatives, m.Activities) {
		matched[id] = true
	}
	return plandetails.Matching{Year: sc.Year, Nodes: sc.Nodes, Matched: matched, ActivityTypes: sc.Choices.ActivityTypes,
		Currency: sc.Currency}, err
}

func (s planDetailsScope) Attributes(ctx context.Context) ([]allwork.Attribute, error) {
	return s.source.Attributes(ctx)
}
