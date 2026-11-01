package gsaplink

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

// summary is what a report says about its table, apart from the rows.
type summary struct {
	Tab, Sort, Currency          string
	Page, PageSize, Total, Count int
	OnlySwbs                     bool
}

func summaryOf(r Report) summary {
	p := summary{Tab: r.Tab, Sort: r.Sort, Currency: r.Currency, Page: r.Page, PageSize: r.PageSize, Total: r.Total,
		OnlySwbs: r.Swbs != nil && r.Waiting == nil && r.Orders == nil && r.Finance == nil}
	if r.Swbs != nil {
		p.Count = r.Swbs.Waiting
	}
	return p
}

func waitingWorld(tb testing.TB) *world {
	tb.Helper()
	w := twoPlans(tb)
	w.approved[ukActivity], w.approved[deActivity] = now.Add(-10*time.Hour), now.AddDate(0, 0, -11)
	return w
}

func TestReport_Swbs_FR001_FR005(t *testing.T) {
	t.Parallel()
	r := waitingWorld(t).report(Table{})

	want := summary{Tab: TabSwbs, Sort: "-waiting", Currency: "USD", Page: 1, PageSize: DefaultPageSize, Total: 3, Count: 3,
		OnlySwbs: true}
	if diff := cmp.Diff(want, summaryOf(r)); diff != "" {
		t.Errorf("report (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff([]string{"L-3", "L-1", "L-2"}, swbsIDs(r.Swbs.Rows)); diff != "" {
		t.Errorf("rows, longest waiting first (-want +got):\n%s", diff)
	}
}

func TestReport_SortAndPage_FR015(t *testing.T) {
	t.Parallel()
	r := waitingWorld(t).report(Table{Tab: TabSwbs, Sort: "lineItem", Page: 2, PageSize: 2})

	want := summary{Tab: TabSwbs, Sort: "lineItem", Currency: "USD", Page: 2, PageSize: 2, Total: 3, Count: 3, OnlySwbs: true}
	if diff := cmp.Diff(want, summaryOf(r)); diff != "" {
		t.Errorf("report (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff([]string{"L-2"}, swbsIDs(r.Swbs.Rows)); diff != "" {
		t.Errorf("page 2 by line item (-want +got):\n%s", diff)
	}
}

func TestReport_UnknownSortFallsBack(t *testing.T) {
	t.Parallel()
	if r := twoPlans(t).report(Table{Tab: TabOrders, Sort: "waiting"}); r.Sort != "lineItem" {
		t.Errorf("sort = %s, want lineItem", r.Sort)
	}
}
