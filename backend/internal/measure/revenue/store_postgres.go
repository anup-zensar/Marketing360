package revenue

import (
	"context"
	"fmt"

	"mapp/internal/platform/authz"
	"mapp/internal/platform/database"
	"mapp/internal/platform/database/sqlcdb"
)

const monthOnly = "2006-01"

var _ Facts = (*PostgresStore)(nil)

type PostgresStore struct {
	db *database.DB
}

func NewPostgresStore(db *database.DB) *PostgresStore {
	return &PostgresStore{db: db}
}

// ReportTotals is 049's one new read: the activities' counted facts, one row per activity (research R2).
func (s *PostgresStore) ReportTotals(ctx context.Context, scope authz.Scope, activityIDs []string, year int) ([]FactTotal, error) {
	rows, err := s.db.Queries(ctx).ReportFactTotals(ctx, sqlcdb.ReportFactTotalsParams{
		ActivityIds: activityIDs, Year: year,
		SeesEverything: scope.Everything, ViewerID: scope.PersonID,
		ViewsPlans: scope.Holds, ViewerTeamIds: scope.TeamIDs,
	})
	if err != nil {
		return nil, fmt.Errorf("sum the facts of %d activities: %w", len(activityIDs), err)
	}
	totals := make([]FactTotal, 0, len(rows))
	for _, r := range rows {
		totals = append(totals, FactTotal{
			ActivityID: r.ActivityID, Stage: r.Stage, WebMetric: r.WebMetric,
			Month: r.Month.Format(monthOnly), Currency: r.Currency,
			Counts: r.Counts, AmountMinor: r.AmountMinor, FactCount: int(r.FactCount),
		})
	}
	return totals, nil
}
