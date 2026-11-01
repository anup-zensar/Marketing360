package performance_test

import (
	"errors"
	"testing"

	"mapp/internal/performance"
	"mapp/internal/platform/authz"
	"mapp/internal/platform/errs"
	"mapp/internal/platform/record"
)

func keysFixture(t *testing.T) (*fixture, performance.Source) {
	t.Helper()
	f := newFixture(t)
	f.store.Keys = map[performance.MatchedOn]map[string][]string{
		performance.MatchedExternalName: {"spring fair": {"A-1"}, "spring fair 2025": {"A-1"}},
	}
	f.store.Activities = map[string]performance.KeyFacts{
		"A-1": {ExternalName: "Spring fair", GoalEntered: true},
	}
	f.service.UseApprovals(approvedStub{"A-1": true})
	return f, f.createSource(t, "Eloqua")
}

func TestClassify_TheCurrentExternalNameMatchesAndIsRecorded_FR033(t *testing.T) {
	t.Parallel()
	f, src := keysFixture(t)

	rows := importRows(map[string]string{"r1": "5"})
	rows[0].Keys = performance.RowKeys{ExternalName: "Spring fair"}
	run, err := f.service.RunImport(t.Context(), alex, performance.ImportInput{
		SourceID: src.ID, FileName: "names.csv", Origin: performance.OriginUpload,
		Headings: []string{"ref", "date", "inqs"}, Rows: rows,
	})
	if err != nil || run.Counts.Matched != 1 {
		t.Fatalf("counts = %+v (%v), want the current external name matched", run.Counts, err)
	}
	facts := f.currentFacts(t, src.ID, "r1")
	if len(facts) != 1 || facts[0].MatchedOn != performance.MatchedExternalName {
		t.Errorf("matched on %v, want externalName recorded as the key used", facts[0].MatchedOn)
	}
}

func TestClassify_ASupersededNameStillMatchesAsAnAlias_FR034(t *testing.T) {
	t.Parallel()
	f, src := keysFixture(t)

	rows := importRows(map[string]string{"r1": "5"})
	rows[0].Keys = performance.RowKeys{ExternalName: "Spring fair 2025"}
	run, err := f.service.RunImport(t.Context(), alex, performance.ImportInput{
		SourceID: src.ID, FileName: "names.csv", Origin: performance.OriginUpload,
		Headings: []string{"ref", "date", "inqs"}, Rows: rows,
	})
	if err != nil || run.Counts.Matched != 1 {
		t.Fatalf("counts = %+v (%v), want the old name still matching through its alias", run.Counts, err)
	}
}

func TestCheckAliasFree_ANameHeldElsewhereIsRefusedWithTheReason_FR035(t *testing.T) {
	t.Parallel()
	f, _ := keysFixture(t)
	if err := f.store.PutAlias(t.Context(), performance.Alias{
		ID: "ENA-1", ActivityID: "A-2", Name: "Autumn push", Meta: record.NewMeta(alex, noon)}); err != nil {
		t.Fatalf("PutAlias: %v", err)
	}

	err := f.service.CheckAliasFree(t.Context(), "A-1", "Autumn push")

	var refusal *errs.Error
	if !errors.As(err, &refusal) || refusal.Code != errs.CodeNameTaken {
		t.Fatalf("CheckAliasFree = %v, want the name refused with its holder named", err)
	}
}

func TestDeleteAlias_RemovesItAndWritesOneLogEntry_FR034(t *testing.T) {
	t.Parallel()
	f, _ := keysFixture(t)
	if err := f.store.PutAlias(t.Context(), performance.Alias{
		ID: "ENA-1", ActivityID: "A-1", Name: "Spring fair 2025", Meta: record.NewMeta(alex, noon)}); err != nil {
		t.Fatalf("PutAlias: %v", err)
	}

	if err := f.service.DeleteAlias(t.Context(), alex, "A-1", "ENA-1"); err != nil {
		t.Fatalf("DeleteAlias: %v", err)
	}
	aliases, err := f.service.Aliases(t.Context(), alex, "A-1")
	if err != nil || len(aliases) != 0 {
		t.Errorf("aliases = %v (%v), want the alias gone", aliases, err)
	}
	entries := f.log(t)
	if len(entries) == 0 || entries[0].Action != "deleted" || entries[0].RecordID != "ENA-1" {
		t.Errorf("latest log entry = %+v, want the alias deletion audited", entries[0])
	}
}

func TestDeleteAlias_NeedsPerformanceLink_FR043(t *testing.T) {
	t.Parallel()
	f, _ := keysFixture(t)

	ctx := authz.With(t.Context(), grantsWithout(authz.PlansView))
	err := f.service.DeleteAlias(ctx, alex, "A-1", "ENA-1")

	var refusal *errs.Error
	if !errors.As(err, &refusal) || refusal.Code != errs.CodeForbidden {
		t.Fatalf("DeleteAlias without performance.link = %v, want forbidden", err)
	}
}
