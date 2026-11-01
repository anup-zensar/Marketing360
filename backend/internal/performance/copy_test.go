package performance_test

import (
	"testing"

	"mapp/internal/performance"
	"mapp/internal/platform/record"
)

// FR-038's exclusion itself lives in the activity area's copy path, outside this feature's folders; this
// proves nothing in this package follows a copy on its own.
func TestKeys_NothingFollowsACopiedActivity_FR038(t *testing.T) {
	t.Parallel()
	f, src := keysFixture(t)
	ref := performance.KeyRef{ActivityID: "A-1", SourceID: src.ID}
	if _, err := f.service.SetIntegrationID(t.Context(), alex, ref, performance.IntegrationIDInput{Value: "elq-77"}); err != nil {
		t.Fatalf("SetIntegrationID: %v", err)
	}
	if err := f.store.PutAlias(t.Context(), performance.Alias{
		ID: "ENA-1", ActivityID: "A-1", Name: "Spring fair 2025", Meta: record.NewMeta(alex, noon)}); err != nil {
		t.Fatalf("PutAlias: %v", err)
	}
	f.store.Activities["A-2"] = performance.KeyFacts{ExternalName: ""}

	area, err := f.service.SourcesArea(t.Context(), alex, "A-2")
	if err != nil {
		t.Fatalf("SourcesArea: %v", err)
	}
	if len(area.Aliases) != 0 || area.Integrated || len(area.Sources) != 0 || area.MediaCodeCount != 0 {
		t.Errorf("the copy carries %+v, want no alias, key, code or fact of the original", area)
	}
}
