-- name: GetAttributionSettings :many
SELECT * FROM attribution_settings WHERE id = 'attribution';

-- name: PutAttributionSettings :exec
INSERT INTO attribution_settings (
    id, model, version, created_at, created_by_id, created_by_name, updated_at, updated_by_id, updated_by_name
) VALUES (
    'attribution', $1, $2, $3, $4, $5, $6, $7, $8
)
ON CONFLICT (id) DO UPDATE SET
    model = excluded.model,
    version = excluded.version,
    updated_at = excluded.updated_at,
    updated_by_id = excluded.updated_by_id,
    updated_by_name = excluded.updated_by_name;
