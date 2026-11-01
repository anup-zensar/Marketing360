package gsaplink

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func serve(t *testing.T, w *world, method, target string) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)
	router := gin.New()
	NewHandler(w.service()).RegisterRoutes(router)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequestWithContext(t.Context(), method, target, nil))
	return recorder
}

func TestHandler_TabSortPage_FR015(t *testing.T) {
	t.Parallel()
	recorder := serve(t, twoPlans(t), http.MethodGet, "/reports/gsap-link?year=2026&tab=swbs&sort=-lineItem&page=1&pageSize=5")

	var r Report
	if err := json.Unmarshal(recorder.Body.Bytes(), &r); err != nil || recorder.Code != http.StatusOK {
		t.Fatalf("answer %d %s, %v", recorder.Code, recorder.Body, err)
	}
	if r.Tab != TabSwbs || r.Sort != "-lineItem" || r.PageSize != 5 || r.Total != 3 {
		t.Errorf("report = tab %s, sort %s, page size %d, total %d", r.Tab, r.Sort, r.PageSize, r.Total)
	}
}

func TestHandler_RejectsUnknownTabSortPageSize_FR015(t *testing.T) {
	t.Parallel()
	for _, query := range []string{"tab=all", "sort=vendor", "tab=orders&sort=-waiting", "page=0", "pageSize=4",
		"pageSize=501", "pageSize=ten", "show=lineItem"} {
		if recorder := serve(t, twoPlans(t), http.MethodGet, "/reports/gsap-link?"+query); recorder.Code != http.StatusBadRequest {
			t.Errorf("%s: answer %d %s, want 400", query, recorder.Code, recorder.Body)
		}
	}
}

func TestRoutes_ReadOnly_FR002(t *testing.T) {
	t.Parallel()
	for _, method := range []string{http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete} {
		if recorder := serve(t, twoPlans(t), method, "/reports/gsap-link"); recorder.Code != http.StatusNotFound {
			t.Errorf("%s: answer %d, want 404: the report has only a GET", method, recorder.Code)
		}
	}
}
