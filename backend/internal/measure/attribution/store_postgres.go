package attribution

import (
	"cmp"
	"context"
	"fmt"

	"mapp/internal/platform/actor"
	"mapp/internal/platform/database"
	"mapp/internal/platform/database/sqlcdb"
	"mapp/internal/platform/record"
)

type PostgresStore struct {
	db *database.DB
}

func NewPostgresStore(db *database.DB) *PostgresStore {
	return &PostgresStore{db: db}
}

func (s *PostgresStore) Settings(ctx context.Context) (Settings, error) {
	rows, err := s.db.Queries(ctx).GetAttributionSettings(ctx)
	if err != nil {
		return Settings{}, fmt.Errorf("select the attribution settings: %w", err)
	}
	if len(rows) == 0 {
		return DefaultSettings(), nil
	}
	r := rows[0]
	return Settings{Model: Model(r.Model), Meta: record.Meta{Version: r.Version, CreatedAt: r.CreatedAt,
		CreatedBy: actor.Actor{ID: r.CreatedByID, Name: r.CreatedByName}, UpdatedAt: r.UpdatedAt,
		UpdatedBy: actor.Actor{ID: r.UpdatedByID, Name: r.UpdatedByName}}}, nil
}

func (s *PostgresStore) PutSettings(ctx context.Context, st Settings) error {
	created, by := cmp.Or(st.CreatedAt, st.UpdatedAt), cmp.Or(st.CreatedBy, st.UpdatedBy)
	err := s.db.Queries(ctx).PutAttributionSettings(ctx, sqlcdb.PutAttributionSettingsParams{
		Model: string(st.Model), Version: st.Version, CreatedAt: created, CreatedByID: by.ID, CreatedByName: by.Name,
		UpdatedAt: st.UpdatedAt, UpdatedByID: st.UpdatedBy.ID, UpdatedByName: st.UpdatedBy.Name,
	})
	if err != nil {
		return fmt.Errorf("save the attribution settings: %w", err)
	}
	return nil
}
