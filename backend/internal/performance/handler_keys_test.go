package performance_test

import (
	"net/http"
	"strings"
	"testing"

	"mapp/internal/platform/apitest"
)

// demoActivityID is seeded by the demo data; the contract run registers keys on it.
func demoActivityID(t *testing.T, alex *apitest.Client) string {
	t.Helper()
	area := alex.Do(t, http.MethodGet, "/api/v1/activities/A-1001/performance/sources", nil)
	if area.Status != http.StatusOK {
		t.Fatalf("the demo activity's area answered %d %s", area.Status, area.Body)
	}
	return "A-1001"
}

func TestIntegrationIds_Contract_FR037(t *testing.T) {
	t.Parallel()
	alex := apitest.New(t).SignIn(t, alexsEmail)
	sourceID := preparedSource(t, alex)
	activityID := demoActivityID(t, alex)
	path := "/api/v1/activities/" + activityID + "/integration-ids/" + sourceID

	created := alex.Do(t, http.MethodPut, path, map[string]string{"value": "elq-77"})
	if created.Status != http.StatusCreated {
		t.Fatalf("PUT = %d %s, want 201", created.Status, created.Body)
	}

	refused := alex.Do(t, http.MethodPut, path, map[string]string{"value": "elq-88"})
	if refused.Status != http.StatusConflict || !strings.Contains(string(refused.Body), "elq-77") {
		t.Fatalf("a second key = %d %s, want 409 naming the one that exists", refused.Status, refused.Body)
	}

	listed := alex.Do(t, http.MethodGet, "/api/v1/activities/"+activityID+"/integration-ids", nil)
	if listed.Status != http.StatusOK || !strings.Contains(string(listed.Body), "elq-77") {
		t.Errorf("GET = %d %s, want the key listed", listed.Status, listed.Body)
	}

	area := alex.Do(t, http.MethodGet, "/api/v1/activities/"+activityID+"/performance/sources", nil)
	if area.Status != http.StatusOK || !strings.Contains(string(area.Body), `"integrated":true`) {
		t.Errorf("area = %d %s, want Is integrated set by the key", area.Status, area.Body)
	}

	if removed := alex.Do(t, http.MethodDelete, path, nil); removed.Status != http.StatusNoContent {
		t.Errorf("DELETE = %d %s, want 204", removed.Status, removed.Body)
	}
}

func TestExternalNameAliases_Contract_FR034(t *testing.T) {
	t.Parallel()
	alex := apitest.New(t).SignIn(t, alexsEmail)
	activityID := demoActivityID(t, alex)

	listed := alex.Do(t, http.MethodGet, "/api/v1/activities/"+activityID+"/external-name-aliases", nil)
	if listed.Status != http.StatusOK {
		t.Fatalf("GET aliases = %d %s, want 200", listed.Status, listed.Body)
	}

	missing := alex.Do(t, http.MethodDelete,
		"/api/v1/activities/"+activityID+"/external-name-aliases/ENA-9999", nil)
	if missing.Status != http.StatusNotFound {
		t.Errorf("deleting an unknown alias = %d %s, want 404", missing.Status, missing.Body)
	}
}
