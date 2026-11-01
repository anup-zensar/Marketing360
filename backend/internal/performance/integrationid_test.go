package performance_test

import (
	"errors"
	"testing"

	"mapp/internal/performance"
	"mapp/internal/platform/errs"
)

func TestSetIntegrationID_AtMostOnePerSource_FR037(t *testing.T) {
	t.Parallel()
	f, src := keysFixture(t)
	ref := performance.KeyRef{ActivityID: "A-1", SourceID: src.ID}

	first, err := f.service.SetIntegrationID(t.Context(), alex, ref, performance.IntegrationIDInput{Value: "elq-77"})
	if err != nil || first.Value != "elq-77" {
		t.Fatalf("SetIntegrationID = %+v (%v), want the key registered", first, err)
	}

	_, err = f.service.SetIntegrationID(t.Context(), alex, ref, performance.IntegrationIDInput{Value: "elq-88"})
	var refusal *errs.Error
	if !errors.As(err, &refusal) || refusal.Code != errs.CodeNotAllowed {
		t.Fatalf("a second key for the source = %v, want a 409 naming the one that exists", err)
	}
}

func TestIntegrationID_EveryChangeIsAudited_FR037(t *testing.T) {
	t.Parallel()
	f, src := keysFixture(t)
	ref := performance.KeyRef{ActivityID: "A-1", SourceID: src.ID}

	if _, err := f.service.SetIntegrationID(t.Context(), alex, ref, performance.IntegrationIDInput{Value: "elq-77"}); err != nil {
		t.Fatalf("SetIntegrationID: %v", err)
	}
	if err := f.service.DeleteIntegrationID(t.Context(), alex, ref); err != nil {
		t.Fatalf("DeleteIntegrationID: %v", err)
	}

	actions := map[string]int{}
	for _, entry := range f.log(t) {
		if entry.RecordType == "integrationId" {
			actions[string(entry.Action)]++
		}
	}
	if actions["created"] != 1 || actions["deleted"] != 1 {
		t.Errorf("integration-id log actions = %v, want the set and the removal each audited", actions)
	}
}

func TestDeleteIntegrationID_RemovingRestoresNothingElse_FR037(t *testing.T) {
	t.Parallel()
	f, src := keysFixture(t)
	other := f.createSource(t, "Marketo")
	for _, source := range []string{src.ID, other.ID} {
		ref := performance.KeyRef{ActivityID: "A-1", SourceID: source}
		if _, err := f.service.SetIntegrationID(t.Context(), alex, ref, performance.IntegrationIDInput{Value: "k-" + source}); err != nil {
			t.Fatalf("SetIntegrationID(%s): %v", source, err)
		}
	}

	if err := f.service.DeleteIntegrationID(t.Context(), alex, performance.KeyRef{ActivityID: "A-1", SourceID: src.ID}); err != nil {
		t.Fatalf("DeleteIntegrationID: %v", err)
	}
	keys, err := f.service.IntegrationIDs(t.Context(), alex, "A-1")
	if err != nil || len(keys) != 1 || keys[0].SourceID != other.ID {
		t.Errorf("keys = %+v (%v), want only the other source's key kept", keys, err)
	}
}

func TestSourcesArea_ShowsKeysFlagsAndTotals_FR031(t *testing.T) {
	t.Parallel()
	f, src := keysFixture(t)
	f.store.Keys[performance.MatchedActivityID] = map[string][]string{"a-1": {"A-1"}}
	ref := performance.KeyRef{ActivityID: "A-1", SourceID: src.ID}
	if _, err := f.service.SetIntegrationID(t.Context(), alex, ref, performance.IntegrationIDInput{Value: "elq-77"}); err != nil {
		t.Fatalf("SetIntegrationID: %v", err)
	}
	f.runImport(t, src.ID, map[string]string{"r1": "5", "r2": "3"})

	area, err := f.service.SourcesArea(t.Context(), alex, "A-1")
	if err != nil {
		t.Fatalf("SourcesArea: %v", err)
	}
	if area.ExternalName != "Spring fair" || !area.GoalEntered || !area.Integrated {
		t.Errorf("area = %+v, want the external name and both flags set", area)
	}
	assertEloquaTotals(t, area.Sources)
}

func assertEloquaTotals(t *testing.T, sources []performance.SourceOfFacts) {
	t.Helper()
	if len(sources) != 1 || sources[0].FactCount != 2 || sources[0].IntegrationID != "elq-77" {
		t.Fatalf("sources = %+v, want Eloqua with two facts and its key", sources)
	}
	if sources[0].LastFactDate != "2026-02-01" {
		t.Errorf("last fact date = %s, want 2026-02-01", sources[0].LastFactDate)
	}
}

func TestSourcesArea_AnActivityOutsideTheScopeAnswersAsMissing_FR044(t *testing.T) {
	t.Parallel()
	f, _ := keysFixture(t)

	_, err := f.service.SourcesArea(t.Context(), alex, "A-9")

	var refusal *errs.Error
	if !errors.As(err, &refusal) || refusal.Code != errs.CodeNotFound {
		t.Fatalf("SourcesArea(unknown) = %v, want it answered as missing", err)
	}
}
