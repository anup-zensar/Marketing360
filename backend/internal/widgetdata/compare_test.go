package widgetdata_test

import (
	"context"
	"testing"

	"mapp/internal/platform/actor"
	"mapp/internal/platform/authz"
	"mapp/internal/platform/date"
	"mapp/internal/widgetdata"
)

// periodReader answers the rows of the period asked for.
type periodReader map[string]widgetdata.Rows

func (p periodReader) Rows(_ context.Context, _ actor.Actor, f widgetdata.Prefilter) (widgetdata.Rows, error) {
	return p[f.From.String()], nil
}

func kpiOf(comparison *widgetdata.ComparisonRequest, filters *widgetdata.Filters) widgetdata.Request {
	return widgetdata.Request{Type: "kpi", Query: widgetdata.Query{Source: "activities",
		Measures: []widgetdata.Measure{{Aggregation: "sum", Field: ptr("plannedCost")}},
		Sort:     widgetdata.Sort{By: "measure", Direction: "desc"}, Limit: 1}, Filters: filters, Comparison: comparison}
}

func TestComparison_TargetAndPreviousPeriod_FR023(t *testing.T) {
	t.Parallel()
	from, to := date.MustParse("2026-07-01"), date.MustParse("2026-12-31")
	filters := &widgetdata.Filters{DateFrom: &from, DateTo: &to}
	fields := widgetdata.BuiltInFields("activities")
	cost := func(s string) widgetdata.Rows {
		return widgetdata.Rows{Items: []widgetdata.Row{widgetdata.RowOf(fields, map[string]widgetdata.Value{"plannedCost": usd(t, s)})}}
	}
	reader := periodReader{"2026-07-01": cost("150.00"), "2025-12-29": cost("120.00"), "2026-01-01": cost("0.00")}
	s := service(map[string]widgetdata.SourceReader{"activities": reader}, nil)
	ctx := signedIn(t, authz.DashboardsView, authz.PlansView)
	tests := []struct {
		name               string
		r                  widgetdata.Request
		value, change, per string
	}{
		{"the share of the target reached", kpiOf(&widgetdata.ComparisonRequest{Kind: "target", Target: ptr("200")}, filters),
			"200.00 USD", "75.0", ""},
		{"a target of zero has no share", kpiOf(&widgetdata.ComparisonRequest{Kind: "target", Target: ptr("0")}, filters),
			"0.00 USD", "—", ""},
		{"the previous period of equal length", kpiOf(&widgetdata.ComparisonRequest{Kind: "previousPeriod"}, filters),
			"120.00 USD", "25.0", "2025-12-29–2026-06-30"},
		{"a previous period of zero has no change", kpiOf(&widgetdata.ComparisonRequest{Kind: "previousPeriod"},
			&widgetdata.Filters{DateFrom: ptr(date.MustParse("2027-01-01")), DateTo: ptr(date.MustParse("2027-12-31"))}),
			"0.00 USD", "—", "2026-01-01–2026-12-31"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			data, err := s.Query(ctx, viewer, tt.r)
			if err != nil {
				t.Fatalf("Query: %v", err)
			}
			value, change, period := comparisonTexts(data.Comparison)
			if value != tt.value || change != tt.change || period != tt.per {
				t.Errorf("comparison = %s, %s, %s; want %s, %s, %s", value, change, period, tt.value, tt.change, tt.per)
			}
		})
	}

	t.Run("the previous period needs a date range", func(t *testing.T) {
		t.Parallel()
		if _, err := s.Query(ctx, viewer, kpiOf(&widgetdata.ComparisonRequest{Kind: "previousPeriod"}, nil)); status(err) != 400 {
			t.Errorf("Query without dates error = %v, want 400", err)
		}
	})
}

func comparisonTexts(c *widgetdata.Comparison) (value, change, period string) {
	value, change = "—", "—"
	if c.Value.Money != nil {
		value = c.Value.Money.Format()
	}
	if c.Change != nil {
		change = c.Change.String()
	}
	if c.Period != nil {
		period = c.Period.From.String() + "–" + c.Period.To.String()
	}
	return value, change, period
}
