package gsaplink

import (
	"cmp"
	"slices"

	"mapp/internal/platform/date"
	"mapp/internal/platform/money"
)

type orderKey struct {
	lineItem, po string
}

// order adds up one purchase order's transactions on one line item, in USD.
type order struct {
	po, invoiced, paid    money.Amount
	lines                 []string
	date                  date.Date
	vendor, request, gsap string
}

func (o *order) add(t Transaction, usd money.Amount) {
	switch t.Type {
	case TypePurchaseOrder:
		o.po = o.po.Add(usd)
		o.lines = append(o.lines, t.LineNumber)
		if o.date.IsZero() || t.PostingDate.Before(o.date) {
			o.date = t.PostingDate
		}
		o.vendor, o.request, o.gsap = cmp.Or(o.vendor, t.Vendor), cmp.Or(o.request, t.PurchaseRequestID),
			cmp.Or(o.gsap, t.GsapPrNumber)
	case TypeInvoice:
		o.invoiced = o.invoiced.Add(usd)
	case TypePayment:
		o.paid = o.paid.Add(usd)
	}
}

// ordersTab gives one row per line item and purchase order: its PO lines, and the invoices (credit notes
// negative) and payments that refer to it, by 015's links (FR-011 – FR-013). Totals cover every row.
func (y year) ordersTab(transactions []Transaction, rates money.Rates) OrdersTab {
	orders, keys := map[orderKey]*order{}, []orderKey{}
	for _, t := range transactions {
		if !y.inView[t.LineItemID] || !y.seesFinance(y.byID[t.LineItemID].PlanID) {
			continue
		}
		k := orderKey{lineItem: t.LineItemID, po: t.PoNumber}
		if orders[k] == nil {
			orders[k], keys = &order{}, append(keys, k)
		}
		orders[k].add(t, usdOf(t.Amount, rates))
	}
	tab := OrdersTab{Rows: make([]OrderRow, 0, len(keys))}
	var total order
	for _, k := range keys {
		o := orders[k]
		if len(o.lines) == 0 {
			continue // invoices or payments of a purchase order whose lines are on another line item
		}
		li := y.byID[k.lineItem]
		slices.SortFunc(o.lines, compareIDs)
		tab.Rows = append(tab.Rows, OrderRow{LineItemID: li.ID, LineItem: li.Name, Path: y.path(li),
			PurchaseRequestID: o.request, GsapPrNumber: o.gsap, PoNumber: k.po, PoLines: slices.Compact(o.lines),
			PoDate: o.date, Vendor: o.vendor, PoAmount: y.inCurrency(o.po), Invoiced: y.inCurrency(o.invoiced),
			Paid: y.inCurrency(o.paid)})
		total.po, total.invoiced, total.paid = total.po.Add(o.po), total.invoiced.Add(o.invoiced), total.paid.Add(o.paid)
	}
	tab.Totals = OrderTotals{PoAmount: y.inCurrency(total.po), Invoiced: y.inCurrency(total.invoiced),
		Paid: y.inCurrency(total.paid)}
	return tab
}
