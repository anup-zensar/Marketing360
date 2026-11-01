package approvalreport_test

import (
	"net/http"
	"testing"
	"time"

	"mapp/internal/measure/approvalreport"
	"mapp/internal/platform/apitest"
)

const (
	alexEmail = "alex.morgan@shell.com"
	reportURL = "/api/v1/reports/approvals"
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

func TestApprovalsReport_Contract_041(t *testing.T) {
	t.Parallel()
	server := apitest.New(t)
	alex := server.SignIn(t, alexEmail)

	var view approvalreport.View
	get(t, alex, reportURL+"?year=2026", http.StatusOK).Decode(t, &view)

	if len(view.StatusMix.Items) != 6 || view.StatusMix.RecordType != approvalreport.RecordActivity {
		t.Errorf("StatusMix = %+v, want six outcomes for activities", view.StatusMix)
	}
	for _, unknown := range []string{"statusType=nonsense", "recordType=nonsense", "outcome=nonsense", "approver=nobody"} {
		get(t, alex, reportURL+"?year=2026&"+unknown, http.StatusOK)
	}
	for _, malformed := range []string{"year=abc", "period=year"} {
		get(t, alex, reportURL+"?"+malformed, http.StatusBadRequest)
	}
	signedOut := server.Do(t, http.MethodGet, reportURL, nil)
	if signedOut.Status != http.StatusUnauthorized {
		t.Errorf("signed out = %d, want 401", signedOut.Status)
	}
}

func TestApprovalsReport_NoReportsView_041(t *testing.T) {
	t.Parallel()
	server := apitest.New(t)
	admin := server.SignIn(t, alexEmail)
	var created struct{ ID string }
	admin.Do(t, http.MethodPost, "/api/v1/roles", map[string]any{"name": "Plans only", "defaultScope": "myTeams",
		"permissions": []string{"plans.view"}}).Decode(t, &created)
	admin.Do(t, http.MethodPut, "/api/v1/people/U-1004/roles", map[string]any{"version": 1,
		"roles": []map[string]any{{"roleId": created.ID, "scope": map[string]any{"type": "myTeams"}}}})

	get(t, server.SignIn(t, "mei.lin@shell.com"), reportURL, http.StatusForbidden)
}

type executive struct {
	Approvals struct {
		Waiting             int        `json:"waiting"`
		OldestWaitingSince  *time.Time `json:"oldestWaitingSince"`
		AverageDecisionDays *string    `json:"averageDecisionDays"`
	} `json:"approvals"`
}

func TestApprovalsReport_EqualsDashboard_041_SC002(t *testing.T) {
	t.Parallel()
	alex := apitest.New(t).SignIn(t, alexEmail)
	var board executive
	get(t, alex, "/api/v1/dashboards/executive?year=2026", http.StatusOK).Decode(t, &board)

	waiting, oldest := 0, (*time.Time)(nil)
	for _, kind := range []approvalreport.RecordType{approvalreport.RecordActivity, approvalreport.RecordPurchaseRequest,
		approvalreport.RecordModifiedPurchaseRequest} {
		var view approvalreport.View
		get(t, alex, reportURL+"?year=2026&statusType="+string(kind), http.StatusOK).Decode(t, &view)
		waiting += view.StatusMix.Waiting.Count
		if since := view.StatusMix.Waiting.OldestSince; since != nil && (oldest == nil || since.Before(*oldest)) {
			oldest = since
		}
	}
	var view approvalreport.View
	get(t, alex, reportURL+"?year=2026", http.StatusOK).Decode(t, &view)

	if waiting != board.Approvals.Waiting {
		t.Errorf("waiting over the three types = %d, want the dashboard's %d", waiting, board.Approvals.Waiting)
	}
	if (oldest == nil) != (board.Approvals.OldestWaitingSince == nil) ||
		(oldest != nil && !oldest.Equal(*board.Approvals.OldestWaitingSince)) {
		t.Errorf("oldest waiting = %v, want the dashboard's %v", oldest, board.Approvals.OldestWaitingSince)
	}
	if got, want := view.Average.Days, board.Approvals.AverageDecisionDays; (got == nil) != (want == nil) || (got != nil && *got != *want) {
		t.Errorf("average days = %v, want the dashboard's %v", got, want)
	}
}
