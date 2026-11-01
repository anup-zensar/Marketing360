package gsaplink

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func TestReport_Swbs_FR001_FR005(t *testing.T) {
	t.Parallel()
	w := twoPlans(t)
	w.approved[ukActivity], w.approved[deActivity] = now.Add(-10*time.Hour), now.AddDate(0, 0, -11)

	r := w.report(Table{})
	if r.Tab != TabSwbs || r.Sort != "-waiting" || r.Page != 1 || r.PageSize != DefaultPageSize || r.Total != 3 {
		t.Errorf("report = tab %s, sort %s, page %d of %d, total %d", r.Tab, r.Sort, r.Page, r.PageSize, r.Total)
	}
	if r.Finance != nil || r.Waiting != nil || r.Orders != nil || r.Swbs == nil || r.Currency != "USD" {
		t.Fatalf("report = %+v; want only the SWBS tab, in USD, without finance coverage", r)
	}
	if diff := cmp.Diff([]string{"L-3", "L-1", "L-2"}, swbsIDs(r.Swbs.Rows)); diff != "" {
		t.Errorf("rows (-want +got):\n%s", diff)
	}

	r = w.report(Table{Tab: TabSwbs, Sort: "lineItem", Page: 2, PageSize: 2})
	if diff := cmp.Diff([]string{"L-2"}, swbsIDs(r.Swbs.Rows)); diff != "" || r.Total != 3 || r.Swbs.Waiting != 3 {
		t.Errorf("page 2 by line item (-want +got):\n%s; total %d, waiting %d", diff, r.Total, r.Swbs.Waiting)
	}
}

func TestReport_UnknownSortFallsBack(t *testing.T) {
	t.Parallel()
	if r := twoPlans(t).report(Table{Tab: TabOrders, Sort: "waiting"}); r.Sort != "lineItem" {
		t.Errorf("sort = %s, want lineItem", r.Sort)
	}
}
