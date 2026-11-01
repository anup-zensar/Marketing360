-- name: GetReportView :one
SELECT id, person_id, report, name, is_default, query, version,
       created_at, created_by_id, created_by_name, updated_at, updated_by_id, updated_by_name
FROM report_views
WHERE id = $1;

-- name: ListReportViews :many
SELECT id, person_id, report, name, is_default, query, version,
       created_at, created_by_id, created_by_name, updated_at, updated_by_id, updated_by_name
FROM report_views
WHERE person_id = $1 AND report = $2
ORDER BY length(id), id COLLATE "C";

-- name: PutReportView :exec
INSERT INTO report_views (id, person_id, report, name, is_default, query, version,
                          created_at, created_by_id, created_by_name, updated_at, updated_by_id, updated_by_name)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
ON CONFLICT (id) DO UPDATE SET
    person_id = excluded.person_id,
    report = excluded.report,
    name = excluded.name,
    is_default = excluded.is_default,
    query = excluded.query,
    version = excluded.version,
    created_at = excluded.created_at,
    created_by_id = excluded.created_by_id,
    created_by_name = excluded.created_by_name,
    updated_at = excluded.updated_at,
    updated_by_id = excluded.updated_by_id,
    updated_by_name = excluded.updated_by_name;

-- name: DeleteReportView :exec
DELETE FROM report_views
WHERE id = $1;
