-- name: InsertDataset :exec
INSERT INTO datasets (
    id, name, description, owner_id, file_name, sheet_name, row_count, version, created_at, created_by_id,
    created_by_name, updated_at, updated_by_id, updated_by_name
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14
);

-- name: UpdateDataset :exec
UPDATE datasets
SET name = $2, description = $3, version = $4, updated_at = $5, updated_by_id = $6, updated_by_name = $7
WHERE id = $1;

-- name: DeleteDataset :exec
DELETE FROM datasets
WHERE id = $1;

-- name: GetDataset :one
SELECT id, name, description, owner_id, file_name, sheet_name, row_count, version, created_at, created_by_id,
    created_by_name, updated_at, updated_by_id, updated_by_name
FROM datasets
WHERE id = $1;

-- name: GetDatasetVisible :one
SELECT id, name, description, owner_id, file_name, sheet_name, row_count, version, created_at, created_by_id,
    created_by_name, updated_at, updated_by_id, updated_by_name
FROM datasets d
WHERE d.id = sqlc.arg(id)
    AND (d.owner_id = sqlc.arg(person_id) OR EXISTS (
        SELECT 1 FROM dataset_shares s JOIN team_memberships m ON m.team_id = s.team_id
        WHERE s.dataset_id = d.id AND m.person_id = sqlc.arg(person_id)));

-- name: ListDatasetsVisible :many
SELECT id, name, description, owner_id, file_name, sheet_name, row_count, version, created_at, created_by_id,
    created_by_name, updated_at, updated_by_id, updated_by_name
FROM datasets d
WHERE d.owner_id = sqlc.arg(person_id) OR EXISTS (
    SELECT 1 FROM dataset_shares s JOIN team_memberships m ON m.team_id = s.team_id
    WHERE s.dataset_id = d.id AND m.person_id = sqlc.arg(person_id))
ORDER BY lower(d.name), d.id;

-- name: CountDatasetsOfOwner :one
SELECT count(*)::integer
FROM datasets
WHERE owner_id = $1;

-- name: DatasetOwner :one
SELECT owner_id
FROM datasets
WHERE id = $1;

-- name: DatasetColumns :many
SELECT dataset_id, position, key, label, kind, currency
FROM dataset_columns
WHERE dataset_id = ANY (sqlc.arg(dataset_ids)::text[])
ORDER BY dataset_id, position;

-- name: InsertDatasetColumns :copyfrom
INSERT INTO dataset_columns (dataset_id, position, key, label, kind, currency)
VALUES ($1, $2, $3, $4, $5, $6);

-- name: DeleteDatasetColumns :exec
DELETE FROM dataset_columns
WHERE dataset_id = $1;

-- name: DatasetRows :many
SELECT row_number, cells
FROM dataset_rows
WHERE dataset_id = sqlc.arg(dataset_id)
    AND row_number BETWEEN sqlc.arg(first_row)::integer AND sqlc.arg(last_row)::integer
ORDER BY row_number;

-- name: InsertDatasetRows :copyfrom
INSERT INTO dataset_rows (dataset_id, row_number, cells)
VALUES ($1, $2, $3);

-- name: DatasetShares :many
SELECT dataset_id, team_id
FROM dataset_shares
WHERE dataset_id = ANY (sqlc.arg(dataset_ids)::text[])
ORDER BY dataset_id, team_id;

-- name: DeleteDatasetShares :exec
DELETE FROM dataset_shares
WHERE dataset_id = $1;

-- name: InsertDatasetShares :exec
INSERT INTO dataset_shares (dataset_id, team_id)
SELECT sqlc.arg(dataset_id), unnest(sqlc.arg(team_ids)::text[]);
