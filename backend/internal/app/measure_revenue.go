package app

import (
	"context"

	"mapp/internal/allwork"
	"mapp/internal/measure"
	"mapp/internal/measure/revenue"
	"mapp/internal/plan"
	"mapp/internal/platform/actor"
	"mapp/internal/platform/authz"
	"mapp/internal/projection"
	"mapp/internal/settings"
	"mapp/internal/stores"
)

// registerRevenue builds feature 049's one service and mounts its handler, then registers the Revenue
// catalogue entry; registerFunnel (called next, app/measure.go) registers Funnel on a service of its own
// (research R1 - one package, two reports) without mounting the handler again.
func registerRevenue(w *wiring, r *measure.Registry) error {
	mount(w.signedIn, revenue.NewHandler(revenueService(w)))
	return r.Register(revenue.NewRevenueReport(revenueService(w)))
}

func registerFunnel(w *wiring, r *measure.Registry) error {
	return r.Register(revenue.NewFunnelReport(revenueService(w)))
}

func revenueService(w *wiring) *revenue.Service {
	deps := revenue.Deps{
		Scoper: w.reports, Facts: w.set.Revenue,
		Goals:      revenueGoals{plans: w.set.Plans, activities: w.set.Activities, settings: w.settings},
		Categories: revenueCategories{settings: w.settings},
		Rates:      w.settings,
	}
	return revenue.NewService(w.set.Tx, deps, w.now)
}

// revenueGoals answers 049's Goals port: each activity's plan model and "set later stages myself" flag, the
// two inputs allwork.Node does not carry (research R3). The nodes already name each activity's plan, so
// plans are read once for the whole report, not once per activity.
type revenueGoals struct {
	plans      stores.PlanStore
	activities stores.ActivityStore
	settings   *settings.Service
}

func (g revenueGoals) ActivityGoals(ctx context.Context, scope authz.Scope, nodes []allwork.Node) (map[string]revenue.ActivityGoals, error) {
	activityIDs := make([]string, 0, len(nodes))
	planOf := make(map[string]string, len(nodes))
	for _, n := range nodes {
		activityIDs = append(activityIDs, n.ID)
		planOf[n.ID] = n.PlanID
	}
	found, err := g.activities.ListInScopeOf(ctx, scope, activityIDs)
	if err != nil {
		return nil, err
	}
	models, err := g.modelsByPlan(ctx, planOf)
	if err != nil {
		return nil, err
	}
	out := make(map[string]revenue.ActivityGoals, len(found))
	for _, a := range found {
		out[a.ID] = revenue.ActivityGoals{SetLaterStagesMyself: a.SetLaterStagesMyself, Model: models[planOf[a.ID]]}
	}
	return out, nil
}

// modelsByPlan reads every plan the activities belong to once, then every distinct model once: bounded by
// the plans and models in view, never by the number of activities (SC-003).
func (g revenueGoals) modelsByPlan(ctx context.Context, planOf map[string]string) (map[string]projection.Model, error) {
	planIDs := uniqueValues(planOf)
	plans := make(map[string]plan.Plan, len(planIDs))
	for _, id := range planIDs {
		p, found, err := g.plans.Get(ctx, id)
		if err != nil {
			return nil, err
		}
		if found {
			plans[id] = p
		}
	}
	models := make(map[string]projection.Model, len(plans))
	cache := map[string]projection.Model{}
	for planID, p := range plans {
		if cached, ok := cache[p.ModelID]; ok {
			models[planID] = cached
			continue
		}
		m, found, err := g.settings.Model(ctx, p.ModelID)
		if err != nil {
			return nil, err
		}
		if found {
			cache[p.ModelID] = projectionModel(m)
			models[planID] = cache[p.ModelID]
		}
	}
	return models, nil
}

func uniqueValues(m map[string]string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(m))
	for _, v := range m {
		if v != "" && !seen[v] {
			seen[v] = true
			out = append(out, v)
		}
	}
	return out
}

// revenueCategories answers 049's Categories port over 047's activity categories.
type revenueCategories struct {
	settings *settings.Service
}

func (c revenueCategories) ListCategories(ctx context.Context) ([]revenue.Category, error) {
	found, err := c.settings.ListActivityCategories(ctx, actor.System)
	if err != nil {
		return nil, err
	}
	out := make([]revenue.Category, 0, len(found))
	for _, cat := range found {
		typeIDs := make([]string, 0, len(cat.ActivityTypes))
		for _, t := range cat.ActivityTypes {
			typeIDs = append(typeIDs, t.ID)
		}
		out = append(out, revenue.Category{ID: cat.ID, Name: cat.Name, IsDefault: cat.IsDefault, ActivityTypeIDs: typeIDs})
	}
	return out, nil
}
