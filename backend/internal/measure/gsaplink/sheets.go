package gsaplink

import (
	"context"

	"mapp/internal/measure"
	"mapp/internal/platform/actor"
	"mapp/internal/platform/sheet"
)

// The report's own address parameters that saved views keep (039): the tab and its sort, never the page.
const (
	paramTab  = "tab"
	paramSort = "sort"
)

// The widgets of the export, in screen order (contract: swbs, swbsBands, waiting, orders).
const (
	widgetSwbs  = "swbs"
	widgetBands = "swbsBands"
)

func (s *Service) Key() string      { return measure.KeyGsapLink }
func (s *Service) Available() bool  { return true }
func (s *Service) Params() []string { return []string{paramTab, paramSort} }

// Sheets gives each widget with every row, the table of the address's tab in its sort, the others in theirs (FR-019).
func (s *Service) Sheets(ctx context.Context, a actor.Actor, f measure.Filter) ([]measure.WidgetSheet, error) {
	t, err := ParseTable(f.Own)
	if err != nil {
		return nil, err
	}
	t = withDefaults(t)
	var sheets []measure.WidgetSheet
	err = s.tx.Read(ctx, func(ctx context.Context) error {
		y, readErr := s.year(ctx, a, f)
		if readErr != nil {
			return readErr
		}
		swbs, readErr := s.swbs(ctx, y, sortOf(t, TabSwbs))
		sheets = []measure.WidgetSheet{
			{Key: widgetSwbs, Sheet: sheet.Sheet{Name: "SWBS not received", Rows: swbsTable(swbs.Rows)}},
			{Key: widgetBands, Sheet: sheet.Sheet{Name: "Waiting by time", Rows: bandsTable(swbs.Bands)}},
		}
		return readErr
	})
	if err != nil {
		return nil, err
	}
	return sheets, nil
}

func sortOf(t Table, tab string) string {
	if t.Tab == tab {
		return t.Sort
	}
	return DefaultSort(tab)
}

func swbsTable(rows []SwbsRow) [][]sheet.Cell {
	table := [][]sheet.Cell{boldRow("Line item ID", "Line item", "Plan", "Campaign", "Initiative", "Activity", "Owner",
		"Company code", "Approved on", "Waiting (hours)", "Over 48 hours")}
	for _, r := range rows {
		table = append(table, []sheet.Cell{sheet.Text(r.LineItemID), sheet.Text(r.LineItem), sheet.Text(r.Path.Plan),
			sheet.Text(r.Path.Campaign), sheet.Text(r.Path.Initiative), sheet.Text(r.Path.Activity), sheet.Text(r.Owner),
			sheet.Text(r.CompanyCode), sheet.DateTime(r.ApprovedAt), sheet.Number(int64(r.WaitingHours)),
			sheet.Text(yesNo(r.OverTurnaround))})
	}
	return table
}

func bandsTable(bands []Band) [][]sheet.Cell {
	table := [][]sheet.Cell{boldRow("Waiting", "Line items")}
	total := 0
	for _, b := range bands {
		table = append(table, []sheet.Cell{sheet.Text(b.Label), sheet.Number(int64(b.Count))})
		total += b.Count
	}
	return append(table, []sheet.Cell{sheet.Bold("Total"), sheet.Number(int64(total))})
}

func boldRow(labels ...string) []sheet.Cell {
	row := make([]sheet.Cell, 0, len(labels))
	for _, l := range labels {
		row = append(row, sheet.Bold(l))
	}
	return row
}

func yesNo(b bool) string {
	if b {
		return "Yes"
	}
	return "No"
}
