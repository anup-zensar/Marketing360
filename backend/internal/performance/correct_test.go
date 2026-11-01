package performance_test

import (
	"errors"
	"testing"

	"mapp/internal/performance"
	"mapp/internal/platform/errs"
)

func openRow(t *testing.T, f *fixture, src performance.Source) performance.ReviewRowView {
	t.Helper()
	f.runImport(t, src.ID, map[string]string{"r1": "oops"})
	rows, err := f.service.ReviewRows(t.Context(), alex, performance.ReviewFilter{Status: "open"})
	if err != nil || len(rows) != 1 {
		t.Fatalf("ReviewRows = %d (%v), want the one open row", len(rows), err)
	}
	return rows[0]
}

func TestDismiss_RequiresAComment_FR024(t *testing.T) {
	t.Parallel()
	f, src := reviewFixture(t)
	row := openRow(t, f, src)

	_, err := f.service.Dismiss(t.Context(), alex, row.ID, performance.DismissInput{Comment: "  "})

	var refusal *errs.Error
	if !errors.As(err, &refusal) || refusal.Code != errs.CodeValidationFailed {
		t.Fatalf("Dismiss without a comment = %v, want validation_failed on the comment", err)
	}
}

func TestDismissAndReopen_BothLandInTheHistory_FR024(t *testing.T) {
	t.Parallel()
	f, src := reviewFixture(t)
	row := openRow(t, f, src)

	dismissed, err := f.service.Dismiss(t.Context(), alex, row.ID, performance.DismissInput{Comment: "test data"})
	if err != nil || dismissed.Status != performance.ReviewDismissed || dismissed.Comment != "test data" {
		t.Fatalf("Dismiss = %+v (%v), want the row dismissed with its comment", dismissed, err)
	}
	reopened, err := f.service.Reopen(t.Context(), alex, row.ID)
	if err != nil || reopened.Status != performance.ReviewOpen || reopened.Comment != "" {
		t.Fatalf("Reopen = %+v (%v), want the row open again with the comment gone", reopened, err)
	}
	assertBothActionsKept(t, f, row.ID)
}

func assertBothActionsKept(t *testing.T, f *fixture, rowID string) {
	t.Helper()
	verbs := map[string]int{}
	for _, entry := range f.log(t) {
		if entry.RecordID == rowID {
			verbs[entry.Verb]++
		}
	}
	if verbs["dismissed"] != 1 || verbs["reopened"] != 1 {
		t.Errorf("history verbs = %v, want both actions kept", verbs)
	}
}

func TestReopen_OnlyADismissedRowReopens_FR024(t *testing.T) {
	t.Parallel()
	f, src := reviewFixture(t)
	row := openRow(t, f, src)

	_, err := f.service.Reopen(t.Context(), alex, row.ID)
	if err == nil {
		t.Fatal("Reopen ran on an open row, want it refused")
	}
}
