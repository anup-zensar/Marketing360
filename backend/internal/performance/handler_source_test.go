package performance_test

import (
	"net/http"
	"strings"
	"testing"

	"mapp/internal/platform/apitest"
)

func TestSources_Contract_FR001_FR002_FR003(t *testing.T) {
	t.Parallel()
	alex := apitest.New(t).SignIn(t, alexsEmail)
	sourceID := preparedSource(t, alex)

	// FR-003: a second active source may not supply the same model stage.
	rival := alex.Do(t, http.MethodPost, "/api/v1/performance/sources", map[string]any{
		"name": "Marketo", "ownerId": "U-1001", "category": "marketingAutomation",
		"deliveryRoute": "upload", "cadence": "monthly", "description": "", "integrationId": "",
	})
	if rival.Status != http.StatusCreated {
		t.Fatalf("create rival = %d %s", rival.Status, rival.Body)
	}
	rivalID := idOf(t, string(rival.Body), "PS-")
	duplicate := alex.Do(t, http.MethodPost, "/api/v1/performance/sources/"+rivalID+"/mappings", map[string]any{
		"columnName": "INQs", "meaning": "stage", "stage": "inq", "modelId": "MOD-1001",
	})
	if duplicate.Status != http.StatusConflict || !strings.Contains(string(duplicate.Body), "Eloqua") {
		t.Fatalf("a second stage supplier = %d %s, want 409 naming Eloqua", duplicate.Status, duplicate.Body)
	}

	// FR-002: once facts exist the source cannot be deleted, and Deactivate is offered.
	file := "Ref,Date,Activity,INQs\nr1,2026-02-01,A-1001,5\n"
	if upload := alex.Do(t, http.MethodPost, "/api/v1/performance/sources/"+sourceID+"/uploads",
		uploadBody(file)); upload.Status != http.StatusCreated {
		t.Fatalf("upload = %d %s", upload.Status, upload.Body)
	}
	refused := alex.Do(t, http.MethodDelete, "/api/v1/performance/sources/"+sourceID, nil)
	if refused.Status != http.StatusConflict || !strings.Contains(string(refused.Body), "Deactivate") {
		t.Fatalf("delete with facts = %d %s, want 409 offering Deactivate", refused.Status, refused.Body)
	}

	assertDeactivateAndCleanDelete(t, alex, sourceID, rivalID)
}

// FR-004: deactivation flips the source at once; the rival without facts deletes cleanly.
func assertDeactivateAndCleanDelete(t *testing.T, alex *apitest.Client, sourceID, rivalID string) {
	t.Helper()
	if off := alex.Do(t, http.MethodPost, "/api/v1/performance/sources/"+sourceID+"/deactivate", nil); off.Status != http.StatusOK ||
		!strings.Contains(string(off.Body), `"status":"inactive"`) {
		t.Errorf("deactivate = %d %s, want the source inactive", off.Status, off.Body)
	}
	if gone := alex.Do(t, http.MethodDelete, "/api/v1/performance/sources/"+rivalID, nil); gone.Status != http.StatusNoContent {
		t.Errorf("delete unused = %d %s, want 204", gone.Status, gone.Body)
	}
}

func TestTemplate_Contract_FR008(t *testing.T) {
	t.Parallel()
	alex := apitest.New(t).SignIn(t, alexsEmail)
	sourceID := preparedSource(t, alex)

	template := alex.Do(t, http.MethodGet, "/api/v1/performance/sources/"+sourceID+"/template", nil)
	if template.Status != http.StatusOK || strings.TrimSpace(string(template.Body)) != "Ref,Date,Activity,INQs" {
		t.Fatalf("template = %d %q, want exactly the mapped columns", template.Status, template.Body)
	}
}
