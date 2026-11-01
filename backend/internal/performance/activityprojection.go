package performance

import (
	"context"
	"time"

	"mapp/internal/platform/date"
	"mapp/internal/platform/money"
	"mapp/internal/projection"
)

// withProjection adds 047's figures to the panel: each stage's projected volume and the plan year split at
// today, projected revenue, and the stamp naming the model behind them (FR-010, FR-012, FR-016). A panel
// whose activity has no goal keeps its reasons and gains nothing.
func (s *Service) withProjection(ctx context.Context, panel Panel, in panelInput) Panel {
	forecast := projection.Project(projection.Inputs{
		Goals: projection.Goals{INQ: in.goals.INQ, MQL: in.goals.MQL, Revenue: in.goals.Revenue,
			SetLaterStagesMyself: in.goals.SetLaterStagesMyself},
		Model: in.goals.Model, PlanYear: in.goals.PlanYear,
		StartDate: in.goals.StartDate, EndDate: in.goals.EndDate, Ongoing: in.goals.Ongoing,
		Today: date.Today(s.now()),
	})
	if !forecast.Present {
		return panel
	}
	stamp := forecast.Stamp
	panel.Projection = &stamp
	today, year := date.Today(s.now()), in.goals.PlanYear
	for i, figure := range panel.Stages {
		series, found := forecast.Find(figure.Stage)
		if !found {
			continue
		}
		split := projection.SplitCounts(monthlyCounts(in.totals, figure.Stage, year), series.Months, today, year)
		total := split.Projected
		panel.Stages[i].Projected, panel.Stages[i].Months = &total, split.Months
	}
	panel.Revenue = s.projectedRevenue(ctx, panel.Revenue, forecast, in)
	return panel
}

func (s *Service) projectedRevenue(ctx context.Context, figure RevenueFigure, forecast projection.Projection,
	in panelInput,
) RevenueFigure {
	today, year := date.Today(s.now()), in.goals.PlanYear
	split := projection.SplitAmounts(s.monthlyRevenue(ctx, in.totals, year), forecast.Revenue, today, year)
	projected := split.Projected
	figure.Projected, figure.Months = &projected, split.Months
	return figure
}

// monthlyCounts lays a stage's counted facts out across the plan year's months, in the engine's own shape,
// so the actual and projected series line up month by month.
func monthlyCounts(totals []StageTotals, stage string, year int) []projection.MonthFigure {
	months := yearMonths(year)
	for _, total := range totals {
		if total.Stage != stage {
			continue
		}
		if i := monthIndex(total.Month, year); i >= 0 {
			months[i].Value += total.Counts
		}
	}
	return months
}

// monthlyRevenue converts each month's revenue facts once at the current rate, exactly as the panel's
// total does (046-C7): a month holding only unconvertible facts stays empty rather than guessing.
func (s *Service) monthlyRevenue(ctx context.Context, totals []StageTotals, year int) []projection.MoneyFigure {
	rates := s.currentRates(ctx)
	months := make([]projection.MoneyFigure, 0, projection.Months)
	sums := make([]money.Amount, projection.Months)
	for _, total := range totals {
		if total.Stage != stageRevenue {
			continue
		}
		i := monthIndex(total.Month, year)
		if i < 0 {
			continue
		}
		converted, err := rates.Convert(money.Money{Amount: money.FromCents(total.AmountMinor),
			Currency: total.Currency}, money.BaseCurrency)
		if err != nil {
			continue
		}
		sums[i] = sums[i].Add(converted.Amount)
	}
	for i, sum := range sums {
		months = append(months, projection.MoneyFigure{Month: monthLabel(year, i),
			Value: money.USD(sum)})
	}
	return months
}

func yearMonths(year int) []projection.MonthFigure {
	months := make([]projection.MonthFigure, 0, projection.Months)
	for i := range projection.Months {
		months = append(months, projection.MonthFigure{Month: monthLabel(year, i)})
	}
	return months
}

func monthLabel(year, index int) string {
	return time.Date(year, time.Month(index+1), 1, 0, 0, 0, 0, time.UTC).Format("2006-01")
}

// monthIndex reads a fact month ("2026-01") as its place in the plan year, or -1 when it belongs to
// another year.
func monthIndex(month string, year int) int {
	at, err := time.Parse("2006-01", month)
	if err != nil || at.Year() != year {
		return -1
	}
	return int(at.Month()) - 1
}
