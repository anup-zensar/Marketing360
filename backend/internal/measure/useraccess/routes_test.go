package useraccess

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func serve(t *testing.T, method, target string) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)
	router := gin.New()
	o := newOffice()
	NewHandler(NewService(fakeTx{}, o, o, func() time.Time { return now })).RegisterRoutes(router)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequestWithContext(t.Context(), method, target, nil))
	return recorder
}

func TestRoutes_ReadOnly_FR004(t *testing.T) {
	t.Parallel()
	if r := serve(t, http.MethodGet, "/reports/user-access?status=all&sort=-lastSignedIn"); r.Code != http.StatusOK {
		t.Errorf("GET = %d %s, want 200", r.Code, r.Body)
	}
	if r := serve(t, http.MethodGet, "/reports/user-access?sort=salary"); r.Code != http.StatusBadRequest {
		t.Errorf("GET with an unknown sort = %d, want 400", r.Code)
	}
	for _, method := range []string{http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete} {
		if r := serve(t, method, "/reports/user-access"); r.Code != http.StatusNotFound {
			t.Errorf("%s = %d, want 404: the report has only a GET", method, r.Code)
		}
	}
}
