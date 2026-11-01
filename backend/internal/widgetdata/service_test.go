package widgetdata_test

import (
	"errors"
	"net/http"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"mapp/internal/platform/authz"
	"mapp/internal/platform/date"
	"mapp/internal/platform/errs"
	"mapp/internal/widgetdata"
)

func ptr[T any](v T) *T { return &v }

func activitiesRows(t *testing.T) widgetdata.Rows {
	t.Helper()
	fields := widgetdata.BuiltInFields("activities")
	row := func(kind, cost string) widgetdata.Row {
		return widgetdata.RowOf(fields, map[string]widgetdata.Value{"activityType": text(kind), "plannedCost": usd(t, cost)})
	}
	return widgetdata.Rows{Items: []widgetdata.Row{row("Event", "100.00"), row("Event", "20.00"), row("Digital", "30.00")},
		Notes: []string{widgetdata.FinanceNote(1, 2)}}
}

func sumByType(source string) widgetdata.Request {
	return widgetdata.Request{Type: "bar", Query: widgetdata.Query{
		Source: source, Dimension: &widgetdata.GroupBy{Field: "activityType"},
		Measures: []widgetdata.Measure{{Aggregation: "sum", Field: ptr("plannedCost")}},
		Sort:     widgetdata.Sort{By: "measure", Direction: "desc"}, Limit: 50,
	}}
}

func status(err error) int {
	if apiErr, ok := errors.AsType[*errs.Error](err); ok {
		return apiErr.Status
	}
	return 0
}

func TestQueryValidation_AllProblems_FR019(t *testing.T) {
	t.Parallel()
	r := widgetdata.Request{Type: "pie", Query: widgetdata.Query{
		Source: "activities", Series: &widgetdata.GroupBy{Field: "team"},
		Measures: []widgetdata.Measure{{Aggregation: "sum", Field: ptr("activity")}, {Aggregation: "count", Field: ptr("nope")}},
		Filters: []widgetdata.Filter{
			{Field: "startDate", Operator: "gt", Values: []string{"1 July"}},
			{Field: "activity", Operator: "between", Values: []string{"a", "b"}},
			{Field: "plannedCost", Operator: "between", Values: []string{"10", "5"}},
			{Field: "plannedCost", Operator: "eq", Values: []string{"1.234"}},
			{Field: "status", Operator: "isEmpty", Values: []string{"x"}},
		},
		Sort: widgetdata.Sort{By: "measure", MeasureIndex: 5, Direction: "up"}, Limit: 0,
	}, Filters: &widgetdata.Filters{DateFrom: ptr(date.MustParse("2026-12-31")), DateTo: ptr(date.MustParse("2026-01-01"))},
		Comparison: &widgetdata.ComparisonRequest{Kind: "target"}}
	reader := &fakeReader{rows: activitiesRows(t)}

	_, err := service(map[string]widgetdata.SourceReader{"activities": reader}, nil).Query(signedIn(t, authz.DashboardsView, authz.PlansView), viewer, r)

	apiErr, ok := errors.AsType[*errs.Error](err)
	if !ok || apiErr.Code != errs.CodeValidationFailed {
		t.Fatalf("Query error = %v, want validation_failed", err)
	}
	fields := map[string]string{}
	for _, f := range apiErr.Fields {
		fields[f.Field] = f.Message
	}
	want := map[string]string{
		"query.dimension":           "Choose what to group by.",
		"query.series":              "A pie chart has no breakdown.",
		"query.measures":            "A pie chart shows one measure.",
		"query.measures[0].field":   "Sum needs a number or money field.",
		"query.measures[1].field":   "This field isn't in the data source.",
		"query.filters[0].values":   "Enter a date as YYYY-MM-DD.",
		"query.filters[1].operator": "This condition doesn't apply to this field.",
		"query.filters[2].values":   "The first value is after the second.",
		"query.filters[3].values":   "Enter an amount in USD with up to 2 decimals.",
		"query.filters[4].values":   "This condition takes no values.",
		"query.sort.measureIndex":   "Choose one of the measures.",
		"query.sort.direction":      "Choose ascending or descending.",
		"query.limit":               "Show 1 to 500 groups.",
		"filters.dateTo":            "The start date is after the end date.",
		"comparison":                "Only a key figure compares.",
	}
	if diff := cmp.Diff(want, fields); diff != "" {
		t.Errorf("problems (-want +got):\n%s", diff)
	}
	if len(reader.asked) != 0 {
		t.Error("a refused query reads nothing")
	}
}

func TestQuery_Access_FR007(t *testing.T) {
	t.Parallel()
	readers := map[string]widgetdata.SourceReader{"activities": &fakeReader{rows: activitiesRows(t)}, "financialTransactions": &fakeReader{}}
	s := service(readers, leadsDataset())
	tests := []struct {
		name   string
		perms  []authz.Permission
		source string
		want   int
	}{
		{"without See or Build dashboards", []authz.Permission{authz.PlansView}, "activities", http.StatusForbidden},
		{"a source the person may not read", []authz.Permission{authz.DashboardsView, authz.PlansView}, "financialTransactions",
			http.StatusForbidden},
		{"a dataset the person may not see", []authz.Permission{authz.DashboardsView}, "DS-2002", http.StatusNotFound},
		{"an unknown source", []authz.Permission{authz.DashboardsView}, "budgets", http.StatusBadRequest},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := s.Query(signedIn(t, tt.perms...), viewer, sumByType(tt.source))
			if got := status(err); got != tt.want {
				t.Errorf("Query error = %v, want %d", err, tt.want)
			}
		})
	}
}

func TestPrefilter_AppliedPerSource_FR022(t *testing.T) {
	t.Parallel()
	dates := cmpopts.EquateComparable(date.Date{})
	from, to := date.MustParse("2026-07-01"), date.MustParse("2026-12-31")
	filters := &widgetdata.Filters{DateFrom: &from, DateTo: &to, PlanIDs: []string{"P-1001"}, TeamIDs: []string{"T-1002"}}
	activities, plans := &fakeReader{rows: activitiesRows(t)}, &fakeReader{}
	budgets := &fakeReader{}
	s := service(map[string]widgetdata.SourceReader{"activities": activities, "plans": plans, "globalBudgets": budgets}, leadsDataset())
	ctx := signedIn(t, authz.DashboardsView, authz.PlansView, authz.GlobalBudgetsView)
	ask := func(source, dimension, field string) widgetdata.Data {
		r := sumByType(source)
		r.Query.Dimension.Field, r.Query.Measures[0].Field, r.Filters = dimension, ptr(field), filters
		data, err := s.Query(ctx, viewer, r)
		if err != nil {
			t.Fatalf("Query %s: %v", source, err)
		}
		return data
	}

	ask("activities", "activityType", "plannedCost")
	ask("plans", "status", "budget")
	budgetsData := ask("globalBudgets", "owner", "amount")
	leads := ask("DS-1001", "c1", "c2")

	if diff := cmp.Diff(widgetdata.Prefilter{From: &from, To: &to, PlanIDs: []string{"P-1001"}, TeamIDs: []string{"T-1002"}},
		activities.asked[0], dates); diff != "" {
		t.Errorf("activities prefilter (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff(widgetdata.Prefilter{PlanIDs: []string{"P-1001"}, TeamIDs: []string{"T-1002"}}, plans.asked[0], dates); diff != "" {
		t.Errorf("plans have no date filter (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff(widgetdata.Prefilter{}, budgets.asked[0], dates); diff != "" || activities.who[0] != viewer.ID {
		t.Errorf("global budgets take no filter (-want +got):\n%s", diff)
	}
	wantNotes := []string{"Not filtered by date", "Not filtered by plan", "Not filtered by team"}
	if diff := cmp.Diff(wantNotes, budgetsData.Notes); diff != "" {
		t.Errorf("global budgets notes (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff(wantNotes, leads.Notes); diff != "" || leads.FiltersApplied != (widgetdata.FiltersApplied{}) {
		t.Errorf("dataset notes (-want +got):\n%s applied %+v", diff, leads.FiltersApplied)
	}
}

func TestQuery_AnswersRowsTotalsAndNotes_FR016(t *testing.T) {
	t.Parallel()
	s := service(map[string]widgetdata.SourceReader{"activities": &fakeReader{rows: activitiesRows(t)}}, nil)

	data, err := s.Query(signedIn(t, authz.DashboardsView, authz.PlansView), viewer, sumByType("activities"))
	if err != nil {
		t.Fatalf("Query: %v", err)
	}

	first := data.Rows[0]
	got := []string{first.Dimension.Label, first.Values[0].Money.Format(), first.Values[0].Share.String(), data.Totals[0].Money.Format()}
	if diff := cmp.Diff([]string{"Event", "120.00 USD", "80.0", "150.00 USD"}, got); diff != "" || len(data.Rows) != 2 {
		t.Errorf("first row and total (-want +got):\n%s rows %d, want 2", diff, len(data.Rows))
	}
	wantColumns := []widgetdata.Column{{Key: "dimension", Label: "Activity type", Kind: widgetdata.KindText},
		{Key: "m0", Label: "Sum of Planned cost", Kind: widgetdata.KindMoney}}
	if diff := cmp.Diff(wantColumns, data.Columns); diff != "" {
		t.Errorf("columns (-want +got):\n%s", diff)
	}
	if len(data.Notes) != 0 || data.FiltersApplied != (widgetdata.FiltersApplied{Date: true, Plan: true, Team: true}) {
		t.Errorf("notes %q (the finance note only with a finance field), applied %+v", data.Notes, data.FiltersApplied)
	}
}

func TestQuery_FinanceNoteWithFinanceFields_FR007(t *testing.T) {
	t.Parallel()
	s := service(map[string]widgetdata.SourceReader{"activities": &fakeReader{rows: activitiesRows(t)}}, nil)
	r := sumByType("activities")
	r.Query.Measures[0].Field = ptr("committed")
	data, _ := s.Query(signedIn(t, authz.DashboardsView, authz.PlansView), viewer, r)
	if diff := cmp.Diff([]string{"Financial figures cover 1 of 2 plans"}, data.Notes); diff != "" {
		t.Errorf("finance notes (-want +got):\n%s", diff)
	}
}

func TestQuery_DatasetMoneyInUSD_FR018(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name, currency, want string
		notes                []string
	}{
		{"converted at the current rate", "EUR", "20.02 USD", []string{}},
		{"a currency no longer in Settings is left out", "XXX", "0.00 USD",
			[]string{"Amounts in a currency no longer in Settings are left out"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ds := leadsDataset()
			table := ds.tables["DS-1001"]
			table.Info.Columns = []widgetdata.DatasetColumn{table.Info.Columns[0], {Key: "c2", Label: "Spend",
				Kind: widgetdata.KindMoney, Currency: tt.currency}}
			table.Rows = []widgetdata.Row{{text("North"), usd(t, "10.00")}, {text("North"), usd(t, "0.01")},
				{text("South"), widgetdata.Value{Empty: true}}}
			ds.tables["DS-1001"] = table
			r := sumByType("DS-1001")
			r.Query.Dimension.Field, r.Query.Measures[0].Field = "c1", ptr("c2")

			data, err := service(nil, ds).Query(signedIn(t, authz.DashboardsView), viewer, r)

			if err != nil || data.Rows[0].Values[0].Money.Format() != tt.want {
				t.Fatalf("dataset sum = %+v, %v; want %s", data.Rows, err, tt.want)
			}
			if diff := cmp.Diff(tt.notes, data.Notes); diff != "" {
				t.Errorf("notes (-want +got):\n%s", diff)
			}
		})
	}
}
