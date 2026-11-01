package purchasereport_test

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"

	"mapp/internal/allwork"
	"mapp/internal/measure"
	"mapp/internal/measure/purchasereport"
	"mapp/internal/measure/purchasereport/purchasereporttest"
	"mapp/internal/platform/money"
)

func reportOf(t *testing.T, f purchasereporttest.Fixture, q purchasereport.Query) purchasereport.View {
	t.Helper()
	view, err := purchasereport.NewService(f.DB, f.Deps).Report(t.Context(), purchasereporttest.Alice(), measure.Filter{}, q)
	if err != nil {
		t.Fatalf("Report() error: %v", err)
	}
	return view
}

func amountText(m *money.Money) string {
	if m == nil {
		return "—"
	}
	return m.Amount.String() + " " + m.Currency
}

func texts(values []*money.Money) []string {
	out := make([]string, 0, len(values))
	for _, v := range values {
		out = append(out, amountText(v))
	}
	return out
}

func TestStatusMix_042_FR007(t *testing.T) {
	t.Parallel()

	mix := reportOf(t, purchasereporttest.New(t), purchasereport.Query{}).StatusMix

	type row struct {
		Status purchasereport.Status
		Count  int
		Amount string
		Share  int
	}
	want := []row{
		{purchasereport.StatusCreated, 1, "700.00 USD", 11}, {purchasereport.StatusSubmitted, 1, "50.00 USD", 11},
		{purchasereport.StatusInReview, 1, "60.00 USD", 11}, {purchasereport.StatusDeclined, 1, "70.00 USD", 11},
		{purchasereport.StatusApproved, 4, "3600.00 USD", 45}, {purchasereport.StatusClosed, 1, "300.00 USD", 11},
	}
	got := make([]row, 0, len(mix.Items))
	for _, item := range mix.Items {
		got = append(got, row{item.Status, item.Count, amountText(item.Amount), item.SharePercent})
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("status items, shares adding to 100, the plan without finance.view left out (-want +got):\n%s", diff)
	}
	if mix.Total != 9 || amountText(mix.TotalAmount) != "4780.00 USD" || mix.WaitingForGSAP != 1 {
		t.Errorf("Total, TotalAmount, WaitingForGSAP = %d, %s, %d, want 9, 4780.00 USD, 1", mix.Total, amountText(mix.TotalAmount), mix.WaitingForGSAP)
	}
}

func TestCards_042_FR006_C5(t *testing.T) {
	t.Parallel()

	cards := reportOf(t, purchasereporttest.New(t), purchasereport.Query{}).Cards

	if cards.PurchaseRequests != 9 || cards.Lines != 10 || cards.Vendors != 4 {
		t.Errorf("counts = %d, %d, %d, want 9 requests, 10 lines, 4 vendors", cards.PurchaseRequests, cards.Lines, cards.Vendors)
	}
	if amountText(cards.Committed) != "3900.00 USD" || amountText(cards.Invoiced) != "700.00 USD" {
		t.Errorf("Committed, Invoiced = %s, %s, want 3900.00 USD, 700.00 USD", amountText(cards.Committed), amountText(cards.Invoiced))
	}
}

func TestCommitment_042_FR009_C1(t *testing.T) {
	t.Parallel()
	view := reportOf(t, purchasereporttest.New(t), purchasereport.Query{})

	byVendor := view.ByVendor
	wantMonths := []string{"2026-03", "2026-04", "2026-05", "none"}
	if diff := cmp.Diff(wantMonths, byVendor.Months); diff != "" {
		t.Errorf("Months (-want +got):\n%s", diff)
	}
	names := []string{}
	for _, s := range byVendor.AllSeries {
		names = append(names, s.Name)
	}
	if diff := cmp.Diff([]string{"Globex", "Acme", "Not filled in"}, names); diff != "" {
		t.Errorf("vendors, largest first, a request without a vendor kept (-want +got):\n%s", diff)
	}
	acme := byVendor.AllSeries[1]
	wantAcme := []string{"1300.00 USD", "0.00 USD", "500.00 USD", "0.00 USD"}
	if diff := cmp.Diff(wantAcme, texts(acme.Values)); diff != "" {
		t.Errorf("Acme per month: Approved and Closed lines only, the hidden plan's left out (-want +got):\n%s", diff)
	}
	owners := view.ByOwner.AllSeries
	if owners[0].Name != "Alice Ng" || texts(owners[0].Values)[0] != "1000.00 USD" {
		t.Errorf("owner series = %v, want Alice Ng first with 1000.00 USD in March", owners[0])
	}
}

func TestChartsAddUpToTheCard_042_FR013(t *testing.T) {
	t.Parallel()
	view := reportOf(t, purchasereporttest.New(t), purchasereport.Query{})

	for name, chart := range map[string]purchasereport.Chart{"vendor": view.ByVendor, "owner": view.ByOwner} {
		var sum money.Amount
		for _, total := range chart.Totals {
			sum = sum.Add(total.Amount)
		}
		if sum.String() != view.Cards.Committed.Amount.String() || amountText(chart.Committed) != amountText(view.Cards.Committed) {
			t.Errorf("%s chart: months add to %s, committed %s, want the card's %s", name, sum, amountText(chart.Committed), amountText(view.Cards.Committed))
		}
	}
}

func TestMonths_042_C2(t *testing.T) {
	t.Parallel()
	f := purchasereporttest.New(t)
	f.Deps.Purchases = purchasereporttest.Purchases{Requests: purchasereporttest.Spread("2026-01-15", "2026-04-02")}

	chart := reportOf(t, f, purchasereport.Query{}).ByVendor

	want := []string{"2026-01", "2026-02", "2026-03", "2026-04"}
	if diff := cmp.Diff(want, chart.Months); diff != "" {
		t.Errorf("every month between the first and the last (-want +got):\n%s", diff)
	}
	if got := texts(chart.AllSeries[0].Values); got[1] != "0.00 USD" || got[2] != "0.00 USD" {
		t.Errorf("months without a line = %v, want a real 0.00", got)
	}
}

func TestTopTenAndOther_042_FR011(t *testing.T) {
	t.Parallel()
	f := purchasereporttest.New(t)
	f.Deps.Purchases = purchasereporttest.Purchases{Requests: purchasereporttest.Vendors(12)}

	chart := reportOf(t, f, purchasereport.Query{}).ByVendor

	if len(chart.AllSeries) != 12 || len(chart.ChartSeries) != 11 {
		t.Fatalf("series = %d all, %d in the chart, want 12 and ten plus Other", len(chart.AllSeries), len(chart.ChartSeries))
	}
	other := chart.ChartSeries[10]
	if other.Name != "Other" || amountText(other.Values[0]) != "30.00 USD" {
		t.Errorf("Other = %s %s, want the two smallest vendors (10 + 20) as 30.00 USD", other.Name, amountText(other.Values[0]))
	}
}

func TestNoFinanceView_042_FR020(t *testing.T) {
	t.Parallel()
	f := purchasereporttest.New(t)
	f.Scope.Choices = allwork.Choices{SeesFinance: false, People: f.Scope.Choices.People}
	f.Deps.Scope = purchasereporttest.Scoper{Result: f.Scope}

	view := reportOf(t, f, purchasereport.Query{})

	if view.Cards.Committed != nil || view.Cards.Invoiced != nil || view.StatusMix.TotalAmount != nil || view.ByVendor.Committed != nil {
		t.Errorf("amounts = %v, %v, want a dash without finance.view", view.Cards, view.StatusMix.TotalAmount)
	}
	if view.Cards.PurchaseRequests != 9 || view.StatusMix.Items[4].Count != 4 {
		t.Errorf("counts must still show, got %d requests and %d approved", view.Cards.PurchaseRequests, view.StatusMix.Items[4].Count)
	}
	if want := (purchasereport.FinanceCovers{Plans: 0, Of: 2}); view.FinanceCovers != want {
		t.Errorf("FinanceCovers = %+v, want %+v", view.FinanceCovers, want)
	}
}

func TestFinanceCovers_042_FR020(t *testing.T) {
	t.Parallel()

	got := reportOf(t, purchasereporttest.New(t), purchasereport.Query{}).FinanceCovers

	if want := (purchasereport.FinanceCovers{Plans: 1, Of: 2}); got != want {
		t.Errorf("FinanceCovers = %+v, want %+v", got, want)
	}
}

func TestFilters_042_FR004_C6(t *testing.T) {
	t.Parallel()
	f := purchasereporttest.New(t)

	tests := []struct {
		q    purchasereport.Query
		want int
	}{
		{purchasereport.Query{Vendors: []string{"Acme"}}, 3},
		{purchasereport.Query{Vendors: []string{"none"}}, 1},
		{purchasereport.Query{Statuses: []purchasereport.Status{purchasereport.StatusApproved}}, 4},
		{purchasereport.Query{RequestOwners: []string{"p-bob"}}, 2},
		{purchasereport.Query{Vendors: []string{"Nobody"}}, 0},
		{purchasereport.Query{Statuses: []purchasereport.Status{"nonsense"}}, 0},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%+v", tt.q), func(t *testing.T) {
			t.Parallel()
			view := reportOf(t, f, tt.q)
			if view.Cards.PurchaseRequests != tt.want || view.StatusMix.Total != tt.want {
				t.Errorf("requests = %d, status total = %d, want %d", view.Cards.PurchaseRequests, view.StatusMix.Total, tt.want)
			}
		})
	}
}

func TestSwitchedOff_042_FR022(t *testing.T) {
	t.Parallel()
	f := purchasereporttest.New(t)
	f.Deps.Purchases = purchasereporttest.Purchases{Off: true}

	view := reportOf(t, f, purchasereport.Query{})

	if view.PurchaseSupport || view.Cards.PurchaseRequests != 0 {
		t.Errorf("PurchaseSupport, requests = %v, %d, want off and none", view.PurchaseSupport, view.Cards.PurchaseRequests)
	}
}
