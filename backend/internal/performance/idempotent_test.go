package performance_test

import (
	"context"
	"testing"

	"mapp/internal/performance"
)

func importRows(values map[string]string) []performance.Row {
	rows := make([]performance.Row, 0, len(values))
	for reference, count := range values {
		rows = append(rows, performance.Row{
			Raw:       map[string]string{"ref": reference, "inqs": count},
			Reference: performance.Cell{Column: "ref", Value: reference},
			Date:      performance.Cell{Column: "date", Value: "2026-02-01"},
			Keys:      performance.RowKeys{ActivityID: "A-1"},
			Readings: []performance.Reading{{
				Measure: performance.Measure{Stage: "inq"},
				Cell:    performance.Cell{Column: "inqs", Value: count},
			}},
		})
	}
	return rows
}

func (f *fixture) runImport(t *testing.T, sourceID string, values map[string]string) performance.Import {
	t.Helper()
	run, err := f.service.RunImport(t.Context(), alex, performance.ImportInput{
		SourceID: sourceID, FileName: "load.csv", Origin: performance.OriginUpload,
		Headings: []string{"ref", "date", "inqs"},
		Rows:     importRows(values),
	})
	if err != nil {
		t.Fatalf("RunImport: %v", err)
	}
	return run
}

func (f *fixture) currentFacts(t *testing.T, sourceID string, references ...string) []performance.Fact {
	t.Helper()
	facts, err := f.store.CurrentFacts(t.Context(), sourceID, references)
	if err != nil {
		t.Fatalf("CurrentFacts: %v", err)
	}
	return facts
}

func newImportFixture(t *testing.T) (*fixture, performance.Source) {
	t.Helper()
	f := newFixture(t)
	f.store.Keys = map[performance.MatchedOn]map[string][]string{
		performance.MatchedActivityID: {"a-1": {"A-1"}},
	}
	f.service.UseApprovals(approvedStub{"A-1": true})
	return f, f.createSource(t, "Eloqua")
}

type approvedStub map[string]bool

func (a approvedStub) ApprovedIDs(_ context.Context, ids []string) (map[string]bool, error) {
	approved := map[string]bool{}
	for _, id := range ids {
		approved[id] = a[id]
	}
	return approved, nil
}

func TestRunImport_AnIdenticalFileChangesNothing_FR013(t *testing.T) {
	t.Parallel()
	f, src := newImportFixture(t)

	first := f.runImport(t, src.ID, map[string]string{"r1": "5", "r2": "3"})
	again := f.runImport(t, src.ID, map[string]string{"r1": "5", "r2": "3"})

	if first.Counts.Matched != 2 || again.Counts.Matched != 2 {
		t.Fatalf("counts = %+v then %+v, want 2 matched in both summaries", first.Counts, again.Counts)
	}
	facts := f.currentFacts(t, src.ID, "r1", "r2")
	if len(facts) != 2 {
		t.Fatalf("after the identical re-upload %d facts are current, want 2: nothing changes (SC-004)", len(facts))
	}
	for _, fact := range facts {
		if fact.ImportID != first.ID {
			t.Errorf("fact %s belongs to import %s, want %s: the re-upload must not restate it",
				fact.ID, fact.ImportID, first.ID)
		}
	}
}

func TestRunImport_ARepeatedReferenceReplacesAndAuditsBothVersions_FR013(t *testing.T) {
	t.Parallel()
	f, src := newImportFixture(t)

	f.runImport(t, src.ID, map[string]string{"r1": "5"})
	second := f.runImport(t, src.ID, map[string]string{"r1": "7"})

	current := f.currentFacts(t, src.ID, "r1")
	if len(current) != 1 || current[0].Count == nil || *current[0].Count != 7 {
		t.Fatalf("current facts = %+v, want one fact counting 7", current)
	}
	replaced := f.factsReplacedByImport(t, second.ID)
	if len(replaced) != 1 || replaced[0].Count == nil || *replaced[0].Count != 5 {
		t.Errorf("the replaced versions = %+v, want the first, counting 5, kept and audited (046-C4)", replaced)
	}
}

func (f *fixture) factsReplacedByImport(t *testing.T, importID string) []performance.Fact {
	t.Helper()
	facts, err := f.store.ListFactsOfImport(t.Context(), importID)
	if err != nil {
		t.Fatalf("ListFactsOfImport: %v", err)
	}
	replaced, err := f.store.FactsReplacedBy(t.Context(), factIDsOf(facts))
	if err != nil {
		t.Fatalf("FactsReplacedBy: %v", err)
	}
	return replaced
}

func factIDsOf(facts []performance.Fact) []string {
	ids := make([]string, 0, len(facts))
	for _, fact := range facts {
		ids = append(ids, fact.ID)
	}
	return ids
}

func TestRunImport_PersonalDataRefusesTheWholeFile_FR012(t *testing.T) {
	t.Parallel()
	f, src := newImportFixture(t)

	_, err := f.service.RunImport(t.Context(), alex, performance.ImportInput{
		SourceID: src.ID, FileName: "load.csv", Origin: performance.OriginUpload,
		Headings: []string{"ref", "date", "inqs", "Email"},
		Rows:     importRows(map[string]string{"r1": "5"}),
	})

	if err == nil {
		t.Fatal("RunImport accepted a file carrying a personal-data column, want a whole-file refusal")
	}
	if facts := f.currentFacts(t, src.ID, "r1"); len(facts) != 0 {
		t.Errorf("%d facts stored from a refused file, want none (SC-007)", len(facts))
	}
	imports, err := f.store.ListImports(t.Context(), performance.ImportFilter{SourceID: src.ID, Limit: 10})
	if err != nil || len(imports) != 0 {
		t.Errorf("a refused file left %d imports listed (%v), want none: it never started", len(imports), err)
	}
}

func TestRunImport_SummaryAndBreakdownAreStored_FR022_046C18(t *testing.T) {
	t.Parallel()
	f, src := newImportFixture(t)

	run := f.runImport(t, src.ID, map[string]string{"r1": "5", "r2": "oops"})

	if !run.Counts.Reconciles() || run.Counts.Matched != 1 || run.Counts.Rejected != 1 {
		t.Fatalf("counts = %+v, want 1 matched and 1 rejected, reconciling", run.Counts)
	}
	groups := f.store.BreakdownOf(run.ID)
	if len(groups) != 1 || groups[0].Stage != "inq" || !groups[0].Counts.Reconciles() {
		t.Errorf("breakdown = %+v, want the inq group carrying both rows, reconciling (046-C18)", groups)
	}
	rows, err := f.store.ListReviewRows(t.Context(), performance.ReviewFilter{Status: "open", SourceID: src.ID, Limit: 10})
	if err != nil || len(rows) != 1 {
		t.Errorf("review rows = %d (%v), want the rejected row waiting for review (FR-023)", len(rows), err)
	}
}

func TestRunImport_AnIdenticalFileAddsNoSecondReviewRow_FR013(t *testing.T) {
	t.Parallel()
	f, src := newImportFixture(t)

	f.runImport(t, src.ID, map[string]string{"r1": "oops"})
	f.runImport(t, src.ID, map[string]string{"r1": "oops"})

	rows, err := f.store.ListReviewRows(t.Context(), performance.ReviewFilter{Status: "open", SourceID: src.ID, Limit: 10})
	if err != nil || len(rows) != 1 {
		t.Fatalf("open review rows = %d (%v), want 1: the identical row is already waiting (SC-004)", len(rows), err)
	}
}
