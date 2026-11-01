package performance_test

import (
	"testing"

	"mapp/internal/performance"
)

// A binned activity's keys resolve to nothing; the row waits for the restore and re-check (046-C24).
func TestRunImport_ARowForABinnedActivityWaitsInTheReviewQueue_FR021(t *testing.T) {
	t.Parallel()
	f, src := newImportFixture(t)
	delete(f.store.Keys[performance.MatchedActivityID], "a-1")

	run := f.runImport(t, src.ID, map[string]string{"r1": "5"})

	if run.Counts.Unmatched != 1 {
		t.Fatalf("counts = %+v, want the row unmatched while its activity sits in the bin", run.Counts)
	}
	rows, err := f.store.ListReviewRows(t.Context(), "open", src.ID, 10)
	if err != nil || len(rows) != 1 {
		t.Fatalf("review rows = %d (%v), want the row waiting for the restore and re-check", len(rows), err)
	}
}
