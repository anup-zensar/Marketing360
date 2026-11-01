package performance_test

import (
	"net/http"
	"strings"
	"testing"

	"mapp/internal/platform/apitest"
)

func TestReviewRows_Contract_FR023_FR024(t *testing.T) {
	t.Parallel()
	alex := apitest.New(t).SignIn(t, alexsEmail)
	sourceID := preparedSource(t, alex)
	file := "Ref,Date,Activity,INQs\nr1,2026-02-01,A-9999,oops\n"
	upload := alex.Do(t, http.MethodPost, "/api/v1/performance/sources/"+sourceID+"/uploads", uploadBody(file))
	if upload.Status != http.StatusCreated {
		t.Fatalf("upload = %d %s", upload.Status, upload.Body)
	}

	listed := alex.Do(t, http.MethodGet, "/api/v1/performance/review-rows?status=open&sourceId="+sourceID, nil)
	if listed.Status != http.StatusOK || !strings.Contains(string(listed.Body), "rejected:invalidCount") {
		t.Fatalf("list = %d %s, want the rejected row with its reason", listed.Status, listed.Body)
	}
	rowID := idOf(t, string(listed.Body), "PRV-")

	noComment := alex.Do(t, http.MethodPost, "/api/v1/performance/review-rows/"+rowID+"/dismiss",
		map[string]string{"comment": "  "})
	if noComment.Status != http.StatusBadRequest {
		t.Errorf("dismiss without a comment = %d %s, want validation_failed", noComment.Status, noComment.Body)
	}

	corrected := alex.Do(t, http.MethodPost, "/api/v1/performance/review-rows/"+rowID+"/correct",
		map[string]any{"values": []map[string]string{{"column": "INQs", "value": "5"}}})
	if corrected.Status != http.StatusOK {
		t.Fatalf("correct = %d %s, want 200 with the row's new verdict", corrected.Status, corrected.Body)
	}

	recheck := alex.Do(t, http.MethodPost, "/api/v1/performance/sources/"+sourceID+"/recheck", nil)
	if recheck.Status != http.StatusOK {
		t.Errorf("recheck = %d %s, want the resolved counts", recheck.Status, recheck.Body)
	}
}

// idOf pulls the first readable id with the given prefix out of a JSON answer.
func idOf(t *testing.T, body, prefix string) string {
	t.Helper()
	at := strings.Index(body, prefix)
	if at < 0 {
		t.Fatalf("no %s id in %s", prefix, body)
	}
	end := at
	for end < len(body) && body[end] != '"' {
		end++
	}
	return body[at:end]
}
