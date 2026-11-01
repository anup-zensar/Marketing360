package useraccess

import (
	"net/url"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestQuery_Allowlist_FR006(t *testing.T) {
	t.Parallel()
	got, err := ParseQuery(url.Values{"q": {" Lubricants "}, "team": {"T-1001", "T-1002"}, "status": {"all"},
		"sort": {"-lastSignedIn"}, "page": {"2"}, "pageSize": {"50"}})
	want := Query{Search: "Lubricants", Teams: []string{"T-1001", "T-1002"}, Status: StatusAll, Sort: "-lastSignedIn",
		Page: 2, PageSize: 50}
	if diff := cmp.Diff(want, got); err != nil || diff != "" {
		t.Errorf("ParseQuery (-want +got):\n%s %v", diff, err)
	}
	defaults, err := ParseQuery(url.Values{})
	if diff := cmp.Diff(Query{Teams: []string{}, Status: StatusActive, Sort: "name", Page: 1, PageSize: 20}, defaults); err != nil || diff != "" {
		t.Errorf("defaults (-want +got):\n%s %v", diff, err)
	}

	for _, bad := range []url.Values{{"q": {strings.Repeat("a", 101)}}, {"team": {"X-1"}}, {"status": {"gone"}},
		{"sort": {"salary"}}, {"page": {"0"}}, {"pageSize": {"4"}}, {"pageSize": {"501"}}} {
		if _, err := ParseQuery(bad); err == nil {
			t.Errorf("ParseQuery(%v): no error, want the value refused", bad)
		}
	}
}
