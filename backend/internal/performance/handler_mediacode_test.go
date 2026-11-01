package performance_test

import (
	"net/http"
	"strings"
	"testing"

	"mapp/internal/platform/apitest"
)

func TestMediaCodes_Contract_FR039_FR042(t *testing.T) {
	t.Parallel()
	alex := apitest.New(t).SignIn(t, alexsEmail)

	read := alex.Do(t, http.MethodGet, "/api/v1/performance/configuration", nil)
	if read.Status != http.StatusOK {
		t.Fatalf("GET configuration = %d %s", read.Status, read.Body)
	}
	var config struct {
		MediaCodeNext int64 `json:"mediaCodeNext"`
		Version       int   `json:"version"`
	}
	read.Decode(t, &config)

	activity := alex.Do(t, http.MethodGet, "/api/v1/activities/A-1001", nil)
	var act struct {
		ActivityType struct {
			ID string `json:"id"`
		} `json:"activityType"`
	}
	activity.Decode(t, &act)

	saved := alex.Do(t, http.MethodPatch, "/api/v1/performance/configuration", map[string]any{
		"activityTypeIds": []string{act.ActivityType.ID},
		"codeAttributes":  []map[string]any{{"name": "Channel", "required": true}},
		"mediaCodeNext":   config.MediaCodeNext,
		"version":         config.Version,
	})
	if saved.Status != http.StatusOK {
		t.Fatalf("PATCH configuration = %d %s", saved.Status, saved.Body)
	}

	missing := alex.Do(t, http.MethodPost, "/api/v1/activities/A-1001/media-codes", map[string]any{})
	if missing.Status != http.StatusBadRequest || !strings.Contains(string(missing.Body), "Channel") {
		t.Fatalf("generate without attributes = %d %s, want every missing attribute named", missing.Status, missing.Body)
	}

	assertGenerateAndDelete(t, alex)
}

func assertGenerateAndDelete(t *testing.T, alex *apitest.Client) {
	t.Helper()
	created := alex.Do(t, http.MethodPost, "/api/v1/activities/A-1001/media-codes", map[string]any{
		"attributes": map[string]string{"Channel": "paid-social"},
	})
	if created.Status != http.StatusCreated || !strings.Contains(string(created.Body), `"code":9000001`) {
		t.Fatalf("generate = %d %s, want code 9000001", created.Status, created.Body)
	}
	codeID := idOf(t, string(created.Body), "MC-")

	area := alex.Do(t, http.MethodGet, "/api/v1/activities/A-1001/media-codes", nil)
	if area.Status != http.StatusOK || !strings.Contains(string(area.Body), `"requiresCodes":true`) {
		t.Errorf("area = %d %s, want the section present for the configured type", area.Status, area.Body)
	}

	if removed := alex.Do(t, http.MethodDelete, "/api/v1/media-codes/"+codeID, nil); removed.Status != http.StatusNoContent {
		t.Errorf("DELETE unused = %d %s, want 204", removed.Status, removed.Body)
	}
}
