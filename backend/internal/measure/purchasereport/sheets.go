package purchasereport

import (
	"cmp"
	"slices"

	"mapp/internal/platform/date"
	"mapp/internal/platform/money"
	"mapp/internal/platform/sheet"
)

const dash = "—"

func moneyCell(m *money.Money) sheet.Cell {
	if m == nil {
		return sheet.Text(dash)
	}
	return sheet.Money(m.Amount)
}

func dateCell(d *date.Date) sheet.Cell {
	if d == nil {
		return sheet.Text(dash)
	}
	return sheet.Date(*d)
}

func textCell(s *string) sheet.Cell {
	if s == nil {
		return sheet.Text(dash)
	}
	return sheet.Text(*s)
}

func heads(names ...string) []sheet.Cell {
	row := make([]sheet.Cell, 0, len(names))
	for _, n := range names {
		row = append(row, sheet.Bold(n))
	}
	return row
}

func cardsSheet(c Cards, currency string) sheet.Sheet {
	return sheet.Sheet{Name: "Summary", Rows: [][]sheet.Cell{
		heads("Card", "Value"),
		{sheet.Text("Purchase requests"), sheet.Number(int64(c.PurchaseRequests))},
		{sheet.Text("Lines"), sheet.Number(int64(c.Lines))},
		{sheet.Text("Vendors"), sheet.Number(int64(c.Vendors))},
		{sheet.Text("Committed (" + currency + ")"), moneyCell(c.Committed)},
		{sheet.Text("Invoiced (" + currency + ")"), moneyCell(c.Invoiced)},
	}}
}

func statusSheet(m StatusMix, currency string) sheet.Sheet {
	rows := [][]sheet.Cell{heads("Status", "Count", "Amount ("+currency+")", "Share (%)")}
	for _, item := range m.Items {
		rows = append(rows, []sheet.Cell{sheet.Text(statusLabel(item.Status)), sheet.Number(int64(item.Count)),
			moneyCell(item.Amount), sheet.Percent(int64(item.SharePercent))})
	}
	rows = append(rows, []sheet.Cell{sheet.Bold("Total"), sheet.Number(int64(m.Total)), moneyCell(m.TotalAmount)})
	return sheet.Sheet{Name: "Purchase requests by status", Rows: rows}
}

func chartSheet(name, first string, c Chart) sheet.Sheet {
	head := []string{first}
	for _, m := range c.Months {
		head = append(head, monthLabel(m))
	}
	rows := [][]sheet.Cell{heads(append(head, "Total")...)}
	for _, s := range c.AllSeries {
		row := []sheet.Cell{sheet.Text(s.Name)}
		var total money.Amount
		for _, v := range s.Values {
			row = append(row, moneyCell(v))
			if v != nil {
				total = total.Add(v.Amount)
			}
		}
		rows = append(rows, append(row, sheet.Money(total)))
	}
	totals := []sheet.Cell{sheet.Bold("Total")}
	for _, t := range c.Totals {
		totals = append(totals, moneyCell(t))
	}
	return sheet.Sheet{Name: name, Rows: append(rows, append(totals, moneyCell(c.Committed)))}
}

// gridSheet is every matching line as one row, under its vendor, purchase request and line item (FR-027).
func (d data) gridSheet(rs []Request, q Query) sheet.Sheet {
	columns := slices.DeleteFunc(slices.Clone(columnsOr(q.Columns)), func(c string) bool { return c == colName })
	head := append([]string{"Vendor", "Purchase request", "Line item"}, columnHeads(columns, d.shown.code)...)
	rows := [][]sheet.Cell{heads(head...)}
	key, order := cmp.Or(q.Sort, colName), cmp.Or(q.Order, orderAsc)
	for _, g := range d.sorted(d.searched(groupsOf(rs), words(q.Search)), key, order) {
		node := d.vendorNode(g)
		for _, r := range node.Requests {
			rows = append(rows, requestRows(node, r, columns)...)
		}
	}
	return sheet.Sheet{Name: "Purchase requests", Rows: rows}
}

func requestRows(v VendorNode, r RequestNode, columns []string) [][]sheet.Cell {
	if len(r.Lines) == 0 {
		prefix := []sheet.Cell{sheet.Text(v.Label), sheet.Text(r.Name), sheet.Text(dash)}
		return [][]sheet.Cell{append(prefix, rowCells(r, LineNode{}, columns)...)}
	}
	rows := make([][]sheet.Cell, 0, len(r.Lines))
	for _, l := range r.Lines {
		prefix := []sheet.Cell{sheet.Text(v.Label), sheet.Text(r.Name), sheet.Text(l.LineItem.Name)}
		rows = append(rows, append(prefix, rowCells(r, l, columns)...))
	}
	return rows
}
