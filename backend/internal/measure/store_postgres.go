package measure

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

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

func (s *PostgresStore) ListViews(ctx context.Context, personID, report string) ([]StoredView, error) {
	rows, err := s.db.Queries(ctx).ListReportViews(ctx, sqlcdb.ListReportViewsParams{PersonID: personID, Report: report})
	if err != nil {
		return nil, fmt.Errorf("select %s views of %s: %w", report, personID, err)
	}
	views := make([]StoredView, 0, len(rows))
	for _, row := range rows {
		views = append(views, viewFromRow(row))
	}
	return views, nil
}

func (s *PostgresStore) GetView(ctx context.Context, id string) (StoredView, bool, error) {
	row, err := s.db.Queries(ctx).GetReportView(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return StoredView{}, false, nil
	}
	if err != nil {
		return StoredView{}, false, fmt.Errorf("select report view %s: %w", id, err)
	}
	return viewFromRow(row), true, nil
}

func (s *PostgresStore) PutView(ctx context.Context, v StoredView) error {
	err := s.db.Queries(ctx).PutReportView(ctx, sqlcdb.PutReportViewParams{
		ID: v.ID, PersonID: v.PersonID, Report: v.Report, Name: v.Name, IsDefault: v.IsDefault, Query: v.Query,
		Version: v.Version, CreatedAt: v.CreatedAt, CreatedByID: v.CreatedBy.ID, CreatedByName: v.CreatedBy.Name,
		UpdatedAt: v.UpdatedAt, UpdatedByID: v.UpdatedBy.ID, UpdatedByName: v.UpdatedBy.Name,
	})
	if err != nil {
		return fmt.Errorf("save report view %s: %w", v.ID, err)
	}
	return nil
}

func (s *PostgresStore) DeleteView(ctx context.Context, id string) error {
	if err := s.db.Queries(ctx).DeleteReportView(ctx, id); err != nil {
		return fmt.Errorf("delete report view %s: %w", id, err)
	}
	return nil
}

func viewFromRow(row sqlcdb.ReportView) StoredView {
	return StoredView{ID: row.ID, PersonID: row.PersonID, Report: row.Report, Name: row.Name,
		IsDefault: row.IsDefault, Query: row.Query, Meta: record.Meta{Version: row.Version, CreatedAt: row.CreatedAt,
			CreatedBy: actor.Actor{ID: row.CreatedByID, Name: row.CreatedByName}, UpdatedAt: row.UpdatedAt,
			UpdatedBy: actor.Actor{ID: row.UpdatedByID, Name: row.UpdatedByName}}}
}
