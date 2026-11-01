package performance_test

import (
	"context"
	"testing"

	"mapp/internal/performance"
	"mapp/internal/platform/money"
)

type figuresStub struct {
	goals performance.GoalFacts
	known bool
}

func (f figuresStub) GoalFacts(_ context.Context, _ string) (performance.GoalFacts, bool, error) {
	return f.goals, f.known, nil
}

func panelFixture(t *testing.T) (*fixture, performance.Source) {
	t.Helper()
	f, src := reviewFixture(t)
	inq := int64(100)
	mql := int64(10)
	f.service.UseFigures(figuresStub{known: true, goals: performance.GoalFacts{
		INQ: &inq, MQL: &mql, PlanYear: 2026, ModelID: "MOD-1",
	}})
	f.service.UseRates(func(context.Context) (money.Rates, error) { return money.Rates{}, nil })
	f.store.Activities = map[string]performance.KeyFacts{"A-1": {}}
	return f, src
}

func TestActivityPanel_ANonApprovedActivityShowsTheReasonAndNoFigures_FR027(t *testing.T) {
	t.Parallel()
	f, src := panelFixture(t)
	f.store.Keys[performance.MatchedActivityID]["a-2"] = []string{"A-2"}
	f.store.Activities = map[string]performance.KeyFacts{"A-1": {}, "A-2": {}}
	rows := importRows(map[string]string{"r1": "5"})
	rows[0].Keys.ActivityID = "A-2"
	if _, err := f.service.RunImport(t.Context(), alex, performance.ImportInput{
		SourceID: src.ID, FileName: "held.csv", Origin: performance.OriginUpload,
		Headings: []string{"ref", "date", "inqs"}, Rows: rows,
	}); err != nil {
		t.Fatalf("RunImport: %v", err)
	}

	panel, err := f.service.ActivityPanel(t.Context(), alex, "A-2")
	if err != nil {
		t.Fatalf("ActivityPanel: %v", err)
	}
	if panel.State != performance.PanelNotApproved {
		t.Fatalf("state = %s, want the FR-027 gate", panel.State)
	}
	if len(panel.Stages) != 0 || len(panel.Sources) != 0 {
		t.Errorf("a gated panel carries %d stages and %d sources, want no figures at all (SC-008)",
			len(panel.Stages), len(panel.Sources))
	}
}

func TestActivityPanel_AnApprovedActivityShowsGoalActualAndConversion_FR029(t *testing.T) {
	t.Parallel()
	f, src := panelFixture(t)
	f.runImport(t, src.ID, map[string]string{"r1": "40", "r2": "40"})
	mqlRows := importRows(map[string]string{"m1": "8"})
	mqlRows[0].Readings[0].Measure = performance.Measure{Stage: "mql"}
	if _, err := f.service.RunImport(t.Context(), alex, performance.ImportInput{
		SourceID: src.ID, FileName: "mql.csv", Origin: performance.OriginUpload,
		Headings: []string{"ref", "date", "mqls"}, Rows: mqlRows,
	}); err != nil {
		t.Fatalf("RunImport: %v", err)
	}

	panel, err := f.service.ActivityPanel(t.Context(), alex, "A-1")
	if err != nil || panel.State != performance.PanelOK {
		t.Fatalf("ActivityPanel = %s (%v), want figures", panel.State, err)
	}
	assertFigures(t, panel)
}

func assertFigures(t *testing.T, panel performance.Panel) {
	t.Helper()
	stages := map[string]performance.StageFigure{}
	for _, stage := range panel.Stages {
		stages[stage.Stage] = stage
	}
	assertStageFigures(t, stages["inq"], stages["mql"])
	if len(panel.Sources) != 1 || panel.Sources[0].Freshness == "" {
		t.Errorf("sources = %+v, want the contributing source with its freshness", panel.Sources)
	}
}

func assertStageFigures(t *testing.T, inq, mql performance.StageFigure) {
	t.Helper()
	if inq.Actual == nil || *inq.Actual != 80 || inq.Goal == nil || *inq.Goal != 100 {
		t.Errorf("inq = %+v, want goal 100 and actual 80", inq)
	}
	if mql.Conversion == nil || *mql.Conversion != 10.0 {
		t.Errorf("mql conversion = %v, want 8/80 as 10.0 (one decimal)", mql.Conversion)
	}
	if inq.Conversion != nil {
		t.Errorf("inq conversion = %v, want none: the first stage has no previous one", inq.Conversion)
	}
}

func TestActivityPanel_EveryEmptyFigureStatesItsReason_FR032(t *testing.T) {
	t.Parallel()
	f, _ := panelFixture(t)

	panel, err := f.service.ActivityPanel(t.Context(), alex, "A-1")
	if err != nil || panel.State != performance.PanelOK {
		t.Fatalf("ActivityPanel = %s (%v)", panel.State, err)
	}
	for _, stage := range panel.Stages {
		if stage.Reason != "noSource" {
			t.Errorf("stage %s reason = %q, want noSource: no active source supplies it", stage.Stage, stage.Reason)
		}
		if stage.Actual != nil {
			t.Errorf("stage %s shows %d with no data, want no sample figure (SC-008)", stage.Stage, *stage.Actual)
		}
	}
	if panel.Revenue.Reason == "" {
		t.Error("the revenue block states no reason while empty")
	}
}

func TestActivityPanel_RevenueConvertsAndCountsTheExcluded_046C7(t *testing.T) {
	t.Parallel()
	f, src := panelFixture(t)
	f.store.Currencies = map[string]bool{"USD": true, "NOK": true}
	rows := importRows(map[string]string{})
	rows = append(rows, revenueRow("r1", "1000.00", "USD"), revenueRow("r2", "500.00", "NOK"))
	if _, err := f.service.RunImport(t.Context(), alex, performance.ImportInput{
		SourceID: src.ID, FileName: "revenue.csv", Origin: performance.OriginUpload,
		Headings: []string{"ref", "date", "revenue", "currency"}, Rows: rows,
	}); err != nil {
		t.Fatalf("RunImport: %v", err)
	}

	panel, err := f.service.ActivityPanel(t.Context(), alex, "A-1")
	if err != nil {
		t.Fatalf("ActivityPanel: %v", err)
	}
	revenue := panel.Revenue
	if revenue.Actual == nil || revenue.Actual.Currency != "USD" || revenue.Actual.Amount.Format() != "1,000.00" {
		t.Errorf("revenue actual = %+v, want the USD fact converted and summed", revenue.Actual)
	}
	if revenue.ExcludedCount != 1 {
		t.Errorf("excluded = %d, want the NOK fact counted out: no rate exists for it (SC-010)", revenue.ExcludedCount)
	}
}

func revenueRow(reference, amount, currency string) performance.Row {
	return performance.Row{
		Raw:       map[string]string{"ref": reference},
		Reference: performance.Cell{Column: "ref", Value: reference},
		Date:      performance.Cell{Column: "date", Value: "2026-02-01"},
		Currency:  performance.Cell{Column: "currency", Value: currency},
		Keys:      performance.RowKeys{ActivityID: "A-1"},
		Readings: []performance.Reading{{
			Measure: performance.Measure{Stage: "revenue"},
			Cell:    performance.Cell{Column: "revenue", Value: amount},
		}},
	}
}

func TestActivityPanel_AFactCountsInTheYearOfItsOwnDate_FR015(t *testing.T) {
	t.Parallel()
	f, src := panelFixture(t)
	f.runImport(t, src.ID, map[string]string{"r1": "40"})
	lastYear := importRows(map[string]string{"r2": "99"})
	lastYear[0].Date = performance.Cell{Column: "date", Value: "2025-11-01"}
	if _, err := f.service.RunImport(t.Context(), alex, performance.ImportInput{
		SourceID: src.ID, FileName: "old.csv", Origin: performance.OriginUpload,
		Headings: []string{"ref", "date", "inqs"}, Rows: lastYear,
	}); err != nil {
		t.Fatalf("RunImport: %v", err)
	}

	panel, err := f.service.ActivityPanel(t.Context(), alex, "A-1")
	if err != nil {
		t.Fatalf("ActivityPanel: %v", err)
	}
	for _, stage := range panel.Stages {
		if stage.Stage == "inq" && (stage.Actual == nil || *stage.Actual != 40) {
			t.Fatalf("inq actual = %v, want 40: the 2025 fact counts in its own year, not the plan year 2026", stage.Actual)
		}
	}
}
