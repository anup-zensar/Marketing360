package purchasereport_test

import (
	"net/http"
	"testing"

	"mapp/internal/measure/purchasereport"
	"mapp/internal/platform/apitest"
)

const (
	alexEmail = "alex.morgan@shell.com"
	reportURL = "/api/v1/reports/purchase-requests"
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

func TestPurchaseRequestsReport_Contract_042(t *testing.T) {
	t.Parallel()
	server := apitest.New(t)
	alex := server.SignIn(t, alexEmail)

	var view purchasereport.View
	get(t, alex, reportURL+"?year=2027", http.StatusOK).Decode(t, &view)

	if len(view.StatusMix.Items) != 6 || !view.PurchaseSupport {
		t.Errorf("StatusMix, PurchaseSupport = %d items, %v, want the six statuses and purchase support on", len(view.StatusMix.Items), view.PurchaseSupport)
	}
	if view.Cards.PurchaseRequests == 0 || len(view.Grid.Vendors) == 0 {
		t.Errorf("Cards, Grid = %+v, want the demo data's purchase requests, validated against the contract", view.Cards)
	}
	for _, query := range []string{"vendor=Nobody", "status=nonsense", "requestOwner=nobody", "sort=amount&order=desc&pageSize=5&page=1",
		"q=stand&columns=id,invoiced,status"} {
		get(t, alex, reportURL+"?year=2027&"+query, http.StatusOK)
	}
	for _, malformed := range []string{"year=abc", "sort=nope", "order=up", "page=0", "pageSize=7", "columns=nonsense", "columns=id,id"} {
		get(t, alex, reportURL+"?"+malformed, http.StatusBadRequest)
	}
	if signedOut := server.Do(t, http.MethodGet, reportURL, nil); signedOut.Status != http.StatusUnauthorized {
		t.Errorf("signed out = %d, want 401", signedOut.Status)
	}
}

func TestPurchaseRequestsReport_NoReportsView_042(t *testing.T) {
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

type board struct {
	PurchaseRequests struct {
		Available      bool `json:"available"`
		Count          int  `json:"count"`
		WaitingForGSAP int  `json:"waitingForGsap"`
		ByStatus       []struct {
			Status string `json:"status"`
			Count  int    `json:"count"`
		} `json:"byStatus"`
	} `json:"purchaseRequests"`
}

func TestPurchaseRequestsReport_EqualsDashboard_042_SC002(t *testing.T) {
	t.Parallel()
	alex := apitest.New(t).SignIn(t, alexEmail)
	var dashboard board
	get(t, alex, "/api/v1/dashboards/executive?year=2027", http.StatusOK).Decode(t, &dashboard)
	var view purchasereport.View
	get(t, alex, reportURL+"?year=2027", http.StatusOK).Decode(t, &view)

	if view.StatusMix.Total != dashboard.PurchaseRequests.Count {
		t.Errorf("status total = %d, want the dashboard's %d", view.StatusMix.Total, dashboard.PurchaseRequests.Count)
	}
	if view.StatusMix.WaitingForGSAP != dashboard.PurchaseRequests.WaitingForGSAP {
		t.Errorf("waiting for GSAP = %d, want the dashboard's %d", view.StatusMix.WaitingForGSAP, dashboard.PurchaseRequests.WaitingForGSAP)
	}
	for i, want := range dashboard.PurchaseRequests.ByStatus {
		if got := view.StatusMix.Items[i]; string(got.Status) != want.Status || got.Count != want.Count {
			t.Errorf("status %d = %s %d, want the dashboard's %s %d", i, got.Status, got.Count, want.Status, want.Count)
		}
	}
}
