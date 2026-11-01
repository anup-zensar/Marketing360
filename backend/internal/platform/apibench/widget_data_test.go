package apibench_test

import (
	"net/http"
	"testing"

	"mapp/internal/platform/apibench"
	"mapp/internal/widgetdata"
)

// BenchmarkWidgetData measures feature 051's SC-004 on the demo data and on the volume: each of the twelve
// widgets of a dashboard answered on its own, together within 3 seconds and one preview within 1 second.
func BenchmarkWidgetData(b *testing.B) {
	for _, data := range []apibench.Data{apibench.Demo, apibench.Volume} {
		server := apibench.New(b, data)
		clients := signIn(b, server)
		for _, w := range dashboardWidgets() {
			b.Run(string(data)+"/"+w.name, func(b *testing.B) {
				clients[alex].Measure(b, http.MethodPost, "/api/v1/data-sources/query", w.request)
			})
		}
	}
}

type widget struct {
	name    string
	request widgetdata.Request
}

func dashboardWidgets() []widget {
	field := func(key string) *string { return &key }
	month := "month"
	of := func(typ, source, dimension string, dateGroup *string, m widgetdata.Measure) widgetdata.Request {
		r := widgetdata.Request{Type: typ, Query: widgetdata.Query{Source: source, Measures: []widgetdata.Measure{m},
			Filters: []widgetdata.Filter{}, Sort: widgetdata.Sort{By: "measure", Direction: "desc"}, Limit: 50}}
		if dimension != "" {
			r.Query.Dimension = &widgetdata.GroupBy{Field: dimension, DateGroup: dateGroup}
		}
		return r
	}
	sum := func(key string) widgetdata.Measure { return widgetdata.Measure{Aggregation: "sum", Field: field(key)} }
	count := widgetdata.Measure{Aggregation: "count"}
	return []widget{
		{"plans budget", of("bar", "plans", "plan", nil, sum("budget"))},
		{"plans committed", of("bar", "plans", "team", nil, sum("committed"))},
		{"campaigns planned", of("bar", "campaigns", "campaign", nil, sum("plannedCost"))},
		{"initiatives left", of("table", "initiatives", "initiative", nil, sum("leftToPlan"))},
		{"activities by type", of("pie", "activities", "activityType", nil, sum("plannedCost"))},
		{"activities by status", of("donut", "activities", "status", nil, count)},
		{"line items by currency", of("bar", "lineItems", "currency", nil, sum("plannedCost"))},
		{"planned by month", of("line", "plannedByMonth", "month", &month, sum("plannedCost"))},
		{"transactions by type", of("bar", "financialTransactions", "type", nil, sum("amount"))},
		{"approvals waiting", of("kpi", "approvals", "", nil, count)},
		{"purchase requests", of("funnel", "purchaseRequests", "status", nil, sum("amount"))},
		{"global budgets", of("bar", "globalBudgets", "globalBudget", nil, sum("allocated"))},
	}
}
