package performancetest

import (
	"context"
	"slices"
	"strconv"
	"strings"

	"mapp/internal/performance"
	"mapp/internal/platform/authz"
)

func (s *Store) ActivityKeyFacts(ctx context.Context, scope authz.Scope, activityID string) (performance.KeyFacts, bool, error) {
	if s.ActivityLookup != nil {
		return s.ActivityLookup(ctx, scope, activityID)
	}
	facts, found := s.Activities[activityID]
	return facts, found, nil
}

func (s *Store) FactTotalsBySource(_ context.Context, activityID string) ([]performance.SourceTotals, error) {
	all, err := s.facts.List()
	if err != nil {
		return nil, err
	}
	bySource := map[string]*performance.SourceTotals{}
	var order []string
	for _, f := range all {
		if f.ActivityID != activityID || f.ReplacedByID != "" {
			continue
		}
		totals, ok := bySource[f.SourceID]
		if !ok {
			src, _, _ := s.sources.Get(f.SourceID)
			totals = &performance.SourceTotals{SourceID: f.SourceID, SourceName: src.Name}
			bySource[f.SourceID] = totals
			order = append(order, f.SourceID)
		}
		totals.FactCount++
		if day := f.FactDate.Format("2006-01-02"); day > totals.LastFactDate {
			totals.LastFactDate = day
		}
	}
	merged := make([]performance.SourceTotals, 0, len(order))
	for _, id := range order {
		merged = append(merged, *bySource[id])
	}
	return merged, nil
}

func (s *Store) ListAliases(_ context.Context, activityID string) ([]performance.Alias, error) {
	found := []performance.Alias{}
	for _, alias := range s.aliases {
		if alias.ActivityID == activityID {
			found = append(found, alias)
		}
	}
	slices.SortFunc(found, func(a, b performance.Alias) int { return strings.Compare(a.Name, b.Name) })
	return found, nil
}

func (s *Store) ListIntegrationIDs(_ context.Context, activityID string) ([]performance.IntegrationID, error) {
	found := []performance.IntegrationID{}
	for _, key := range s.integrations {
		if key.ActivityID == activityID {
			found = append(found, key)
		}
	}
	slices.SortFunc(found, func(a, b performance.IntegrationID) int { return strings.Compare(a.SourceID, b.SourceID) })
	return found, nil
}

func (s *Store) RemoveAlias(_ context.Context, activityID, aliasID string) (bool, error) {
	alias, found := s.aliases[aliasID]
	if !found || alias.ActivityID != activityID {
		return false, nil
	}
	delete(s.aliases, aliasID)
	return true, nil
}

func (s *Store) RemoveIntegrationID(_ context.Context, activityID, sourceID string) (bool, error) {
	for id, key := range s.integrations {
		if key.ActivityID == activityID && key.SourceID == sourceID {
			delete(s.integrations, id)
			return true, nil
		}
	}
	return false, nil
}

func (s *Store) HeldActivityIDs(_ context.Context, sourceID string) ([]string, error) {
	all, err := s.facts.List()
	if err != nil {
		return nil, err
	}
	seen := map[string]bool{}
	held := []string{}
	for _, f := range all {
		if f.SourceID == sourceID && f.Held && f.ReplacedByID == "" && f.ActivityID != "" && !seen[f.ActivityID] {
			seen[f.ActivityID] = true
			held = append(held, f.ActivityID)
		}
	}
	slices.Sort(held)
	return held, nil
}

func (s *Store) ActivityTotals(_ context.Context, _ authz.Scope, activityID string, year int) ([]performance.StageTotals, error) {
	all, err := s.facts.List()
	if err != nil {
		return nil, err
	}
	byGroup := map[string]*performance.StageTotals{}
	var order []string
	for _, f := range all {
		if f.ActivityID != activityID || f.Held || f.ReplacedByID != "" || f.FactDate.Year() != year {
			continue
		}
		key, group := totalsGroupOf(f)
		totals, found := byGroup[key]
		if !found {
			totals = &group
			byGroup[key] = totals
			order = append(order, key)
		}
		addFact(totals, f)
	}
	slices.Sort(order)
	merged := make([]performance.StageTotals, 0, len(order))
	for _, key := range order {
		merged = append(merged, *byGroup[key])
	}
	return merged, nil
}

func (s *Store) LatestImportStatuses(_ context.Context, sourceIDs []string) (map[string]string, error) {
	all, err := s.imports.List()
	if err != nil {
		return nil, err
	}
	type latest struct {
		at     string
		status string
	}
	newest := map[string]latest{}
	for _, run := range all {
		if !slices.Contains(sourceIDs, run.SourceID) {
			continue
		}
		key := run.CreatedAt.Format("2006-01-02T15:04:05") + run.ID
		if held, ok := newest[run.SourceID]; !ok || key > held.at {
			newest[run.SourceID] = latest{at: key, status: string(run.Status)}
		}
	}
	statuses := map[string]string{}
	for id, held := range newest {
		statuses[id] = held.status
	}
	return statuses, nil
}

func totalsGroupOf(f performance.Fact) (string, performance.StageTotals) {
	currency := ""
	if f.Amount != nil {
		currency = f.Amount.Currency
	}
	month := f.FactDate.Format("2006-01")
	key := f.Stage + "|" + string(f.WebMetric) + "|" + month + "|" + currency
	return key, performance.StageTotals{Stage: f.Stage, WebMetric: string(f.WebMetric), Month: month, Currency: currency}
}

func addFact(totals *performance.StageTotals, f performance.Fact) {
	if f.Count != nil {
		totals.Counts += *f.Count
	}
	if f.Amount != nil {
		totals.AmountMinor += f.Amount.Amount.Cents()
	}
	totals.FactCount++
}

func (s *Store) GetConfiguration(_ context.Context) (performance.Configuration, bool, error) {
	if s.configuration == nil {
		return performance.Configuration{}, false, nil
	}
	return *s.configuration, true, nil
}

func (s *Store) PutConfiguration(_ context.Context, c performance.Configuration) error {
	s.configuration = &c
	return nil
}

func (s *Store) MediaCodeUsed(_ context.Context, code int64) (bool, error) {
	all, err := s.facts.List()
	if err != nil {
		return false, err
	}
	for _, f := range all {
		if f.MatchedOn == performance.MatchedMediaCode && f.MatchedKey == strconv.FormatInt(code, 10) {
			return true, nil
		}
	}
	return false, nil
}
