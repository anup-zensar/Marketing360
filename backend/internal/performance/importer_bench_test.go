package performance_test

import (
	"context"
	"strconv"
	"testing"

	"mapp/internal/performance"
	"mapp/internal/platform/authz"
)

func bigFile(rows, activities int) ([]performance.Row, performance.Env) {
	resolved := performance.Resolved{performance.MatchedActivityID: map[string][]string{}}
	for i := 0; i < activities; i++ {
		id := "A-" + strconv.Itoa(i)
		resolved[performance.MatchedActivityID]["a-"+strconv.Itoa(i)] = []string{id}
	}
	file := make([]performance.Row, 0, rows)
	for i := 0; i < rows; i++ {
		file = append(file, performance.Row{
			Reference: performance.Cell{Column: "ref", Value: "r" + strconv.Itoa(i)},
			Date:      performance.Cell{Column: "date", Value: "2026-02-01"},
			Keys:      performance.RowKeys{ActivityID: "A-" + strconv.Itoa(i%activities)},
			Readings: []performance.Reading{{
				Measure: performance.Measure{Stage: "inq"},
				Cell:    performance.Cell{Column: "inqs", Value: strconv.Itoa(i % 97)},
			}},
		})
	}
	return file, performance.Env{
		Today:      today,
		Currencies: knownCcy,
		Resolved:   resolved,
		Approved:   func(string) bool { return true },
	}
}

// BenchmarkClassify_SC009 is plan.md's target: a 50,000-row file against 5,000 in-scope activities,
// classified and summarised in well under the 10-second budget.
func BenchmarkClassify_SC009(b *testing.B) {
	rows, env := bigFile(50_000, 5_000)
	for b.Loop() {
		results := performance.Classify(rows, env)
		summary := performance.Summarise(results)
		if !summary.Counts.Reconciles() || summary.Counts.Read != 50_000 {
			b.Fatalf("summary = %+v", summary.Counts)
		}
	}
}

func TestRunImport_MatchQueriesStayConstantWhateverTheFileSize_SC009(t *testing.T) {
	t.Parallel()

	calls := map[int]int{}
	for _, size := range []int{10, 1000} {
		service, counter, sourceID := countedService(t)
		rows, _ := bigFile(size, 5)
		for i := range rows {
			rows[i].Keys.ActivityID = "A-1"
		}
		if _, err := service.RunImport(t.Context(), alex, performance.ImportInput{
			SourceID: sourceID, FileName: "big.csv", Origin: performance.OriginUpload,
			Headings: []string{"ref", "date", "inqs"}, Rows: rows,
		}); err != nil {
			t.Fatalf("RunImport(%d rows): %v", size, err)
		}
		calls[size] = counter.matchCalls
	}

	if calls[10] != calls[1000] || calls[10] == 0 {
		t.Fatalf("match queries: %d for 10 rows, %d for 1000 - SC-009 requires a constant count", calls[10], calls[1000])
	}
}

// countedService builds the fixture's service around a store that counts its matching queries.
func countedService(t *testing.T) (*performance.Service, *matchCounter, string) {
	t.Helper()
	f := newFixture(t)
	f.store.Keys = map[performance.MatchedOn]map[string][]string{
		performance.MatchedActivityID: {"a-1": {"A-1"}},
	}
	counter := &matchCounter{Store: f.store}
	service := performance.NewService(f.tools, counter)
	service.UseApprovals(approvedStub{"A-1": true})
	src, err := service.CreateSource(t.Context(), alex, performance.SourceInput{
		Name: "Big", OwnerID: alex.ID, Category: performance.CategoryCRM,
		DeliveryRoute: performance.RouteUpload, Cadence: performance.CadenceMonthly,
	})
	if err != nil {
		t.Fatalf("CreateSource: %v", err)
	}
	return service, counter, src.ID
}

type matchCounter struct {
	performance.Store
	matchCalls int
}

func (m *matchCounter) MatchByActivityID(ctx context.Context, scope authz.Scope, keys []string) (map[string][]string, error) {
	m.matchCalls++
	return m.Store.MatchByActivityID(ctx, scope, keys)
}
