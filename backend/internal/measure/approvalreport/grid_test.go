package approvalreport_test

import (
	"net/url"
	"slices"
	"testing"

	"github.com/google/go-cmp/cmp"

	"mapp/internal/measure/approvalreport"
	"mapp/internal/measure/approvalreport/approvalreporttest"
)

func ids(rows []approvalreport.Row) []string {
	list := make([]string, 0, len(rows))
	for _, r := range rows {
		list = append(list, keyOf(r))
	}
	return list
}

func TestGridQuery_041_FR013_FR015(t *testing.T) {
	t.Parallel()
	f := approvalreporttest.New(t)

	tests := []struct {
		name string
		q    approvalreport.Query
		want []string
	}{
		{name: "record type", q: approvalreport.Query{RecordTypes: []approvalreport.RecordType{approvalreport.RecordPurchaseRequest}},
			want: []string{"pr-waiting|p-carol|waiting|1"}},
		{name: "approver", q: approvalreport.Query{Approvers: []string{"p-bob"}},
			want: []string{"act-approved-any|p-bob|approved|1", "act-skipped-level|p-bob|approved|2"}},
		{name: "outcome not needed", q: approvalreport.Query{Outcomes: []approvalreport.Outcome{approvalreport.OutcomeNotNeeded}},
			want: []string{"act-approved-any|p-carol|notNeeded|1", "act-declined-all|p-dan|notNeeded|1"}},
		{name: "search words over the text columns", q: approvalreport.Query{Search: "event STAND"},
			want: []string{"pr-waiting|p-carol|waiting|1"}},
		{name: "search a comment", q: approvalreport.Query{Search: "out of budget"},
			want: []string{"act-declined-all|p-erin|declined|1"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := ids(allRows(t, f, tt.q))
			slices.Sort(got)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("rows (-want +got):\n%s", diff)
			}
		})
	}
}

func TestUnknownValues_041_C10(t *testing.T) {
	t.Parallel()
	f := approvalreporttest.New(t)

	for _, q := range []approvalreport.Query{
		{RecordTypes: []approvalreport.RecordType{"nonsense"}},
		{Approvers: []string{"nobody"}},
		{Outcomes: []approvalreport.Outcome{"nonsense"}},
	} {
		grid := reportOf(t, f, q).Grid
		if grid.Total != 0 || len(grid.Items) != 0 {
			t.Errorf("Query %+v: Total, Items = %d, %d, want an empty page and no error", q, grid.Total, len(grid.Items))
		}
	}
}

func TestApproverFilterGridOnly_041_C9(t *testing.T) {
	t.Parallel()
	f := approvalreporttest.New(t)

	plain := reportOf(t, f, approvalreport.Query{})
	narrowed := reportOf(t, f, approvalreport.Query{Approvers: []string{"p-bob"}})

	if diff := cmp.Diff(plain.StatusMix, narrowed.StatusMix); diff != "" {
		t.Errorf("StatusMix changed with the approver filter (-plain +narrowed):\n%s", diff)
	}
	if diff := cmp.Diff(plain.Average, narrowed.Average); diff != "" {
		t.Errorf("Average changed with the approver filter (-plain +narrowed):\n%s", diff)
	}
}

func TestSort_AcrossAllRows_041_FR015(t *testing.T) {
	t.Parallel()
	f := approvalreporttest.New(t)

	tests := []struct {
		sort, order string
		first, last string
	}{
		{sort: "itemId", order: "asc", first: "act-approved-any", last: "pr-waiting"},
		{sort: "itemId", order: "desc", first: "pr-waiting", last: "act-approved-any"},
		{sort: "itemName", order: "asc", first: "act-approved-any", last: "pr-waiting"},
		{sort: "approver", order: "asc", first: "act-cancelled", last: "act-automatic"},
		{sort: "comments", order: "desc", first: "act-declined-all", last: "pr-waiting"},
		{sort: "submittedOn", order: "desc", first: "pr-waiting", last: "act-declined-all"},
		{sort: "amount", order: "desc", first: "pr-waiting"},
	}
	for _, tt := range tests {
		t.Run(tt.sort+" "+tt.order, func(t *testing.T) {
			t.Parallel()
			rows := allRows(t, f, approvalreport.Query{Sort: tt.sort, Order: tt.order})
			if rows[0].ItemID != tt.first || (tt.last != "" && rows[len(rows)-1].ItemID != tt.last) {
				t.Errorf("first, last = %s, %s, want %s, %s", rows[0].ItemID, rows[len(rows)-1].ItemID, tt.first, tt.last)
			}
		})
	}
}

func TestSort_TiesByItemThenLevel_041_FR015(t *testing.T) {
	t.Parallel()
	rows := allRows(t, approvalreporttest.New(t), approvalreport.Query{Sort: "outcome", Order: "asc"})

	var approved []string
	for _, r := range rows {
		if r.Outcome == approvalreport.OutcomeApproved {
			approved = append(approved, keyOf(r))
		}
	}
	want := []string{"act-approved-any|p-bob|approved|1", "act-skipped-level|p-bob|approved|2"}
	if diff := cmp.Diff(want, approved); diff != "" {
		t.Errorf("equal outcomes in item order (-want +got):\n%s", diff)
	}
}

func TestPaging_041_FR015(t *testing.T) {
	t.Parallel()
	f := approvalreporttest.New(t)

	first := reportOf(t, f, approvalreport.Query{PageSize: 5, Sort: "itemId", Order: "asc"}).Grid
	third := reportOf(t, f, approvalreport.Query{PageSize: 5, Page: 3, Sort: "itemId", Order: "asc"}).Grid
	beyond := reportOf(t, f, approvalreport.Query{PageSize: 5, Page: 9}).Grid
	defaults := reportOf(t, f, approvalreport.Query{}).Grid

	if first.Total != 11 || len(first.Items) != 5 || first.Page != 1 || first.PageSize != 5 {
		t.Errorf("first page = total %d, %d rows, page %d of size %d, want 11, 5, 1, 5",
			first.Total, len(first.Items), first.Page, first.PageSize)
	}
	if len(third.Items) != 1 || third.Items[0].ItemID != "pr-waiting" {
		t.Errorf("third page = %v, want the last row alone", ids(third.Items))
	}
	if len(beyond.Items) != 0 || beyond.Total != 11 {
		t.Errorf("a page past the end = %d rows of %d, want none of 11", len(beyond.Items), beyond.Total)
	}
	if defaults.PageSize != 50 || defaults.Page != 1 {
		t.Errorf("defaults = page %d of size %d, want 1 of 50", defaults.Page, defaults.PageSize)
	}
}

func TestChoices_041_FR013_FR014(t *testing.T) {
	t.Parallel()

	choices := reportOf(t, approvalreporttest.New(t), approvalreport.Query{Outcomes: []approvalreport.Outcome{"declined"}}).Grid.Choices

	wantTypes := []approvalreport.RecordType{approvalreport.RecordActivity, approvalreport.RecordPurchaseRequest}
	if diff := cmp.Diff(wantTypes, choices.RecordTypes); diff != "" {
		t.Errorf("RecordTypes: those with rows, whatever the filters (-want +got):\n%s", diff)
	}
	if len(choices.Outcomes) != 7 || choices.Outcomes[6] != approvalreport.OutcomeNotNeeded {
		t.Errorf("Outcomes = %v, want the six words and Not needed", choices.Outcomes)
	}
	if len(choices.Approvers) != 5 || choices.Approvers[0].Name != "Alice Ng" {
		t.Errorf("Approvers = %v, want the five people with a row, by name", choices.Approvers)
	}
}

func TestParseQuery_041_C10(t *testing.T) {
	t.Parallel()

	good, err := approvalreport.ParseQuery(url.Values{"sort": {"amount"}, "order": {"asc"}, "pageSize": {"25"}, "page": {"2"},
		"recordType": {"activity", "x"}, "approver": {"p-bob"}, "outcome": {"waiting"}, "q": {" brand "}})
	if err != nil || good.Sort != "amount" || good.PageSize != 25 || good.Page != 2 || good.Search != "brand" || len(good.RecordTypes) != 2 {
		t.Errorf("ParseQuery(good) = %+v, %v, want the values kept and unknown filter values allowed", good, err)
	}
	for _, bad := range []url.Values{{"sort": {"nope"}}, {"order": {"up"}}, {"page": {"0"}}, {"page": {"x"}}, {"pageSize": {"7"}}} {
		if _, err := approvalreport.ParseQuery(bad); err == nil {
			t.Errorf("ParseQuery(%v) = nil error, want a problem", bad)
		}
	}
}
