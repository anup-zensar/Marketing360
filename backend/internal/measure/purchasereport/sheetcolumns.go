package purchasereport

import "mapp/internal/platform/sheet"

func columnHeads(columns []string, currency string) []string {
	labels := map[string]string{colID: "ID", colStatus: "Status", colAmount: "Amount (" + currency + ")", colOwner: "Owner",
		colCreated: "Created on", colDelivery: "Delivery date", colSent: "Sent on", colPONumber: "Purchase order",
		colPOLine: "PO line", colPOLineStatus: "PO line status", colInvoiced: "Invoiced (" + currency + ")",
		colWorkflow: "Approval workflow"}
	heads := make([]string, 0, len(columns))
	for _, c := range columns {
		heads = append(heads, labels[c])
	}
	return heads
}

// rowCells gives each chosen column the line's own value where it has one, and the request's otherwise.
func rowCells(r RequestNode, l LineNode, columns []string) []sheet.Cell {
	cells := make([]sheet.Cell, 0, len(columns))
	for _, c := range columns {
		if cell, ok := requestCell(r, c); ok {
			cells = append(cells, cell)
			continue
		}
		cells = append(cells, lineCell(l, c))
	}
	return cells
}

func requestCell(r RequestNode, column string) (sheet.Cell, bool) {
	switch column {
	case colID:
		return sheet.Text(r.ID), true
	case colStatus:
		return sheet.Text(statusLabel(r.Status)), true
	case colOwner:
		return sheet.Text(r.Owner.Name), true
	case colCreated:
		return sheet.Date(r.CreatedOn), true
	case colWorkflow:
		return textCell(r.Workflow), true
	}
	return sheet.Cell{}, false
}

func lineCell(l LineNode, column string) sheet.Cell {
	switch column {
	case colAmount:
		return moneyCell(l.Amount)
	case colDelivery:
		return dateCell(l.DeliveryDate)
	case colSent:
		return dateCell(l.SentOn)
	case colPONumber:
		return textCell(l.PONumber)
	case colPOLine:
		return textCell(l.POLine)
	case colPOLineStatus:
		return textCell(l.POLineStatus)
	case colInvoiced:
		return moneyCell(l.Invoiced)
	}
	return sheet.Text(dash)
}

func statusLabel(s Status) string {
	switch s {
	case StatusCreated:
		return "Created"
	case StatusSubmitted:
		return "Submitted"
	case StatusInReview:
		return "In review"
	case StatusDeclined:
		return "Declined"
	case StatusApproved:
		return "Approved"
	case StatusClosed:
		return "Closed"
	}
	return string(s)
}
