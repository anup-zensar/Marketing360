package performance_test

import (
	"testing"

	"mapp/internal/performance"
	"mapp/internal/platform/authz"
)

func reviewFixture(t *testing.T) (*fixture, performance.Source) {
	t.Helper()
	f, src := newImportFixture(t)
	f.store.Keys[performance.MatchedActivityID]["a-shared"] = []string{"A-2", "A-3"}
	return f, src
}

func TestReviewRows_EveryRejectedAndUnmatchedRowAppearsWithItsReason_FR023(t *testing.T) {
	t.Parallel()
	f, src := reviewFixture(t)
	rows := importRows(map[string]string{"r1": "oops"})
	rows = append(rows, importRows(map[string]string{"r2": "5"})...)
	rows[1].Keys.ActivityID = "A-missing"
	if _, err := f.service.RunImport(t.Context(), alex, performance.ImportInput{
		SourceID: src.ID, FileName: "mixed.csv", Origin: performance.OriginUpload,
		Headings: []string{"ref", "date", "inqs"}, Rows: rows,
	}); err != nil {
		t.Fatalf("RunImport: %v", err)
	}

	found, err := f.service.ReviewRows(t.Context(), alex, performance.ReviewFilter{Status: "open"})
	if err != nil || len(found) != 2 {
		t.Fatalf("ReviewRows = %d (%v), want the rejected and the unmatched row (SC-003)", len(found), err)
	}
	reasons := map[string]bool{}
	for _, row := range found {
		reasons[row.Reason] = true
	}
	if !reasons["rejected:invalidCount"] || !reasons["unmatched:unknown"] {
		t.Errorf("reasons = %v, want each row carrying its own stated reason", reasons)
	}
}

func TestReviewRows_FindableByReasonUploadAndKey_FR023(t *testing.T) {
	t.Parallel()
	f, src := reviewFixture(t)
	run := f.runImport(t, src.ID, map[string]string{"r1": "oops", "r2": "worse"})

	byReason, err := f.service.ReviewRows(t.Context(), alex, performance.ReviewFilter{Reason: "rejected:invalidCount"})
	if err != nil || len(byReason) != 2 {
		t.Fatalf("by reason = %d (%v), want both rejected rows", len(byReason), err)
	}
	byImport, err := f.service.ReviewRows(t.Context(), alex, performance.ReviewFilter{ImportID: run.ID})
	if err != nil || len(byImport) != 2 {
		t.Fatalf("by upload = %d (%v), want both rows of the load", len(byImport), err)
	}
	byKey, err := f.service.ReviewRows(t.Context(), alex, performance.ReviewFilter{KeyText: "r1"})
	if err != nil || len(byKey) != 1 {
		t.Fatalf("by key = %d (%v), want the one row whose raw content carries r1", len(byKey), err)
	}
}

func TestReviewRows_CandidatesAreFilteredToTheReadersScope_046C15(t *testing.T) {
	t.Parallel()
	f, src := reviewFixture(t)
	rows := importRows(map[string]string{"r1": "5"})
	rows[0].Keys.ActivityID = "A-shared"
	if _, err := f.service.RunImport(t.Context(), alex, performance.ImportInput{
		SourceID: src.ID, FileName: "ambiguous.csv", Origin: performance.OriginUpload,
		Headings: []string{"ref", "date", "inqs"}, Rows: rows,
	}); err != nil {
		t.Fatalf("RunImport: %v", err)
	}
	// The reader's matching tables reach A-2 alone, so A-3 must not be named to them.
	delete(f.store.Keys[performance.MatchedActivityID], "a-3")
	f.store.Keys[performance.MatchedActivityID]["a-2"] = []string{"A-2"}

	found, err := f.service.ReviewRows(t.Context(), alex, performance.ReviewFilter{Status: "open"})
	if err != nil || len(found) != 1 {
		t.Fatalf("ReviewRows = %d (%v), want the ambiguous row still listed", len(found), err)
	}
	if len(found[0].Candidates) != 1 || found[0].Candidates[0] != "A-2" {
		t.Errorf("candidates = %v, want only the activity within the reader's scope named", found[0].Candidates)
	}
	if found[0].Reason != "unmatched:ambiguous" {
		t.Errorf("reason = %s, want the row keeping its stated reason", found[0].Reason)
	}
}

func TestReviewRows_NeedPerformanceImport_FR043(t *testing.T) {
	t.Parallel()
	f, _ := reviewFixture(t)

	ctx := authz.With(t.Context(), grantsWithout(authz.PlansView))
	if _, err := f.service.ReviewRows(ctx, alex, performance.ReviewFilter{}); err == nil {
		t.Fatal("ReviewRows answered without performance.import, want it to fail closed")
	}
}
