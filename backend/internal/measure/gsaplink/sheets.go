package gsaplink

import (
	"context"
	"strings"

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
	widgetSwbs    = "swbs"
	widgetBands   = "swbsBands"
	widgetWaiting = "waiting"
	widgetOrders  = "orders"
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
		sheets, readErr = s.sheets(ctx, y, t)
		return readErr
	})
	if err != nil {
		return nil, err
	}
	return sheets, nil
}

func (s *Service) sheets(ctx context.Context, y year, t Table) ([]measure.WidgetSheet, error) {
	swbs, err := s.swbs(ctx, y, sortOf(t, TabSwbs))
	if err != nil {
		return nil, err
	}
	waiting, err := s.waiting(ctx, y, sortOf(t, TabWaiting))
	if err != nil {
		return nil, err
	}
	orders, err := s.orders(ctx, y, sortOf(t, TabOrders))
	if err != nil {
		return nil, err
	}
	code := y.scope.Currency.Code
	return []measure.WidgetSheet{
		{Key: widgetSwbs, Sheet: sheet.Sheet{Name: "SWBS not received", Rows: swbsTable(swbs.Rows)}},
		{Key: widgetBands, Sheet: sheet.Sheet{Name: "Waiting by time", Rows: bandsTable(swbs.Bands)}},
		{Key: widgetWaiting, Sheet: sheet.Sheet{Name: "Waiting for GSAP", Rows: waitingTable(waiting.Rows, code)}},
		{Key: widgetOrders, Sheet: sheet.Sheet{Name: "Purchase orders", Rows: ordersTable(orders, code)}},
	}, nil
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

// waitingTable leaves an amount the person may not see empty, as the screen shows "—".
func waitingTable(rows []WaitingRow, code string) [][]sheet.Cell {
	table := [][]sheet.Cell{boldRow("Purchase request ID", "Purchase request", "Vendor", "Owner", "Plan", "Campaign",
		"Initiative", "Activity", "Routed on", "Routing", "GSAP PR number", "Amount ("+code+")", "Waiting (working days)",
		"Over 2 working days")}
	for _, r := range rows {
		amount := sheet.Cell{}
		if r.Amount != nil {
			amount = sheet.Money(r.Amount.Amount)
		}
		table = append(table, []sheet.Cell{sheet.Text(r.PurchaseRequestID), sheet.Text(r.Name), sheet.Text(r.Vendor),
			sheet.Text(r.Owner), sheet.Text(r.Path.Plan), sheet.Text(r.Path.Campaign), sheet.Text(r.Path.Initiative),
			sheet.Text(r.Path.Activity), sheet.DateTime(r.RoutedAt), sheet.Text(routingWords(r.Routing)),
			sheet.Text(r.GsapPrNumber), amount, sheet.Number(int64(r.WaitingWorkingDays)), sheet.Text(yesNo(r.OverTurnaround))})
	}
	return table
}

func routingWords(routing string) string {
	if routing == RoutingModification {
		return "Modification"
	}
	return "Original"
}

// ordersTable ends with the totals of every row (FR-013).
func ordersTable(tab OrdersTab, code string) [][]sheet.Cell {
	table := [][]sheet.Cell{boldRow("Line item ID", "Line item", "Plan", "Campaign", "Initiative", "Activity",
		"Purchase request", "GSAP PR number", "PO number", "PO lines", "PO date", "Vendor", "PO amount ("+code+")",
		"Invoiced ("+code+")", "Paid ("+code+")")}
	for _, r := range tab.Rows {
		table = append(table, []sheet.Cell{sheet.Text(r.LineItemID), sheet.Text(r.LineItem), sheet.Text(r.Path.Plan),
			sheet.Text(r.Path.Campaign), sheet.Text(r.Path.Initiative), sheet.Text(r.Path.Activity),
			sheet.Text(r.PurchaseRequestID), sheet.Text(r.GsapPrNumber), sheet.Text(r.PoNumber),
			sheet.Text(strings.Join(r.PoLines, ", ")), sheet.Date(r.PoDate), sheet.Text(r.Vendor),
			sheet.Money(r.PoAmount.Amount), sheet.Money(r.Invoiced.Amount), sheet.Money(r.Paid.Amount)})
	}
	amounts := []sheet.Cell{sheet.Money(tab.Totals.PoAmount.Amount), sheet.Money(tab.Totals.Invoiced.Amount),
		sheet.Money(tab.Totals.Paid.Amount)}
	totals := append(make([]sheet.Cell, len(table[0])-len(amounts)), amounts...)
	totals[0] = sheet.Bold("Total")
	return append(table, totals)
}
