package performance_test

import (
	"testing"

	"mapp/internal/performance"
)

func factsFixture(t *testing.T) *fixture {
	t.Helper()
	f, src := keysFixture(t)
	f.store.Keys[performance.MatchedActivityID] = map[string][]string{"a-1": {"A-1"}}
	f.runImport(t, src.ID, map[string]string{"r1": "5", "r2": "3"})
	return f
}

func TestActivityFacts_CarryEveryColumnFR030Names(t *testing.T) {
	t.Parallel()
	f := factsFixture(t)

	facts, err := f.service.ActivityFacts(t.Context(), alex, performance.FactQuery{ActivityID: "A-1"})
	if err != nil || len(facts) != 2 {
		t.Fatalf("ActivityFacts = %d (%v), want both facts", len(facts), err)
	}
	fact := facts[0]
	if fact.FactDate.IsZero() || fact.SourceName != "Eloqua" || fact.Stage != "inq" ||
		fact.MatchedOn != performance.MatchedActivityID || fact.ImportID == "" {
		t.Errorf("fact = %+v, want date, source, stage, the key matched on and the originating upload", fact)
	}
}

func TestActivityFacts_SearchNarrowsByKeyAndUpload_FR030(t *testing.T) {
	t.Parallel()
	f := factsFixture(t)

	byReference, err := f.service.ActivityFacts(t.Context(), alex,
		performance.FactQuery{ActivityID: "A-1", Search: "r1"})
	if err != nil || len(byReference) != 1 {
		t.Fatalf("search r1 = %d (%v), want the one fact whose reference carries it", len(byReference), err)
	}
	bySource, err := f.service.ActivityFacts(t.Context(), alex,
		performance.FactQuery{ActivityID: "A-1", Search: "eloqua"})
	if err != nil || len(bySource) != 2 {
		t.Fatalf("search eloqua = %d (%v), want both facts of the source", len(bySource), err)
	}
}

func TestActivityFacts_AnActivityOutsideTheScopeAnswersAsMissing_FR044(t *testing.T) {
	t.Parallel()
	f := factsFixture(t)

	if _, err := f.service.ActivityFacts(t.Context(), alex, performance.FactQuery{ActivityID: "A-9"}); err == nil {
		t.Fatal("ActivityFacts answered for an unknown activity, want it missing")
	}
}
