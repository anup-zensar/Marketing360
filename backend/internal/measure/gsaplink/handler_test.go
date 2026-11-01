package gsaplink_test

import (
	"net/http"
	"testing"

	"mapp/internal/measure/gsaplink"
	"mapp/internal/platform/apitest"
)

const (
	alex   = "alex.morgan@shell.com"
	carlos = "carlos.mendes@shell.com"
	priya  = "priya.raman@shell.com"
	mei    = "mei.lin@shell.com"
	report = "/api/v1/reports/gsap-link?year=2026"
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

func reportOf(t *testing.T, c *apitest.Client, path string) gsaplink.Report {
	t.Helper()
	var r gsaplink.Report
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

// TestGsapLink_Contract reads every tab on the demo year; apitest checks each answer against the contract.
func TestGsapLink_Contract(t *testing.T) {
	t.Parallel()
	carlosSees := apitest.New(t).SignIn(t, carlos)

	if r := reportOf(t, carlosSees, report); r.Tab != gsaplink.TabSwbs || r.Swbs == nil || r.Total == 0 {
		t.Errorf("default tab = %s with %d rows, want the SWBS tab of the demo year", r.Tab, r.Total)
	}
	if r := reportOf(t, carlosSees, report+"&tab=waiting&sort=amount"); r.Waiting == nil || r.Finance == nil {
		t.Errorf("waiting tab = %+v, want its cards and finance coverage", r)
	}
	get(t, carlosSees, report+"&tab=all", http.StatusBadRequest)
	get(t, carlosSees, report+"&tab=orders&sort=-waiting", http.StatusBadRequest)
}

func TestGsapLink_OrdersInReportCurrency_Contract(t *testing.T) {
	t.Parallel()
	r := reportOf(t, apitest.New(t).SignIn(t, carlos), report+"&tab=orders&sort=-poAmount&currency=EUR&page=1&pageSize=5")

	if r.Orders == nil || r.Total == 0 {
		t.Fatalf("orders tab = %+v, want the demo purchase orders", r)
	}
	if r.Currency != "EUR" || r.Orders.Totals.PoAmount.Currency != "EUR" {
		t.Errorf("currency = %s, totals in %s; want EUR", r.Currency, r.Orders.Totals.PoAmount.Currency)
	}
}

func TestReport_OnlyPlansInScope_FR003(t *testing.T) {
	t.Parallel()
	server := apitest.New(t)

	everything := reportOf(t, server.SignIn(t, carlos), report+"&tab=orders")
	mine := reportOf(t, server.SignIn(t, priya), report+"&tab=orders")

	if mine.Finance == nil || everything.Finance == nil || mine.Finance.Of == 0 || mine.Finance.Of >= everything.Finance.Of {
		t.Errorf("Priya has %+v plans in view, Carlos (All markets) %+v: want only Priya's markets", mine.Finance,
			everything.Finance)
	}
}

func TestGsapLink_NoReportsView_FR003(t *testing.T) {
	t.Parallel()
	server := apitest.New(t)
	giveMeiOnly(t, server, "plans.view")

	get(t, server.SignIn(t, mei), report, http.StatusForbidden)
}

// TestGsapLinkExport_OwnParameters_FR019: the export reads the report's tab and sort and refuses one of another tab.
func TestGsapLinkExport_OwnParameters_FR019(t *testing.T) {
	t.Parallel()
	carlosSees := apitest.New(t).SignIn(t, carlos)
	export := "/api/v1/reports/gsap-link/export?year=2026&format=csv&widget=swbs"

	if r := carlosSees.Do(t, http.MethodGet, export+"&tab=swbs&sort=lineItem", nil); r.Status != http.StatusOK {
		t.Errorf("export sorted by line item = %d %s, want 200", r.Status, r.Body)
	}
	get(t, carlosSees, export+"&sort=vendor", http.StatusBadRequest)
}

func TestGsapLinkExport_NeedsDataExport(t *testing.T) {
	t.Parallel()
	server := apitest.New(t)
	giveMeiOnly(t, server, "reports.view")
	meiSees := server.SignIn(t, mei)

	get(t, meiSees, report, http.StatusOK)
	get(t, meiSees, "/api/v1/reports/gsap-link/export?year=2026&format=xlsx", http.StatusForbidden)
}
