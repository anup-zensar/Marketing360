package planbudget_test

import (
	"net/http"
	"slices"
	"testing"

	"mapp/internal/measure"
	"mapp/internal/measure/planbudget"
	"mapp/internal/platform/apitest"
)

func TestPlanBudgetContract(t *testing.T) {
	t.Parallel()
	planner := apitest.New(t).SignIn(t, "priya.raman@shell.com")
	for _, path := range []string{
		"/api/v1/reports/plan-budget?year=2026",
		"/api/v1/reports/plan-budget?year=2026&period=quarter&currency=EUR",
		"/api/v1/reports/plan-budget/ranked?year=2026",
		"/api/v1/reports/plan-budget/ranked?year=2026&show=activity&group=owner&sort=actual&order=asc&size=25&q=a",
		"/api/v1/reports/plan-budget/export?year=2026&format=xlsx",
	} {
		if r := planner.Do(t, http.MethodGet, path, nil); r.Status != http.StatusOK {
			t.Errorf("GET %s = %d %s, want 200", path, r.Status, r.Body)
		}
	}
	bad := planner.Do(t, http.MethodGet, "/api/v1/reports/plan-budget/ranked?sort=name&size=7", nil, apitest.SkipRequestValidation())
	if bad.Status != http.StatusBadRequest {
		t.Errorf("unknown sort and size = %d %s, want 400", bad.Status, bad.Body)
	}
	var catalogue measure.ReportCatalogue
	planner.Do(t, http.MethodGet, "/api/v1/reports", nil).Decode(t, &catalogue)
	if !slices.ContainsFunc(catalogue.Items, func(e measure.ReportEntry) bool { return e.Key == measure.KeyPlanBudget && e.Available }) {
		t.Errorf("catalogue %+v: Plan budget is built", catalogue.Items)
	}
}

func TestPlanBudget_DemoCardsAddUp_040_SC003(t *testing.T) {
	t.Parallel()
	planner := apitest.New(t).SignIn(t, "priya.raman@shell.com")
	var report planbudget.Report
	planner.Do(t, http.MethodGet, "/api/v1/reports/plan-budget?year=2026", nil).Decode(t, &report)
	var list planbudget.Ranked
	planner.Do(t, http.MethodGet, "/api/v1/reports/plan-budget/ranked?year=2026&show=plan", nil).Decode(t, &list)

	if report.Cards.Planned == nil || len(list.Groups) != 1 || shown(list.Groups[0].Totals.Planned) != shown(report.Cards.Planned) {
		t.Errorf("planned cost: cards %s, ranked plans %+v; want equal", shown(report.Cards.Planned), list.Groups)
	}
}
