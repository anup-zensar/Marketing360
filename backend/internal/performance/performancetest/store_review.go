package performancetest

import (
	"context"
	"slices"
	"strings"
	"time"

	"mapp/internal/performance"
	"mapp/internal/platform/actor"
	"mapp/internal/platform/authz"
)

func (s *Store) ListReviewRows(_ context.Context, f performance.ReviewFilter) ([]performance.ReviewRowView, error) {
	all, err := s.reviews.List()
	if err != nil {
		return nil, err
	}
	views := []performance.ReviewRowView{}
	for _, r := range all {
		if !reviewMatches(r, f) {
			continue
		}
		src, _, _ := s.sources.Get(r.SourceID)
		views = append(views, performance.ReviewRowView{ReviewRow: r, SourceName: src.Name})
	}
	if f.Limit > 0 && len(views) > f.Limit {
		views = views[:f.Limit]
	}
	return views, nil
}

func reviewMatches(r performance.ReviewRow, f performance.ReviewFilter) bool {
	return reviewFieldsMatch(r, f) && reviewContentMatches(r, f) && reviewDatesMatch(r, f)
}

func reviewFieldsMatch(r performance.ReviewRow, f performance.ReviewFilter) bool {
	return reviewStatusMatches(r, f) && (f.SourceID == "" || r.SourceID == f.SourceID) &&
		(f.ID == "" || r.ID == f.ID) && (f.ImportID == "" || r.ImportID == f.ImportID)
}

func reviewStatusMatches(r performance.ReviewRow, f performance.ReviewFilter) bool {
	return (f.Status == "" || string(r.Status) == f.Status) &&
		(f.StatusNot == "" || string(r.Status) != f.StatusNot)
}

func reviewContentMatches(r performance.ReviewRow, f performance.ReviewFilter) bool {
	if f.Reason != "" && r.Reason != f.Reason {
		return false
	}
	return f.KeyText == "" || strings.Contains(strings.ToLower(string(r.Raw)), strings.ToLower(f.KeyText))
}

func reviewDatesMatch(r performance.ReviewRow, f performance.ReviewFilter) bool {
	if f.Since != nil && r.CreatedAt.Before(*f.Since) {
		return false
	}
	return f.Until == nil || !r.CreatedAt.After(*f.Until)
}

func (s *Store) PutReviewRow(ctx context.Context, r performance.ReviewRow) error {
	return s.reviews.Put(ctx, r)
}

func (s *Store) ListMediaCodes(_ context.Context, _ authz.Scope, activityID string) ([]performance.MediaCode, error) {
	all, err := s.codes.List()
	if err != nil {
		return nil, err
	}
	found := []performance.MediaCode{}
	for _, c := range all {
		if c.ActivityID == activityID {
			found = append(found, c)
		}
	}
	return found, nil
}

// TakeNextCode advances the configuration's counter, exactly as the PostgreSQL query does.
func (s *Store) TakeNextCode(_ context.Context, _ actor.Actor, _ time.Time) (int64, error) {
	if s.configuration != nil {
		code := s.configuration.MediaCodeNext
		s.configuration.MediaCodeNext++
		return code, nil
	}
	code := s.NextCode
	s.NextCode++
	return code, nil
}

func (s *Store) PutMediaCode(ctx context.Context, c performance.MediaCode) error {
	return s.codes.Put(ctx, c)
}

func (s *Store) PutIntegrationID(_ context.Context, k performance.IntegrationID) error {
	s.integrations[k.ID] = k
	key := strings.ToLower(strings.TrimSpace(k.Value))
	if s.Keys[performance.MatchedIntegrationID] == nil {
		s.Keys[performance.MatchedIntegrationID] = map[string][]string{}
	}
	s.Keys[performance.MatchedIntegrationID][key] = append(
		s.Keys[performance.MatchedIntegrationID][key], k.ActivityID)
	return nil
}

func (s *Store) idsOf(all []string) ([]string, error) { return all, nil }

func (s *Store) FactIDsOf(_ context.Context, activityIDs []string) ([]string, error) {
	all, err := s.facts.List()
	if err != nil {
		return nil, err
	}
	var ids []string
	for _, f := range all {
		if slices.Contains(activityIDs, f.ActivityID) {
			ids = append(ids, f.ID)
		}
	}
	return s.idsOf(ids)
}

func (s *Store) MediaCodeIDsOf(_ context.Context, activityIDs []string) ([]string, error) {
	all, err := s.codes.List()
	if err != nil {
		return nil, err
	}
	var ids []string
	for _, c := range all {
		if slices.Contains(activityIDs, c.ActivityID) {
			ids = append(ids, c.ID)
		}
	}
	return s.idsOf(ids)
}

func (s *Store) AliasIDsOf(_ context.Context, activityIDs []string) ([]string, error) {
	var ids []string
	for _, a := range s.aliases {
		if slices.Contains(activityIDs, a.ActivityID) {
			ids = append(ids, a.ID)
		}
	}
	return s.idsOf(ids)
}

func (s *Store) IntegrationIDsOf(_ context.Context, activityIDs []string) ([]string, error) {
	var ids []string
	for _, k := range s.integrations {
		if slices.Contains(activityIDs, k.ActivityID) {
			ids = append(ids, k.ID)
		}
	}
	return s.idsOf(ids)
}

func (s *Store) GetFact(_ context.Context, id string) (performance.Fact, bool, error) {
	return s.facts.Get(id)
}

func (s *Store) GetMediaCode(_ context.Context, id string) (performance.MediaCode, bool, error) {
	return s.codes.Get(id)
}

func (s *Store) GetAlias(_ context.Context, id string) (performance.Alias, bool, error) {
	a, ok := s.aliases[id]
	return a, ok, nil
}

func (s *Store) GetIntegrationID(_ context.Context, id string) (performance.IntegrationID, bool, error) {
	k, ok := s.integrations[id]
	return k, ok, nil
}

func (s *Store) DeleteFacts(ctx context.Context, ids []string) error {
	for _, id := range ids {
		if err := s.facts.Delete(ctx, id); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) DeleteMediaCodes(ctx context.Context, ids []string) error {
	for _, id := range ids {
		if err := s.codes.Delete(ctx, id); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) DeleteAliases(_ context.Context, ids []string) error {
	for _, id := range ids {
		delete(s.aliases, id)
	}
	return nil
}

func (s *Store) DeleteIntegrationIDs(_ context.Context, ids []string) error {
	for _, id := range ids {
		delete(s.integrations, id)
	}
	return nil
}

func (s *Store) PutAlias(_ context.Context, a performance.Alias) error {
	s.aliases[a.ID] = a
	return nil
}

func (s *Store) ExternalNameHolders(_ context.Context, name string) ([]string, error) {
	key := strings.ToLower(strings.TrimSpace(name))
	var holders []string
	for _, a := range s.aliases {
		if strings.ToLower(strings.TrimSpace(a.Name)) == key {
			holders = append(holders, a.ActivityID)
		}
	}
	holders = append(holders, s.Keys[performance.MatchedExternalName][key]...)
	return holders, nil
}

func (s *Store) ListFactsOfImport(_ context.Context, importID string) ([]performance.Fact, error) {
	all, err := s.facts.List()
	if err != nil {
		return nil, err
	}
	found := []performance.Fact{}
	for _, f := range all {
		if f.ImportID == importID {
			found = append(found, f)
		}
	}
	return found, nil
}

func (s *Store) FactsReplacedBy(_ context.Context, replacerIDs []string) ([]performance.Fact, error) {
	all, err := s.facts.List()
	if err != nil {
		return nil, err
	}
	found := []performance.Fact{}
	for _, f := range all {
		if f.ReplacedByID != "" && slices.Contains(replacerIDs, f.ReplacedByID) {
			found = append(found, f)
		}
	}
	return found, nil
}

func (s *Store) PutBreakdown(_ context.Context, importID string, groups []performance.Breakdown) error {
	if s.breakdowns == nil {
		s.breakdowns = map[string][]performance.Breakdown{}
	}
	s.breakdowns[importID] = slices.Clone(groups)
	return nil
}

func (s *Store) BreakdownOf(importID string) []performance.Breakdown {
	return s.breakdowns[importID]
}

func (s *Store) Breakdowns(_ context.Context, importIDs []string) (map[string][]performance.Breakdown, error) {
	groups := map[string][]performance.Breakdown{}
	for _, id := range importIDs {
		if held, ok := s.breakdowns[id]; ok {
			groups[id] = held
		}
	}
	return groups, nil
}

func (s *Store) CurrencyCodes(_ context.Context) (map[string]bool, error) {
	if s.Currencies == nil {
		return map[string]bool{"USD": true}, nil
	}
	return s.Currencies, nil
}
