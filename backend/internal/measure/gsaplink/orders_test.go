package gsaplink

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"mapp/internal/allwork"
	"mapp/internal/measure"
	"mapp/internal/platform/date"
)

const (
	poShared = "4500001"
	poDE     = "4500002"
)

// withOrders: PO 4500001 has lines 10 and 20 on L-1, with an invoice, a credit note and a payment, and line 30 on L-2;
// L-2 also has an invoice of a PO with no line on it; PO 4500002 is on L-3 of the Germany plan.
func withOrders(w *world) *world {
	tx := func(id, kind, lineItem, po, line, amount, code string) Transaction {
		return Transaction{ID: id, Type: kind, LineItemID: lineItem, PoNumber: po, LineNumber: line,
			PostingDate: date.New(2026, 2, 1), Amount: amountIn(w.tb, amount, code)}
	}
	first := tx("FT-1", TypePurchaseOrder, "L-1", poShared, "20", "100.00", "EUR")
	first.Vendor, first.PurchaseRequestID, first.GsapPrNumber = "Northlight Events GmbH", "PR-1001", "1000777"
	second := tx("FT-2", TypePurchaseOrder, "L-1", poShared, "10", "50.00", "USD")
	second.PostingDate = date.New(2026, 1, 12)
	w.transactions = []Transaction{first, second,
		tx("FT-3", TypeInvoice, "L-1", poShared, "", "30.00", "USD"),
		tx("FT-4", TypeInvoice, "L-1", poShared, "", "-10.00", "USD"),
		tx("FT-5", TypePayment, "L-1", poShared, "", "20.00", "USD"),
		tx("FT-6", TypePurchaseOrder, "L-2", poShared, "30", "5.00", "USD"),
		tx("FT-7", TypeInvoice, "L-2", "4500009", "", "7.00", "USD"),
		tx("FT-8", TypePurchaseOrder, "L-3", poDE, "10", "40.00", "USD"),
	}
	return w
}

func ordersOf(w *world) OrdersTab {
	rates, _ := w.Rates(w.tb.Context())
	return w.year().ordersTab(w.transactions, rates)
}

func orderKeys(rows []OrderRow) []string {
	keys := []string{}
	for _, r := range rows {
		keys = append(keys, r.LineItemID+" "+r.PoNumber)
	}
	return keys
}

func TestOrders_SumsPerLineItemAndPO_FR012(t *testing.T) {
	t.Parallel()
	w := withOrders(twoPlans(t))

	want := OrderRow{LineItemID: "L-1", LineItem: "Stand build", Path: Path{PlanID: ukPlan, Plan: "UK plan",
		Campaign: "Launch", Initiative: "Retail", ActivityID: ukActivity, Activity: "Stand"}, PurchaseRequestID: "PR-1001",
		GsapPrNumber: "1000777", PoNumber: poShared, PoLines: []string{"10", "20"}, PoDate: date.New(2026, 1, 12),
		Vendor: "Northlight Events GmbH", PoAmount: amountIn(t, "250.00", "USD"), Invoiced: amountIn(t, "20.00", "USD"),
		Paid: amountIn(t, "20.00", "USD")}
	if diff := cmp.Diff(want, ordersOf(w).Rows[0], byValue()); diff != "" {
		t.Errorf("L-1's purchase order (-want +got):\n%s", diff)
	}
}

func TestOrders_PoOnTwoLineItems_FR011(t *testing.T) {
	t.Parallel()
	want := []string{"L-1 " + poShared, "L-2 " + poShared, "L-3 " + poDE}
	if diff := cmp.Diff(want, orderKeys(ordersOf(withOrders(twoPlans(t))).Rows)); diff != "" {
		t.Errorf("rows (-want +got):\n%s", diff)
	}
}

func TestOrders_TotalsAllPages_FR013(t *testing.T) {
	t.Parallel()
	r := withOrders(twoPlans(t)).report(Table{Tab: TabOrders, PageSize: 1})

	want := OrderTotals{PoAmount: amountIn(t, "295.00", "USD"), Invoiced: amountIn(t, "20.00", "USD"),
		Paid: amountIn(t, "20.00", "USD")}
	if diff := cmp.Diff(want, r.Orders.Totals, byValue()); diff != "" || r.Total != 3 || len(r.Orders.Rows) != 1 {
		t.Errorf("totals of %d rows on a page of %d (-want +got):\n%s", r.Total, len(r.Orders.Rows), diff)
	}
	if diff := cmp.Diff([]string{"L-3 " + poDE}, orderKeys(r.Orders.Rows)); diff != "" {
		t.Errorf("first page, by line item then PO (-want +got):\n%s", diff)
	}
}

func TestOrders_OnlyFinanceViewPlans_FR004(t *testing.T) {
	t.Parallel()
	w := withOrders(twoPlans(t))
	w.choices.FinancePlans = []string{ukPlan}

	r := w.report(Table{Tab: TabOrders})
	if diff := cmp.Diff([][]string{{"L-1", "L-2"}}, w.asked); diff != "" {
		t.Errorf("line items read (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff(&measure.FinanceCoverage{Plans: 1, Of: 2}, r.Finance); diff != "" || r.Total != 2 {
		t.Errorf("%d rows, coverage (-want +got):\n%s", r.Total, diff)
	}

	w.choices, w.asked = allwork.Choices{}, nil
	if r = w.report(Table{Tab: TabOrders}); r.Total != 0 || w.asked != nil || r.Finance.Plans != 0 {
		t.Errorf("without finance.view: %d rows, reads %v, coverage %+v; want none", r.Total, w.asked, r.Finance)
	}
}

func TestAmounts_InReportCurrency_FR014(t *testing.T) {
	t.Parallel()
	w := withOrders(twoPlans(t))
	w.currency = allwork.Currency{Code: "EUR", Rate: rate(t, "0.5")}

	r := w.report(Table{Tab: TabOrders})
	row := r.Orders.Rows[1]
	if diff := cmp.Diff(amountIn(t, "125.00", "EUR"), row.PoAmount, byValue()); diff != "" || r.Currency != "EUR" {
		t.Errorf("L-1's PO amount in %s (-want +got):\n%s", r.Currency, diff)
	}
	if diff := cmp.Diff(amountIn(t, "147.50", "EUR"), r.Orders.Totals.PoAmount, byValue()); diff != "" {
		t.Errorf("total (-want +got):\n%s", diff)
	}
}
