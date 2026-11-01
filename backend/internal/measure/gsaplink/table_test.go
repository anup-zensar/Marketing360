package gsaplink

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

// sortTable is three line items: L-10 and L-9 approved together, L-2 earlier.
func sortTable() []SwbsRow {
	return []SwbsRow{
		{LineItemID: "L-10", LineItem: "banner", Owner: "Mei", ApprovedAt: now.Add(-time.Hour)},
		{LineItemID: "L-9", LineItem: "Banner", Owner: "alex", ApprovedAt: now.Add(-time.Hour)},
		{LineItemID: "L-2", LineItem: "Catering", Owner: "Priya", ApprovedAt: now.Add(-50 * time.Hour)},
	}
}

func TestTable_SortAndPage_FR015(t *testing.T) {
	t.Parallel()
	tests := []struct {
		sort string
		want []string
	}{
		{sort: DefaultSort(TabSwbs), want: []string{"L-2", "L-9", "L-10"}},
		{sort: "waiting", want: []string{"L-9", "L-10", "L-2"}},
		{sort: "lineItem", want: []string{"L-9", "L-10", "L-2"}},
		{sort: "-lineItem", want: []string{"L-2", "L-9", "L-10"}},
		{sort: "owner", want: []string{"L-9", "L-10", "L-2"}},
		{sort: "-approvedAt", want: []string{"L-9", "L-10", "L-2"}},
	}
	for _, tt := range tests {
		rows := sortTable()
		sortRows(rows, swbsColumns(), tt.sort, swbsID)
		if diff := cmp.Diff(tt.want, swbsIDs(rows)); diff != "" {
			t.Errorf("sort %s (-want +got):\n%s", tt.sort, diff)
		}
	}
	rows := sortTable()
	sortRows(rows, swbsColumns(), DefaultSort(TabSwbs), swbsID)
	if diff := cmp.Diff([]string{"L-10"}, swbsIDs(pageOf(rows, 2, 2))); diff != "" {
		t.Errorf("page 2 of 2 rows (-want +got):\n%s", diff)
	}
	if got := pageOf(rows, 3, 2); len(got) != 0 {
		t.Errorf("a page after the last = %v, want none", got)
	}
}

func TestTable_EveryColumnSorts_FR015(t *testing.T) {
	t.Parallel()
	for _, tab := range []string{TabSwbs, TabWaiting, TabOrders} {
		for _, key := range SortKeys(tab) {
			if !ValidSort(tab, key) || !ValidSort(tab, descending+key) {
				t.Errorf("%s: %s is listed but not valid", tab, key)
			}
		}
		if ValidSort(tab, "name") || ValidSort(tab, "--waiting") {
			t.Errorf("%s: an unknown sort is valid", tab)
		}
	}
	if DefaultSort(TabOrders) != "lineItem" || DefaultSort(TabWaiting) != "-waiting" {
		t.Errorf("default sorts = %s, %s", DefaultSort(TabOrders), DefaultSort(TabWaiting))
	}
}
