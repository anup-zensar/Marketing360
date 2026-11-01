package plandetails

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"mapp/internal/allwork"
)

func serve(t *testing.T, scoped *fakeScoped, target string) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)
	router := gin.New()
	NewHandler(NewService(fakeTx{}, scoped)).RegisterRoutes(router)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequestWithContext(t.Context(), http.MethodGet, target, nil))
	return recorder
}

func TestHandler_ShowDefaultsToActivity_FR001(t *testing.T) {
	t.Parallel()
	scoped := &fakeScoped{matching: companyCodeYear(t).matching(), attributes: []allwork.Attribute{companyCode()}}

	recorder := serve(t, scoped, "/reports/plan-details?year=2026")

	var r Report
	if err := json.Unmarshal(recorder.Body.Bytes(), &r); err != nil || recorder.Code != http.StatusOK {
		t.Fatalf("answer %d %s, %v", recorder.Code, recorder.Body, err)
	}
	if r.Show != LevelActivity || r.Cards.Records != 28 || len(r.Attributes) != 1 {
		t.Errorf("report = %+v, want the Activity level of the 28 activities", r)
	}
}

func TestHandler_MalformedFilter_400(t *testing.T) {
	t.Parallel()
	recorder := serve(t, &fakeScoped{}, "/reports/plan-details?show=lineItem&period=year")

	if recorder.Code != http.StatusBadRequest {
		t.Errorf("answer %d %s, want 400 with the problems of show and period", recorder.Code, recorder.Body)
	}
}
