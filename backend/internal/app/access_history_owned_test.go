package app_test

import (
	"net/http"
	"slices"
	"testing"

	"mapp/internal/platform/apitest"
)

const overviewDashboard = "DB-1001"

func logRecordIDs(t *testing.T, c *apitest.Client, recordID string) []string {
	t.Helper()
	var page struct{ Items []struct{ RecordID string } }
	c.Do(t, http.MethodGet, "/api/v1/activity-log?limit=50&recordId="+recordID, nil).Decode(t, &page)
	ids := []string{}
	for _, e := range page.Items {
		ids = append(ids, e.RecordID)
	}
	return ids
}

func binNames(t *testing.T, c *apitest.Client) []string {
	t.Helper()
	names := []string{}
	for _, e := range binOf(t, c) {
		names = append(names, e.RecordName)
	}
	return names
}

func TestHistory_OwnedRecordsLogToOwnerAndAuditors_FR009_FR045(t *testing.T) {
	t.Parallel()
	server := apitest.New(t)
	priya, alex, daniel := server.SignIn(t, priyasEmail), server.SignIn(t, alexsEmail), server.SignIn(t, danielsEmail)
	shareLeads(t, priya, []string{"T-1001"})
	if answer := priya.Do(t, http.MethodDelete, "/api/v1/custom-dashboards/"+overviewDashboard, nil); answer.Status != http.StatusOK {
		t.Fatalf("deleting %s = %d %s", overviewDashboard, answer.Status, answer.Body)
	}

	for _, id := range []string{eventLeads, overviewDashboard} {
		if got := logRecordIDs(t, priya, id); len(got) == 0 {
			t.Errorf("the owner sees no log entry of %s", id)
		}
		if got := logRecordIDs(t, alex, id); len(got) == 0 {
			t.Errorf("an administrator, who sees the whole log, sees no log entry of %s", id)
		}
		if got := logRecordIDs(t, daniel, id); len(got) != 0 {
			t.Errorf("Daniel, a member of a shared team, sees the log entries %q of %s", got, id)
		}
	}
}

func TestRecycleBin_OwnedRecordsOnlyToTheOwner_FR010(t *testing.T) {
	t.Parallel()
	server := apitest.New(t)
	priya, alex := server.SignIn(t, priyasEmail), server.SignIn(t, alexsEmail)
	var deleted struct{ RecycleBinEntry struct{ ID string } }
	priya.Do(t, http.MethodDelete, "/api/v1/custom-dashboards/"+overviewDashboard, nil).Decode(t, &deleted)
	restore := "/api/v1/recycle-bin/" + deleted.RecycleBinEntry.ID + "/restore"

	if names := binNames(t, priya); !slices.Contains(names, "Lubricants Europe overview") {
		t.Errorf("Priya's recycle bin = %q, want her dashboard in it", names)
	}
	if names := binNames(t, alex); slices.Contains(names, "Lubricants Europe overview") {
		t.Errorf("the administrator's recycle bin = %q, want Priya's dashboard left out", names)
	}
	if answer := alex.Do(t, http.MethodPost, restore, nil); answer.Status != http.StatusNotFound {
		t.Errorf("the administrator restoring it = %d %s, want 404", answer.Status, answer.Body)
	}
	if answer := priya.Do(t, http.MethodPost, restore, nil); answer.Status != http.StatusOK {
		t.Errorf("Priya restoring it = %d %s, want 200", answer.Status, answer.Body)
	}
	if answer := priya.Do(t, http.MethodGet, "/api/v1/custom-dashboards/"+overviewDashboard, nil); answer.Status != http.StatusOK {
		t.Errorf("the restored dashboard = %d %s, want 200", answer.Status, answer.Body)
	}
}
