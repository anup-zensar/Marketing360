package performance_test

import (
	"testing"
	"time"

	"mapp/internal/performance"
)

func twoNamedLoads(t *testing.T) *fixture {
	t.Helper()
	f, src := newImportFixture(t)
	for _, name := range []string{"january.csv", "february.csv"} {
		if _, err := f.service.RunImport(t.Context(), alex, performance.ImportInput{
			SourceID: src.ID, FileName: name, Origin: performance.OriginUpload,
			Headings: []string{"ref", "date", "inqs"},
			Rows:     importRows(map[string]string{name: "5"}),
		}); err != nil {
			t.Fatalf("RunImport(%s): %v", name, err)
		}
	}
	return f
}

func TestImports_FindableByFileUploaderStatusAndDate_FR016(t *testing.T) {
	t.Parallel()
	f := twoNamedLoads(t)
	dayAfter := noon.AddDate(0, 0, 1)
	dayBefore := noon.AddDate(0, 0, -1)

	tests := []struct {
		name   string
		filter performance.ImportFilter
		want   int
	}{
		{name: "FR-016 by file name", filter: performance.ImportFilter{Search: "january"}, want: 1},
		{name: "FR-016 by uploader name", filter: performance.ImportFilter{Search: "morgan"}, want: 2},
		{name: "FR-016 by status, none failed", filter: performance.ImportFilter{Status: "failed"}, want: 0},
		{name: "FR-016 by status, both completed", filter: performance.ImportFilter{Status: "completed"}, want: 2},
		{name: "FR-016 by date, nothing after today", filter: performance.ImportFilter{Since: &dayAfter}, want: 0},
		{name: "FR-016 by date, nothing before today", filter: performance.ImportFilter{Until: &dayBefore}, want: 0},
		{name: "FR-016 a stranger's name finds nothing", filter: performance.ImportFilter{Search: "nobody"}, want: 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			found, err := f.service.Imports(t.Context(), alex, tt.filter)
			if err != nil || len(found) != tt.want {
				t.Errorf("Imports(%+v) = %d loads (%v), want %d", tt.filter, len(found), err, tt.want)
			}
		})
	}
}

func TestSources_LatestLoadThatFellOverShowsFailed_FR005(t *testing.T) {
	t.Parallel()
	broken := failingServiceOver(t)
	if _, err := broken.service.RunImport(t.Context(), alex, performance.ImportInput{
		SourceID: broken.sourceID, FileName: "web.csv", Origin: performance.OriginUpload,
		Headings: []string{"ref", "date", "inqs"},
		Rows:     importRows(map[string]string{"r1": "5"}),
	}); err == nil {
		t.Fatal("RunImport on a broken store succeeded, want the outage surfaced")
	}

	sources, err := broken.service.Sources(t.Context(), alex)
	if err != nil || len(sources) != 1 {
		t.Fatalf("Sources() = %d sources (%v), want the one that failed", len(sources), err)
	}
	if got := sources[0].FreshnessAt(noon.Add(time.Hour)); got != performance.FreshnessFailed {
		t.Errorf("freshness in Settings = %s, want %s: the failed load shows beyond the panel", got,
			performance.FreshnessFailed)
	}
}
