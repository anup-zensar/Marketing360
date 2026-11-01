package performance

import (
	"context"
	"fmt"

	"mapp/internal/history"
	"mapp/internal/platform/actor"
	"mapp/internal/platform/authz"
	"mapp/internal/platform/errs"
)

// DataSourcesArea is the activity's Performance › Data sources section (FR-031). The external name is
// feature 006's field: read here, never written (046-C14).
type DataSourcesArea struct {
	ExternalName   string          `json:"externalName"`
	Aliases        []Alias         `json:"aliases"`
	MediaCodeCount int             `json:"mediaCodeCount"`
	Integrated     bool            `json:"integrated"`
	GoalEntered    bool            `json:"goalEntered"`
	Sources        []SourceOfFacts `json:"sources"`
}

type SourceOfFacts struct {
	SourceID      string `json:"sourceId"`
	SourceName    string `json:"sourceName"`
	IntegrationID string `json:"integrationId"`
	FactCount     int    `json:"factCount"`
	LastFactDate  string `json:"lastFactDate"`
}

func (s *Service) SourcesArea(ctx context.Context, _ actor.Actor, activityID string) (DataSourcesArea, error) {
	var area DataSourcesArea
	err := s.tx.Read(ctx, func(ctx context.Context) error {
		facts, readErr := s.visibleActivity(ctx, activityID)
		if readErr != nil {
			return readErr
		}
		area, readErr = s.areaOf(ctx, activityID, facts)
		return readErr
	})
	if err != nil {
		return DataSourcesArea{}, fmt.Errorf("read the data sources area of %s: %w", activityID, err)
	}
	return area, nil
}

// visibleActivity is the area's scope gate: an activity outside the reader's scope answers as missing,
// exactly like one that does not exist (constitution non-negotiable 7).
func (s *Service) visibleActivity(ctx context.Context, activityID string) (KeyFacts, error) {
	facts, found, err := s.repo.ActivityKeyFacts(ctx, authz.ScopeOf(ctx), activityID)
	if err != nil {
		return KeyFacts{}, err
	}
	if !found {
		return KeyFacts{}, errs.NotFound(msgActivityNotFound)
	}
	return facts, nil
}

func (s *Service) areaOf(ctx context.Context, activityID string, facts KeyFacts) (DataSourcesArea, error) {
	aliases, err := s.repo.ListAliases(ctx, activityID)
	if err != nil {
		return DataSourcesArea{}, err
	}
	keys, err := s.repo.ListIntegrationIDs(ctx, activityID)
	if err != nil {
		return DataSourcesArea{}, err
	}
	codes, err := s.repo.ListMediaCodes(ctx, authz.ScopeOf(ctx), activityID)
	if err != nil {
		return DataSourcesArea{}, err
	}
	totals, err := s.repo.FactTotalsBySource(ctx, activityID)
	if err != nil {
		return DataSourcesArea{}, err
	}
	return DataSourcesArea{
		ExternalName: facts.ExternalName, Aliases: aliases, MediaCodeCount: len(codes),
		Integrated: len(keys) > 0, GoalEntered: facts.GoalEntered,
		Sources: sourcesOf(keys, totals),
	}, nil
}

// sourcesOf merges the registered integration ids with the fact totals: a source appears when it holds
// either, so a key without data and data without a key both show.
func sourcesOf(keys []IntegrationID, totals []SourceTotals) []SourceOfFacts {
	bySource := map[string]*SourceOfFacts{}
	var order []string
	for _, total := range totals {
		bySource[total.SourceID] = &SourceOfFacts{SourceID: total.SourceID, SourceName: total.SourceName,
			FactCount: total.FactCount, LastFactDate: total.LastFactDate}
		order = append(order, total.SourceID)
	}
	for _, key := range keys {
		if found, ok := bySource[key.SourceID]; ok {
			found.IntegrationID = key.Value
			continue
		}
		bySource[key.SourceID] = &SourceOfFacts{SourceID: key.SourceID, IntegrationID: key.Value}
		order = append(order, key.SourceID)
	}
	merged := make([]SourceOfFacts, 0, len(order))
	for _, id := range order {
		merged = append(merged, *bySource[id])
	}
	return merged
}

func (s *Service) Aliases(ctx context.Context, _ actor.Actor, activityID string) ([]Alias, error) {
	var aliases []Alias
	err := s.tx.Read(ctx, func(ctx context.Context) error {
		if _, readErr := s.visibleActivity(ctx, activityID); readErr != nil {
			return readErr
		}
		var readErr error
		aliases, readErr = s.repo.ListAliases(ctx, activityID)
		return readErr
	})
	if err != nil {
		return nil, fmt.Errorf("list the aliases of %s: %w", activityID, err)
	}
	return aliases, nil
}

// DeleteAlias stops the old name matching; the confirmation wording carries that consequence (FR-034).
func (s *Service) DeleteAlias(ctx context.Context, a actor.Actor, activityID, aliasID string) error {
	if err := authz.Require(ctx, PermissionLink); err != nil {
		return err
	}
	return s.tx.Write(ctx, func(ctx context.Context) error {
		if _, err := s.visibleActivity(ctx, activityID); err != nil {
			return err
		}
		alias, found, err := s.repo.GetAlias(ctx, aliasID)
		if err != nil {
			return err
		}
		if !found || alias.ActivityID != activityID {
			return errs.NotFound(msgAliasNotFound)
		}
		if _, removeErr := s.repo.RemoveAlias(ctx, activityID, aliasID); removeErr != nil {
			return removeErr
		}
		_, err = s.history.Log(ctx, a, history.LogInput{
			Kind: aliasKind, RecordID: alias.ID, RecordName: alias.Name,
			ParentID: activityID, Action: history.ActionDeleted,
		})
		return err
	})
}
