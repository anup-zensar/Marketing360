package planbudget

import (
	"context"
	"math"

	"mapp/internal/measure"
	"mapp/internal/platform/actor"
	"mapp/internal/platform/date"
	"mapp/internal/platform/money"
	"mapp/internal/platform/sheet"
)

const (
	widgetCards   = "cards"
	widgetPeriods = "periods"
	widgetRanked  = "ranked"
	noValue       = "—"
)

func levelWords() map[string]string {
	return map[string]string{"plan": "Plan", "campaign": "Campaign", "initiative": "Initiative", "activity": "Activity",
		"lineItem": "Line item"}
}

// sheets are the report's widgets as 039's export writes them; the ranked list has every page (FR-016).
func (s *Service) sheets(ctx context.Context, a actor.Actor, f measure.Filter) ([]measure.WidgetSheet, error) {
	q, err := ParseRanked(f.Own)
	if err != nil {
		return nil, err
	}
	q.Page, q.Size = 1, math.MaxInt32
	var out []measure.WidgetSheet
	err = s.tx.Read(ctx, func(ctx context.Context) error {
		r, recordsErr := s.records(ctx, a, f)
		if recordsErr != nil {
			return recordsErr
		}
		periods, _ := r.periods(periodOf(f), date.Today(s.now()))
		out = []measure.WidgetSheet{
			{Key: widgetCards, Sheet: cardsSheet(r.cards())},
			{Key: widgetPeriods, Sheet: periodsSheet(periods)},
			{Key: widgetRanked, Sheet: rankedSheet(r.ranked(levelOf(f), q))},
		}
		return nil
	})
	return out, err
}

func cardsSheet(c Cards) sheet.Sheet {
	rows := [][]sheet.Cell{{sheet.Bold("Figure"), sheet.Bold("Value")}}
	for _, row := range []struct {
		label string
		value sheet.Cell
	}{
		{"Budget", moneyCell(c.Budget)}, {"Planned cost", moneyCell(c.Planned)}, {"Committed", moneyCell(c.Committed)},
		{"Actual", moneyCell(c.Actual)}, {"Budget used (%)", percentCell(c.BudgetUsed)},
		{"Left to plan", moneyCell(c.LeftToPlan)}, {"Unspent budget", moneyCell(c.Unspent)},
		{"Budget spent (%)", percentCell(c.BudgetSpent)}, {"Not yet committed", moneyCell(c.NotYetCommitted)},
		{"% committed", percentCell(c.CommittedShare)},
	} {
		rows = append(rows, []sheet.Cell{sheet.Text(row.label), row.value})
	}
	return sheet.Sheet{Name: "Cards", Rows: rows}
}

func periodsSheet(periods []Period) sheet.Sheet {
	rows := [][]sheet.Cell{boldRow("Period", "Budget", "Planned cost", "Actual", "Budget to date", "Planned cost to date",
		"Actual to date")}
	for _, p := range periods {
		rows = append(rows, []sheet.Cell{sheet.Text(p.Key), sheet.Money(p.Budget.Amount), sheet.Money(p.Planned.Amount),
			moneyCell(p.Actual), sheet.Money(p.BudgetToDate.Amount), sheet.Money(p.PlannedToDate.Amount),
			moneyCell(p.ActualToDate)})
	}
	if len(periods) > 0 {
		last := periods[len(periods)-1]
		rows = append(rows, []sheet.Cell{sheet.Bold("Total"), sheet.Money(last.BudgetToDate.Amount),
			sheet.Money(last.PlannedToDate.Amount), moneyCell(lastActual(periods))})
	}
	return sheet.Sheet{Name: "Periods", Rows: rows}
}

func lastActual(periods []Period) *money.Money {
	var last *money.Money
	for _, p := range periods {
		if p.ActualToDate != nil {
			last = p.ActualToDate
		}
	}
	return last
}

func rankedSheet(r Ranked) sheet.Sheet {
	rows := [][]sheet.Cell{boldRow("Group", "Level", "Name", "Budget", "Planned cost", "Committed", "Actual",
		"Left to plan", "Unspent budget", "Over-allocated")}
	for _, g := range r.Groups {
		if g.Label != "" {
			rows = append(rows, totalsRow(g))
		}
		for _, e := range g.Entries {
			rows = append(rows, entryRow(g.Label, e, ""))
			for _, child := range e.Children {
				rows = append(rows, entryRow(g.Label, child, "  "))
			}
		}
	}
	return sheet.Sheet{Name: "Ranked list", Rows: rows}
}

func totalsRow(g Group) []sheet.Cell {
	f := g.Totals
	return []sheet.Cell{sheet.Bold(g.Label), sheet.Bold("Total"), sheet.Text(""), moneyCell(f.Budget), moneyCell(f.Planned),
		moneyCell(f.Committed), moneyCell(f.Actual), moneyCell(f.LeftToPlan), moneyCell(f.Unspent), sheet.Text("")}
}

func entryRow(group string, e Entry, indent string) []sheet.Cell {
	f := e.Figures
	over := ""
	if f.OverAllocated {
		over = "Over-allocated"
	}
	return []sheet.Cell{sheet.Text(group), sheet.Text(levelWords()[e.Level]), sheet.Text(indent + e.Name),
		moneyCell(f.Budget), moneyCell(f.Planned), moneyCell(f.Committed), moneyCell(f.Actual), moneyCell(f.LeftToPlan),
		moneyCell(f.Unspent), sheet.Text(over)}
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

func percentCell(p *int) sheet.Cell {
	if p == nil {
		return sheet.Text(noValue)
	}
	return sheet.Number(int64(*p))
}
