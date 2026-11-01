package dataset_test

import (
	"net/http"
	"slices"
	"testing"

	"mapp/internal/dataset"
	"mapp/internal/history"
	"mapp/internal/platform/apitest"
)

func TestUpdateDataset_OwnerAndVersion_FR044(t *testing.T) {
	t.Parallel()
	server := apitest.New(t)
	priya, view := created(t, server, "Event leads 2026")
	path := datasetsPath + "/" + view.ID
	decoded[dataset.View](t, priya.Do(t, http.MethodPut, path+"/sharing", map[string]any{"version": 1, "teamIds": []string{europe}}),
		http.StatusOK)

	renamed := decoded[dataset.View](t, priya.Do(t, http.MethodPatch, path, map[string]any{"version": 2, "name": "Leads",
		"columns": []map[string]any{{"position": 3, "label": "Cost (EUR)"}}}), http.StatusOK)
	stale := priya.Do(t, http.MethodPatch, path, map[string]any{"version": 2, "name": "Other"})
	invalid := priya.Do(t, http.MethodPatch, path, map[string]any{"version": 3, "name": ""})
	retype := priya.Do(t, http.MethodPatch, path, map[string]any{"version": 3,
		"columns": []map[string]any{{"position": 2, "label": "Leads", "kind": "text"}}}, apitest.SkipRequestValidation())
	byDaniel := server.SignIn(t, danielsEmail).Do(t, http.MethodPatch, path, map[string]any{"version": 3, "name": "Daniel's"})
	byMei := server.SignIn(t, meisEmail).Do(t, http.MethodPatch, path, map[string]any{"version": 3, "name": "Mei's"})

	if renamed.Name != "Leads" || renamed.Columns[3].Label != "Cost (EUR)" || renamed.Columns[3].Kind != dataset.KindMoney ||
		renamed.Version != 3 {
		t.Errorf("renamed = %+v", renamed)
	}
	if got := statuses(stale, invalid, retype, byDaniel, byMei); !slices.Equal(got, []int{409, 400, 400, 403, 404}) {
		t.Errorf("refusals = %v, want 409, 400, 400, 403, 404", got)
	}
}

func TestShareDataset_OwnTeamsOnly_FR046(t *testing.T) {
	t.Parallel()
	server := apitest.New(t)
	priya, view := created(t, server, "Event leads 2026")
	path := datasetsPath + "/" + view.ID + "/sharing"

	notHers := priya.Do(t, http.MethodPut, path, map[string]any{"version": 1, "teamIds": []string{china}})
	shared := decoded[dataset.View](t, priya.Do(t, http.MethodPut, path, map[string]any{"version": 1, "teamIds": []string{europe}}),
		http.StatusOK)
	stale := priya.Do(t, http.MethodPut, path, map[string]any{"version": 1, "teamIds": []string{}})
	byDaniel := server.SignIn(t, danielsEmail).Do(t, http.MethodPut, path, map[string]any{"version": 2, "teamIds": []string{}})
	byMei := server.SignIn(t, meisEmail).Do(t, http.MethodPut, path, map[string]any{"version": 2, "teamIds": []string{}})

	if len(shared.SharedWith) != 1 || shared.SharedWith[0].ID != europe || shared.Version != 2 {
		t.Errorf("shared = %+v", shared)
	}
	if got := statuses(notHers, stale, byDaniel, byMei); !slices.Equal(got, []int{400, 409, 403, 404}) {
		t.Errorf("refusals = %v, want 400, 409, 403, 404", got)
	}
}

func TestDeleteDataset_ToRecycleBin_FR045(t *testing.T) {
	t.Parallel()
	server := apitest.New(t)
	priya, view := created(t, server, "Event leads 2026")
	path := datasetsPath + "/" + view.ID
	decoded[dataset.View](t, priya.Do(t, http.MethodPut, path+"/sharing", map[string]any{"version": 1, "teamIds": []string{europe}}),
		http.StatusOK)

	byDaniel := server.SignIn(t, danielsEmail).Do(t, http.MethodDelete, path, nil)
	result := decoded[history.DeleteResult](t, priya.Do(t, http.MethodDelete, path, nil), http.StatusOK)
	gone := priya.Do(t, http.MethodGet, path, nil)
	again := priya.Do(t, http.MethodDelete, path, nil)

	if result.RecycleBinEntry.RecordType != "dataset" || result.RecycleBinEntry.RecordName != "Event leads 2026" {
		t.Errorf("bin entry = %+v", result.RecycleBinEntry)
	}
	if got := statuses(byDaniel, gone, again); !slices.Equal(got, []int{403, 404, 404}) {
		t.Errorf("refusals = %v, want 403, 404, 404", got)
	}
}

func TestListDatasetRows_PagesAsStored_FR040(t *testing.T) {
	t.Parallel()
	server := apitest.New(t)
	priya, view := created(t, server, "Event leads 2026")
	path := datasetsPath + "/" + view.ID + "/rows"

	page := decoded[dataset.RowPage](t, priya.Do(t, http.MethodGet, path+"?offset=1&limit=1", nil), http.StatusOK)
	all := decoded[dataset.RowPage](t, priya.Do(t, http.MethodGet, path, nil), http.StatusOK)
	badLimit := priya.Do(t, http.MethodGet, path+"?limit=0", nil, apitest.SkipRequestValidation())
	notNumber := priya.Do(t, http.MethodGet, path+"?offset=x", nil, apitest.SkipRequestValidation())
	byMei := server.SignIn(t, meisEmail).Do(t, http.MethodGet, path, nil)

	if len(page.Rows) != 1 || *page.Rows[0][0] != "Fleet Day" || *page.Rows[0][3] != "9200.50" || page.Total != 2 {
		t.Errorf("page = %+v", page)
	}
	if len(all.Rows) != 2 || all.Limit != 50 || all.Offset != 0 {
		t.Errorf("default page = %+v", all)
	}
	if got := statuses(badLimit, notNumber, byMei); !slices.Equal(got, []int{400, 400, 404}) {
		t.Errorf("refusals = %v, want 400, 400, 404", got)
	}
}
