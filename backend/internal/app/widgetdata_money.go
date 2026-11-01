package app

import (
	"context"

	"mapp/internal/budget"
	"mapp/internal/financials"
	"mapp/internal/platform/actor"
	"mapp/internal/widgetdata"
)

var transactionWords = map[financials.Type]string{ //nolint:gochecknoglobals // the source's type names
	financials.TypePurchaseOrder: "Purchase order line", financials.TypeInvoice: "Invoice",
	financials.TypePayment: "Payment", financials.TypeAccrual: "Accrual",
}

// transactionsReader answers the GSAP transactions of the line items whose plan's financial data a sees.
type transactionsReader struct {
	planning   widgetPlanning
	financials *financials.Service
}

func (r transactionsReader) Rows(ctx context.Context, a actor.Actor, f widgetdata.Prefilter) (widgetdata.Rows, error) {
	d, err := r.planning.read(ctx, a, f, false)
	if err != nil {
		return widgetdata.Rows{}, err
	}
	out := widgetdata.Rows{Items: []widgetdata.Row{}, Notes: d.notes}
	ids := d.financeLineItems()
	if len(ids) == 0 {
		return out, nil
	}
	linked, err := r.financials.TransactionsOnLineItems(ctx, a, ids)
	if err != nil {
		return widgetdata.Rows{}, err
	}
	fields := widgetdata.BuiltInFields("financialTransactions")
	for _, t := range linked {
		if within(&t.PostingDate, f) {
			out.Items = append(out.Items, widgetdata.RowOf(fields, d.transactionRow(t)))
		}
	}
	return out, nil
}

func (d *planningData) transactionRow(t financials.LinkedTransaction) values {
	lineItem := d.byID[t.LineItemID]
	activity := d.byID[parentOf(lineItem)]
	v := values{
		fieldType: text(transactionWords[t.Type]), fieldPostingDate: widgetdata.DateValue(&t.PostingDate),
		fieldAmount: widgetdata.MoneyValue(t.USD), fieldVendor: text(t.VendorName), fieldDocumentNumber: text(t.DocumentNumber),
		fieldLineItem: d.nameOf(t.LineItemID), fieldActivity: d.nameOf(parentOf(lineItem)),
		fieldCampaign: d.parentName(d.byID[parentOf(activity)]),
	}
	if lineItem != nil {
		v[fieldPlan] = d.nameOf(lineItem.PlanID)
		v[fieldTeam] = widgetdata.TextsValue(d.teamNamesOf(d.byID[lineItem.PlanID].TeamIDs))
	}
	return v
}

// globalBudgetsReader answers the global budgets as the Global budgets page lists them (014).
type globalBudgetsReader struct {
	budgets *budget.Service
}

func (r globalBudgetsReader) Rows(ctx context.Context, a actor.Actor, _ widgetdata.Prefilter) (widgetdata.Rows, error) {
	page, err := r.budgets.List(ctx, a, nil)
	if err != nil {
		return widgetdata.Rows{}, err
	}
	fields := widgetdata.BuiltInFields("globalBudgets")
	out := widgetdata.Rows{Items: make([]widgetdata.Row, 0, len(page.Items))}
	for _, b := range page.Items {
		f := b.Figures
		v := values{fieldGlobalBudget: text(b.Name), fieldYear: widgetdata.WholeValue(int64(b.Year)),
			fieldAmount: widgetdata.MoneyValue(&f.Amount.Amount), fieldAllocated: widgetdata.MoneyValue(&f.Allocated.Amount),
			fieldUnallocated: widgetdata.MoneyValue(&f.Unallocated.Amount), fieldPlannedCost: widgetdata.MoneyValue(&f.PlannedCost.Amount)}
		if b.Owner != nil {
			v[fieldOwner] = text(b.Owner.Name)
		}
		out.Items = append(out.Items, widgetdata.RowOf(fields, v))
	}
	return out, nil
}
