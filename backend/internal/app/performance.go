package app

import (
	"context"

	"mapp/internal/activity"
	"mapp/internal/history"
	"mapp/internal/performance"
	"mapp/internal/performance/simulatedfeed"
	"mapp/internal/plan"
	"mapp/internal/platform/actor"
	"mapp/internal/platform/money"
	"mapp/internal/projection"
	"mapp/internal/settings"
	"mapp/internal/stores"
)

const feedSeed = 46

// performanceApprovals answers 046-C1: In progress and Complete are display forms of approved.
type performanceApprovals struct {
	activities interface {
		ListByIDs(ctx context.Context, ids []string) ([]activity.Activity, error)
	}
}

func (p performanceApprovals) ApprovedIDs(ctx context.Context, ids []string) (map[string]bool, error) {
	found, err := p.activities.ListByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}
	approved := make(map[string]bool, len(found))
	for _, a := range found {
		approved[a.ID] = a.Status == activity.StatusApproved
	}
	return approved, nil
}

// heldRelease late-binds the release hook: approvals are wired before the performance area exists.
type heldRelease struct {
	service *performance.Service
}

func (r *heldRelease) activityApproved(ctx context.Context, a actor.Actor, activityID string) error {
	if r.service == nil {
		return nil
	}
	_, err := r.service.Release(ctx, a, activityID)
	return err
}

// performanceArea builds feature 046 after the planning areas, whose activities its facts match to.
func (w *wiring) performanceArea(release *heldRelease) error {
	tools := history.Tools{Tx: w.set.Tx, IDs: w.generator, Log: w.history, Now: w.now}
	service := performance.NewService(tools, w.set.Performance)
	service.UseApprovals(performanceApprovals{activities: w.set.Activities})
	service.UseFeed(simulatedfeed.New(feedSeed))
	service.UseFigures(performanceFigures{
		activities: w.set.Activities, campaigns: w.set.Campaigns, plans: w.set.Plans, settings: w.settings})
	service.UseRates(w.settings.Rates)
	release.service = service
	if err := w.register(service.Kinds()); err != nil {
		return err
	}
	mount(w.signedIn, performance.NewHandler(service, w.now))
	return nil
}

// performanceFigures answers the panel's goal side: the goals as shown (typed, or the model's waterfall -
// domain model 4.5), with the plan's model and year.
type performanceFigures struct {
	activities stores.ActivityStore
	campaigns  stores.CampaignStore
	plans      stores.PlanStore
	settings   *settings.Service
}

func (p performanceFigures) GoalFacts(ctx context.Context, activityID string) (performance.GoalFacts, bool, error) {
	act, found, err := p.activities.Get(ctx, activityID)
	if err != nil || !found {
		return performance.GoalFacts{}, false, err
	}
	owning, found, err := p.planOf(ctx, act.InitiativeID)
	if err != nil || !found {
		return performance.GoalFacts{}, false, err
	}
	facts := performance.GoalFacts{INQ: act.INQ, ModelID: owning.ModelID,
		ActivityTypeID: act.ActivityTypeID, PlanYear: owning.Year,
		SetLaterStagesMyself: act.SetLaterStagesMyself,
		StartDate:            act.StartDate, EndDate: act.EndDate, Ongoing: act.Ongoing}
	model, found, err := p.settings.Model(ctx, facts.ModelID)
	if err != nil {
		return facts, false, err
	}
	if found {
		facts.Model = projectionModel(model)
	}
	if act.SetLaterStagesMyself {
		facts.MQL, facts.Revenue = act.MQL, act.Revenue
		return facts, true, nil
	}
	return p.calculated(ctx, facts, act.INQ)
}

func (p performanceFigures) calculated(ctx context.Context, facts performance.GoalFacts, inq *int64) (performance.GoalFacts, bool, error) {
	if inq == nil {
		return facts, true, nil
	}
	model, found, err := p.settings.Model(ctx, facts.ModelID)
	if err != nil || !found {
		return facts, err == nil, err
	}
	// A waterfall too large for money keeps the goals empty rather than failing the panel, exactly as the
	// activity's own Goals section shows them (goals.go).
	goals, waterfallErr := settings.Waterfall(*inq, model)
	if waterfallErr != nil {
		return facts, true, nil //nolint:nilerr // an overflowing goal is shown as absent, not as a failure
	}
	usd := money.USD(goals.Revenue)
	facts.MQL, facts.Revenue = &goals.MQL, &usd
	return facts, true, nil
}

// projectionModel is the model as 047's engine reads it: the same parameters, stamped with when the model
// last changed so every projected figure can be reproduced (047 FR-012).
func projectionModel(m settings.Model) projection.Model {
	return projection.Model{
		Name: m.Name, InqToMqlRate: m.InqToMqlRate, MqlToDealRate: m.MqlToDealRate,
		AverageDealSize:        m.AverageDealSize,
		InqToMqlVelocityMonths: m.InqToMqlVelocityMonths, MqlToDealVelocityMonths: m.MqlToDealVelocityMonths,
		ParametersChangedAt: m.UpdatedAt,
	}
}

func (p performanceFigures) planOf(ctx context.Context, initiativeID string) (plan.Plan, bool, error) {
	initiative, found, err := p.campaigns.GetInitiative(ctx, initiativeID)
	if err != nil || !found {
		return plan.Plan{}, false, err
	}
	owner, found, err := p.campaigns.GetCampaign(ctx, initiative.CampaignID)
	if err != nil || !found {
		return plan.Plan{}, false, err
	}
	return p.plans.Get(ctx, owner.PlanID)
}
