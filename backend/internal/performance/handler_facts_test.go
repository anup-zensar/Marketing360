package performance_test

import (
	"net/http"
	"strings"
	"testing"

	"mapp/internal/platform/apitest"
)

func TestActivityFacts_Contract_FR030(t *testing.T) {
	t.Parallel()
	alex := apitest.New(t).SignIn(t, alexsEmail)
	sourceID := preparedSource(t, alex)
	file := "Ref,Date,Activity,INQs\nr1,2026-02-01,A-1001,5\n"
	upload := alex.Do(t, http.MethodPost, "/api/v1/performance/sources/"+sourceID+"/uploads", uploadBody(file))
	if upload.Status != http.StatusCreated {
		t.Fatalf("upload = %d %s", upload.Status, upload.Body)
	}

	listed := alex.Do(t, http.MethodGet, "/api/v1/activities/A-1001/performance/facts", nil)
	if listed.Status != http.StatusOK {
		t.Fatalf("facts = %d %s", listed.Status, listed.Body)
	}
	body := string(listed.Body)
	for _, want := range []string{`"factDate":"2026-02-01"`, `"stage":"inq"`, `"matchedOn":"activityId"`, `"sourceName":"Eloqua"`} {
		if !strings.Contains(body, want) {
			t.Errorf("the fact list misses %s: %s", want, body)
		}
	}

	exported := alex.Do(t, http.MethodGet, "/api/v1/activities/A-1001/performance/facts/export", nil)
	if exported.Status != http.StatusOK || !strings.Contains(string(exported.Body), "Key matched on") {
		t.Errorf("export = %d %s, want the CSV with its headings", exported.Status, exported.Body)
	}
}
