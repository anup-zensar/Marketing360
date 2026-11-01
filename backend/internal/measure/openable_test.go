package measure_test

import (
	"net/http"
	"testing"

	"mapp/internal/platform/apitest"
)

func TestReportsOfAnotherRole_FrameAndExportRefused_039_FR003(t *testing.T) {
	t.Parallel()
	server := apitest.New(t)
	planner, admin := server.SignIn(t, priyaEmail), server.SignIn(t, alexEmail)
	frameOf := "/api/v1/reports/frame?report=userAccess"
	export := "/api/v1/reports/user-access/export?format=xlsx"

	for _, path := range []string{frameOf, export} {
		if r := planner.Do(t, http.MethodGet, path, nil); r.Status != http.StatusForbidden {
			t.Errorf("a planner GET %s = %d %s, want 403: User access needs access.manage", path, r.Status, r.Body)
		}
	}
	if r := admin.Do(t, http.MethodGet, frameOf, nil); r.Status != http.StatusOK {
		t.Errorf("the Administrator GET %s = %d %s, want 200", frameOf, r.Status, r.Body)
	}
}
