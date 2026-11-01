package app_test

import (
	"net/http"
	"testing"

	"github.com/google/go-cmp/cmp"

	"mapp/internal/customdashboard"
	"mapp/internal/dataset"
	"mapp/internal/people"
	"mapp/internal/platform/apitest"
)

const lubricantsEurope = "T-1001"

func newTeamWithPriya(t *testing.T, alex *apitest.Client) string {
	t.Helper()
	var team people.Team
	answer := alex.Do(t, http.MethodPost, "/api/v1/teams",
		people.TeamInput{Name: "Event agencies", Colour: "blue", MemberIDs: []string{"U-1002"}})
	if answer.Status != http.StatusCreated {
		t.Fatalf("creating a team = %d %s", answer.Status, answer.Body)
	}
	answer.Decode(t, &team)
	return team.ID
}

func shareOverview(t *testing.T, priya *apitest.Client, teamIDs []string) {
	t.Helper()
	var stored customdashboard.View
	priya.Do(t, http.MethodGet, "/api/v1/custom-dashboards/"+overviewDashboard, nil).Decode(t, &stored)
	answer := priya.Do(t, http.MethodPut, "/api/v1/custom-dashboards/"+overviewDashboard+"/sharing",
		customdashboard.SharingInput{Version: &stored.Version, TeamIDs: teamIDs})
	if answer.Status != http.StatusOK {
		t.Fatalf("sharing %s = %d %s", overviewDashboard, answer.Status, answer.Body)
	}
}

func sharedTeamIDs(t *testing.T, priya *apitest.Client) (dashboard, data []string) {
	t.Helper()
	var d customdashboard.View
	var s dataset.View
	priya.Do(t, http.MethodGet, "/api/v1/custom-dashboards/"+overviewDashboard, nil).Decode(t, &d)
	priya.Do(t, http.MethodGet, "/api/v1/datasets/"+eventLeads, nil).Decode(t, &s)
	dashboard, data = []string{}, []string{}
	for _, team := range d.SharedWith {
		dashboard = append(dashboard, team.ID)
	}
	for _, team := range s.SharedWith {
		data = append(data, team.ID)
	}
	return dashboard, data
}

// D-379: deleting a team takes its dashboard and dataset shares with it, and restoring the team does not
// bring them back.
func TestDeleteTeam_SharesLeaveWithTheTeam_D379(t *testing.T) {
	t.Parallel()
	server := apitest.New(t)
	alex, priya := server.SignIn(t, alexsEmail), server.SignIn(t, priyasEmail)
	agencies := newTeamWithPriya(t, alex)
	shareOverview(t, priya, []string{lubricantsEurope, agencies})
	shareLeads(t, priya, []string{agencies})

	deleted := alex.Do(t, http.MethodDelete, "/api/v1/teams/"+agencies, nil)
	if deleted.Status != http.StatusOK {
		t.Fatalf("deleting the shared team = %d %s, want 200", deleted.Status, deleted.Body)
	}
	var bin struct{ RecycleBinEntry struct{ ID string } }
	deleted.Decode(t, &bin)
	if answer := alex.Do(t, http.MethodPost, "/api/v1/recycle-bin/"+bin.RecycleBinEntry.ID+"/restore", nil); answer.Status != http.StatusOK {
		t.Fatalf("restoring the team = %d %s", answer.Status, answer.Body)
	}

	dashboard, data := sharedTeamIDs(t, priya)
	if diff := cmp.Diff([]string{lubricantsEurope}, dashboard); diff != "" {
		t.Errorf("%s shared with (-want +got):\n%s", overviewDashboard, diff)
	}
	if diff := cmp.Diff([]string{}, data); diff != "" {
		t.Errorf("%s shared with (-want +got):\n%s", eventLeads, diff)
	}
}

func TestRestoreDashboard_AfterItsTeamWasDeleted_D379(t *testing.T) {
	t.Parallel()
	server := apitest.New(t)
	alex, priya := server.SignIn(t, alexsEmail), server.SignIn(t, priyasEmail)
	agencies := newTeamWithPriya(t, alex)
	shareOverview(t, priya, []string{lubricantsEurope, agencies})
	var bin struct{ RecycleBinEntry struct{ ID string } }
	priya.Do(t, http.MethodDelete, "/api/v1/custom-dashboards/"+overviewDashboard, nil).Decode(t, &bin)
	if answer := alex.Do(t, http.MethodDelete, "/api/v1/teams/"+agencies, nil); answer.Status != http.StatusOK {
		t.Fatalf("deleting the shared team = %d %s, want 200", answer.Status, answer.Body)
	}

	if answer := priya.Do(t, http.MethodPost, "/api/v1/recycle-bin/"+bin.RecycleBinEntry.ID+"/restore", nil); answer.Status != http.StatusOK {
		t.Fatalf("restoring the dashboard = %d %s, want 200", answer.Status, answer.Body)
	}
	if dashboard, _ := sharedTeamIDs(t, priya); !cmp.Equal([]string{lubricantsEurope}, dashboard) {
		t.Errorf("the restored %s is shared with %q, want only %s", overviewDashboard, dashboard, lubricantsEurope)
	}
}
