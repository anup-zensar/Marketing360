package projection_test

import (
	"slices"
	"testing"
	"time"

	"mapp/internal/platform/date"
	"mapp/internal/platform/money"
	"mapp/internal/projection"
)

var changedAt = time.Date(2026, 1, 2, 9, 0, 0, 0, time.UTC)

// lubricants is the domain model's example model (§4.5) with 047's velocities: 2 %, 5 %, 10,000 USD a
// deal, one month from INQ to MQL and two more to a deal.
func lubricants() projection.Model {
	return projection.Model{
		Name:         "Lubricants 1.0",
		InqToMqlRate: money.PercentFromHundredths(200), MqlToDealRate: money.PercentFromHundredths(500),
		AverageDealSize:        money.USD(money.FromCents(1_000_000)),
		InqToMqlVelocityMonths: 1, MqlToDealVelocityMonths: 2,
		ParametersChangedAt: changedAt,
	}
}

func day(s string) *date.Date {
	d := date.MustParse(s)
	return &d
}

func inputs(inq int64) projection.Inputs {
	return projection.Inputs{
		Goals:    projection.Goals{INQ: &inq},
		Model:    lubricants(),
		PlanYear: 2026,
		Today:    date.MustParse("2026-07-15"),
	}
}

func values(s projection.Series) []int64 {
	out := make([]int64, 0, len(s.Months))
	for _, m := range s.Months {
		out = append(out, m.Value)
	}
	return out
}

// SC-001: the hand-calculated table. A goal of 365,000 INQ over the whole of 2026 is 1,000 a day, so every
// month is its day count × 1,000 and each later stage can be checked by hand.
func TestProject_MatchesTheHandCalculatedTable_SC001(t *testing.T) {
	t.Parallel()
	in := inputs(365_000)
	in.StartDate, in.EndDate = day("2026-01-01"), day("2026-12-31")

	got := projection.Project(in)

	wantINQ := []int64{31_000, 28_000, 31_000, 30_000, 31_000, 30_000, 31_000, 31_000, 30_000, 31_000, 30_000, 31_000}
	// 2 % of each month, placed one month later; December's share falls outside the plan year and is dropped.
	wantMQL := []int64{0, 620, 560, 620, 600, 620, 600, 620, 620, 600, 620, 600}
	// 5 % of each MQL month, placed two months later; the last two fall outside the year.
	wantDeals := []int64{0, 0, 0, 31, 28, 31, 30, 31, 30, 31, 31, 30}
	for _, want := range []struct {
		stage  string
		months []int64
		total  int64
	}{
		{projection.StageINQ, wantINQ, 365_000},
		{projection.StageMQL, wantMQL, 6_680},
		{projection.StageDeals, wantDeals, 273},
	} {
		series, found := got.Find(want.stage)
		if !found || !slices.Equal(values(series), want.months) || series.Total != want.total {
			t.Errorf("%s = %v (total %d), want %v (total %d)", want.stage, values(series), series.Total,
				want.months, want.total)
		}
	}
	if got.Total.Amount.String() != "2730000.00" || got.Total.Currency != "USD" {
		t.Errorf("projected revenue = %s %s, want 2730000.00 USD (273 deals × 10,000)", got.Total.Amount, got.Total.Currency)
	}
	if got.Stamp.ModelName != "Lubricants 1.0" || !got.Stamp.ParametersChangedAt.Equal(changedAt) {
		t.Errorf("stamp = %+v, want the model and when its parameters last changed (FR-012)", got.Stamp)
	}
}

func TestProject_NothingLandsPastThePlanYear_FR008(t *testing.T) {
	t.Parallel()
	in := inputs(365_000)
	in.Model.InqToMqlVelocityMonths, in.Model.MqlToDealVelocityMonths = 12, 0

	got := projection.Project(in)

	mql, _ := got.Find(projection.StageMQL)
	if mql.Total != 0 {
		t.Errorf("MQL total = %d, want 0: a shift of a whole year leaves the plan year empty, never bunched into December", mql.Total)
	}
	if len(mql.Months) != projection.Months {
		t.Errorf("months = %d, want %d: the year keeps its shape", len(mql.Months), projection.Months)
	}
}

func TestProject_PhasesFromTheActivitysOwnDays_FR006(t *testing.T) {
	t.Parallel()
	// The goal belongs to the activity, so a shorter activity packs the same goal into fewer months: the
	// months outside it are empty and the year still holds the whole goal.
	tests := []struct {
		name    string
		start   *date.Date
		end     *date.Date
		ongoing bool
		wantJan bool
		wantDec bool
	}{
		{name: "FR-006 a whole year is phased across every day", wantJan: true, wantDec: true},
		{name: "FR-006 an activity starting late gets nothing before it", start: day("2026-07-01"), wantDec: true},
		{name: "FR-006 an activity ending early gets nothing after it", end: day("2026-06-30"), wantJan: true},
		{name: "FR-006 an ongoing activity runs to the year's end", start: day("2026-07-01"), end: day("2026-08-31"),
			ongoing: true, wantDec: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			in := inputs(365_000)
			in.StartDate, in.EndDate, in.Ongoing = tt.start, tt.end, tt.ongoing

			inq, _ := projection.Project(in).Find(projection.StageINQ)

			months := values(inq)
			if (months[0] > 0) != tt.wantJan || (months[11] > 0) != tt.wantDec || inq.Total != 365_000 {
				t.Errorf("INQ = %v (total %d), want January filled %t, December filled %t and the whole goal phased",
					months, inq.Total, tt.wantJan, tt.wantDec)
			}
		})
	}
}

func TestProject_AMidMonthStartGetsOnlyItsOwnDays_FR006(t *testing.T) {
	t.Parallel()
	in := inputs(12_000)
	in.StartDate = day("2026-01-21") // 11 days of January, 345 of the year

	inq, _ := projection.Project(in).Find(projection.StageINQ)

	months := values(inq)
	if months[0] != 12_000*11/345 || inq.Total != 12_000 {
		t.Errorf("INQ = %v (total %d), want January to carry its 11 days' share and the year to hold the whole goal",
			months, inq.Total)
	}
}

func TestProject_WithoutAGoalNothingIsProjected_FR013(t *testing.T) {
	t.Parallel()
	in := inputs(0)
	in.Goals = projection.Goals{}

	got := projection.Project(in)

	if got.Present || len(got.Stages) != 0 {
		t.Errorf("projection = %+v, want nothing to show: an activity without a goal shows a reason, never a zero", got)
	}
}

func TestProject_TypedStagesIgnoreTheModel_FR009(t *testing.T) {
	t.Parallel()
	in := inputs(365_000)
	mql := int64(365_000)
	revenue := money.USD(money.FromCents(36_500_000))
	in.Goals = projection.Goals{INQ: in.Goals.INQ, MQL: &mql, Revenue: &revenue, SetLaterStagesMyself: true}

	got := projection.Project(in)

	typedMQL, found := got.Find(projection.StageMQL)
	if !found || typedMQL.Total != 365_000 || typedMQL.Months[0].Value != 31_000 {
		t.Errorf("MQL = %v (total %d), want the typed goal phased on its own, not 2 %% of INQ one month later",
			values(typedMQL), typedMQL.Total)
	}
	if _, hasDeals := got.Find(projection.StageDeals); hasDeals {
		t.Error("a typed activity projected deals, want only the stages the activity itself types")
	}
	if got.Total != revenue {
		t.Errorf("revenue = %+v, want the typed revenue goal %+v", got.Total, revenue)
	}
}

func TestProject_AnActivityOutsideThePlanYearProjectsNothing(t *testing.T) {
	t.Parallel()
	in := inputs(365_000)
	in.StartDate, in.EndDate = day("2027-01-01"), day("2027-12-31")

	got := projection.Project(in)

	if !got.Present || len(got.Stages) != 0 {
		t.Errorf("projection = %+v, want a goal with no days in the plan year to project nothing", got)
	}
}
