-- name: InsertCustomDashboard :exec
INSERT INTO custom_dashboards (
    id, name, description, owner_id, filter_date_from, filter_date_to, filter_plan_ids, filter_team_ids, version,
    created_at, created_by_id, created_by_name, updated_at, updated_by_id, updated_by_name
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15
);

-- name: UpdateCustomDashboard :execrows
UPDATE custom_dashboards
SET name = sqlc.arg(name),
    description = sqlc.arg(description),
    filter_date_from = sqlc.narg(filter_date_from),
    filter_date_to = sqlc.narg(filter_date_to),
    filter_plan_ids = sqlc.arg(filter_plan_ids),
    filter_team_ids = sqlc.arg(filter_team_ids),
    version = sqlc.arg(version),
    updated_at = sqlc.arg(updated_at),
    updated_by_id = sqlc.arg(updated_by_id),
    updated_by_name = sqlc.arg(updated_by_name)
WHERE id = sqlc.arg(id) AND version = sqlc.arg(previous_version);

-- name: DeleteCustomDashboard :exec
DELETE FROM custom_dashboards
WHERE id = $1;

-- name: GetCustomDashboard :one
SELECT id, name, description, owner_id, filter_date_from, filter_date_to, filter_plan_ids, filter_team_ids, version,
    created_at, created_by_id, created_by_name, updated_at, updated_by_id, updated_by_name
FROM custom_dashboards
WHERE id = $1;

-- name: GetCustomDashboardVisible :one
SELECT d.id, d.name, d.description, d.owner_id, d.filter_date_from, d.filter_date_to, d.filter_plan_ids,
    d.filter_team_ids, d.version, d.created_at, d.created_by_id, d.created_by_name, d.updated_at, d.updated_by_id,
    d.updated_by_name
FROM custom_dashboards d
WHERE d.id = sqlc.arg(id)
    AND (d.owner_id = sqlc.arg(person_id) OR EXISTS (
        SELECT 1
        FROM custom_dashboard_shares s
        JOIN team_memberships m ON m.team_id = s.team_id
        WHERE s.dashboard_id = d.id AND m.person_id = sqlc.arg(person_id)));

-- name: ListCustomDashboardsVisible :many
SELECT d.id, d.name, d.description, d.owner_id, d.filter_date_from, d.filter_date_to, d.filter_plan_ids,
    d.filter_team_ids, d.version, d.created_at, d.created_by_id, d.created_by_name, d.updated_at, d.updated_by_id,
    d.updated_by_name
FROM custom_dashboards d
WHERE d.owner_id = sqlc.arg(person_id) OR EXISTS (
    SELECT 1
    FROM custom_dashboard_shares s
    JOIN team_memberships m ON m.team_id = s.team_id
    WHERE s.dashboard_id = d.id AND m.person_id = sqlc.arg(person_id))
ORDER BY length(d.id), d.id COLLATE "C";

-- name: CustomDashboardWidgets :many
SELECT dashboard_id, id, position, type, title, subtitle, grid_x, grid_y, grid_w, grid_h, source, query, display, note
FROM custom_dashboard_widgets
WHERE dashboard_id = ANY (sqlc.arg(dashboard_ids)::text[])
ORDER BY dashboard_id, position;

-- name: InsertCustomDashboardWidgets :copyfrom
INSERT INTO custom_dashboard_widgets (
    dashboard_id, id, position, type, title, subtitle, grid_x, grid_y, grid_w, grid_h, source, query, display, note
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14
);

-- name: DeleteCustomDashboardWidgets :exec
DELETE FROM custom_dashboard_widgets
WHERE dashboard_id = $1;

-- name: CustomDashboardShares :many
SELECT dashboard_id, team_id
FROM custom_dashboard_shares
WHERE dashboard_id = ANY (sqlc.arg(dashboard_ids)::text[])
ORDER BY dashboard_id, length(team_id), team_id COLLATE "C";

-- name: DeleteCustomDashboardShares :exec
DELETE FROM custom_dashboard_shares
WHERE dashboard_id = $1;

-- name: InsertCustomDashboardShares :exec
INSERT INTO custom_dashboard_shares (dashboard_id, team_id)
SELECT sqlc.arg(dashboard_id)::text, unnest(sqlc.arg(team_ids)::text[]);

-- name: DashboardsUsingDataset :many
SELECT d.id, d.name
FROM custom_dashboards d
WHERE EXISTS (SELECT 1 FROM custom_dashboard_widgets w WHERE w.dashboard_id = d.id AND w.source = sqlc.arg(dataset_id))
    AND (d.owner_id = sqlc.arg(person_id) OR EXISTS (
        SELECT 1
        FROM custom_dashboard_shares s
        JOIN team_memberships m ON m.team_id = s.team_id
        WHERE s.dashboard_id = d.id AND m.person_id = sqlc.arg(person_id)))
ORDER BY lower(d.name), d.id;

-- name: CustomDashboardOwner :one
SELECT owner_id
FROM custom_dashboards
WHERE id = $1;

-- name: CustomDashboardNameTaken :one
SELECT EXISTS (
    SELECT 1
    FROM custom_dashboards
    WHERE owner_id = sqlc.arg(owner_id) AND lower(name) = lower(sqlc.arg(name)) AND id <> sqlc.arg(except_id)
);
