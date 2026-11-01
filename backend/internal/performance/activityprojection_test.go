package performance_test

import (
	"testing"
	"time"

	"mapp/internal/performance"
	"mapp/internal/platform/date"
	"mapp/internal/platform/money"
	"mapp/internal/projection"
)

func projectingModel() projection.Model {
	return projection.Model{
		Name:         "Lubricants 1.0",
		InqToMqlRate: money.PercentFromHundredths(200), MqlToDealRate: money.PercentFromHundredths(500),
		AverageDealSize:        money.USD(money.FromCents(1_000_000)),
		InqToMqlVelocityMonths: 1, MqlToDealVelocityMonths: 2,
		ParametersChangedAt: time.Date(2026, 1, 2, 9, 0, 0, 0, time.UTC),
	}
}

// projectingFixture is the panel fixture with a goal the model can project from: 365,000 INQ across the
// whole of the plan year, the SC-001 table's own numbers.
func projectingFixture(t *testing.T) (*fixture, performance.Source) {
	t.Helper()
	f, src := panelFixture(t)
	inq := int64(365_000)
	start, end := date.MustParse("2026-01-01"), date.MustParse("2026-12-31")
	f.service.UseFigures(figuresStub{known: true, goals: performance.GoalFacts{
		INQ: &inq, PlanYear: 2026, ModelID: "MOD-1", Model: projectingModel(),
		StartDate: &start, EndDate: &end,
	}})
	return f, src
}

func TestActivityPanel_ProjectsEachStageAndStampsTheModel_FR012_FR016(t *testing.T) {
	t.Parallel()
	f, _ := projectingFixture(t)

	panel, err := f.service.ActivityPanel(t.Context(), alex, "A-1")
	if err != nil {
		t.Fatalf("ActivityPanel: %v", err)
	}

	if panel.Projection == nil || panel.Projection.ModelName != "Lubricants 1.0" {
		t.Fatalf("stamp = %+v, want the model named so a reviewer can reproduce the figure", panel.Projection)
	}
	inq := stageOf(t, panel, "inq")
	if inq.Projected == nil || len(inq.Months) != projection.Months {
		t.Fatalf("INQ = %+v, want a projected figure and the plan year's months", inq)
	}
	// The clock sits in March 2026, so January and February are behind today and the rest is still expected.
	if *inq.Projected != 365_000-31_000-28_000 {
		t.Errorf("projected INQ = %d, want the goal less the two whole months already behind today (FR-010)", *inq.Projected)
	}
	if panel.Revenue.Projected == nil || panel.Revenue.Projected.Currency != "USD" {
		t.Errorf("projected revenue = %+v, want the deals priced at the model's average deal size", panel.Revenue.Projected)
	}
	assertSplitAtToday(t, inq)
}

// assertSplitAtToday checks the one rule the months carry: actual behind today, projected ahead of it, both
// in the month in progress (FR-010).
func assertSplitAtToday(t *testing.T, figure performance.StageFigure) {
	t.Helper()
	if march := figure.Months[2]; !march.InProgress || march.Actual == nil || march.Projected == nil {
		t.Errorf("March = %+v, want the month in progress to carry both figures", march)
	}
	if january := figure.Months[0]; january.Projected != nil || january.Actual == nil {
		t.Errorf("January = %+v, want a whole month behind today to be actual only", january)
	}
}

func TestActivityPanel_ProjectionsAreDerivedNeverStored_FR011_SC002(t *testing.T) {
	t.Parallel()
	f, _ := projectingFixture(t)
	before, err := f.service.ActivityPanel(t.Context(), alex, "A-1")
	if err != nil {
		t.Fatalf("ActivityPanel: %v", err)
	}

	doubled := projectingModel()
	doubled.InqToMqlRate = money.PercentFromHundredths(400)
	inq := int64(365_000)
	start, end := date.MustParse("2026-01-01"), date.MustParse("2026-12-31")
	f.service.UseFigures(figuresStub{known: true, goals: performance.GoalFacts{
		INQ: &inq, PlanYear: 2026, ModelID: "MOD-1", Model: doubled, StartDate: &start, EndDate: &end,
	}})

	after, err := f.service.ActivityPanel(t.Context(), alex, "A-1")
	if err != nil {
		t.Fatalf("ActivityPanel after the rate change: %v", err)
	}
	wasMQL, nowMQL := stageOf(t, before, "mql"), stageOf(t, after, "mql")
	if wasMQL.Projected == nil || nowMQL.Projected == nil || *nowMQL.Projected <= *wasMQL.Projected {
		t.Errorf("projected MQL went from %v to %v, want doubling the rate to show at once: nothing is stored",
			wasMQL.Projected, nowMQL.Projected)
	}
}

func TestActivityPanel_WithoutAGoalNothingIsProjected_FR013(t *testing.T) {
	t.Parallel()
	f, _ := panelFixture(t)
	f.service.UseFigures(figuresStub{known: true, goals: performance.GoalFacts{PlanYear: 2026, ModelID: "MOD-1"}})

	panel, err := f.service.ActivityPanel(t.Context(), alex, "A-1")
	if err != nil {
		t.Fatalf("ActivityPanel: %v", err)
	}

	if panel.Projection != nil {
		t.Errorf("stamp = %+v, want none: there is nothing to reproduce", panel.Projection)
	}
	inq := stageOf(t, panel, "inq")
	if inq.Projected != nil || inq.Reason == "" {
		t.Errorf("INQ = %+v, want no projected figure and a reason, never a zero", inq)
	}
}

func stageOf(t *testing.T, panel performance.Panel, stage string) performance.StageFigure {
	t.Helper()
	for _, figure := range panel.Stages {
		if figure.Stage == stage {
			return figure
		}
	}
	t.Fatalf("the panel carries no %s stage: %+v", stage, panel.Stages)
	return performance.StageFigure{}
}

func TestActivityPanel_TheActualSeriesCountsOnlyCountedFacts_FR014(t *testing.T) {
	t.Parallel()
	f, src := projectingFixture(t)
	f.runImport(t, src.ID, map[string]string{"r1": "40"})

	panel, err := f.service.ActivityPanel(t.Context(), alex, "A-1")
	if err != nil {
		t.Fatalf("ActivityPanel: %v", err)
	}

	inq := stageOf(t, panel, "inq")
	february := inq.Months[1]
	if february.Actual == nil || *february.Actual != 40 {
		t.Fatalf("February actual = %v, want the 40 the loaded fact carries in its own month", february.Actual)
	}
	counted := int64(0)
	for _, month := range inq.Months {
		if month.Actual != nil {
			counted += *month.Actual
		}
	}
	if counted != 40 {
		t.Errorf("the actual series totals %d, want only 046's counted facts — held, unmatched and replaced ones count nowhere", counted)
	}
}

func TestActivityPanel_TodayComesFromTheInjectedClock_FR017(t *testing.T) {
	t.Parallel()
	f, _ := projectingFixture(t)

	panel, err := f.service.ActivityPanel(t.Context(), alex, "A-1")
	if err != nil {
		t.Fatalf("ActivityPanel: %v", err)
	}

	// The fixture clock stands at 12 March 2026, so March is the month in progress and nothing else is.
	inq := stageOf(t, panel, "inq")
	for i, month := range inq.Months {
		if month.InProgress != (i == 2) {
			t.Errorf("%s inProgress = %t, want only March to be in progress under the injected clock",
				month.Month, month.InProgress)
		}
	}
}
