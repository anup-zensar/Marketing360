// Package performancetest is the in-memory fake of the performance store, for tests that build services or
// the whole API without a database.
package performancetest

import (
	"context"
	"slices"
	"strings"
	"time"

	"mapp/internal/performance"
	"mapp/internal/platform/actor"
	"mapp/internal/platform/authz"
	"mapp/internal/platform/fake"
)

var _ performance.Store = (*Store)(nil)

type Store struct {
	sources  *fake.Table[performance.Source]
	mappings *fake.Table[performance.Mapping]
	facts    *fake.Table[performance.Fact]
	imports  *fake.Table[performance.Import]
	reviews  *fake.Table[performance.ReviewRow]
	codes    *fake.Table[performance.MediaCode]
	// Keys stands in for the matching tables: a key, lower-cased, to the activities it reaches.
	Keys map[performance.MatchedOn]map[string][]string
	// NextCode is the media-code counter the configuration holds in PostgreSQL.
	NextCode int64
	// Currencies stands in for the currencies table; tests set what the organisation has configured.
	Currencies map[string]bool
	// Activities stands in for the activities table: key facts by activity id, visible to every scope.
	Activities map[string]performance.KeyFacts
	// ActivityLookup, when set, answers ActivityKeyFacts from the shared fakes instead (storestest wires it).
	ActivityLookup func(ctx context.Context, scope authz.Scope, activityID string) (performance.KeyFacts, bool, error)
	// ResolveActivityIDs, when set, matches activity-id keys against the shared fakes (storestest wires it).
	ResolveActivityIDs func(ctx context.Context, scope authz.Scope, keys []string) (map[string][]string, error)
	breakdowns         map[string][]performance.Breakdown
	aliases            map[string]performance.Alias
	integrations       map[string]performance.IntegrationID
	configuration      *performance.Configuration
}

func NewStore(db *fake.DB) *Store {
	return &Store{
		sources:      fake.TableOf(db, "performance_sources", func(s performance.Source) string { return s.ID }),
		mappings:     fake.TableOf(db, "performance_source_mappings", func(m performance.Mapping) string { return m.ID }),
		facts:        fake.TableOf(db, "performance_facts", func(f performance.Fact) string { return f.ID }),
		imports:      fake.TableOf(db, "performance_imports", func(i performance.Import) string { return i.ID }),
		reviews:      fake.TableOf(db, "performance_review_rows", func(r performance.ReviewRow) string { return r.ID }),
		codes:        fake.TableOf(db, "media_codes", func(c performance.MediaCode) string { return c.ID }),
		Keys:         map[performance.MatchedOn]map[string][]string{},
		NextCode:     9000001,
		aliases:      map[string]performance.Alias{},
		integrations: map[string]performance.IntegrationID{},
	}
}

func (s *Store) ListSources(_ context.Context, id string) ([]performance.SourceView, error) {
	all, err := s.sources.List()
	if err != nil {
		return nil, err
	}
	views := make([]performance.SourceView, 0, len(all))
	for _, src := range all {
		if id != "" && src.ID != id {
			continue
		}
		views = append(views, performance.SourceView{Source: src, FactCount: s.factsOf(src.ID)})
	}
	slices.SortFunc(views, func(a, b performance.SourceView) int {
		return strings.Compare(strings.ToLower(a.Name), strings.ToLower(b.Name))
	})
	return views, nil
}

func (s *Store) factsOf(sourceID string) int {
	all, err := s.facts.List()
	if err != nil {
		return 0
	}
	count := 0
	for _, f := range all {
		if f.SourceID == sourceID {
			count++
		}
	}
	return count
}

func (s *Store) PutSource(ctx context.Context, src performance.Source) error {
	return s.sources.Put(ctx, src)
}

func (s *Store) DeleteSource(ctx context.Context, id string) (bool, error) {
	if s.factsOf(id) > 0 {
		return false, nil
	}
	if _, found, err := s.sources.Get(id); err != nil || !found {
		return false, err
	}
	return true, s.sources.Delete(ctx, id)
}

func (s *Store) ListMappings(_ context.Context, sourceIDs []string) ([]performance.Mapping, error) {
	all, err := s.mappings.List()
	if err != nil {
		return nil, err
	}
	found := make([]performance.Mapping, 0, len(all))
	for _, m := range all {
		if len(sourceIDs) == 0 || slices.Contains(sourceIDs, m.SourceID) {
			found = append(found, m)
		}
	}
	return found, nil
}

func (s *Store) PutMapping(ctx context.Context, m performance.Mapping) error {
	return s.mappings.Put(ctx, m)
}

func (s *Store) DeleteMapping(ctx context.Context, id string) error {
	return s.mappings.Delete(ctx, id)
}

func (s *Store) ActiveSuppliers(ctx context.Context, modelID string) ([]performance.Supplier, error) {
	mappings, err := s.ListMappings(ctx, nil)
	if err != nil {
		return nil, err
	}
	suppliers := []performance.Supplier{}
	for _, m := range mappings {
		if m.ModelID != modelID || !m.SuppliedBy() {
			continue
		}
		src, found, srcErr := s.sources.Get(m.SourceID)
		if srcErr != nil || !found || src.Status != performance.StatusActive {
			continue
		}
		suppliers = append(suppliers, performance.Supplier{
			MappingID: m.ID, SourceID: src.ID, SourceName: src.Name, Meaning: m.Meaning, Stage: m.Stage,
		})
	}
	return suppliers, nil
}

func (s *Store) match(on performance.MatchedOn, keys []string) map[string][]string {
	found := map[string][]string{}
	for _, k := range keys {
		key := strings.ToLower(strings.TrimSpace(k))
		if hit, ok := s.Keys[on][key]; ok {
			found[key] = hit
		}
	}
	return found
}

func (s *Store) MatchByActivityID(ctx context.Context, scope authz.Scope, keys []string) (map[string][]string, error) {
	if s.ResolveActivityIDs != nil {
		return s.ResolveActivityIDs(ctx, scope, keys)
	}
	return s.match(performance.MatchedActivityID, keys), nil
}

func (s *Store) MatchByMediaCode(_ context.Context, _ authz.Scope, codes []int64) (map[string][]string, error) {
	keys := make([]string, 0, len(codes))
	for _, c := range codes {
		keys = append(keys, strings.TrimSpace(strings.ToLower(itoa(c))))
	}
	return s.match(performance.MatchedMediaCode, keys), nil
}

func (s *Store) MatchByIntegrationID(_ context.Context, _ authz.Scope, _ string, keys []string) (map[string][]string, error) {
	return s.match(performance.MatchedIntegrationID, keys), nil
}

func (s *Store) MatchByExternalName(_ context.Context, _ authz.Scope, keys []string) (map[string][]string, error) {
	return s.match(performance.MatchedExternalName, keys), nil
}

func (s *Store) ListFactsForActivity(_ context.Context, _ authz.Scope, q performance.FactQuery) ([]performance.FactView, error) {
	all, err := s.facts.List()
	if err != nil {
		return nil, err
	}
	views := []performance.FactView{}
	for _, f := range all {
		view, kept := s.factView(f, q)
		if kept {
			views = append(views, view)
		}
	}
	if q.Limit > 0 && len(views) > q.Limit {
		views = views[:q.Limit]
	}
	return views, nil
}

func (s *Store) factView(f performance.Fact, q performance.FactQuery) (performance.FactView, bool) {
	if f.ActivityID != q.ActivityID || f.ReplacedByID != "" || (q.CountedOnly && f.Held) {
		return performance.FactView{}, false
	}
	src, _, _ := s.sources.Get(f.SourceID)
	if q.Search != "" && !factMatches(f, src.Name, q.Search) {
		return performance.FactView{}, false
	}
	return performance.FactView{Fact: f, SourceName: src.Name}, true
}

func factMatches(f performance.Fact, sourceName, search string) bool {
	search = strings.ToLower(search)
	for _, hay := range []string{sourceName, f.Stage, string(f.WebMetric), string(f.MatchedOn), f.ImportID, f.SourceReference} {
		if strings.Contains(strings.ToLower(hay), search) {
			return true
		}
	}
	return false
}

func (s *Store) PutFact(ctx context.Context, f performance.Fact) error {
	return s.facts.Put(ctx, f)
}

func (s *Store) CurrentFacts(_ context.Context, sourceID string, references []string) ([]performance.Fact, error) {
	all, err := s.facts.List()
	if err != nil {
		return nil, err
	}
	found := []performance.Fact{}
	for _, f := range all {
		if f.SourceID == sourceID && slices.Contains(references, f.SourceReference) && f.ReplacedByID == "" {
			found = append(found, f)
		}
	}
	return found, nil
}

func (s *Store) ReleaseHeld(ctx context.Context, activityID string, by actor.Actor, at time.Time) (int, error) {
	all, err := s.facts.List()
	if err != nil {
		return 0, err
	}
	released := 0
	for _, f := range all {
		if f.ActivityID != activityID || !f.Held {
			continue
		}
		f.Held = false
		f.Version++
		f.UpdatedAt, f.UpdatedBy = at, by
		if err := s.facts.Put(ctx, f); err != nil {
			return released, err
		}
		released++
	}
	return released, nil
}

func (s *Store) ListImports(_ context.Context, f performance.ImportFilter) ([]performance.Import, error) {
	all, err := s.imports.List()
	if err != nil {
		return nil, err
	}
	found := []performance.Import{}
	for _, i := range all {
		if importMatches(i, f) {
			found = append(found, i)
		}
	}
	if f.Limit > 0 && len(found) > f.Limit {
		found = found[:f.Limit]
	}
	return found, nil
}

func importMatches(i performance.Import, f performance.ImportFilter) bool {
	identified := (f.ID == "" || i.ID == f.ID) && (f.SourceID == "" || i.SourceID == f.SourceID) &&
		(f.Status == "" || string(i.Status) == f.Status)
	return identified && importDated(i, f) && importSearched(i, f.Search)
}

func importDated(i performance.Import, f performance.ImportFilter) bool {
	if f.Since != nil && i.CreatedAt.Before(*f.Since) {
		return false
	}
	return f.Until == nil || !i.CreatedAt.After(*f.Until)
}

func importSearched(i performance.Import, search string) bool {
	if search == "" {
		return true
	}
	search = strings.ToLower(search)
	return strings.Contains(strings.ToLower(i.FileName), search) ||
		strings.Contains(strings.ToLower(i.CreatedBy.Name), search)
}

func (s *Store) PutImport(ctx context.Context, i performance.Import) error {
	return s.imports.Put(ctx, i)
}
