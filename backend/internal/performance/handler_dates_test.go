package performance

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"mapp/internal/platform/errs"
)

func queryContext(t *testing.T, rawQuery string) *gin.Context {
	t.Helper()
	gin.SetMode(gin.ReleaseMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/?"+rawQuery, nil)
	return c
}

func TestDateRangeOf_EveryBadDateReportedTogether_FR016_FR023(t *testing.T) {
	t.Parallel()
	_, _, err := dateRangeOf(queryContext(t, "since=not-a-date&until=31-12-2026"))

	refusal, ok := errors.AsType[*errs.Error](err)
	if !ok || len(refusal.Fields) != 2 {
		t.Fatalf("dateRangeOf() = %v, want since and until refused together as YYYY-MM-DD", err)
	}
}

func TestDateRangeOf_UntilCoversItsWholeDay_FR016(t *testing.T) {
	t.Parallel()
	since, until, err := dateRangeOf(queryContext(t, "since=2026-03-12&until=2026-03-12"))
	if err != nil || since == nil || until == nil {
		t.Fatalf("dateRangeOf() = %v %v (%v), want both ends parsed", since, until, err)
	}
	if !until.After(*since) || until.Day() != 12 {
		t.Errorf("until = %v, want the last instant of the same day so one-day ranges match that day", until)
	}
}
