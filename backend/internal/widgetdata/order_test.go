package widgetdata

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

var shapes = Source{Key: "test", Fields: []Field{
	{Key: "type", Label: "Type", Kind: KindText}, {Key: "team", Label: "Team", Kind: KindText, Multi: true},
	{Key: "cost", Label: "Cost", Kind: KindMoney}, {Key: "start", Label: "Start", Kind: KindDate},
}}

func shapeRows(t *testing.T) []Row {
	t.Helper()
	row := func(kind, cost, start string, teams ...string) Row {
		return Row{TextValue(kind), TextsValue(teams), moneyValue(t, cost), dateValue(t, start)}
	}
	return []Row{
		row("Event", "50.00", "2026-01-10", "Europe"), row("Event", "30.00", "2026-02-10", "Asia"),
		row("Digital", "60.00", "2026-01-20", "Europe", "Asia"), row("Print", "10.00", "2026-03-05", "Europe"),
		row("Radio", "10.00", "2026-03-06", "Asia"), row("", "5.00", "2026-03-07"),
	}
}

func field(key string) *string { return &key }

func run(t *testing.T, r Request) outcome {
	t.Helper()
	p, err := compile(r, shapes)
	if err != nil {
		t.Fatalf("compile: %v", err)
	}
	out, err := p.run(shapeRows(t))
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	return out
}

func barOf(sort Sort, limit int, rest bool) Request {
	return Request{Type: typeBar, Query: Query{Dimension: &GroupBy{Field: "type"},
		Measures: []Measure{{Aggregation: aggSum, Field: field("cost")}}, Sort: sort, Limit: limit, GroupRest: rest}}
}

func rowTexts(rows []DataRow) []string {
	out := []string{}
	for _, r := range rows {
		text := r.Dimension.Label
		if r.Series != nil {
			text += "/" + r.Series.Label
		}
		for _, v := range r.Values {
			if v.Money != nil {
				text += " " + v.Money.Amount.String()
			}
			if v.Number != nil {
				text += " " + v.Number.String()
			}
			if v.Share != nil {
				text += " " + v.Share.String() + "%"
			}
		}
		out = append(out, text)
	}
	return out
}

func TestOrder_LimitAndOther_FR020(t *testing.T) {
	t.Parallel()
	byCost := Sort{By: sortByMeasure, Direction: directionDesc}
	tests := []struct {
		name  string
		r     Request
		want  []string
		count int
	}{
		{"largest first, ties by label, Not set by its value", barOf(byCost, 50, false),
			[]string{"Event 80.00 48.5%", "Digital 60.00 36.4%", "Print 10.00 6.1%", "Radio 10.00 6.1%", "Not set 5.00 3.0%"}, 5},
		{"the top two", barOf(byCost, 2, false), []string{"Event 80.00 48.5%", "Digital 60.00 36.4%"}, 5},
		{"the rest as Other", barOf(byCost, 2, true),
			[]string{"Event 80.00 48.5%", "Digital 60.00 36.4%", "Other 25.00 15.2%"}, 5},
		{"by the group ascending, Not set last", barOf(Sort{By: sortByDimension, Direction: directionAsc}, 50, false),
			[]string{"Digital 60.00 36.4%", "Event 80.00 48.5%", "Print 10.00 6.1%", "Radio 10.00 6.1%", "Not set 5.00 3.0%"}, 5},
		{"by the group descending, Not set still last", barOf(Sort{By: sortByDimension, Direction: directionDesc}, 50, false),
			[]string{"Radio 10.00 6.1%", "Print 10.00 6.1%", "Event 80.00 48.5%", "Digital 60.00 36.4%", "Not set 5.00 3.0%"}, 5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			out := run(t, tt.r)
			got := rowTexts(out.rows)
			if diff := cmp.Diff(tt.want, got); diff != "" || out.groupCount != tt.count {
				t.Errorf("rows (-want +got):\n%s groupCount = %d, want %d", diff, out.groupCount, tt.count)
			}
		})
	}
}

func TestShares_OfGroupOrTotal_FR021(t *testing.T) {
	t.Parallel()
	r := Request{Type: typeBar, Query: Query{Dimension: &GroupBy{Field: "start", DateGroup: field(bucketMonth)},
		Series: &GroupBy{Field: "team"}, Measures: []Measure{{Aggregation: aggCount}},
		Sort: Sort{By: sortByDimension, Direction: directionAsc}, Limit: 50}}

	out := run(t, r)

	want := []string{
		"2026-01-01/Asia 1 33.3%", "2026-01-01/Europe 2 66.7%",
		"2026-02-01/Asia 1 100.0%",
		"2026-03-01/Asia 1 33.3%", "2026-03-01/Europe 1 33.3%", "2026-03-01/Not set 1 33.3%",
	}
	if diff := cmp.Diff(want, rowTexts(out.rows)); diff != "" {
		t.Errorf("rows (-want +got):\n%s", diff)
	}
	if !out.multi || out.totals[0].Number.String() != "6" || out.totals[0].Share != nil {
		t.Errorf("multi = %v, totals = %+v; want a several-teams note and 6 records without a share", out.multi, out.totals[0])
	}

	avg := barOf(Sort{By: sortByDimension, Direction: directionAsc}, 50, false)
	avg.Query.Measures[0].Aggregation = aggAvg
	if out := run(t, avg); out.rows[0].Values[0].Share != nil {
		t.Errorf("an average has no share, got %v", out.rows[0].Values[0].Share)
	}
}

func TestOrder_SeriesBeyondTheCapJoinOther_FR020(t *testing.T) {
	t.Parallel()
	r := Request{Type: typeBar, Query: Query{Dimension: &GroupBy{Field: "start", DateGroup: field(bucketYear)},
		Series: &GroupBy{Field: "type"}, Measures: []Measure{{Aggregation: aggSum, Field: field("cost")}},
		Sort: Sort{By: sortByDimension, Direction: directionAsc}, Limit: 50}}
	p, err := compile(r, shapes)
	if err != nil {
		t.Fatalf("compile: %v", err)
	}
	p.seriesCap = 2

	out, err := p.run(shapeRows(t))
	if err != nil {
		t.Fatalf("run: %v", err)
	}

	want := []string{"2026-01-01/Digital 60.00 36.4%", "2026-01-01/Event 80.00 48.5%", "2026-01-01/Other 25.00 15.2%"}
	if diff := cmp.Diff(want, rowTexts(out.rows)); diff != "" {
		t.Errorf("rows (-want +got):\n%s", diff)
	}
}

func TestOrder_UngroupedAnswersOneRow_FR020(t *testing.T) {
	t.Parallel()
	kpi := Request{Type: typeKpi, Query: Query{Measures: []Measure{{Aggregation: aggSum, Field: field("cost")}},
		Sort: Sort{By: sortByMeasure, Direction: directionDesc}, Limit: 1}}
	p, err := compile(kpi, shapes)
	if err != nil {
		t.Fatalf("compile: %v", err)
	}
	out, err := p.run(nil)
	if err != nil || len(out.rows) != 1 || out.rows[0].Dimension != nil || out.rows[0].Values[0].Money.Amount.String() != "0.00" {
		t.Errorf("an ungrouped sum of no rows = %+v, %v; want one row of 0.00", out.rows, err)
	}
}
