package revenue

import (
	"context"
	"math"
	"net/url"

	"mapp/internal/measure"
	"mapp/internal/platform/actor"
	"mapp/internal/platform/errs"
	"mapp/internal/platform/money"
	"mapp/internal/platform/sheet"
)

const (
	widgetCards   = "cards"
	widgetPeriods = "periods"
	widgetRanked  = "ranked"
	noValue       = "—"

	sheetCards       = "Cards"
	sheetPeriods     = "Periods"
	sheetRanked      = "Ranked list"
	labelPerformance = "Performance (%)"
)

func levelWords() map[string]string {
	return map[string]string{"plan": "Plan", "campaign": "Campaign", "initiative": "Initiative", "activity": "Activity"}
}

// revenueSheets are Revenue's widgets as 039's export writes them; the ranked list has every page (FR-007).
func (s *Service) revenueSheets(ctx context.Context, a actor.Actor, f measure.Filter) ([]measure.WidgetSheet, error) {
	report, err := s.Report(ctx, a, f)
	if err != nil {
		return nil, err
	}
	q, err := unboundedRanked(f.Own, []string{SortActualRevenue, SortProjectedRevenue}, SortProjectedRevenue)
	if err != nil {
		return nil, err
	}
	ranked, err := s.RankedRevenue(ctx, a, f, q)
	if err != nil {
		return nil, err
	}
	return []measure.WidgetSheet{
		{Key: widgetCards, Sheet: revenueCardsSheet(report.Cards)},
		{Key: widgetPeriods, Sheet: revenuePeriodsSheet(report.Periods)},
		{Key: widgetRanked, Sheet: revenueRankedSheet(ranked)},
	}, nil
}

func unboundedRanked(own url.Values, sorts []string, defaultSort string) (RankedQuery, error) {
	var problems errs.Problems
	q := parseRanked(own, sorts, defaultSort, &problems)
	q.Page, q.Size = 1, math.MaxInt32
	return q, problems.Err()
}

func revenueCardsSheet(c RevenueCards) sheet.Sheet {
	rows := [][]sheet.Cell{{sheet.Bold("Figure"), sheet.Bold("Value")}}
	for _, row := range []struct {
		label string
		value sheet.Cell
	}{
		{"Revenue goal", moneyCardCell(c.Goal)}, {"Actual revenue", moneyCardCell(c.Actual)},
		{"Actual or projected", moneyCardCell(c.Projected)}, {labelPerformance, numberCardCell(c.Performance)},
	} {
		rows = append(rows, []sheet.Cell{sheet.Text(row.label), row.value})
	}
	return sheet.Sheet{Name: sheetCards, Rows: rows}
}

func revenuePeriodsSheet(periods []RevenuePeriod) sheet.Sheet {
	rows := [][]sheet.Cell{boldRow("Period", "Goal", "Actual", "Projected")}
	for _, p := range periods {
		rows = append(rows, []sheet.Cell{sheet.Text(p.Key), moneyCell(p.Goal), moneyCell(p.Actual), moneyCell(p.Projected)})
	}
	return sheet.Sheet{Name: sheetPeriods, Rows: rows}
}

func revenueRankedSheet(r RevenueRanked) sheet.Sheet {
	rows := [][]sheet.Cell{boldRow("Group", "Level", "Name", "Goal", "Actual", "Projected", labelPerformance,
		"Actual cost", "ROI (%)")}
	for _, g := range r.Groups {
		for _, e := range g.Entries {
			rows = append(rows, revenueEntryRow(g.Label, e, ""))
			for _, child := range e.Children {
				rows = append(rows, revenueEntryRow(g.Label, child, "  "))
			}
		}
	}
	return sheet.Sheet{Name: sheetRanked, Rows: rows}
}

func revenueEntryRow(group string, e RevenueEntry, indent string) []sheet.Cell {
	f := e.Figures
	return []sheet.Cell{sheet.Text(group), sheet.Text(levelWords()[e.Level]), sheet.Text(indent + e.Name),
		moneyCell(f.Goal), moneyCell(f.Actual), moneyCell(f.Projected), numberCell(f.Performance),
		moneyCell(f.ActualCost), numberCell(f.ROI)}
}

func boldRow(labels ...string) []sheet.Cell {
	row := make([]sheet.Cell, 0, len(labels))
	for _, label := range labels {
		row = append(row, sheet.Bold(label))
	}
	return row
}

func moneyCell(m *money.Money) sheet.Cell {
	if m == nil {
		return sheet.Text(noValue)
	}
	return sheet.Money(m.Amount)
}

func moneyCardCell(c MoneyCard) sheet.Cell { return moneyCell(c.Value) }

func numberCell(n *float64) sheet.Cell {
	if n == nil {
		return sheet.Text(noValue)
	}
	return sheet.Number(int64(*n))
}

func numberCardCell(c NumberCard) sheet.Cell { return numberCell(c.Value) }

// funnelSheets are Funnel stages' widgets as 039's export writes them (FR-007).
func (s *Service) funnelSheets(ctx context.Context, a actor.Actor, f measure.Filter) ([]measure.WidgetSheet, error) {
	report, err := s.FunnelReport(ctx, a, f, f.Own.Get(paramStage))
	if err != nil {
		return nil, err
	}
	q, err := unboundedRanked(f.Own, []string{SortActual, SortProjected}, SortProjected)
	if err != nil {
		return nil, err
	}
	q.Stage = report.Stage
	ranked, err := s.RankedFunnel(ctx, a, f, q)
	if err != nil {
		return nil, err
	}
	return []measure.WidgetSheet{
		{Key: widgetCards, Sheet: funnelCardsSheet(report.Stage, report.Cards)},
		{Key: widgetPeriods, Sheet: funnelPeriodsSheet(report.Periods)},
		{Key: widgetRanked, Sheet: funnelRankedSheet(ranked)},
	}, nil
}

func funnelCardsSheet(stage string, c FunnelCards) sheet.Sheet {
	name := map[string]string{StageINQ: "INQ", StageMQL: "MQL"}[stage]
	rows := [][]sheet.Cell{{sheet.Bold("Figure"), sheet.Bold("Value")}}
	for _, row := range []struct {
		label string
		value sheet.Cell
	}{
		{name + " goal", numberCardCell(c.Goal)}, {name + " actual", numberCardCell(c.Actual)},
		{name + " actual or projected", numberCardCell(c.Projected)},
		{labelPerformance, numberCardCell(c.Performance)}, {"Conversion (%)", numberCardCell(c.Conversion)},
	} {
		rows = append(rows, []sheet.Cell{sheet.Text(row.label), row.value})
	}
	return sheet.Sheet{Name: sheetCards, Rows: rows}
}

func funnelPeriodsSheet(periods []FunnelPeriod) sheet.Sheet {
	rows := [][]sheet.Cell{boldRow("Period", "Goal", "Actual", "Projected")}
	for _, p := range periods {
		rows = append(rows, []sheet.Cell{sheet.Text(p.Key), int64Cell(p.Goal), int64Cell(p.Actual), int64Cell(p.Projected)})
	}
	return sheet.Sheet{Name: sheetPeriods, Rows: rows}
}

func funnelRankedSheet(r FunnelRanked) sheet.Sheet {
	rows := [][]sheet.Cell{boldRow("Group", "Level", "Name", "Goal", "Actual", "Projected", labelPerformance,
		"Conversion (%)")}
	for _, g := range r.Groups {
		for _, e := range g.Entries {
			rows = append(rows, funnelEntryRow(g.Label, e, ""))
			for _, child := range e.Children {
				rows = append(rows, funnelEntryRow(g.Label, child, "  "))
			}
		}
	}
	return sheet.Sheet{Name: sheetRanked, Rows: rows}
}

func funnelEntryRow(group string, e FunnelEntry, indent string) []sheet.Cell {
	f := e.Figures
	return []sheet.Cell{sheet.Text(group), sheet.Text(levelWords()[e.Level]), sheet.Text(indent + e.Name),
		int64Cell(f.Goal), int64Cell(f.Actual), int64Cell(f.Projected), numberCell(f.Performance), numberCell(f.Conversion)}
}

func int64Cell(n *int64) sheet.Cell {
	if n == nil {
		return sheet.Text(noValue)
	}
	return sheet.Number(*n)
}
