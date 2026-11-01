package customdashboard_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"mapp/internal/customdashboard"
	"mapp/internal/platform/apitest"
)

// The handler tests run the whole API on the demo data, where DB-1001 "Lubricants Europe overview" is Priya's
// dashboard shared with Lubricants Europe (T-1001): Daniel and Sara are members, Mei is not.
const (
	priyaEmail  = "priya.raman@shell.com"
	danielEmail = "daniel.okoro@shell.com"
	meiEmail    = "mei.lin@shell.com"
	saraEmail   = "sara.novak@shell.com"
	demoPath    = "/api/v1/custom-dashboards/DB-1001"
	listPath    = "/api/v1/custom-dashboards"
)

func decode[T any](t *testing.T, r *apitest.Response, status int) T {
	t.Helper()
	var v T
	if r.Status != status {
		t.Fatalf("status = %d %s, want %d", r.Status, r.Body, status)
	}
	if err := json.Unmarshal(r.Body, &v); err != nil {
		t.Fatalf("decode %s: %v", r.Body, err)
	}
	return v
}

func assertError(t *testing.T, r *apitest.Response, status int, code string) {
	t.Helper()
	if r.Status != status || !strings.Contains(string(r.Body), `"code":"`+code+`"`) {
		t.Errorf("answer = %d %s, want %d %s", r.Status, r.Body, status, code)
	}
}

func newDashboard(name string) map[string]any {
	return map[string]any{
		"name": name, "description": "", "filters": map[string]any{"dateFrom": nil, "dateTo": nil, "planIds": []string{}, "teamIds": []string{}},
		"widgets": []any{map[string]any{
			"id": "w1", "type": "kpi", "title": "Activities", "subtitle": "", "note": "",
			"layout": map[string]int{"x": 0, "y": 0, "w": 3, "h": 2}, "display": map[string]any{"numberFormat": "number"},
			"query": map[string]any{"source": "activities", "dimension": nil, "series": nil,
				"measures": []any{map[string]any{"aggregation": "count", "field": nil, "label": ""}}, "filters": []any{},
				"sort": map[string]any{"by": "measure", "measureIndex": 0, "direction": "desc"}, "limit": 50, "groupRest": false},
		}},
	}
}

func TestListCustomDashboards_MineAndSharedWithMe_FR003(t *testing.T) {
	t.Parallel()
	server := apitest.New(t)
	names := func(email string) string {
		list := decode[customdashboard.List](t, server.SignIn(t, email).Do(t, http.MethodGet, listPath, nil), http.StatusOK)
		return fmt.Sprintf("mine %v shared %v", ids(list.Mine), ids(list.SharedWithMe))
	}

	got := []string{names(priyaEmail), names(danielEmail), names(meiEmail)}

	want := []string{"mine [DB-1001] shared []", "mine [] shared [DB-1001]", "mine [] shared []"}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("lists mismatch (-want +got):\n%s", diff)
	}
}

func ids(list []customdashboard.Summary) []string {
	out := []string{}
	for _, s := range list {
		out = append(out, s.ID)
	}
	return out
}

func TestGetCustomDashboard_200_404_FR005(t *testing.T) {
	t.Parallel()
	server := apitest.New(t)

	view := decode[customdashboard.View](t, server.SignIn(t, saraEmail).Do(t, http.MethodGet, demoPath, nil), http.StatusOK)
	mei := server.SignIn(t, meiEmail)

	if view.IsOwner || view.Access.Edit.Available || view.Access.Copy.Available || len(view.Widgets) != 6 ||
		view.Owner.Name != "Priya Raman" || len(view.ShareableTeams) != 0 {
		t.Errorf("Sara's view = %+v", view)
	}
	assertError(t, mei.Do(t, http.MethodGet, demoPath, nil), http.StatusNotFound, "not_found")
	assertError(t, mei.Do(t, http.MethodGet, listPath+"/DB-9999", nil), http.StatusNotFound, "not_found")
}

func TestCreateCustomDashboard_201_400_403_409_FR008_FR033(t *testing.T) {
	t.Parallel()
	server := apitest.New(t)
	priya := server.SignIn(t, priyaEmail)
	broken := newDashboard("")
	broken["widgets"].([]any)[0].(map[string]any)["layout"] = map[string]int{"x": 11, "y": 0, "w": 3, "h": 2}

	created := decode[customdashboard.View](t, priya.Do(t, http.MethodPost, listPath, newDashboard("Activities")), http.StatusCreated)
	refused := priya.Do(t, http.MethodPost, listPath, broken)
	taken := priya.Do(t, http.MethodPost, listPath, newDashboard("lubricants europe overview"))
	viewer := server.SignIn(t, saraEmail).Do(t, http.MethodPost, listPath, newDashboard("Mine"))

	if created.ID != "DB-1002" || !created.IsOwner || created.Version != 1 || len(created.Widgets) != 1 {
		t.Errorf("created = %+v", created)
	}
	if want := `"fields":[{"field":"name","message":"Enter a name."},{"field":"widgets[0].layout",` +
		`"message":"This widget goes past the grid's right edge."}]`; !strings.Contains(string(refused.Body), want) {
		t.Errorf("refused = %d %s, want every problem at once", refused.Status, refused.Body)
	}
	assertError(t, taken, http.StatusConflict, "name_taken")
	assertError(t, viewer, http.StatusForbidden, "forbidden")
}

func TestCreateCustomDashboard_413_FR008(t *testing.T) {
	t.Parallel()
	priya := apitest.New(t).SignIn(t, priyaEmail)
	body := newDashboard("Large")
	body["description"] = strings.Repeat("x", customdashboard.MaxBodyBytes)

	assertError(t, priya.Do(t, http.MethodPost, listPath, body, apitest.SkipRequestValidation()),
		http.StatusRequestEntityTooLarge, "too_large")
}

func TestUpdateCustomDashboard_200_403_404_409_FR006_FR012(t *testing.T) {
	t.Parallel()
	server := apitest.New(t)
	priya := server.SignIn(t, priyaEmail)
	rename := map[string]any{"version": 1, "name": "Lubricants at a glance"}

	notOwner := server.SignIn(t, danielEmail).Do(t, http.MethodPatch, demoPath, rename)
	notSeen := server.SignIn(t, meiEmail).Do(t, http.MethodPatch, demoPath, rename)
	changed := decode[customdashboard.View](t, priya.Do(t, http.MethodPatch, demoPath, rename), http.StatusOK)
	stale := priya.Do(t, http.MethodPatch, demoPath, map[string]any{"version": 1, "description": "Again"})

	assertError(t, notOwner, http.StatusForbidden, "forbidden")
	assertError(t, notSeen, http.StatusNotFound, "not_found")
	if changed.Name != "Lubricants at a glance" || changed.Version != 2 || len(changed.Widgets) != 6 {
		t.Errorf("changed = %q v%d with %d widgets", changed.Name, changed.Version, len(changed.Widgets))
	}
	assertError(t, stale, http.StatusConflict, "version_conflict")
}

func TestDeleteCustomDashboard_200_403_404_FR010(t *testing.T) {
	t.Parallel()
	server := apitest.New(t)

	notOwner := server.SignIn(t, danielEmail).Do(t, http.MethodDelete, demoPath, nil)
	notSeen := server.SignIn(t, meiEmail).Do(t, http.MethodDelete, demoPath, nil)
	deleted := server.SignIn(t, priyaEmail).Do(t, http.MethodDelete, demoPath, nil)

	assertError(t, notOwner, http.StatusForbidden, "forbidden")
	assertError(t, notSeen, http.StatusNotFound, "not_found")
	if deleted.Status != http.StatusOK || !strings.Contains(string(deleted.Body), `"recordId":"DB-1001"`) {
		t.Errorf("deleted = %d %s", deleted.Status, deleted.Body)
	}
}

func TestCopyCustomDashboard_201_404_409_FR011(t *testing.T) {
	t.Parallel()
	server := apitest.New(t)
	daniel := server.SignIn(t, danielEmail)

	copied := decode[customdashboard.View](t, daniel.Do(t, http.MethodPost, demoPath+"/copy", nil), http.StatusCreated)
	named := decode[customdashboard.View](t, daniel.Do(t, http.MethodPost, demoPath+"/copy", map[string]string{"name": "Mine"}),
		http.StatusCreated)
	taken := daniel.Do(t, http.MethodPost, demoPath+"/copy", map[string]string{"name": "mine"})
	notSeen := server.SignIn(t, meiEmail).Do(t, http.MethodPost, demoPath+"/copy", nil)

	got := []any{copied.Name, copied.Owner.Name, copied.IsOwner, len(copied.SharedWith), named.Name}
	if diff := cmp.Diff([]any{"Copy of Lubricants Europe overview", "Daniel Okoro", true, 0, "Mine"}, got); diff != "" {
		t.Errorf("copies mismatch (-want +got):\n%s", diff)
	}
	assertError(t, taken, http.StatusConflict, "name_taken")
	assertError(t, notSeen, http.StatusNotFound, "not_found")
}

func TestShareCustomDashboard_200_400_403_FR046(t *testing.T) {
	t.Parallel()
	server := apitest.New(t)
	priya := server.SignIn(t, priyaEmail)
	path := demoPath + "/sharing"

	notHers := priya.Do(t, http.MethodPut, path, map[string]any{"version": 1, "teamIds": []string{"T-1001", "T-1002"}})
	notOwner := server.SignIn(t, danielEmail).Do(t, http.MethodPut, path, map[string]any{"version": 1, "teamIds": []string{}})
	private := decode[customdashboard.View](t, priya.Do(t, http.MethodPut, path, map[string]any{"version": 1, "teamIds": []string{}}),
		http.StatusOK)
	gone := server.SignIn(t, saraEmail).Do(t, http.MethodGet, demoPath, nil)

	assertError(t, notHers, http.StatusBadRequest, "validation_failed")
	assertError(t, notOwner, http.StatusForbidden, "forbidden")
	if len(private.SharedWith) != 0 || private.Version != 2 || len(private.ShareableTeams) != 1 {
		t.Errorf("private = %+v", private)
	}
	assertError(t, gone, http.StatusNotFound, "not_found")
}
