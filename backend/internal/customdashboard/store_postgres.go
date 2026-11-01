package customdashboard

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

// PostgresStore is the Store on PostgreSQL; every visible read filters by the person in SQL.
type PostgresStore struct {
	db *database.DB
}

func NewPostgresStore(db *database.DB) *PostgresStore {
	return &PostgresStore{db: db}
}

func (s *PostgresStore) ListVisible(ctx context.Context, personID string) ([]Dashboard, error) {
	rows, err := s.db.Queries(ctx).ListCustomDashboardsVisible(ctx, personID)
	if err != nil {
		return nil, fmt.Errorf("select visible custom dashboards: %w", err)
	}
	return s.withParts(ctx, rows)
}

func (s *PostgresStore) GetVisible(ctx context.Context, personID, id string) (Dashboard, bool, error) {
	row, err := s.db.Queries(ctx).GetCustomDashboardVisible(ctx, sqlcdb.GetCustomDashboardVisibleParams{ID: id, PersonID: personID})
	return s.one(ctx, row, err)
}

func (s *PostgresStore) Get(ctx context.Context, id string) (Dashboard, bool, error) {
	row, err := s.db.Queries(ctx).GetCustomDashboard(ctx, id)
	return s.one(ctx, row, err)
}

func (s *PostgresStore) Insert(ctx context.Context, d Dashboard) error {
	err := s.db.Queries(ctx).InsertCustomDashboard(ctx, sqlcdb.InsertCustomDashboardParams{
		ID: d.ID, Name: d.Name, Description: d.Description, OwnerID: d.OwnerID,
		FilterDateFrom: database.DateColumn(d.Filters.DateFrom), FilterDateTo: database.DateColumn(d.Filters.DateTo),
		FilterPlanIds: d.Filters.PlanIDs, FilterTeamIds: d.Filters.TeamIDs, Version: d.Version, CreatedAt: d.CreatedAt,
		CreatedByID: d.CreatedBy.ID, CreatedByName: d.CreatedBy.Name, UpdatedAt: d.UpdatedAt,
		UpdatedByID: d.UpdatedBy.ID, UpdatedByName: d.UpdatedBy.Name,
	})
	if err != nil {
		return fmt.Errorf("insert custom dashboard %s: %w", d.ID, err)
	}
	if err := s.insertWidgets(ctx, d.ID, d.Widgets); err != nil {
		return err
	}
	return s.insertShares(ctx, d.ID, d.SharedWith)
}

// Update saves the dashboard's own fields; widgets and shares change through their Replace methods.
func (s *PostgresStore) Update(ctx context.Context, d Dashboard, previousVersion int) error {
	updated, err := s.db.Queries(ctx).UpdateCustomDashboard(ctx, sqlcdb.UpdateCustomDashboardParams{
		ID: d.ID, PreviousVersion: previousVersion, Name: d.Name, Description: d.Description,
		FilterDateFrom: database.DateColumn(d.Filters.DateFrom), FilterDateTo: database.DateColumn(d.Filters.DateTo),
		FilterPlanIds: d.Filters.PlanIDs, FilterTeamIds: d.Filters.TeamIDs, Version: d.Version,
		UpdatedAt: d.UpdatedAt, UpdatedByID: d.UpdatedBy.ID, UpdatedByName: d.UpdatedBy.Name,
	})
	if err != nil {
		return fmt.Errorf("update custom dashboard %s: %w", d.ID, err)
	}
	if updated == 0 {
		return fmt.Errorf("update custom dashboard %s: version %d is no longer stored", d.ID, previousVersion)
	}
	return nil
}

func (s *PostgresStore) Delete(ctx context.Context, id string) error {
	if err := s.db.Queries(ctx).DeleteCustomDashboard(ctx, id); err != nil {
		return fmt.Errorf("delete custom dashboard %s: %w", id, err)
	}
	return nil
}

func (s *PostgresStore) ReplaceWidgets(ctx context.Context, id string, widgets []Widget) error {
	if err := s.db.Queries(ctx).DeleteCustomDashboardWidgets(ctx, id); err != nil {
		return fmt.Errorf("delete widgets of custom dashboard %s: %w", id, err)
	}
	return s.insertWidgets(ctx, id, widgets)
}

func (s *PostgresStore) ReplaceShares(ctx context.Context, id string, teamIDs []string) error {
	if err := s.db.Queries(ctx).DeleteCustomDashboardShares(ctx, id); err != nil {
		return fmt.Errorf("delete shares of custom dashboard %s: %w", id, err)
	}
	return s.insertShares(ctx, id, teamIDs)
}

func (s *PostgresStore) UsingDataset(ctx context.Context, personID, datasetID string) ([]Ref, error) {
	rows, err := s.db.Queries(ctx).DashboardsUsingDataset(ctx, sqlcdb.DashboardsUsingDatasetParams{
		DatasetID: &datasetID, PersonID: personID,
	})
	if err != nil {
		return nil, fmt.Errorf("select custom dashboards using %s: %w", datasetID, err)
	}
	refs := make([]Ref, 0, len(rows))
	for _, row := range rows {
		refs = append(refs, Ref{ID: row.ID, Name: row.Name})
	}
	return refs, nil
}

func (s *PostgresStore) NameTaken(ctx context.Context, ownerID, name, exceptID string) (bool, error) {
	taken, err := s.db.Queries(ctx).CustomDashboardNameTaken(ctx, sqlcdb.CustomDashboardNameTakenParams{
		OwnerID: ownerID, Name: name, ExceptID: exceptID,
	})
	if err != nil {
		return false, fmt.Errorf("check the custom dashboard names of %s: %w", ownerID, err)
	}
	return taken, nil
}

// OwnerOf lets history show a dashboard's log and bin entries to its owner.
func (s *PostgresStore) OwnerOf(ctx context.Context, id string) (ownerID string, found bool, err error) {
	ownerID, err = s.db.Queries(ctx).CustomDashboardOwner(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", false, nil
	}
	if err != nil {
		return "", false, fmt.Errorf("select the owner of custom dashboard %s: %w", id, err)
	}
	return ownerID, true, nil
}

func (s *PostgresStore) one(ctx context.Context, row sqlcdb.CustomDashboard, err error) (Dashboard, bool, error) {
	if errors.Is(err, pgx.ErrNoRows) {
		return Dashboard{}, false, nil
	}
	if err != nil {
		return Dashboard{}, false, fmt.Errorf("select custom dashboard: %w", err)
	}
	list, err := s.withParts(ctx, []sqlcdb.CustomDashboard{row})
	if err != nil {
		return Dashboard{}, false, err
	}
	return list[0], true, nil
}

// withParts reads the widgets and shares of every row in two statements.
func (s *PostgresStore) withParts(ctx context.Context, rows []sqlcdb.CustomDashboard) ([]Dashboard, error) {
	ids := make([]string, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.ID)
	}
	widgetRows, err := s.db.Queries(ctx).CustomDashboardWidgets(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("select custom dashboard widgets: %w", err)
	}
	shareRows, err := s.db.Queries(ctx).CustomDashboardShares(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("select custom dashboard shares: %w", err)
	}
	widgets, err := widgetsByDashboard(widgetRows)
	if err != nil {
		return nil, err
	}
	shares := map[string][]string{}
	for _, row := range shareRows {
		shares[row.DashboardID] = append(shares[row.DashboardID], row.TeamID)
	}
	list := make([]Dashboard, 0, len(rows))
	for _, row := range rows {
		list = append(list, dashboardFromRow(row, widgets[row.ID], shares[row.ID]))
	}
	return list, nil
}

func (s *PostgresStore) insertWidgets(ctx context.Context, id string, widgets []Widget) error {
	rows := make([]sqlcdb.InsertCustomDashboardWidgetsParams, 0, len(widgets))
	for position, w := range widgets {
		row, err := widgetRow(id, position, w)
		if err != nil {
			return err
		}
		rows = append(rows, row)
	}
	if _, err := s.db.Queries(ctx).InsertCustomDashboardWidgets(ctx, rows); err != nil {
		return fmt.Errorf("insert widgets of custom dashboard %s: %w", id, err)
	}
	return nil
}

func (s *PostgresStore) insertShares(ctx context.Context, id string, teamIDs []string) error {
	err := s.db.Queries(ctx).InsertCustomDashboardShares(ctx, sqlcdb.InsertCustomDashboardSharesParams{
		DashboardID: id, TeamIds: teamIDs,
	})
	if err != nil {
		return fmt.Errorf("insert shares of custom dashboard %s: %w", id, err)
	}
	return nil
}

func dashboardFromRow(row sqlcdb.CustomDashboard, widgets []Widget, shares []string) Dashboard {
	return normalised(Dashboard{
		ID: row.ID, Name: row.Name, Description: row.Description, OwnerID: row.OwnerID, SharedWith: shares,
		Filters: Filters{DateFrom: database.DateOf(row.FilterDateFrom), DateTo: database.DateOf(row.FilterDateTo),
			PlanIDs: row.FilterPlanIds, TeamIDs: row.FilterTeamIds},
		Widgets: widgets,
		Meta: record.Meta{
			Version: row.Version, CreatedAt: row.CreatedAt, CreatedBy: actor.Actor{ID: row.CreatedByID, Name: row.CreatedByName},
			UpdatedAt: row.UpdatedAt, UpdatedBy: actor.Actor{ID: row.UpdatedByID, Name: row.UpdatedByName},
		},
	})
}
