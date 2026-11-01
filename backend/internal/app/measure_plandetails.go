package app

import (
	"context"
	"slices"

	"mapp/internal/measure"
	"mapp/internal/measure/plandetails"
	"mapp/internal/platform/actor"
)

// registerPlanDetails registers and mounts the Plan details report (feature 043).
func registerPlanDetails(w *wiring, r *measure.Registry) error {
	service := plandetails.NewService(w.set.Tx, planDetailsScope{reports: w.reports})
	mount(w.signedIn, plandetails.NewHandler(service))
	return r.Register(service)
}

// planDetailsScope is the report frame's step (measure.Service.Scope, 039 research R1, R2) as 043 reads it,
// attributes included, so they are not read twice.
type planDetailsScope struct {
	reports *measure.Service
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
		Currency: sc.Currency, Attributes: sc.Attributes}, err
}
