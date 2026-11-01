package app

import (
	"context"
	"slices"
	"strings"

	"mapp/internal/history"
	"mapp/internal/measure"
	"mapp/internal/platform/authz"
	"mapp/internal/platform/money"
	"mapp/internal/settings"
	"mapp/internal/stores"
)

// measureArea wires the Measure module (039): one register function per report, each in its own
// measure_<report>.go, so the lane that builds a report changes only that file (D-369).
func (w *wiring) measureArea() error {
	registry := measure.NewRegistry()
	tools := history.Tools{Tx: w.set.Tx, IDs: w.generator, Log: w.history, Now: w.now}
	w.reports = measure.NewService(tools, measure.Deps{Planning: w.allWork, Plans: reportPlans{repo: w.set.Plans},
		Currencies: reportCurrencies{settings: w.settings}, Views: w.set.Measure}, registry)
	if err := w.history.Register(w.reports.ViewKind()); err != nil {
		return err
	}
	registers := []func(*wiring, *measure.Registry) error{
		registerPlanBudget, registerApprovals, registerPurchaseRequests, registerPlanDetails, registerGsapLink,
		registerUserAccess, registerSummary, registerRevenue, registerFunnel,
	}
	for _, register := range registers {
		if err := register(w, registry); err != nil {
			return err
		}
	}
	mount(w.signedIn, measure.NewHandler(w.reports))
	return w.attributionArea()
}

// reportPlans are the plans in the person's reports.view scope, read by the scoped plan store (039 research R2).
type reportPlans struct {
	repo stores.PlanStore
}

func (p reportPlans) InReportsScope(ctx context.Context) ([]string, error) {
	list, err := p.repo.ListInScope(ctx, authz.PermissionScopeOf(ctx, authz.ReportsView))
	ids := make([]string, 0, len(list))
	for _, plan := range list {
		ids = append(ids, plan.ID)
	}
	return ids, err
}

func (p reportPlans) YearsInReportsScope(ctx context.Context) ([]int, error) {
	list, err := p.repo.ListInScope(ctx, authz.PermissionScopeOf(ctx, authz.ReportsView))
	years := []int{}
	for _, plan := range list {
		if !slices.Contains(years, plan.Year) {
			years = append(years, plan.Year)
		}
	}
	return years, err
}

type reportCurrencies struct {
	settings *settings.Service
}

func (c reportCurrencies) Active(ctx context.Context) ([]measure.Currency, error) {
	all, err := c.settings.Currencies(ctx)
	active := []measure.Currency{}
	for _, currency := range all {
		if currency.Active {
			active = append(active, measure.Currency{Code: currency.Code, Name: currency.Name})
		}
	}
	slices.SortFunc(active, func(x, y measure.Currency) int { return strings.Compare(x.Code, y.Code) })
	return active, err
}

func (c reportCurrencies) RateOf(ctx context.Context, code string) (money.Rate, error) {
	rates, err := c.settings.Rates(ctx)
	if err != nil {
		return money.Rate{}, err
	}
	return rates.Of(code)
}
