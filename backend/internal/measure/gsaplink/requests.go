package gsaplink

import (
	"slices"

	"mapp/internal/allwork"
	"mapp/internal/platform/money"
)

// waitingTab lists the purchase requests waiting for GSAP on line items in view (FR-008 – FR-010); amounts only
// for plans whose financial data the person may see (FR-004).
func (y year) waitingTab(requests []Request, rates money.Rates) WaitingTab {
	tab := WaitingTab{Rows: make([]WaitingRow, 0, len(requests))}
	var total money.Amount
	for _, r := range requests {
		li, found := y.lineItemOf(r)
		if !found {
			continue
		}
		days := workingDays(r.RoutedAt, y.now)
		row := WaitingRow{PurchaseRequestID: r.ID, Name: r.Name, Vendor: r.Vendor, Owner: r.Owner, Path: y.path(li),
			RoutedAt: r.RoutedAt, Routing: r.Routing, GsapPrNumber: r.GsapPrNumber, WaitingWorkingDays: days,
			OverTurnaround: days > gsapTurnaroundDays}
		if y.seesFinance(r.PlanID) {
			usd := requestUSD(r, rates)
			amount := y.inCurrency(usd)
			row.Amount, total = &amount, total.Add(usd)
		}
		tab.Rows = append(tab.Rows, row)
		if row.OverTurnaround {
			tab.OverTurnaround++
		}
	}
	tab.Waiting = len(tab.Rows)
	if y.coverage().Plans > 0 {
		amount := y.inCurrency(total)
		tab.Amount = &amount
	}
	return tab
}

// lineItemOf is the request's first line item in view, which places it in the plan tree.
func (y year) lineItemOf(r Request) (allwork.Node, bool) {
	i := slices.IndexFunc(r.Lines, func(l RequestLine) bool { return y.inView[l.LineItemID] })
	if i < 0 {
		return allwork.Node{}, false
	}
	n, found := y.byID[r.Lines[i].LineItemID]
	return n, found
}

func requestUSD(r Request, rates money.Rates) money.Amount {
	var usd money.Amount
	for _, l := range r.Lines {
		usd = usd.Add(usdOf(l.Amount, rates))
	}
	return usd
}
