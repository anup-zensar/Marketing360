package gsaplink

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"mapp/internal/measure"
	"mapp/internal/platform/date"
	"mapp/internal/platform/money"
)

// friday is the Friday before now, at 16:00 UTC: three working days before Wednesday noon.
var friday = time.Date(2026, 3, 13, 16, 0, 0, 0, time.UTC)

// byValue compares amounts and dates by value: their fields are unexported.
func byValue() cmp.Option {
	return cmp.Options{cmp.Comparer(func(x, y money.Amount) bool { return x.Cmp(y) == 0 }),
		cmp.Comparer(func(x, y date.Date) bool { return x.Compare(y) == 0 })}
}

func amountIn(tb testing.TB, amount, code string) money.Money {
	tb.Helper()
	a, err := money.ParseAmount(amount)
	if err != nil {
		tb.Fatalf("ParseAmount(%q) error = %v", amount, err)
	}
	return money.Money{Amount: a, Currency: code}
}

// withRequests: PR-1001 on the UK plan, 100.00 EUR and 50.00 USD, routed Friday; PR-1002 a top-up on the Germany
// plan routed two hours ago; PR-1003 on a line item out of view.
func withRequests(w *world) *world {
	w.requests = []Request{
		{ID: "PR-1001", Name: "PR for Stand build", Vendor: "Brightside Media BV", Owner: "Priya Raman", PlanID: ukPlan,
			Routing: RoutingOriginal, RoutedAt: friday, Lines: []RequestLine{
				{LineItemID: "L-1", Amount: amountIn(w.tb, "100.00", "EUR")},
				{LineItemID: "L-2", Amount: amountIn(w.tb, "50.00", "USD")},
			}},
		{ID: "PR-1002", Name: "Top-up for Booth", Vendor: "Northlight Events GmbH", Owner: "Mei Lin", PlanID: dePlan,
			GsapPrNumber: "1000123", Routing: RoutingModification, RoutedAt: now.Add(-2 * time.Hour),
			Lines: []RequestLine{{LineItemID: "L-3", Amount: amountIn(w.tb, "10.00", "USD")}}},
		{ID: "PR-1003", Name: "Elsewhere", PlanID: ukPlan, Routing: RoutingOriginal, RoutedAt: friday,
			Lines: []RequestLine{{LineItemID: "L-9", Amount: amountIn(w.tb, "1.00", "USD")}}},
	}
	return w
}

func waitingOf(w *world) WaitingTab {
	rates, _ := w.Rates(w.tb.Context())
	return w.year().waitingTab(w.requests, rates)
}

func TestWaiting_Row_FR008(t *testing.T) {
	t.Parallel()
	w := withRequests(twoPlans(t))
	ukAmount, deAmount := amountIn(t, "250.00", "USD"), amountIn(t, "10.00", "USD")

	want := []WaitingRow{
		{PurchaseRequestID: "PR-1001", Name: "PR for Stand build", Vendor: "Brightside Media BV", Owner: "Priya Raman",
			Path: Path{PlanID: ukPlan, Plan: "UK plan", Campaign: "Launch", Initiative: "Retail", ActivityID: ukActivity,
				Activity: "Stand"}, RoutedAt: friday, Routing: RoutingOriginal, Amount: &ukAmount, WaitingWorkingDays: 3,
			OverTurnaround: true},
		{PurchaseRequestID: "PR-1002", Name: "Top-up for Booth", Vendor: "Northlight Events GmbH", Owner: "Mei Lin",
			Path: Path{PlanID: dePlan, Plan: "Germany plan", Campaign: "Fleet", Initiative: "Trucks", ActivityID: deActivity,
				Activity: "Expo"}, RoutedAt: now.Add(-2 * time.Hour), Routing: RoutingModification, GsapPrNumber: "1000123",
			Amount: &deAmount},
	}
	if diff := cmp.Diff(want, waitingOf(w).Rows, byValue()); diff != "" {
		t.Errorf("rows (-want +got):\n%s", diff)
	}
}

func TestWaiting_Cards_FR010(t *testing.T) {
	t.Parallel()
	tab := waitingOf(withRequests(twoPlans(t)))
	total := amountIn(t, "260.00", "USD")

	if diff := cmp.Diff(&total, tab.Amount, byValue()); diff != "" || tab.Waiting != 2 || tab.OverTurnaround != 1 {
		t.Errorf("cards = Waiting %d, Over 2 working days %d, Amount waiting (-want +got):\n%s", tab.Waiting,
			tab.OverTurnaround, diff)
	}
}

func TestWaiting_AmountOnlyWithFinanceView_FR004(t *testing.T) {
	t.Parallel()
	w := withRequests(twoPlans(t))
	w.choices.FinancePlans = []string{ukPlan}

	r := w.report(Table{Tab: TabWaiting})
	ukOnly := amountIn(t, "250.00", "USD")
	if diff := cmp.Diff(&ukOnly, r.Waiting.Amount, byValue()); diff != "" {
		t.Errorf("Amount waiting covers the UK plan only (-want +got):\n%s", diff)
	}
	if r.Waiting.Rows[1].Amount != nil || r.Waiting.Rows[0].Amount == nil {
		t.Errorf("amounts = %v, %v; want the UK request's only", r.Waiting.Rows[0].Amount, r.Waiting.Rows[1].Amount)
	}
	if diff := cmp.Diff(&measure.FinanceCoverage{Plans: 1, Of: 2}, r.Finance); diff != "" {
		t.Errorf("coverage (-want +got):\n%s", diff)
	}

	w.choices.SeesFinance = false
	if r = w.report(Table{Tab: TabWaiting}); r.Waiting.Amount != nil || r.Waiting.Rows[0].Amount != nil {
		t.Errorf("without finance.view: amount %v, row %v; want none", r.Waiting.Amount, r.Waiting.Rows[0].Amount)
	}
}
