package plandetails

import (
	"context"
	"slices"

	"mapp/internal/measure"
	"mapp/internal/platform/actor"
	"mapp/internal/platform/sheet"
)

// The service is the report 039's registry lists and exports (measure.Report).

func (s *Service) Key() string     { return measure.KeyPlanDetails }
func (s *Service) Available() bool { return true }

// Sheets gives every widget of the level f.Show as one sheet, in screen order, with every value (FR-017).
func (s *Service) Sheets(ctx context.Context, a actor.Actor, f measure.Filter) ([]measure.WidgetSheet, error) {
	r, err := s.Report(ctx, a, f)
	if err != nil {
		return nil, err
	}
	widgets := slices.Clone(r.BuiltIn)
	for _, row := range r.Attributes {
		widgets = append(widgets, row.Widgets...)
	}
	sheets := make([]measure.WidgetSheet, 0, len(widgets))
	for _, w := range widgets {
		sheets = append(sheets, measure.WidgetSheet{Key: w.Key,
			Sheet: sheet.Sheet{Name: w.Title, Rows: tableOf(w, recordsWord(r.Show), r.Currency)}})
	}
	return sheets, nil
}

// tableOf is a widget's table view: Value, the records and, for planned cost, the amount; then the total row.
func tableOf(w Widget, records, currency string) [][]sheet.Cell {
	header := []sheet.Cell{sheet.Bold("Value"), sheet.Bold(records)}
	if w.Measure == MeasurePlannedCost {
		header = append(header, sheet.Bold("Planned cost ("+currency+")"))
	}
	rows := [][]sheet.Cell{header}
	for _, v := range slices.Concat(w.Values, []Value{w.Total}) {
		row := []sheet.Cell{sheet.Text(v.Label), sheet.Number(int64(v.Count))}
		if v.PlannedCost != nil {
			row = append(row, sheet.Money(v.PlannedCost.Amount))
		}
		rows = append(rows, row)
	}
	return rows
}
