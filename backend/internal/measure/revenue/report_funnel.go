package revenue

import (
	"context"

	"mapp/internal/measure"
	"mapp/internal/platform/actor"
	"mapp/internal/platform/date"
	"mapp/internal/projection"
)

// FunnelReport is one funnel stage against goal (US2): the stage selector, its cards and period series.
func (s *Service) FunnelReport(ctx context.Context, a actor.Actor, f measure.Filter, stage string) (FunnelReport, error) {
	var out FunnelReport
	err := s.tx.Read(ctx, func(ctx context.Context) error {
		sc, views, err := s.scopeAndActivities(ctx, a, f, categoryOf(f.Own))
		if err != nil {
			return err
		}
		available := availableStages(views)
		chosen := stageOf(stage, available)
		w := window{today: date.Today(s.now()), year: sc.Year}
		out = buildFunnelReport(funnelReportInput{views: views, stage: chosen, available: available, period: f.Period, w: w})
		return nil
	})
	return out, err
}

// availableStages lists every stage any activity's model allows, in funnel order (049-C11); deals is never
// offered (049-C12).
func availableStages(views []activityView) []string {
	seen := map[string]bool{}
	for _, v := range views {
		for _, stage := range []string{StageINQ, StageMQL} {
			if _, found := v.projection.Find(stage); found {
				seen[stage] = true
			}
		}
	}
	stages := make([]string, 0, 2)
	for _, stage := range []string{StageINQ, StageMQL} {
		if seen[stage] {
			stages = append(stages, stage)
		}
	}
	return stages
}

func stageOf(requested string, available []string) string {
	for _, s := range available {
		if s == requested {
			return requested
		}
	}
	if len(available) > 0 {
		return available[0]
	}
	return StageINQ
}

// funnelReportInput groups FunnelReport's working parts so building it stays within the argument limit.
type funnelReportInput struct {
	views     []activityView
	stage     string
	available []string
	period    string
	w         window
}

func buildFunnelReport(in funnelReportInput) FunnelReport {
	sums, uncovered := funnelMonths(in.views, in.stage, in.w)
	periods := funnelPeriods(sums, in.period, in.w)
	anyGoal := false
	for _, v := range in.views {
		if _, found := v.projection.Find(in.stage); found && v.projection.Present {
			anyGoal = true
			break
		}
	}
	out := FunnelReport{Stage: in.stage, AvailableStages: in.available, UncoveredCount: uncovered,
		Period: periodShown(in.period), Periods: periods}
	out.Cards = funnelCards(periods, anyGoal, previousStageTotal(in.views, in.stage, in.w))
	if stamp := stampOf(in.views); stamp != nil {
		out.Projection = stamp
	}
	if len(in.views) == 0 {
		out.Reason = ReasonNoData
	} else if !anyGoal {
		out.Reason = ReasonNoGoal
	}
	return out
}

// funnelMonthlySums is the twelve-month totals of the selected stage's volume: actual, projected and goal.
type funnelMonthlySums struct {
	actual, projected, goal [projection.Months]int64
}

// funnelMonths sums the selected stage's actual, projected and goal volumes; uncovered counts the plans
// whose model does not allow this stage at all (049-C11, FR-018).
func funnelMonths(views []activityView, stage string, w window) (funnelMonthlySums, int) {
	var sums funnelMonthlySums
	coveredPlans, allPlans := map[string]bool{}, map[string]bool{}
	for _, v := range views {
		allPlans[v.node.PlanID] = true
		series, found := v.projection.Find(stage)
		if !found {
			continue
		}
		coveredPlans[v.node.PlanID] = true
		addFunnelMonths(&sums, v, series, w)
	}
	return sums, uncoveredCount(allPlans, coveredPlans)
}

func addFunnelMonths(sums *funnelMonthlySums, v activityView, series projection.Series, w window) {
	split := projection.SplitCounts(v.facts.counts[series.Stage], series.Months, w.today, w.year)
	for i, m := range split.Months {
		if m.Actual != nil {
			sums.actual[i] += *m.Actual
		}
		if m.Projected != nil {
			sums.projected[i] += *m.Projected
		}
	}
	for i, m := range series.Months {
		if i < len(sums.goal) {
			sums.goal[i] += m.Value
		}
	}
}

func uncoveredCount(allPlans, coveredPlans map[string]bool) int {
	uncovered := 0
	for plan := range allPlans {
		if !coveredPlans[plan] {
			uncovered++
		}
	}
	return uncovered
}

// previousStageTotal is the stage before the selected one's actual total, for conversion (049-C13).
func previousStageTotal(views []activityView, stage string, w window) *int64 {
	previous, found := previousStage(stage)
	if !found {
		return nil
	}
	var total int64
	hasPrevious := false
	for _, v := range views {
		series, found := v.projection.Find(previous)
		if !found {
			continue
		}
		split := projection.SplitCounts(v.facts.counts[previous], series.Months, w.today, w.year)
		total += split.Actual
		hasPrevious = true
	}
	if !hasPrevious {
		return nil
	}
	return &total
}

func previousStage(stage string) (string, bool) {
	if stage == StageMQL {
		return StageINQ, true
	}
	return "", false
}

// funnelPeriods groups the twelve months by month or quarter, splitting at today's period (047 FR-010,
// 049-C4).
func funnelPeriods(sums funnelMonthlySums, period string, w window) []FunnelPeriod {
	bounds := periodsOf(w, period)
	out := make([]FunnelPeriod, 0, len(bounds))
	for _, b := range bounds {
		out = append(out, funnelPeriodOf(sums, b))
	}
	return out
}

func funnelPeriodOf(sums funnelMonthlySums, b periodBounds) FunnelPeriod {
	var a, p, g int64
	hasActual, hasProjected := false, false
	for i := b.first; i < b.first+b.size; i++ {
		a, p, g = a+sums.actual[i], p+sums.projected[i], g+sums.goal[i]
		if b.todayIndex >= 0 && i <= b.todayIndex {
			hasActual = true
		}
		if b.todayIndex < 0 || i >= b.todayIndex {
			hasProjected = true
		}
	}
	fp := FunnelPeriod{Key: periodKey(b.year, b.first, b.size), InProgress: b.inProgress(),
		State: periodState(b.first, b.size, b.todayIndex)}
	if g != 0 {
		fp.Goal = &g
	}
	if hasActual {
		fp.Actual = &a
	}
	if hasProjected {
		fp.Projected = &p
	}
	return fp
}
