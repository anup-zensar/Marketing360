package measure_test

import (
	"net/http"
	"strings"
	"testing"

	"mapp/internal/measure"
	"mapp/internal/platform/apitest"
)

const (
	alexEmail   = "alex.morgan@shell.com"
	carlosEmail = "carlos.mendes@shell.com"
	frameOfYear = "/api/v1/reports/frame?report=planBudget&year=2026"
)

func call(t *testing.T, c *apitest.Client, path string, status int) *apitest.Response {
	t.Helper()
	var opts []apitest.RequestOption
	if status == http.StatusBadRequest || status == http.StatusNotFound {
		opts = append(opts, apitest.SkipRequestValidation())
	}
	r := c.Do(t, http.MethodGet, path, nil, opts...)
	if r.Status != status {
		t.Fatalf("GET %s = %d %s, want %d", path, r.Status, r.Body, status)
	}
	return r
}

func TestReportsCatalogue_Contract_FR001(t *testing.T) {
	t.Parallel()
	var catalogue measure.ReportCatalogue
	call(t, apitest.New(t).SignIn(t, carlosEmail), "/api/v1/reports", http.StatusOK).Decode(t, &catalogue)

	if len(catalogue.Items) < 8 || catalogue.Items[0].Key != measure.KeyPlanBudget {
		t.Errorf("catalogue = %+v, want the reports in the catalogue's order", catalogue.Items)
	}
}

// Every demo persona sees only its markets: no plan, owner or team of another market is named (SC-002).
func TestReportsFrame_ScopePerPersona_SC002(t *testing.T) {
	t.Parallel()
	server := apitest.New(t)
	var everything measure.Frame
	call(t, server.SignIn(t, carlosEmail), frameOfYear, http.StatusOK).Decode(t, &everything)
	for _, email := range []string{"priya.raman@shell.com", "daniel.okoro@shell.com", "mei.lin@shell.com",
		"lea.santos@shell.com", "sara.novak@shell.com"} {
		r := call(t, server.SignIn(t, email), frameOfYear, http.StatusOK)
		var mine measure.Frame
		r.Decode(t, &mine)
		own := map[string]bool{}
		for _, p := range mine.Choices.Plans {
			own[p.ID] = true
		}
		for _, p := range everything.Choices.Plans {
			if !own[p.ID] && strings.Contains(string(r.Body), p.ID) {
				t.Errorf("%s's frame names %s, a plan outside their scope", email, p.ID)
			}
		}
		if mine.InView.Plans > everything.InView.Plans {
			t.Errorf("%s sees %d plans, more than All markets' %d", email, mine.InView.Plans, everything.InView.Plans)
		}
	}
}

func TestReportsExport_FR022(t *testing.T) {
	t.Parallel()
	server := apitest.New(t)
	carlos := server.SignIn(t, carlosEmail)

	excel := call(t, carlos, "/api/v1/reports/plan-budget/export?year=2026", http.StatusOK)

	if !strings.Contains(excel.Header.Get("Content-Disposition"), "SP%26A%20One%20%E2%80%93%20Plan%20budget") {
		t.Errorf("Content-Disposition = %q, want the report's file name", excel.Header.Get("Content-Disposition"))
	}
	call(t, carlos, "/api/v1/reports/plan-budget/export?format=csv", http.StatusBadRequest)
	call(t, carlos, "/api/v1/reports/plan-budget/export?format=csv&widget=statusCount", http.StatusBadRequest)

	admin := server.SignIn(t, alexEmail)
	var created struct{ ID string }
	admin.Do(t, http.MethodPost, "/api/v1/roles", map[string]any{"name": "No export", "defaultScope": "myTeams",
		"permissions": []string{"plans.view", "reports.view"}}).Decode(t, &created)
	admin.Do(t, http.MethodPut, "/api/v1/people/U-1004/roles", map[string]any{"version": 1,
		"roles": []map[string]any{{"roleId": created.ID, "scope": map[string]any{"type": "myTeams"}}}})
	call(t, server.SignIn(t, "mei.lin@shell.com"), "/api/v1/reports/plan-budget/export", http.StatusForbidden)
}

func TestReportsFrame_Refusals(t *testing.T) {
	t.Parallel()
	server := apitest.New(t)
	carlos := server.SignIn(t, carlosEmail)

	call(t, carlos, frameOfYear+"&period=year", http.StatusBadRequest)
	call(t, carlos, "/api/v1/reports/frame?report=nope", http.StatusNotFound)

	admin := server.SignIn(t, alexEmail)
	var created struct{ ID string }
	admin.Do(t, http.MethodPost, "/api/v1/roles", map[string]any{"name": "Plans only", "defaultScope": "myTeams",
		"permissions": []string{"plans.view"}}).Decode(t, &created)
	given := admin.Do(t, http.MethodPut, "/api/v1/people/U-1004/roles", map[string]any{"version": 1,
		"roles": []map[string]any{{"roleId": created.ID, "scope": map[string]any{"type": "myTeams"}}}})
	if given.Status != http.StatusOK {
		t.Fatalf("giving Mei the role = %d %s", given.Status, given.Body)
	}
	mei := server.SignIn(t, "mei.lin@shell.com")
	call(t, mei, frameOfYear, http.StatusForbidden)
	call(t, mei, "/api/v1/reports", http.StatusForbidden)
}
