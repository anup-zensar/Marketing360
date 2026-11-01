package performance

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"time"

	"mapp/internal/platform/authz"
	"mapp/internal/platform/database/sqlcdb"
)

// KeyFacts is what the Data sources area reads of the activity itself (FR-031); a scoped read, so an
// activity outside the scope answers as missing.
type KeyFacts struct {
	ExternalName string
	GoalEntered  bool
}

type SourceTotals struct {
	SourceID     string
	SourceName   string
	FactCount    int
	LastFactDate string
}

func (s *PostgresStore) ActivityKeyFacts(ctx context.Context, scope authz.Scope, activityID string) (KeyFacts, bool, error) {
	a := argsOf(scope)
	row, err := s.db.Queries(ctx).ActivityKeyFacts(ctx, sqlcdb.ActivityKeyFactsParams{
		ActivityID: activityID, SeesEverything: a.seesEverything, ViewerID: a.viewerID,
		ViewsPlans: a.viewsPlans, ViewerTeamIds: a.teamIDs,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return KeyFacts{}, false, nil
	}
	if err != nil {
		return KeyFacts{}, false, fmt.Errorf("select the key facts of %s: %w", activityID, err)
	}
	return KeyFacts{ExternalName: row.ExternalName, GoalEntered: row.GoalEntered}, true, nil
}

func (s *PostgresStore) FactTotalsBySource(ctx context.Context, activityID string) ([]SourceTotals, error) {
	rows, err := s.db.Queries(ctx).FactTotalsBySource(ctx, activityID)
	if err != nil {
		return nil, fmt.Errorf("select the fact totals of %s: %w", activityID, err)
	}
	totals := make([]SourceTotals, 0, len(rows))
	for _, r := range rows {
		totals = append(totals, SourceTotals{SourceID: r.SourceID, SourceName: r.SourceName,
			FactCount: int(r.FactCount), LastFactDate: r.LastFactDate.Format(time.DateOnly)})
	}
	return totals, nil
}

func (s *PostgresStore) ListAliases(ctx context.Context, activityID string) ([]Alias, error) {
	rows, err := s.db.Queries(ctx).ListAliasesOfActivity(ctx, activityID)
	if err != nil {
		return nil, fmt.Errorf("select the aliases of %s: %w", activityID, err)
	}
	aliases := make([]Alias, 0, len(rows))
	for _, n := range rows {
		aliases = append(aliases, Alias{ID: n.ID, ActivityID: n.ActivityID, Name: n.Name,
			Meta: metaOf(n.Version, metaRow{n.CreatedAt, n.CreatedByID, n.CreatedByName,
				n.UpdatedAt, n.UpdatedByID, n.UpdatedByName})})
	}
	return aliases, nil
}

func (s *PostgresStore) ListIntegrationIDs(ctx context.Context, activityID string) ([]IntegrationID, error) {
	rows, err := s.db.Queries(ctx).ListIntegrationIDsOfActivity(ctx, activityID)
	if err != nil {
		return nil, fmt.Errorf("select the integration ids of %s: %w", activityID, err)
	}
	keys := make([]IntegrationID, 0, len(rows))
	for _, k := range rows {
		keys = append(keys, IntegrationID{ID: k.ID, ActivityID: k.ActivityID, SourceID: k.SourceID, Value: k.Value,
			Meta: metaOf(k.Version, metaRow{k.CreatedAt, k.CreatedByID, k.CreatedByName,
				k.UpdatedAt, k.UpdatedByID, k.UpdatedByName})})
	}
	return keys, nil
}

func (s *PostgresStore) RemoveIntegrationID(ctx context.Context, activityID, sourceID string) (bool, error) {
	removed, err := s.db.Queries(ctx).DeleteIntegrationID(ctx, sqlcdb.DeleteIntegrationIDParams{
		ActivityID: activityID, SourceID: sourceID,
	})
	if err != nil {
		return false, fmt.Errorf("delete the integration id of %s for %s: %w", activityID, sourceID, err)
	}
	return removed > 0, nil
}

func (s *PostgresStore) RemoveAlias(ctx context.Context, activityID, aliasID string) (bool, error) {
	removed, err := s.db.Queries(ctx).DeleteAlias(ctx, sqlcdb.DeleteAliasParams{ID: aliasID, ActivityID: activityID})
	if err != nil {
		return false, fmt.Errorf("delete alias %s: %w", aliasID, err)
	}
	return removed > 0, nil
}
