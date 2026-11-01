package performance_test

import (
	"errors"
	"testing"

	"mapp/internal/performance"
	"mapp/internal/platform/authz"
	"mapp/internal/platform/errs"
)

func TestUndoImport_RestoresReplacedVersions_FR017(t *testing.T) {
	t.Parallel()
	f, src := newImportFixture(t)

	f.runImport(t, src.ID, map[string]string{"r1": "5"})
	second := f.runImport(t, src.ID, map[string]string{"r1": "7"})

	undone, err := f.service.UndoImport(t.Context(), alex, second.ID)
	if err != nil {
		t.Fatalf("UndoImport: %v", err)
	}
	if undone.Status != performance.ImportUndone {
		t.Errorf("the import is %s, want it listed as undone", undone.Status)
	}
	current := f.currentFacts(t, src.ID, "r1")
	if len(current) != 1 || current[0].Count == nil || *current[0].Count != 5 {
		t.Fatalf("current facts = %+v, want the first version counting 5 restored", current)
	}
	if facts, _ := f.store.ListFactsOfImport(t.Context(), second.ID); len(facts) != 0 {
		t.Errorf("the undone import still holds %d facts, want none", len(facts))
	}
}

func TestUndoImport_LeavesALaterUploadsFactsAlone_FR017(t *testing.T) {
	t.Parallel()
	f, src := newImportFixture(t)

	f.runImport(t, src.ID, map[string]string{"r1": "5"})
	second := f.runImport(t, src.ID, map[string]string{"r1": "7"})
	f.runImport(t, src.ID, map[string]string{"r1": "9"})

	if _, err := f.service.UndoImport(t.Context(), alex, second.ID); err != nil {
		t.Fatalf("UndoImport: %v", err)
	}
	current := f.currentFacts(t, src.ID, "r1")
	if len(current) != 1 || current[0].Count == nil || *current[0].Count != 9 {
		t.Fatalf("current facts = %+v, want only the later upload's fact counting 9", current)
	}
}

func TestUndoImport_NeedsImportAndSettingsManage_046C16(t *testing.T) {
	t.Parallel()
	f, src := newImportFixture(t)
	run := f.runImport(t, src.ID, map[string]string{"r1": "5"})

	ctx := authz.With(t.Context(), grantsWithout(authz.PerformanceImport))
	_, err := f.service.UndoImport(ctx, alex, run.ID)

	var refusal *errs.Error
	if !errors.As(err, &refusal) || refusal.Code != errs.CodeForbidden {
		t.Fatalf("UndoImport without settings.manage = %v, want forbidden: undo restates figures (046-C16)", err)
	}
}

func TestUndoImport_AnUndoneImportIsNotUndoneAgain_FR016(t *testing.T) {
	t.Parallel()
	f, src := newImportFixture(t)
	run := f.runImport(t, src.ID, map[string]string{"r1": "5"})

	if _, err := f.service.UndoImport(t.Context(), alex, run.ID); err != nil {
		t.Fatalf("first UndoImport: %v", err)
	}
	_, err := f.service.UndoImport(t.Context(), alex, run.ID)
	if err == nil {
		t.Fatal("UndoImport ran twice, want the second refused: the import is already undone")
	}
}
