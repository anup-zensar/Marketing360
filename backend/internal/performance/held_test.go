package performance_test

import (
	"testing"

	"mapp/internal/performance"
)

func TestRelease_ClearsOnlyTheApprovedActivitysHold_FR020(t *testing.T) {
	t.Parallel()
	f, src := newImportFixture(t)
	f.store.Keys[performance.MatchedActivityID]["a-2"] = []string{"A-2"}

	held := f.runImport(t, src.ID, map[string]string{"r1": "5"})
	if held.Counts.Matched != 1 {
		t.Fatalf("counts = %+v, want r1 matched to the approved A-1", held.Counts)
	}
	rows := importRows(map[string]string{"r2": "7"})
	rows[0].Keys.ActivityID = "A-2"
	run, err := f.service.RunImport(t.Context(), alex, performance.ImportInput{
		SourceID: src.ID, FileName: "held.csv", Origin: performance.OriginUpload,
		Headings: []string{"ref", "date", "inqs"}, Rows: rows,
	})
	if err != nil || run.Counts.Held != 1 {
		t.Fatalf("counts = %+v (%v), want r2 held: A-2 does not count as approved (046-C1)", run.Counts, err)
	}

	released, err := f.service.Release(t.Context(), alex, "A-2")
	if err != nil || released != 1 {
		t.Fatalf("Release(A-2) = %d (%v), want the one held fact released", released, err)
	}
	for _, fact := range f.currentFacts(t, src.ID, "r1", "r2") {
		if fact.Held {
			t.Errorf("fact %s on %s is still held after the approval", fact.ID, fact.ActivityID)
		}
		if !fact.Counts() {
			t.Errorf("fact %s does not count after release (FR-020)", fact.ID)
		}
	}
}

func TestRelease_LeavesOtherActivitiesHeldFactsAlone_FR020(t *testing.T) {
	t.Parallel()
	f, src := newImportFixture(t)
	f.store.Keys[performance.MatchedActivityID]["a-2"] = []string{"A-2"}
	f.store.Keys[performance.MatchedActivityID]["a-3"] = []string{"A-3"}

	for reference, activity := range map[string]string{"r2": "A-2", "r3": "A-3"} {
		rows := importRows(map[string]string{reference: "7"})
		rows[0].Keys.ActivityID = activity
		if _, err := f.service.RunImport(t.Context(), alex, performance.ImportInput{
			SourceID: src.ID, FileName: reference + ".csv", Origin: performance.OriginUpload,
			Headings: []string{"ref", "date", "inqs"}, Rows: rows,
		}); err != nil {
			t.Fatalf("RunImport(%s): %v", reference, err)
		}
	}

	if _, err := f.service.Release(t.Context(), alex, "A-2"); err != nil {
		t.Fatalf("Release: %v", err)
	}
	facts := f.currentFacts(t, src.ID, "r2", "r3")
	for _, fact := range facts {
		if fact.ActivityID == "A-3" && !fact.Held {
			t.Errorf("A-3's fact was released by A-2's approval")
		}
	}
}
