package approvalreport

import (
	"mapp/internal/platform/date"
	"mapp/internal/platform/sheet"
)

func approvalsSheet(rows []Row, currency string) sheet.Sheet {
	out := [][]sheet.Cell{{sheet.Bold("Item ID"), sheet.Bold("Name"), sheet.Bold("Record type"), sheet.Bold("Approver"),
		sheet.Bold("Decided on"), sheet.Bold("Status"), sheet.Bold("Submitted on"), sheet.Bold("Level"),
		sheet.Bold("Purchase order"), sheet.Bold("Amount (" + currency + ")"), sheet.Bold("Owner"), sheet.Bold("Comments")}}
	for _, r := range rows {
		out = append(out, []sheet.Cell{sheet.Text(r.ItemID), sheet.Text(r.ItemName), sheet.Text(recordTypeLabel(r.RecordType)),
			text(approverName(r)), dateCell(r.DecidedOn), sheet.Text(outcomeLabel(r.Outcome)), sheet.Date(r.SubmittedOn),
			sheet.Number(int64(r.Level)), text(orDash(r.PurchaseOrderNumber)), amountCell(r), sheet.Text(r.Owner.Name),
			sheet.Text(r.Comments)})
	}
	return sheet.Sheet{Name: "Approvals", Rows: out}
}

func text(s string) sheet.Cell { return sheet.Text(s) }

func approverName(r Row) string {
	if r.Approver == nil {
		return NothingText
	}
	return r.Approver.Name
}

func orDash(s *string) string {
	if s == nil {
		return NothingText
	}
	return *s
}

func dateCell(d *date.Date) sheet.Cell {
	if d == nil {
		return sheet.Text(NothingText)
	}
	return sheet.Date(*d)
}

func amountCell(r Row) sheet.Cell {
	if r.Amount == nil {
		return sheet.Text(NothingText)
	}
	return sheet.Money(r.Amount.Amount)
}
