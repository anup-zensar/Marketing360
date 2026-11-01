package useraccess

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"mapp/internal/platform/authz"
)

func TestRows_NoRole_FR005(t *testing.T) {
	t.Parallel()
	r := newOffice().report(t, Query{Search: "mei"})
	if len(r.Rows) != 1 || len(r.Rows[0].Roles) != 0 || r.Rows[0].Email != "mei.lin@shell.com" {
		t.Errorf("rows = %+v; want Mei, listed with no role", r.Rows)
	}
}

func TestRows_InactiveTeamMarked_FR005(t *testing.T) {
	t.Parallel()
	r := newOffice().report(t, Query{Search: "priya"})
	want := []authz.TeamRef{team("T-1009", "Archive", false), team("T-1001", "Lubricants Europe", true)}
	if diff := cmp.Diff(want, r.Rows[0].Teams); diff != "" {
		t.Errorf("Priya's teams, alphabetical with the inactive one marked (-want +got):\n%s", diff)
	}
}

func TestFilter_SearchTeamStatus_FR006(t *testing.T) {
	t.Parallel()
	o := newOffice()
	tests := []struct {
		name string
		q    Query
		want []string
	}{
		{name: "active by default", q: Query{}, want: []string{"Alex Morgan", "Carlos Mendes", "Mei Lin", "Priya Raman"}},
		{name: "a team's name, any case", q: Query{Search: "LUBRICANTS"}, want: []string{"Priya Raman"}},
		{name: "a role", q: Query{Search: "finance"}, want: []string{"Carlos Mendes"}},
		{name: "an e-mail", q: Query{Search: "alex.morgan@"}, want: []string{"Alex Morgan"}},
		{name: "a team", q: Query{Teams: []string{"T-1002"}}, want: []string{"Carlos Mendes", "Mei Lin"}},
		{name: "inactive", q: Query{Status: StatusInactive}, want: []string{"Daniel Okoro"}},
		{name: "all", q: Query{Status: StatusAll, Teams: []string{"T-1001"}}, want: []string{"Daniel Okoro", "Priya Raman"}},
	}
	for _, tt := range tests {
		if diff := cmp.Diff(tt.want, names(o.report(t, tt.q).Rows)); diff != "" {
			t.Errorf("%s (-want +got):\n%s", tt.name, diff)
		}
	}
}

func TestSort_NeverLast_FR006(t *testing.T) {
	t.Parallel()
	o := newOffice()
	tests := map[string][]string{
		"lastSignedIn":  {"Alex Morgan", "Carlos Mendes", "Priya Raman", "Mei Lin"},
		"-lastSignedIn": {"Priya Raman", "Carlos Mendes", "Alex Morgan", "Mei Lin"},
		"-name":         {"Priya Raman", "Mei Lin", "Carlos Mendes", "Alex Morgan"},
		"roles":         {"Mei Lin", "Alex Morgan", "Carlos Mendes", "Priya Raman"},
	}
	for sort, want := range tests {
		if diff := cmp.Diff(want, names(o.report(t, Query{Sort: sort}).Rows)); diff != "" {
			t.Errorf("sort %s (-want +got):\n%s", sort, diff)
		}
	}
}

func TestPage_FR006(t *testing.T) {
	t.Parallel()
	r := newOffice().report(t, Query{Page: 2, PageSize: 3})
	if diff := cmp.Diff([]string{"Priya Raman"}, names(r.Rows)); diff != "" || r.Total != 4 || r.Cards.People != 4 {
		t.Errorf("page 2 of 3 rows, total %d, People %d (-want +got):\n%s", r.Total, r.Cards.People, diff)
	}
}
