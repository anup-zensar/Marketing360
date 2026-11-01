package measure_test

import (
	"net/http"
	"testing"

	"mapp/internal/measure"
	"mapp/internal/platform/apitest"
)

const (
	priyaEmail      = "priya.raman@shell.com"
	planBudgetViews = "/api/v1/reports/plan-budget/views"
)

type request struct {
	method, path string
	body         any
}

func send(t *testing.T, c *apitest.Client, req request, status int) *apitest.Response {
	t.Helper()
	var opts []apitest.RequestOption
	if status == http.StatusNotFound || status == http.StatusBadRequest {
		opts = append(opts, apitest.SkipRequestValidation())
	}
	r := c.Do(t, req.method, req.path, req.body, opts...)
	if r.Status != status {
		t.Fatalf("%s %s = %d %s, want %d", req.method, req.path, r.Status, r.Body, status)
	}
	return r
}

func codeOf(t *testing.T, r *apitest.Response) string {
	t.Helper()
	var envelope struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	r.Decode(t, &envelope)
	return envelope.Error.Code
}

func TestReportViewsContract_FR024_FR028(t *testing.T) {
	t.Parallel()
	server := apitest.New(t)
	planner, other := server.SignIn(t, priyaEmail), server.SignIn(t, carlosEmail)

	var created measure.ReportView
	send(t, planner, request{http.MethodPost, planBudgetViews, map[string]any{"name": "Europe", "query": "year=2026", "isDefault": true}}, http.StatusCreated).Decode(t, &created)
	var mine, theirs measure.ReportViewPage
	send(t, planner, request{http.MethodGet, planBudgetViews, nil}, http.StatusOK).Decode(t, &mine)
	send(t, other, request{http.MethodGet, planBudgetViews, nil}, http.StatusOK).Decode(t, &theirs)
	if mine.Total != 1 || mine.Items[0].ID != created.ID || theirs.Total != 0 {
		t.Errorf("Priya lists %d views, Carlos %d: want 1 and 0", mine.Total, theirs.Total)
	}

	one := planBudgetViews + "/" + created.ID
	var renamed measure.ReportView
	send(t, planner, request{http.MethodPatch, one, map[string]any{"version": 1, "name": "Europe 2026"}}, http.StatusOK).Decode(t, &renamed)
	stale := send(t, planner, request{http.MethodPatch, one, map[string]any{"version": 1, "isDefault": false}}, http.StatusConflict)
	if renamed.Version != 2 || codeOf(t, stale) != "version_conflict" {
		t.Errorf("rename to version %d; a stale version answered %s", renamed.Version, stale.Body)
	}
	invalid := send(t, planner, request{http.MethodPost, planBudgetViews, map[string]any{"name": " ", "query": ""}}, http.StatusBadRequest)
	if codeOf(t, invalid) != "validation_failed" {
		t.Errorf("a view without a name: %s, want validation_failed", invalid.Body)
	}
	send(t, other, request{http.MethodDelete, one, nil}, http.StatusNotFound)
	send(t, planner, request{http.MethodGet, "/api/v1/reports/sales-pipeline/views", nil}, http.StatusNotFound)

	var deleted measure.ReportView
	send(t, planner, request{http.MethodDelete, one, nil}, http.StatusOK).Decode(t, &deleted)
	if deleted.Name != "Europe 2026" {
		t.Errorf("deleted %+v: want the view as it was, for Undo", deleted)
	}
	send(t, planner, request{http.MethodDelete, one, nil}, http.StatusNotFound)
}

// A planner may keep a view of a report their roles do not open: it is theirs, and shows nothing they may not see.
func TestReportViews_OfAReportTheyCannotOpen_039_EdgeCase(t *testing.T) {
	t.Parallel()
	planner := apitest.New(t).SignIn(t, priyaEmail)
	path := "/api/v1/reports/user-access/views"

	send(t, planner, request{http.MethodPost, path, map[string]any{"name": "Kept", "query": ""}}, http.StatusCreated)
	var page measure.ReportViewPage
	send(t, planner, request{http.MethodGet, path, nil}, http.StatusOK).Decode(t, &page)
	if page.Total != 1 {
		t.Errorf("Priya's User access views = %d, want 1", page.Total)
	}
}
