package performance_test

import (
	"context"
	"errors"
	"strconv"
	"testing"

	"mapp/internal/performance"
	"mapp/internal/performance/performancetest"
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

// failingFacts makes the first fact write fall over, so the load's failure path is reachable on the fake.
type failingFacts struct {
	performance.Store
	failed bool
}

func (f *failingFacts) PutFact(ctx context.Context, fact performance.Fact) error {
	if !f.failed {
		f.failed = true
		return errOutage
	}
	return f.Store.PutFact(ctx, fact)
}

var errOutage = errors.New("the database went away")

func TestRunImport_ALoadThatFellOverIsRecordedAsFailed_046C17(t *testing.T) {
	t.Parallel()
	broken := failingServiceOver(t)

	rows, _ := bigFile(3, 5)
	for i := range rows {
		rows[i].Keys.ActivityID = "A-1"
	}
	_, err := broken.service.RunImport(t.Context(), alex, performance.ImportInput{
		SourceID: broken.sourceID, FileName: "broken.csv", Origin: performance.OriginUpload,
		Headings: []string{"ref", "date", "inqs"}, Rows: rows,
	})
	if err == nil {
		t.Fatal("RunImport survived a dead fact store, want the error surfaced")
	}

	imports, listErr := broken.store.ListImports(t.Context(), performance.ImportFilter{SourceID: broken.sourceID, Limit: 10})
	if listErr != nil || len(imports) != 1 {
		t.Fatalf("imports = %d (%v), want the one failed load recorded (046-C17)", len(imports), listErr)
	}
	if imports[0].Status != performance.ImportFailed || imports[0].Counts.Read != 3 {
		t.Errorf("recorded = %+v, want status failed carrying the 3 rows it had read", imports[0])
	}
}

type brokenFixture struct {
	service  *performance.Service
	store    *performancetest.Store
	sourceID string
}

func failingServiceOver(t *testing.T) brokenFixture {
	t.Helper()
	f := newFixture(t)
	f.store.Keys = map[performance.MatchedOn]map[string][]string{
		performance.MatchedActivityID: {"a-1": {"A-1"}},
	}
	wrapped := &failingFacts{Store: f.store}
	service := performance.NewService(f.tools, wrapped)
	service.UseApprovals(approvedStub{"A-1": true})
	src, err := service.CreateSource(t.Context(), alex, performance.SourceInput{
		Name: "Fragile", OwnerID: alex.ID, Category: performance.CategoryCRM,
		DeliveryRoute: performance.RouteUpload, Cadence: performance.CadenceMonthly,
	})
	if err != nil {
		t.Fatalf("CreateSource: %v", err)
	}
	return brokenFixture{service: service, store: f.store, sourceID: src.ID}
}
