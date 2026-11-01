package gsaplink_test

import (
	"net/http"
	"testing"

	"mapp/internal/platform/apitest"
	"mapp/internal/platform/money"
)

// financialRow is what the test reads of the Financials page's rows (015).
type financialRow struct {
	Type          string       `json:"type"`
	PoNumber      string       `json:"poNumber"`
	AmountDisplay *money.Money `json:"amountDisplay"`
}

// TestOrders_EqualLineItemFinancials_SC003: every row's PO amount, invoiced and paid equal the sums of that PO's
// transactions on the line item's Financials, to the cent, with both in US dollars.
func TestOrders_EqualLineItemFinancials_SC003(t *testing.T) {
	t.Parallel()
	carlosSees := apitest.New(t).SignIn(t, carlos)
	if r := carlosSees.Do(t, http.MethodPatch, "/api/v1/me/preferences", map[string]any{"displayCurrency": "USD"}); r.Status != http.StatusOK {
		t.Fatalf("display currency = %d %s", r.Status, r.Body)
	}
	r := reportOf(t, carlosSees, report+"&tab=orders&currency=USD&pageSize=500")
	if r.Orders == nil || len(r.Orders.Rows) == 0 {
		t.Fatalf("orders = %+v, want the demo purchase orders", r.Orders)
	}
	for _, row := range r.Orders.Rows {
		var page struct{ Items []financialRow }
		get(t, carlosSees, "/api/v1/financials/transactions?recordId="+row.LineItemID+"&pageSize=200", http.StatusOK).
			Decode(t, &page)
		sums := map[string]money.Amount{}
		for _, item := range page.Items {
			if item.PoNumber == row.PoNumber && item.AmountDisplay != nil {
				sums[item.Type] = sums[item.Type].Add(item.AmountDisplay.Amount)
			}
		}
		got := []money.Amount{row.PoAmount.Amount, row.Invoiced.Amount, row.Paid.Amount}
		want := []money.Amount{sums["purchaseOrder"], sums["invoice"], sums["payment"]}
		for i := range got {
			if got[i].Cmp(want[i]) != 0 {
				t.Errorf("%s PO %s: figure %d = %s, Financials %s", row.LineItemID, row.PoNumber, i, got[i], want[i])
			}
		}
	}
}
