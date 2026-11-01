package performance_test

import (
	"testing"

	"mapp/internal/performance"
)

func TestCorrect_AFixedRowResolvesAndItsFactCounts_FR024(t *testing.T) {
	t.Parallel()
	f, src := reviewFixture(t)
	row := openRow(t, f, src)

	saved, err := f.service.Correct(t.Context(), alex, row.ID, performance.CorrectInput{
		Values: []performance.CorrectCell{{Column: "inqs", Value: "5"}},
	})
	if err != nil {
		t.Fatalf("Correct: %v", err)
	}
	if saved.Status != performance.ReviewResolved {
		t.Errorf("the corrected row is %s, want resolved: the fixed count now matches", saved.Status)
	}
	facts := f.currentFacts(t, src.ID, "r1")
	if len(facts) != 1 || facts[0].Count == nil || *facts[0].Count != 5 {
		t.Fatalf("facts = %+v, want the corrected count stored", facts)
	}
	if facts[0].ImportID == "" {
		t.Error("the re-checked fact names no originating upload")
	}
}

func TestCorrect_RefusesAColumnTheRowNeverCarried_FR024(t *testing.T) {
	t.Parallel()
	f, src := reviewFixture(t)
	row := openRow(t, f, src)

	_, err := f.service.Correct(t.Context(), alex, row.ID, performance.CorrectInput{
		Values: []performance.CorrectCell{{Column: "mqls", Value: "5"}},
	})
	if err == nil {
		t.Fatal("Correct accepted a column the row never carried, want a refusal naming it")
	}
}

func TestPersist_ALaterLoadSupersedesACorrectionInProgress_FR026(t *testing.T) {
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
	open, err := f.service.ReviewRows(t.Context(), alex, performance.ReviewFilter{Status: "open"})
	if err != nil || len(open) != 1 {
		t.Fatalf("open rows = %d (%v), want the unmatched row waiting", len(open), err)
	}
	// A correction is in progress: the key is fixed by hand, but the activity is still unknown, so the
	// row stays open and corrected.
	if _, correctErr := f.service.Correct(t.Context(), alex, open[0].ID, performance.CorrectInput{
		Keys: &performance.RowKeys{ActivityID: "A-still-unknown"},
	}); correctErr != nil {
		t.Fatalf("Correct: %v", correctErr)
	}

	f.runImport(t, src.ID, map[string]string{"r1": "9"})

	assertSourceWon(t, f, src, open[0].ID)
}

func assertSourceWon(t *testing.T, f *fixture, src performance.Source, rowID string) {
	t.Helper()
	after, err := f.service.ReviewRowByID(t.Context(), alex, rowID)
	if err != nil || after.Status != performance.ReviewResolved {
		t.Fatalf("the corrected row is %+v (%v), want it resolved: the source wins (FR-026)", after.Status, err)
	}
	facts := f.currentFacts(t, src.ID, "r1")
	if len(facts) != 1 || facts[0].Count == nil || *facts[0].Count != 9 {
		t.Fatalf("facts = %+v, want only the later load's figure of 9", facts)
	}
}
