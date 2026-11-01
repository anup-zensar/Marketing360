-- name: ListActivityCategories :many
SELECT sqlc.embed(activity_categories),
    coalesce((SELECT array_agg(t.activity_type_id ORDER BY t.activity_type_id COLLATE "C")
        FROM activity_category_types t
        WHERE t.category_id = activity_categories.id), '{}')::text[] AS activity_type_ids
FROM activity_categories
WHERE (sqlc.narg(id)::text IS NULL OR id = sqlc.narg(id))
ORDER BY length(id), id COLLATE "C";

-- name: PutActivityCategory :exec
INSERT INTO activity_categories (
    id, name, is_default, active, version, created_at, created_by_id, created_by_name, updated_at,
    updated_by_id, updated_by_name
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
)
ON CONFLICT (id) DO UPDATE SET
    name = excluded.name,
    is_default = excluded.is_default,
    active = excluded.active,
    version = excluded.version,
    created_at = excluded.created_at,
    created_by_id = excluded.created_by_id,
    created_by_name = excluded.created_by_name,
    updated_at = excluded.updated_at,
    updated_by_id = excluded.updated_by_id,
    updated_by_name = excluded.updated_by_name;

-- name: DeleteActivityCategory :exec
DELETE FROM activity_categories
WHERE id = $1;

-- name: ClearActivityCategoryTypes :exec
DELETE FROM activity_category_types
WHERE category_id = $1;

-- name: AssignActivityCategoryType :exec
INSERT INTO activity_category_types (activity_type_id, category_id)
VALUES ($1, $2)
ON CONFLICT (activity_type_id) DO UPDATE SET category_id = excluded.category_id;

-- name: ListActivityCategoryTypes :many
SELECT activity_type_id, category_id
FROM activity_category_types
ORDER BY activity_type_id COLLATE "C";
