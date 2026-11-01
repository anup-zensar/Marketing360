package customdashboard

import (
	"strings"
	"testing"
)

func TestValidateQuery_FieldsShapesAndValues_FR016_FR020(t *testing.T) {
	t.Parallel()
	filter := func(operator string, values ...string) []Filter {
		return []Filter{{Field: "startDate", Operator: operator, Values: values}}
	}
	const at = "widgets[0].query."
	tests := []struct {
		name   string
		change func(q *Query)
		want   map[string]string
	}{
		{name: "a dataset source", change: func(q *Query) { q.Source = "DS-1001" }, want: map[string]string{}},
		{name: "an unknown source", change: func(q *Query) { q.Source = "invoices" },
			want: map[string]string{at + "source": msgUnknownSource}},
		{name: "a group without a field", change: func(q *Query) { q.Dimension.Field = "" },
			want: map[string]string{at + "dimension.field": msgFieldName}},
		{name: "an unknown date group", change: func(q *Query) { q.Dimension.DateGroup = ptr("fortnight") },
			want: map[string]string{at + "dimension.dateGroup": msgDateGroup}},
		{name: "a breakdown equal to the group", change: func(q *Query) { q.Series = &GroupBy{Field: "activityType"} },
			want: map[string]string{at + "series": msgSameAsGroup}},
		{name: "the same date field in another bucket", change: func(q *Query) {
			q.Dimension = &GroupBy{Field: "startDate", DateGroup: ptr("month")}
			q.Series = &GroupBy{Field: "startDate", DateGroup: ptr("year")}
		}, want: map[string]string{}},
		{name: "count of records", change: func(q *Query) { q.Measures[0] = Measure{Aggregation: "count"} },
			want: map[string]string{}},
		{name: "sum without a field and a long label", change: func(q *Query) {
			q.Measures[0] = Measure{Aggregation: "sum", Label: strings.Repeat("l", 41)}
		}, want: map[string]string{at + "measures[0].field": msgMeasureField, at + "measures[0].label": msgLabelTooLong}},
		{name: "an unknown aggregation", change: func(q *Query) { q.Measures[0].Aggregation = "median" },
			want: map[string]string{at + "measures[0].aggregation": msgAggregation}},
		{name: "between two dates", change: func(q *Query) { q.Filters = filter("between", "2026-01-01", "2026-06-30") },
			want: map[string]string{}},
		{name: "between in the wrong order", change: func(q *Query) { q.Filters = filter("between", "12.5", "3") },
			want: map[string]string{at + "filters[0].values": msgBetweenOrder}},
		{name: "between with one value", change: func(q *Query) { q.Filters = filter("between", "3") },
			want: map[string]string{at + "filters[0].values": msgBetweenValues}},
		{name: "is with two values", change: func(q *Query) { q.Filters = filter("eq", "a", "b") },
			want: map[string]string{at + "filters[0].values": msgOneValue}},
		{name: "is one of with 51 values", change: func(q *Query) { q.Filters = filter("in", make([]string, 51)...) },
			want: map[string]string{at + "filters[0].values": msgInValues}},
		{name: "is empty with a value", change: func(q *Query) { q.Filters = filter("isEmpty", "x") },
			want: map[string]string{at + "filters[0].values": msgNoValues}},
		{name: "contains 101 characters", change: func(q *Query) { q.Filters = filter("contains", strings.Repeat("c", 101)) },
			want: map[string]string{at + "filters[0].values": msgContainsLength}},
		{name: "an empty value", change: func(q *Query) { q.Filters = filter("ne", "") },
			want: map[string]string{at + "filters[0].values": msgValueLength}},
		{name: "an unknown operator and no field", change: func(q *Query) { q.Filters = []Filter{{Operator: "like", Values: []string{}}} },
			want: map[string]string{at + "filters[0].operator": msgOperator, at + "filters[0].field": msgFieldName}},
		{name: "eleven filters", change: func(q *Query) { q.Filters = make([]Filter, 11) },
			want: map[string]string{at + "filters": msgTooManyFilters}},
		{name: "sort by a measure that is not there", change: func(q *Query) { q.Sort.MeasureIndex = 1 },
			want: map[string]string{at + "sort.measureIndex": msgSortMeasure}},
		{name: "sort by something else", change: func(q *Query) { q.Sort = Sort{By: "label", Direction: "up"} },
			want: map[string]string{at + "sort.by": msgSortBy, at + "sort.direction": msgSortDirection}},
		{name: "a limit of 501", change: func(q *Query) { q.Limit = 501 }, want: map[string]string{at + "limit": msgLimit}},
		{name: "a limit of 0", change: func(q *Query) { q.Limit = 0 }, want: map[string]string{at + "limit": msgLimit}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			w := validBar("w1", 0, 0)
			tt.change(w.Query)
			assertProblems(t, validDashboard(w), tt.want)
		})
	}
}

func TestValidateDisplay_OptionsByType_FR024_FR025(t *testing.T) {
	t.Parallel()
	const at = "widgets[0].display."
	tests := []struct {
		name   string
		change func(w *Widget)
		want   map[string]string
	}{
		{name: "every bar option", change: func(w *Widget) {
			w.Query.Series = &GroupBy{Field: "campaign"}
			w.Display = Display{Palette: ptr("planning"), Legend: &Legend{Show: true, Position: "top"}, DataLabels: ptr(true),
				AxisTitles: &AxisTitles{X: "Type"}, Gridlines: ptr(false), Stack: ptr("percent"), Orientation: ptr("horizontal"),
				NumberFormat: ptr("compact"), Decimals: ptr(1), Zoom: ptr(true), ReferenceLine: &ReferenceLine{Value: "250000", Label: "Target"}}
		}, want: map[string]string{}},
		{name: "a pie that is smooth and stacked", change: func(w *Widget) {
			w.Type, w.Display = typePie, Display{Smooth: ptr(true), Stack: ptr("none")}
		}, want: map[string]string{at + "smooth": "A pie chart doesn't use this option.", at + "stack": "A pie chart doesn't use this option."}},
		{name: "unknown values", change: func(w *Widget) {
			w.Display = Display{Palette: ptr("rainbow"), Legend: &Legend{Position: "middle"}, Decimals: ptr(5),
				AxisTitles: &AxisTitles{Y2: strings.Repeat("y", 41)}, ReferenceLine: &ReferenceLine{Value: "1,000"}}
		}, want: map[string]string{at + "palette": msgChooseOption, at + "legend.position": msgChooseOption,
			at + "decimals": msgDecimals, at + "axisTitles.y2": msgLabelTooLong, at + "referenceLine.value": msgDecimal}},
		{name: "stacked without a breakdown", change: func(w *Widget) { w.Display = Display{Stack: ptr("stacked")} },
			want: map[string]string{at + "stack": msgStackNeedsSeries}},
		{name: "combo styles one per measure", change: func(w *Widget) {
			w.Type = typeCombo
			w.Query.Measures = append(w.Query.Measures, Measure{Aggregation: "count"})
			w.Display = Display{MeasureStyles: []MeasureStyle{{As: "bar", Axis: "left"}}}
		}, want: map[string]string{at + "measureStyles": msgMeasureStyles}},
		{name: "a gauge whose minimum is not below its maximum", change: func(w *Widget) {
			w.Type, w.Query.Dimension = typeGauge, nil
			w.Display = Display{Gauge: &GaugeOptions{Min: "100", Max: ptr("99.5")}}
		}, want: map[string]string{at + "gauge": msgGaugeRange}},
		{name: "a gauge rounding its maximum", change: func(w *Widget) {
			w.Type, w.Query.Dimension = typeGauge, nil
			w.Display = Display{Gauge: &GaugeOptions{Min: "-5"}}
		}, want: map[string]string{}},
		{name: "a key figure against a target without one", change: func(w *Widget) {
			w.Type, w.Query.Dimension = typeKpi, nil
			w.Display = Display{Kpi: &KpiOptions{Comparison: "target"}}
		}, want: map[string]string{at + "kpi.target": msgEnterTarget}},
		{name: "table columns repeated and of a missing breakdown", change: func(w *Widget) {
			w.Type = typeTable
			w.Display = Display{Table: &TableOptions{Columns: []string{"m0", "m0", "series"}}}
		}, want: map[string]string{at + "table.columns": msgTableColumns}},
		{name: "a text note with a palette", change: func(w *Widget) {
			w.Type, w.Query, w.Display = typeText, nil, Display{Palette: ptr("categorical")}
		}, want: map[string]string{at + "palette": "A text note doesn't use this option."}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			w := validBar("w1", 0, 0)
			tt.change(&w)
			assertProblems(t, validDashboard(w), tt.want)
		})
	}
}
