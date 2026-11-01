package app_test

import (
	"net/http"
	"slices"
	"testing"

	"github.com/google/go-cmp/cmp"

	"mapp/internal/dashboard"
	"mapp/internal/platform/apitest"
	"mapp/internal/platform/money"
	"mapp/internal/widgetdata"
)

const (
	widgetQuery  = "/api/v1/data-sources/query"
	leaEmail     = "lea.santos@shell.com"
	lubricantsUK = "FY2027 Lubricants UK & Ireland"
)

func byPlan(source string, measures ...widgetdata.Measure) widgetdata.Request {
	return widgetdata.Request{Type: "table", Query: widgetdata.Query{
		Source: source, Dimension: &widgetdata.GroupBy{Field: "plan"}, Measures: measures, Filters: []widgetdata.Filter{},
		Sort: widgetdata.Sort{By: "dimension", Direction: "asc"}, Limit: 50,
	}}
}

func measure(aggregation, field string) widgetdata.Measure {
	if field == "" {
		return widgetdata.Measure{Aggregation: aggregation}
	}
	return widgetdata.Measure{Aggregation: aggregation, Field: &field}
}

func ask(t *testing.T, c *apitest.Client, r widgetdata.Request, status int) widgetdata.Data {
	t.Helper()
	answer := c.Do(t, http.MethodPost, widgetQuery, r)
	if answer.Status != status {
		t.Fatalf("POST %s %s = %d %s, want %d", widgetQuery, r.Query.Source, answer.Status, answer.Body, status)
	}
	var data widgetdata.Data
	if status == http.StatusOK {
		answer.Decode(t, &data)
	}
	return data
}

func groupLabels(d widgetdata.Data) []string {
	labels := []string{}
	for _, r := range d.Rows {
		labels = append(labels, r.Dimension.Label)
	}
	return labels
}

func moneyText(c widgetdata.Cell) string {
	if c.Money == nil {
		return "—"
	}
	return c.Money.Format()
}

func TestQuery_OnlyThePlansInScope_FR007(t *testing.T) {
	t.Parallel()
	server := apitest.New(t)
	sources := []string{"plans", "campaigns", "initiatives", "activities", "lineItems", "plannedByMonth", "approvals", "purchaseRequests"}
	tests := []struct {
		email string
		want  []string
	}{
		{priyasEmail, []string{"FY2026 Lubricants Europe", lubricantsUK}},
		{meisEmail, []string{"FY2026 Mobility China"}},
		{sarasEmail, []string{"FY2026 Lubricants Europe", "FY2026 Mobility China", lubricantsUK}},
	}
	for _, tt := range tests {
		c := server.SignIn(t, tt.email)
		for _, source := range sources {
			t.Run(tt.email+" "+source, func(t *testing.T) {
				t.Parallel()
				data := ask(t, c, byPlan(source, measure("count", "")), http.StatusOK)
				for _, plan := range groupLabels(data) {
					if !slices.Contains(tt.want, plan) {
						t.Errorf("%s sees %q in %s; only %q are in scope", tt.email, plan, source, tt.want)
					}
				}
			})
		}
		t.Run(tt.email+" plans", func(t *testing.T) {
			t.Parallel()
			if diff := cmp.Diff(tt.want, groupLabels(ask(t, c, byPlan("plans", measure("count", "")), http.StatusOK))); diff != "" {
				t.Errorf("plans of %s (-want +got):\n%s", tt.email, diff)
			}
		})
	}
}

func TestQuery_FinanceHiddenOutsideFinanceScope_FR007(t *testing.T) {
	t.Parallel()
	server := apitest.New(t)
	lea := server.SignIn(t, leaEmail)
	spend := byPlan("lineItems", measure("sum", "plannedCost"), measure("sum", "committed"))

	before := ask(t, lea, spend, http.StatusOK)
	ask(t, lea, byPlan("financialTransactions", measure("count", "")), http.StatusForbidden)
	for _, r := range before.Rows {
		if r.Values[1].Money != nil {
			t.Errorf("without See financial data, %s's committed = %s, want empty", r.Dimension.Label, moneyText(r.Values[1]))
		}
	}

	server.GrantRole(t, leaEmail, "viewer", "teams", "Lubricants Europe")
	after := ask(t, lea, spend, http.StatusOK)
	committed := map[string]string{}
	for _, r := range after.Rows {
		committed[r.Dimension.Label] = moneyText(r.Values[1])
	}
	if committed["FY2026 Mobility China"] != "—" || committed["FY2026 Lubricants Europe"] == "—" {
		t.Errorf("committed by plan = %v, want Lubricants only", committed)
	}
	if diff := cmp.Diff([]string{"Financial figures cover 2 of 4 plans"}, after.Notes); diff != "" {
		t.Errorf("notes (-want +got):\n%s", diff)
	}
	transactions := ask(t, lea, byPlan("financialTransactions", measure("count", "")), http.StatusOK)
	if diff := cmp.Diff([]string{"FY2026 Lubricants Europe"}, groupLabels(transactions)); diff != "" && len(transactions.Rows) > 0 {
		t.Errorf("transactions by plan (-want +got):\n%s", diff)
	}
}

func TestQuery_EqualsTheRecordPages_SC002(t *testing.T) {
	t.Parallel()
	server := apitest.New(t)
	for _, email := range []string{carlosEmail, priyasEmail, sarasEmail} {
		t.Run(email, func(t *testing.T) {
			t.Parallel()
			c := server.SignIn(t, email)
			var board dashboard.View
			if r := c.Do(t, http.MethodGet, "/api/v1/dashboards/executive?year=2026&currency=USD", nil); r.Status == http.StatusOK {
				r.Decode(t, &board)
			}
			want := map[string][]string{}
			for _, p := range board.Plans {
				f := p.Figures
				want[p.Plan.Name] = []string{moneyOf(f.Budget), moneyOf(f.PlannedCost), moneyOf(f.LeftToPlan)}
			}
			r := byPlan("plans", measure("sum", "budget"), measure("sum", "plannedCost"), measure("sum", "leftToPlan"))
			r.Query.Filters = []widgetdata.Filter{{Field: "year", Operator: "eq", Values: []string{"2026"}}}
			got := map[string][]string{}
			for _, row := range ask(t, c, r, http.StatusOK).Rows {
				got[row.Dimension.Label] = []string{moneyText(row.Values[0]), moneyText(row.Values[1]), moneyText(row.Values[2])}
			}
			if diff := cmp.Diff(want, got); diff != "" || len(got) == 0 {
				t.Errorf("plan figures, dashboard (-) and widget (+):\n%s", diff)
			}
		})
	}
}

func moneyOf(m *money.Money) string {
	if m == nil {
		return "—"
	}
	return m.Format()
}
