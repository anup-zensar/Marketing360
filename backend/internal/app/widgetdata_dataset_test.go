package app_test

import (
	"net/http"
	"testing"

	"github.com/google/go-cmp/cmp"

	"mapp/internal/dataset"
	"mapp/internal/platform/apitest"
	"mapp/internal/widgetdata"
)

const eventLeads = "DS-1001"

func leadsByEvent() widgetdata.Request {
	return widgetdata.Request{Type: "table", Query: widgetdata.Query{
		Source: eventLeads, Dimension: &widgetdata.GroupBy{Field: "c1"}, Measures: []widgetdata.Measure{measure("count", "")},
		Filters: []widgetdata.Filter{}, Sort: widgetdata.Sort{By: "dimension", Direction: "asc"}, Limit: 50,
	}}
}

func shareLeads(t *testing.T, priya *apitest.Client, teamIDs []string) {
	t.Helper()
	var stored dataset.View
	priya.Do(t, http.MethodGet, "/api/v1/datasets/"+eventLeads, nil).Decode(t, &stored)
	answer := priya.Do(t, http.MethodPut, "/api/v1/datasets/"+eventLeads+"/sharing",
		dataset.SharingInput{Version: &stored.Version, TeamIDs: teamIDs})
	if answer.Status != http.StatusOK {
		t.Fatalf("sharing %s = %d %s", eventLeads, answer.Status, answer.Body)
	}
}

func TestDataset_UsedByTheDashboardsReadingIt_FR045(t *testing.T) {
	t.Parallel()
	priya := apitest.New(t).SignIn(t, priyasEmail)
	var stored dataset.View

	priya.Do(t, http.MethodGet, "/api/v1/datasets/"+eventLeads, nil).Decode(t, &stored)

	want := []dataset.DashboardRef{{ID: "DB-1001", Name: "Lubricants Europe overview"}}
	if diff := cmp.Diff(want, stored.UsedBy); diff != "" {
		t.Errorf("%s used by (-want +got):\n%s", eventLeads, diff)
	}
}

func TestQuery_DatasetSharedOrNot_FR044(t *testing.T) {
	t.Parallel()
	server := apitest.New(t)
	priya, sara := server.SignIn(t, priyasEmail), server.SignIn(t, sarasEmail)

	if got := len(ask(t, priya, leadsByEvent(), http.StatusOK).Rows); got != 12 {
		t.Errorf("the owner's widget has %d events, want 12", got)
	}
	ask(t, sara, leadsByEvent(), http.StatusNotFound)

	shareLeads(t, priya, []string{"T-1001"})
	if got := len(ask(t, sara, leadsByEvent(), http.StatusOK).Rows); got != 12 {
		t.Errorf("shared with her team, Sara's widget has %d events, want 12", got)
	}

	if answer := priya.Do(t, http.MethodDelete, "/api/v1/datasets/"+eventLeads, nil); answer.Status != http.StatusOK {
		t.Fatalf("deleting %s = %d %s", eventLeads, answer.Status, answer.Body)
	}
	ask(t, priya, leadsByEvent(), http.StatusNotFound)
	ask(t, sara, leadsByEvent(), http.StatusNotFound)
}
