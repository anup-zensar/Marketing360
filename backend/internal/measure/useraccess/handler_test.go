package useraccess_test

import (
	"net/http"
	"slices"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"mapp/internal/measure/useraccess"
	"mapp/internal/platform/apitest"
	"mapp/internal/platform/authz"
)

const (
	alex   = "alex.morgan@shell.com"
	priya  = "priya.raman@shell.com"
	mei    = "mei.lin@shell.com"
	report = "/api/v1/reports/user-access"
)

func get(t *testing.T, c *apitest.Client, path string, status int) *apitest.Response {
	t.Helper()
	var opts []apitest.RequestOption
	if status == http.StatusBadRequest {
		opts = append(opts, apitest.SkipRequestValidation())
	}
	r := c.Do(t, http.MethodGet, path, nil, opts...)
	if r.Status != status {
		t.Fatalf("GET %s = %d %s, want %d", path, r.Status, r.Body, status)
	}
	return r
}

func reportOf(t *testing.T, c *apitest.Client, path string) useraccess.Report {
	t.Helper()
	var r useraccess.Report
	get(t, c, path, http.StatusOK).Decode(t, &r)
	return r
}

func TestUserAccess_Contract(t *testing.T) {
	t.Parallel()
	alexSees := apitest.New(t).SignIn(t, alex)

	r := reportOf(t, alexSees, report)
	if r.Total == 0 || r.Cards.People != r.Total || r.Sort != "name" || r.Page != 1 || r.PageSize != 20 {
		t.Errorf("report = %d people, cards %+v, sort %s, page %d of %d", r.Total, r.Cards, r.Sort, r.Page, r.PageSize)
	}
	reportOf(t, alexSees, report+"?q=shell&team=T-1001&status=all&sort=-lastSignedIn&page=1&pageSize=5")
	get(t, alexSees, report+"?sort=salary", http.StatusBadRequest)
}

func TestUserAccess_NeedsAccessManage_FR002(t *testing.T) {
	t.Parallel()
	server := apitest.New(t)

	get(t, server.SignIn(t, priya), report, http.StatusForbidden)
	get(t, server.SignIn(t, priya), "/api/v1/reports/user-access/export?format=xlsx", http.StatusForbidden)
}

// TestRows_EqualPeopleTab_SC003: every row's roles, scopes and teams are Settings › Roles and access › People's.
func TestRows_EqualPeopleTab_SC003(t *testing.T) {
	t.Parallel()
	alexSees := apitest.New(t).SignIn(t, alex)
	var tab struct {
		Items []struct {
			Person authz.PersonRef
			Teams  []authz.TeamRef
			Roles  []authz.RoleAssignment
		}
	}
	get(t, alexSees, "/api/v1/access/people", http.StatusOK).Decode(t, &tab)
	r := reportOf(t, alexSees, report+"?status=all&pageSize=500")

	if len(r.Rows) != len(tab.Items) {
		t.Fatalf("%d rows, the People tab %d people", len(r.Rows), len(tab.Items))
	}
	byName := func(a, b authz.TeamRef) int { return strings.Compare(strings.ToLower(a.Name), strings.ToLower(b.Name)) }
	for _, want := range tab.Items {
		i := slices.IndexFunc(r.Rows, func(row useraccess.Row) bool { return row.Person.ID == want.Person.ID })
		if i < 0 {
			t.Errorf("%s is not listed", want.Person.Name)
			continue
		}
		slices.SortFunc(want.Teams, byName)
		if diff := cmp.Diff(want.Roles, r.Rows[i].Roles); diff != "" {
			t.Errorf("%s's roles (-People tab +report):\n%s", want.Person.Name, diff)
		}
		if diff := cmp.Diff(want.Teams, r.Rows[i].Teams); diff != "" {
			t.Errorf("%s's teams (-People tab +report):\n%s", want.Person.Name, diff)
		}
	}
}

// TestSignIn_ShowsInReport_SC002: a sign-in shows on the next load; who has not signed in shows Never.
func TestSignIn_ShowsInReport_SC002(t *testing.T) {
	t.Parallel()
	server := apitest.New(t)
	server.SignIn(t, priya)
	r := reportOf(t, server.SignIn(t, alex), report+"?pageSize=500")

	at := func(email string) bool {
		i := slices.IndexFunc(r.Rows, func(row useraccess.Row) bool { return row.Email == email })
		return i >= 0 && r.Rows[i].LastSignedInAt != nil
	}
	if !at(priya) || !at(alex) || at(mei) {
		t.Errorf("signed in: Priya %v, Alex %v, Mei %v; want Priya and Alex only", at(priya), at(alex), at(mei))
	}
}
