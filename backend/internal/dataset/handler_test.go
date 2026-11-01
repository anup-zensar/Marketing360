package dataset_test

import (
	"encoding/base64"
	"net/http"
	"slices"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"mapp/internal/dataset"
	"mapp/internal/platform/apitest"
)

const (
	priyasEmail  = "priya.raman@shell.com"
	danielsEmail = "daniel.okoro@shell.com"
	meisEmail    = "mei.lin@shell.com"
	sarasEmail   = "sara.novak@shell.com"
	datasetsPath = "/api/v1/datasets"
	europe       = "T-1001"
	china        = "T-1002"
	createdName  = "Agency leads"
	leadsCSV     = "Event,Date,Leads,Cost\nExpo Hamburg,2026-01-22,140,18500\nFleet Day,12/02/2026,85,9200.50\n"
)

func content(text string) string {
	return base64.StdEncoding.EncodeToString([]byte(text))
}

func createBody(name, csv string) map[string]any {
	return map[string]any{"fileName": "leads.csv", "content": content(csv), "name": name, "description": "From the agency",
		"columns": []map[string]any{
			{"position": 0, "label": "Event", "kind": "text", "currency": nil},
			{"position": 1, "label": "Date", "kind": "date", "currency": nil},
			{"position": 2, "label": "Leads", "kind": "number", "currency": nil},
			{"position": 3, "label": "Agency cost", "kind": "money", "currency": "EUR"},
		}}
}

func decoded[T any](t *testing.T, r *apitest.Response, status int) T {
	t.Helper()
	if r.Status != status {
		t.Fatalf("status %d %s, want %d", r.Status, r.Body, status)
	}
	var v T
	r.Decode(t, &v)
	return v
}

func statuses(responses ...*apitest.Response) []int {
	got := make([]int, 0, len(responses))
	for _, r := range responses {
		got = append(got, r.Status)
	}
	return got
}

// created signs Priya in and stores a dataset of hers, not shared, next to the seeded DS-1001.
func created(t *testing.T, server *apitest.Server) (*apitest.Client, dataset.View) {
	t.Helper()
	priya := server.SignIn(t, priyasEmail)
	return priya, decoded[dataset.View](t, priya.Do(t, http.MethodPost, datasetsPath, createBody(createdName, leadsCSV)),
		http.StatusCreated)
}

func same(t *testing.T, what string, want, got any) {
	t.Helper()
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("%s (-want +got):\n%s", what, diff)
	}
}

func TestPreviewDataset_ProposesColumns_FR041(t *testing.T) {
	t.Parallel()
	server := apitest.New(t)
	priya := server.SignIn(t, priyasEmail)

	preview := decoded[dataset.Preview](t, priya.Do(t, http.MethodPost, datasetsPath+"/preview",
		map[string]any{"fileName": "leads.csv", "content": content(leadsCSV)}), http.StatusOK)
	unreadable := priya.Do(t, http.MethodPost, datasetsPath+"/preview", map[string]any{"fileName": "leads.txt", "content": content("A\n1\n")})
	tooLarge := priya.Do(t, http.MethodPost, datasetsPath+"/preview",
		map[string]any{"fileName": "big.csv", "content": strings.Repeat("A", dataset.MaxUploadBodyBytes)}, apitest.SkipRequestValidation())
	bySara := server.SignIn(t, sarasEmail).Do(t, http.MethodPost, datasetsPath+"/preview",
		map[string]any{"fileName": "leads.csv", "content": content(leadsCSV)})

	kinds := []dataset.Kind{}
	for _, c := range preview.Columns {
		kinds = append(kinds, c.ProposedKind)
	}
	if !slices.Equal(kinds, []dataset.Kind{dataset.KindText, dataset.KindDate, dataset.KindNumber, dataset.KindMoney}) || preview.RowCount != 2 {
		t.Errorf("preview = %+v", preview)
	}
	if got := statuses(unreadable, tooLarge, bySara); !slices.Equal(got, []int{400, 413, 403}) {
		t.Errorf("refusals = %v, want 400, 413, 403", got)
	}
}

func TestCreateDataset_StoresAndRefuses_FR042_FR043(t *testing.T) {
	t.Parallel()
	server := apitest.New(t)
	priya, view := created(t, server)

	badCells := priya.Do(t, http.MethodPost, datasetsPath, createBody("Bad", leadsCSV+"Trade Fair,someday,n/a,1\n"))
	taken := priya.Do(t, http.MethodPost, datasetsPath, createBody("AGENCY LEADS", leadsCSV))
	bySara := server.SignIn(t, sarasEmail).Do(t, http.MethodPost, datasetsPath, createBody("Sara's", leadsCSV))

	if !view.IsOwner || view.RowCount != 2 || len(view.Columns) != 4 || view.Owner.Name != "Priya Raman" || len(view.UsedBy) != 0 {
		t.Errorf("created = %+v", view)
	}
	if got := statuses(badCells, taken, bySara); !slices.Equal(got, []int{400, 409, 403}) {
		t.Errorf("refusals = %v, want 400, 409, 403", got)
	}
	if !strings.Contains(string(badCells.Body), "Row 4, Date: 'someday' is not a date") ||
		!strings.Contains(string(badCells.Body), "Row 4, Leads: 'n/a' is not a number.") {
		t.Errorf("cell problems = %s, want both at once", badCells.Body)
	}
}

func TestListDatasets_MineAndSharedWithMe_FR040(t *testing.T) {
	t.Parallel()
	server := apitest.New(t)
	priya, view := created(t, server)
	decoded[dataset.View](t, priya.Do(t, http.MethodPut, datasetsPath+"/"+view.ID+"/sharing",
		map[string]any{"version": 1, "teamIds": []string{europe}}), http.StatusOK)

	asPriya := decoded[dataset.List](t, priya.Do(t, http.MethodGet, datasetsPath, nil), http.StatusOK)
	asDaniel := decoded[dataset.List](t, server.SignIn(t, danielsEmail).Do(t, http.MethodGet, datasetsPath, nil), http.StatusOK)
	asMei := decoded[dataset.List](t, server.SignIn(t, meisEmail).Do(t, http.MethodGet, datasetsPath, nil), http.StatusOK)
	bySara := server.SignIn(t, sarasEmail).Do(t, http.MethodGet, datasetsPath, nil)

	hasView := func(list []dataset.Summary) bool {
		return slices.ContainsFunc(list, func(s dataset.Summary) bool { return s.ID == view.ID })
	}
	if !hasView(asPriya.Mine) || !hasView(asDaniel.SharedWithMe) || hasView(asMei.Mine) || hasView(asMei.SharedWithMe) {
		t.Errorf("Priya %+v, Daniel %+v, Mei %+v", asPriya, asDaniel, asMei)
	}
	if bySara.Status != http.StatusForbidden {
		t.Errorf("Sara without Build dashboards = %d, want 403", bySara.Status)
	}
}

func TestGetDataset_VisibleOnly_FR044(t *testing.T) {
	t.Parallel()
	server := apitest.New(t)
	priya, view := created(t, server)
	daniel, mei := server.SignIn(t, danielsEmail), server.SignIn(t, meisEmail)
	path := datasetsPath + "/" + view.ID

	beforeShare := daniel.Do(t, http.MethodGet, path, nil)
	decoded[dataset.View](t, priya.Do(t, http.MethodPut, path+"/sharing", map[string]any{"version": 1, "teamIds": []string{europe}}),
		http.StatusOK)
	asDaniel := decoded[dataset.View](t, daniel.Do(t, http.MethodGet, path, nil), http.StatusOK)
	asMei := mei.Do(t, http.MethodGet, path, nil)
	missing := priya.Do(t, http.MethodGet, datasetsPath+"/DS-9999", nil)

	if got := statuses(beforeShare, asMei, missing); !slices.Equal(got, []int{404, 404, 404}) {
		t.Errorf("unseen = %v, want 404 each", got)
	}
	if asDaniel.IsOwner || asDaniel.Access.Edit.Available || asDaniel.Access.Edit.Reason == "" || len(asDaniel.ShareableTeams) != 0 {
		t.Errorf("as Daniel = %+v, want view-only", asDaniel)
	}
}
