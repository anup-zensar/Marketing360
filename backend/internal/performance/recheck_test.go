package performance_test

import (
	"testing"

	"mapp/internal/performance"
)

func TestRecheckSource_AnOpenRowResolvesOnceItsKeyExists_FR025(t *testing.T) {
	t.Parallel()
	f, src := reviewFixture(t)
	rows := importRows(map[string]string{"r1": "5"})
	rows[0].Keys.ActivityID = "A-new"
	f.runImport(t, src.ID, map[string]string{})
	if _, err := f.service.RunImport(t.Context(), alex, performance.ImportInput{
		SourceID: src.ID, FileName: "early.csv", Origin: performance.OriginUpload,
		Headings: []string{"ref", "date", "inqs"}, Rows: rows,
	}); err != nil {
		t.Fatalf("RunImport: %v", err)
	}

	f.store.Keys[performance.MatchedActivityID]["a-new"] = []string{"A-1"}
	resolved, err := f.service.RecheckSource(t.Context(), alex, src.ID)
	if err != nil || resolved.Matched != 1 {
		t.Fatalf("RecheckSource = %+v (%v), want the waiting row matched now", resolved, err)
	}

	open, err := f.service.ReviewRows(t.Context(), alex, performance.ReviewFilter{Status: "open"})
	if err != nil || len(open) != 0 {
		t.Errorf("open rows = %d (%v), want none: the row resolved", len(open), err)
	}
	facts := f.currentFacts(t, src.ID, "r1")
	if len(facts) != 1 || !facts[0].Counts() {
		t.Fatalf("facts = %+v, want the re-checked row counting", facts)
	}
}

func TestRecheck_RunsAutomaticallyAtEachLoadOfTheSource_FR025(t *testing.T) {
	t.Parallel()
	f, src := reviewFixture(t)
	rows := importRows(map[string]string{"r1": "5"})
	rows[0].Keys.ActivityID = "A-new"
	if _, err := f.service.RunImport(t.Context(), alex, performance.ImportInput{
		SourceID: src.ID, FileName: "early.csv", Origin: performance.OriginUpload,
		Headings: []string{"ref", "date", "inqs"}, Rows: rows,
	}); err != nil {
		t.Fatalf("RunImport: %v", err)
	}
	f.store.Keys[performance.MatchedActivityID]["a-new"] = []string{"A-1"}

	f.runImport(t, src.ID, map[string]string{"r2": "3"})

	open, err := f.service.ReviewRows(t.Context(), alex, performance.ReviewFilter{Status: "open"})
	if err != nil || len(open) != 0 {
		t.Errorf("open rows = %d (%v), want the earlier row resolved by the later load's re-check", len(open), err)
	}
}

func TestRecheck_ReleasesHeldFactsWhoseActivityNowCountsAsApproved_FR025(t *testing.T) {
	t.Parallel()
	f, src := reviewFixture(t)
	approvals := approvedStub{"A-1": true, "A-2": false}
	f.service.UseApprovals(approvals)
	f.store.Keys[performance.MatchedActivityID]["a-2"] = []string{"A-2"}
	rows := importRows(map[string]string{"r1": "5"})
	rows[0].Keys.ActivityID = "A-2"
	if _, err := f.service.RunImport(t.Context(), alex, performance.ImportInput{
		SourceID: src.ID, FileName: "held.csv", Origin: performance.OriginUpload,
		Headings: []string{"ref", "date", "inqs"}, Rows: rows,
	}); err != nil {
		t.Fatalf("RunImport: %v", err)
	}

	// The restore-from-bin case: the activity counts as approved again, but no approval event fired.
	approvals["A-2"] = true
	if _, err := f.service.RecheckSource(t.Context(), alex, src.ID); err != nil {
		t.Fatalf("RecheckSource: %v", err)
	}

	facts := f.currentFacts(t, src.ID, "r1")
	if len(facts) != 1 || facts[0].Held {
		t.Fatalf("facts = %+v, want the hold released by the re-check", facts)
	}
}

func TestRecheck_LeavesRowsOfAnUndoneImportAlone_FR016(t *testing.T) {
	t.Parallel()
	f, src := reviewFixture(t)
	rows := importRows(map[string]string{"r1": "5"})
	rows[0].Keys.ActivityID = "A-new"
	run, err := f.service.RunImport(t.Context(), alex, performance.ImportInput{
		SourceID: src.ID, FileName: "early.csv", Origin: performance.OriginUpload,
		Headings: []string{"ref", "date", "inqs"}, Rows: rows,
	})
	if err != nil {
		t.Fatalf("RunImport: %v", err)
	}
	if _, err := f.service.UndoImport(t.Context(), alex, run.ID); err != nil {
		t.Fatalf("UndoImport: %v", err)
	}

	f.store.Keys[performance.MatchedActivityID]["a-new"] = []string{"A-1"}
	if _, err := f.service.RecheckSource(t.Context(), alex, src.ID); err != nil {
		t.Fatalf("RecheckSource: %v", err)
	}

	if facts := f.currentFacts(t, src.ID, "r1"); len(facts) != 0 {
		t.Errorf("facts = %+v, want none: a fact must never hang on an undone upload", facts)
	}
}
