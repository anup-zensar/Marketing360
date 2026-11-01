package plandetails_test

import (
	"net/http"
	"testing"

	"mapp/internal/measure/plandetails"
	"mapp/internal/platform/apitest"
)

const (
	alex   = "alex.morgan@shell.com"
	carlos = "carlos.mendes@shell.com"
	priya  = "priya.raman@shell.com"
	mei    = "mei.lin@shell.com"
	report = "/api/v1/reports/plan-details?year=2026"
)

func get(t *testing.T, c *apitest.Client, path string, status int) *apitest.Response {
	t.Helper()
	var opts []apitest.RequestOption
	if status == http.StatusBadRequest {
		opts = append(opts, apitest.SkipRequestValidation())
	}
	r := c.Do(t, http.MethodGet, path, nil, opts...)
	if r.Status != status {
		t.Fatalf("GET %s = %d %s, want %d", path, r.Status, r.Body, status)
	}
	return r
}

func reportOf(t *testing.T, c *apitest.Client, path string) plandetails.Report {
	t.Helper()
	var r plandetails.Report
	get(t, c, path, http.StatusOK).Decode(t, &r)
	return r
}

// giveMeiOnly gives Mei (U-1004) a new role with the permissions only, through the access API.
func giveMeiOnly(t *testing.T, server *apitest.Server, permissions ...string) {
	t.Helper()
	admin := server.SignIn(t, alex)
	role := admin.Do(t, http.MethodPost, "/api/v1/roles", map[string]any{"name": "Report test", "defaultScope": "myTeams",
		"permissions": permissions})
	var created struct{ ID string }
	role.Decode(t, &created)
	given := admin.Do(t, http.MethodPut, "/api/v1/people/U-1004/roles", map[string]any{"version": 1,
		"roles": []map[string]any{{"roleId": created.ID, "scope": map[string]any{"type": "myTeams"}}}})
	if given.Status != http.StatusOK {
		t.Fatalf("giving Mei the role = %d %s", given.Status, given.Body)
	}
}

func TestPlanDetails_Contract(t *testing.T) {
	t.Parallel()
	carlosSees := apitest.New(t).SignIn(t, carlos)

	for _, show := range []string{"", "&show=plan", "&show=campaign", "&show=initiative", "&period=quarter&currency=EUR"} {
		r := reportOf(t, carlosSees, report+show)
		if r.Cards.Records == 0 {
			t.Errorf("%s: no record, want the demo year's", show)
		}
	}
	get(t, carlosSees, report+"&show=lineItem", http.StatusBadRequest)
}

func TestReport_OnlyPlansInScope_FR003(t *testing.T) {
	t.Parallel()
	server := apitest.New(t)

	everything := reportOf(t, server.SignIn(t, carlos), report)
	mine := reportOf(t, server.SignIn(t, priya), report)

	if mine.Cards.Records == 0 || mine.Cards.Records >= everything.Cards.Records {
		t.Errorf("Priya counts %d activities, Carlos (All markets) %d: want only Priya's markets", mine.Cards.Records,
			everything.Cards.Records)
	}
}

func TestPlanDetails_NoReportsView_FR003(t *testing.T) {
	t.Parallel()
	server := apitest.New(t)
	giveMeiOnly(t, server, "plans.view")

	refused := get(t, server.SignIn(t, mei), report, http.StatusForbidden)

	if refused.Status != http.StatusForbidden {
		t.Errorf("without reports.view = %d, want 403", refused.Status)
	}
}

func TestReport_NoFinanceView_ShowsPlannedCost_FR004(t *testing.T) {
	t.Parallel()
	server := apitest.New(t)
	giveMeiOnly(t, server, "plans.view", "reports.view")

	r := reportOf(t, server.SignIn(t, mei), report)

	if r.Cards.Records == 0 || r.Cards.PlannedCost.Amount.IsZero() {
		t.Errorf("cards = %+v, want Mei's activities with their planned cost although she may not see financial data",
			r.Cards)
	}
}
