package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"mapp/internal/platform/database/sqlcdb"
)

// RecordSignIn writes only the person's sign-in row, never the person (045 FR-010).
func (s *PostgresStore) RecordSignIn(ctx context.Context, personID string, at time.Time) error {
	if err := s.db.Queries(ctx).RecordSignIn(ctx, sqlcdb.RecordSignInParams{PersonID: personID, At: at}); err != nil {
		return fmt.Errorf("record the sign-in of %s: %w", personID, err)
	}
	return nil
}

func (s *PostgresStore) LastSignIns(ctx context.Context) (map[string]time.Time, error) {
	rows, err := s.db.Queries(ctx).ListLastSignIns(ctx)
	if err != nil {
		return nil, fmt.Errorf("select the last sign-ins: %w", err)
	}
	last := make(map[string]time.Time, len(rows))
	for _, r := range rows {
		last[r.PersonID] = r.LastSignedInAt
	}
	return last, nil
}

func (s *PostgresStore) SignInsRecordedSince(ctx context.Context) (time.Time, error) {
	since, err := s.db.Queries(ctx).GetSignInsRecordedSince(ctx)
	if errors.Is(err, pgx.ErrNoRows) {
		return time.Time{}, nil
	}
	return since, err
}
