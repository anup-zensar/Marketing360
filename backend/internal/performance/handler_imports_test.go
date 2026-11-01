package performance_test

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"mapp/internal/performance"
	"mapp/internal/platform/apitest"
)

const alexsEmail = "alex.morgan@shell.com"

type importAnswer struct {
	ID     string             `json:"id"`
	Status string             `json:"status"`
	Counts performance.Counts `json:"counts"`
}

func decoded[T any](t *testing.T, response *apitest.Response, status int) T {
	t.Helper()
	if response.Status != status {
		t.Fatalf("status %d %s, want %d", response.Status, response.Body, status)
	}
	var v T
	response.Decode(t, &v)
	return v
}

func uploadBody(file string) map[string]string {
	return map[string]string{
		"fileName": "load.csv",
		"content":  base64.StdEncoding.EncodeToString([]byte(file)),
	}
}

// preparedSource creates a source and its mappings through the API, exactly as a person would.
func preparedSource(t *testing.T, alex *apitest.Client) string {
	t.Helper()
	source := decoded[map[string]any](t, alex.Do(t, http.MethodPost, "/api/v1/performance/sources", map[string]any{
		"name": "Eloqua", "ownerId": "U-1001", "category": "marketingAutomation",
		"deliveryRoute": "upload", "cadence": "monthly", "description": "", "integrationId": "",
	}), http.StatusCreated)
	id, _ := source["id"].(string)
	mappings := []map[string]any{
		{"columnName": "Ref", "meaning": "reference"},
		{"columnName": "Date", "meaning": "factDate"},
		{"columnName": "Activity", "meaning": "key", "keyType": "activityId"},
		{"columnName": "INQs", "meaning": "stage", "stage": "inq", "modelId": "MOD-1001"},
	}
	for _, m := range mappings {
		response := alex.Do(t, http.MethodPost, "/api/v1/performance/sources/"+id+"/mappings", m)
		if response.Status != http.StatusCreated {
			t.Fatalf("mapping %v = %d %s", m, response.Status, response.Body)
		}
	}
	return id
}

func TestUploadPerformanceFile_Contract_FR009_FR022(t *testing.T) {
	t.Parallel()
	alex := apitest.New(t).SignIn(t, alexsEmail)
	sourceID := preparedSource(t, alex)
	file := "Ref,Date,Activity,INQs\nr1,2026-02-01,A-9999,5\nr2,not-a-date,A-9999,3\n"

	created := decoded[importAnswer](t, alex.Do(t, http.MethodPost,
		"/api/v1/performance/sources/"+sourceID+"/uploads", uploadBody(file)), http.StatusCreated)

	// A-9999 is no demo activity, so the good row reads as unmatched; matching itself is proved by the
	// matcher's own tests, and this test proves the translation.
	if created.Status != "completed" || !created.Counts.Reconciles() || created.Counts.Read != 2 {
		t.Fatalf("upload = %+v, want a completed load whose counts account for both rows", created)
	}
	if created.Counts.Rejected != 1 || created.Counts.Unmatched != 1 {
		t.Errorf("counts = %+v, want the unreadable date rejected and the unknown key unmatched", created.Counts)
	}

	assertFindableAndUndoable(t, alex, sourceID, created.ID)
}

func assertFindableAndUndoable(t *testing.T, alex *apitest.Client, sourceID, importID string) {
	t.Helper()
	listed := alex.Do(t, http.MethodGet, "/api/v1/performance/imports?sourceId="+sourceID, nil)
	if listed.Status != http.StatusOK || !strings.Contains(string(listed.Body), importID) {
		t.Errorf("the upload is not findable in the list: %d %s", listed.Status, listed.Body)
	}
	undone := alex.Do(t, http.MethodPost, fmt.Sprintf("/api/v1/performance/imports/%s/undo", importID), nil)
	if undone.Status != http.StatusOK || !strings.Contains(string(undone.Body), `"undone"`) {
		t.Errorf("undo = %d %s, want the upload listed as undone (FR-017)", undone.Status, undone.Body)
	}
}

func TestUploadPerformanceFile_PersonalDataRefused_Contract_FR012(t *testing.T) {
	t.Parallel()
	alex := apitest.New(t).SignIn(t, alexsEmail)
	sourceID := preparedSource(t, alex)
	file := "Ref,Date,Activity,INQs,Email\nr1,2026-02-01,A-9999,5,x@shell.com\n"

	refused := alex.Do(t, http.MethodPost,
		"/api/v1/performance/sources/"+sourceID+"/uploads", uploadBody(file))

	if refused.Status != http.StatusConflict || !strings.Contains(string(refused.Body), "Email") {
		t.Fatalf("a personal-data file = %d %s, want 409 naming the column (SC-007)", refused.Status, refused.Body)
	}
}
