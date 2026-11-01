package dataset

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"mapp/internal/platform/actor"
	"mapp/internal/platform/database"
	"mapp/internal/platform/database/sqlcdb"
	"mapp/internal/platform/record"
)

// PostgresStore keeps datasets on PostgreSQL; the visible reads filter by the person in SQL (data model §8).
type PostgresStore struct {
	db *database.DB
}

func NewPostgresStore(db *database.DB) *PostgresStore {
	return &PostgresStore{db: db}
}

func (s *PostgresStore) Get(ctx context.Context, id string) (Dataset, bool, error) {
	row, err := s.db.Queries(ctx).GetDataset(ctx, id)
	return s.one(ctx, row, err)
}

func (s *PostgresStore) GetVisible(ctx context.Context, personID, id string) (Dataset, bool, error) {
	row, err := s.db.Queries(ctx).GetDatasetVisible(ctx, sqlcdb.GetDatasetVisibleParams{ID: id, PersonID: personID})
	return s.one(ctx, row, err)
}

func (s *PostgresStore) one(ctx context.Context, row sqlcdb.Dataset, err error) (Dataset, bool, error) {
	if errors.Is(err, pgx.ErrNoRows) {
		return Dataset{}, false, nil
	}
	if err != nil {
		return Dataset{}, false, fmt.Errorf("select dataset: %w", err)
	}
	list, err := s.complete(ctx, []sqlcdb.Dataset{row})
	if err != nil {
		return Dataset{}, false, err
	}
	return list[0], true, nil
}

func (s *PostgresStore) ListVisible(ctx context.Context, personID string) ([]Dataset, error) {
	rows, err := s.db.Queries(ctx).ListDatasetsVisible(ctx, personID)
	if err != nil {
		return nil, fmt.Errorf("select visible datasets: %w", err)
	}
	return s.complete(ctx, rows)
}

// complete adds every dataset's columns and shares, read in two more statements.
func (s *PostgresStore) complete(ctx context.Context, rows []sqlcdb.Dataset) ([]Dataset, error) {
	ids := make([]string, 0, len(rows))
	byID := map[string]*Dataset{}
	list := make([]Dataset, len(rows))
	for i, r := range rows {
		list[i] = datasetOf(r)
		ids = append(ids, r.ID)
		byID[r.ID] = &list[i]
	}
	q := s.db.Queries(ctx)
	columns, err := q.DatasetColumns(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("select dataset columns: %w", err)
	}
	for _, c := range columns {
		byID[c.DatasetID].Columns = append(byID[c.DatasetID].Columns, Column{
			Key: c.Key, Position: c.Position, Label: c.Label, Kind: Kind(c.Kind), Currency: c.Currency,
		})
	}
	shares, err := q.DatasetShares(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("select dataset shares: %w", err)
	}
	for _, sh := range shares {
		byID[sh.DatasetID].SharedWith = append(byID[sh.DatasetID].SharedWith, sh.TeamID)
	}
	return list, nil
}

func datasetOf(r sqlcdb.Dataset) Dataset {
	return Dataset{
		ID: r.ID, Name: r.Name, Description: r.Description, OwnerID: r.OwnerID, FileName: r.FileName,
		SheetName: r.SheetName, RowCount: r.RowCount, Columns: []Column{}, SharedWith: []string{},
		Meta: record.Meta{
			Version: r.Version, CreatedAt: r.CreatedAt, CreatedBy: actor.Actor{ID: r.CreatedByID, Name: r.CreatedByName},
			UpdatedAt: r.UpdatedAt, UpdatedBy: actor.Actor{ID: r.UpdatedByID, Name: r.UpdatedByName},
		},
	}
}

func (s *PostgresStore) CountOfOwner(ctx context.Context, ownerID string) (int, error) {
	count, err := s.db.Queries(ctx).CountDatasetsOfOwner(ctx, ownerID)
	if err != nil {
		return 0, fmt.Errorf("count datasets: %w", err)
	}
	return count, nil
}

func (s *PostgresStore) OwnerOf(ctx context.Context, id string) (string, bool, error) {
	owner, err := s.db.Queries(ctx).DatasetOwner(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", false, nil
	}
	if err != nil {
		return "", false, fmt.Errorf("select dataset owner: %w", err)
	}
	return owner, true, nil
}

func (s *PostgresStore) Rows(ctx context.Context, id string) ([]Row, error) {
	return s.RowsPage(ctx, id, 0, maxRows)
}

func (s *PostgresStore) RowsPage(ctx context.Context, id string, offset, limit int) ([]Row, error) {
	stored, err := s.db.Queries(ctx).DatasetRows(ctx, sqlcdb.DatasetRowsParams{
		DatasetID: id, FirstRow: offset + 1, LastRow: offset + limit,
	})
	if err != nil {
		return nil, fmt.Errorf("select dataset rows: %w", err)
	}
	rows := make([]Row, 0, len(stored))
	for _, r := range stored {
		var row Row
		if err := json.Unmarshal(r.Cells, &row); err != nil {
			return nil, fmt.Errorf("decode dataset row %d: %w", r.RowNumber, err)
		}
		rows = append(rows, row)
	}
	return rows, nil
}

func (s *PostgresStore) Insert(ctx context.Context, d Dataset, rows []Row) error {
	return s.db.Write(ctx, func(ctx context.Context) error {
		q := s.db.Queries(ctx)
		if err := q.InsertDataset(ctx, insertParams(d)); err != nil {
			return fmt.Errorf("insert dataset: %w", err)
		}
		if err := s.insertColumns(ctx, d.ID, d.Columns); err != nil {
			return err
		}
		if err := s.ReplaceShares(ctx, d.ID, d.SharedWith); err != nil {
			return err
		}
		return s.insertRows(ctx, d.ID, rows)
	})
}

func insertParams(d Dataset) sqlcdb.InsertDatasetParams {
	return sqlcdb.InsertDatasetParams{
		ID: d.ID, Name: d.Name, Description: d.Description, OwnerID: d.OwnerID, FileName: d.FileName,
		SheetName: d.SheetName, RowCount: d.RowCount, Version: d.Version, CreatedAt: d.CreatedAt,
		CreatedByID: d.CreatedBy.ID, CreatedByName: d.CreatedBy.Name, UpdatedAt: d.UpdatedAt,
		UpdatedByID: d.UpdatedBy.ID, UpdatedByName: d.UpdatedBy.Name,
	}
}

func (s *PostgresStore) insertColumns(ctx context.Context, id string, columns []Column) error {
	params := make([]sqlcdb.InsertDatasetColumnsParams, 0, len(columns))
	for _, c := range columns {
		params = append(params, sqlcdb.InsertDatasetColumnsParams{
			DatasetID: id, Position: c.Position, Key: c.Key, Label: c.Label, Kind: string(c.Kind), Currency: c.Currency,
		})
	}
	if _, err := s.db.Queries(ctx).InsertDatasetColumns(ctx, params); err != nil {
		return fmt.Errorf("insert dataset columns: %w", err)
	}
	return nil
}

func (s *PostgresStore) insertRows(ctx context.Context, id string, rows []Row) error {
	params := make([]sqlcdb.InsertDatasetRowsParams, 0, len(rows))
	for i, r := range rows {
		cells, err := json.Marshal(r)
		if err != nil {
			return fmt.Errorf("encode dataset row %d: %w", i+1, err)
		}
		params = append(params, sqlcdb.InsertDatasetRowsParams{DatasetID: id, RowNumber: i + 1, Cells: cells})
	}
	if _, err := s.db.Queries(ctx).InsertDatasetRows(ctx, params); err != nil {
		return fmt.Errorf("insert dataset rows: %w", err)
	}
	return nil
}

func (s *PostgresStore) Update(ctx context.Context, d Dataset) error {
	err := s.db.Queries(ctx).UpdateDataset(ctx, sqlcdb.UpdateDatasetParams{
		ID: d.ID, Name: d.Name, Description: d.Description, Version: d.Version, UpdatedAt: d.UpdatedAt,
		UpdatedByID: d.UpdatedBy.ID, UpdatedByName: d.UpdatedBy.Name,
	})
	if err != nil {
		return fmt.Errorf("update dataset: %w", err)
	}
	return nil
}

// UpdateLabels writes the columns again: renaming two columns to each other's names would break the unique
// name index row by row.
func (s *PostgresStore) UpdateLabels(ctx context.Context, id string, columns []Column) error {
	return s.db.Write(ctx, func(ctx context.Context) error {
		if err := s.db.Queries(ctx).DeleteDatasetColumns(ctx, id); err != nil {
			return fmt.Errorf("delete dataset columns: %w", err)
		}
		return s.insertColumns(ctx, id, columns)
	})
}

func (s *PostgresStore) ReplaceShares(ctx context.Context, id string, teamIDs []string) error {
	q := s.db.Queries(ctx)
	if err := q.DeleteDatasetShares(ctx, id); err != nil {
		return fmt.Errorf("delete dataset shares: %w", err)
	}
	if err := q.InsertDatasetShares(ctx, sqlcdb.InsertDatasetSharesParams{DatasetID: id, TeamIds: teamIDs}); err != nil {
		return fmt.Errorf("insert dataset shares: %w", err)
	}
	return nil
}

func (s *PostgresStore) Delete(ctx context.Context, id string) error {
	if err := s.db.Queries(ctx).DeleteDataset(ctx, id); err != nil {
		return fmt.Errorf("delete dataset: %w", err)
	}
	return nil
}
