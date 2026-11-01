-- name: ListPerformanceSources :many
SELECT sqlc.embed(performance_sources),
    (SELECT count(*) FROM performance_facts f WHERE f.source_id = performance_sources.id)::integer AS fact_count
FROM performance_sources
WHERE sqlc.narg(id)::text IS NULL OR performance_sources.id = sqlc.narg(id)
ORDER BY lower(performance_sources.name COLLATE "C"), performance_sources.id COLLATE "C";

-- Freshness is derived, and a source that has never loaded simply has no row here, so nothing is NULL.
-- name: SourceFreshness :many
SELECT i.source_id, max(i.created_at)::timestamptz AS last_loaded_at, max(i.data_through)::date AS data_through
FROM performance_imports i
WHERE i.status = 'completed' AND i.data_through IS NOT NULL
GROUP BY i.source_id;

-- name: PutPerformanceSource :exec
INSERT INTO performance_sources (
    id, name, description, owner_id, category, status, delivery_route, cadence, integration_id,
    version, created_at, created_by_id, created_by_name, updated_at, updated_by_id, updated_by_name
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
ON CONFLICT (id) DO UPDATE SET
    name = excluded.name, description = excluded.description, owner_id = excluded.owner_id,
    category = excluded.category, status = excluded.status, delivery_route = excluded.delivery_route,
    cadence = excluded.cadence, integration_id = excluded.integration_id, version = excluded.version,
    updated_at = excluded.updated_at, updated_by_id = excluded.updated_by_id,
    updated_by_name = excluded.updated_by_name;

-- name: DeletePerformanceSource :execrows
DELETE FROM performance_sources
WHERE performance_sources.id = sqlc.arg(source_id)
    AND NOT EXISTS (SELECT 1 FROM performance_facts f WHERE f.source_id = performance_sources.id);

-- name: ListPerformanceMappings :many
SELECT sqlc.embed(performance_source_mappings)
FROM performance_source_mappings
WHERE source_id = ANY (sqlc.arg(source_ids)::text[])
ORDER BY source_id COLLATE "C", lower(column_name COLLATE "C");

-- name: ActiveSuppliersForModel :many
SELECT m.id, m.source_id, m.meaning, m.stage, s.name AS source_name
FROM performance_source_mappings m
JOIN performance_sources s ON s.id = m.source_id
WHERE m.model_id = sqlc.arg(model_id)
    AND m.meaning = ANY (ARRAY['stage', 'revenue'])
    AND s.status = 'active'
ORDER BY m.id COLLATE "C";

-- name: PutPerformanceMapping :exec
INSERT INTO performance_source_mappings (
    id, source_id, column_name, meaning, key_type, stage, web_metric, model_id,
    version, created_at, created_by_id, created_by_name, updated_at, updated_by_id, updated_by_name
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
ON CONFLICT (id) DO UPDATE SET
    column_name = excluded.column_name, meaning = excluded.meaning, key_type = excluded.key_type,
    stage = excluded.stage,
    web_metric = excluded.web_metric, model_id = excluded.model_id, version = excluded.version,
    updated_at = excluded.updated_at, updated_by_id = excluded.updated_by_id,
    updated_by_name = excluded.updated_by_name;

-- name: DeletePerformanceMapping :exec
DELETE FROM performance_source_mappings WHERE id = $1;

-- name: MatchActivitiesByID :many
SELECT id AS activity_id, lower(id) AS key
FROM activities
WHERE lower(id) = ANY (sqlc.arg(keys)::text[])
    AND activity_in_scope(id, sqlc.arg(sees_everything)::boolean, sqlc.arg(viewer_id)::text,
        sqlc.arg(views_plans)::boolean, sqlc.arg(viewer_team_ids)::text[]);

-- name: MatchActivitiesByMediaCode :many
SELECT c.activity_id, c.code::text AS key
FROM media_codes c
WHERE c.code = ANY (sqlc.arg(codes)::bigint[])
    AND activity_in_scope(c.activity_id, sqlc.arg(sees_everything)::boolean, sqlc.arg(viewer_id)::text,
        sqlc.arg(views_plans)::boolean, sqlc.arg(viewer_team_ids)::text[]);

-- name: MatchActivitiesByIntegrationID :many
SELECT k.activity_id, k.value AS key
FROM activity_integration_ids k
WHERE k.source_id = sqlc.arg(source_id)
    AND lower(k.value COLLATE "C") = ANY (sqlc.arg(keys)::text[])
    AND activity_in_scope(k.activity_id, sqlc.arg(sees_everything)::boolean, sqlc.arg(viewer_id)::text,
        sqlc.arg(views_plans)::boolean, sqlc.arg(viewer_team_ids)::text[]);

-- name: MatchActivitiesByExternalName :many
SELECT a.id AS activity_id, lower(a.external_name COLLATE "C") AS key
FROM activities a
WHERE a.external_name <> '' AND lower(a.external_name COLLATE "C") = ANY (sqlc.arg(keys)::text[])
    AND activity_in_scope(a.id, sqlc.arg(sees_everything)::boolean, sqlc.arg(viewer_id)::text,
        sqlc.arg(views_plans)::boolean, sqlc.arg(viewer_team_ids)::text[])
UNION
SELECT n.activity_id, lower(n.name COLLATE "C") AS key
FROM activity_external_name_aliases n
WHERE lower(n.name COLLATE "C") = ANY (sqlc.arg(keys)::text[])
    AND activity_in_scope(n.activity_id, sqlc.arg(sees_everything)::boolean, sqlc.arg(viewer_id)::text,
        sqlc.arg(views_plans)::boolean, sqlc.arg(viewer_team_ids)::text[]);

-- name: ListFactsForActivity :many
SELECT sqlc.embed(performance_facts), s.name AS source_name
FROM performance_facts
JOIN performance_sources s ON s.id = performance_facts.source_id
WHERE performance_facts.activity_id = sqlc.arg(activity_id)
    AND performance_facts.replaced_by_id IS NULL
    AND (NOT sqlc.arg(counted_only)::boolean OR NOT performance_facts.held)
    AND (sqlc.narg(search)::text IS NULL
         OR s.name ILIKE '%' || sqlc.narg(search) || '%'
         OR performance_facts.stage ILIKE '%' || sqlc.narg(search) || '%'
         OR performance_facts.web_metric ILIKE '%' || sqlc.narg(search) || '%'
         OR performance_facts.matched_on ILIKE '%' || sqlc.narg(search) || '%'
         OR performance_facts.import_id ILIKE '%' || sqlc.narg(search) || '%'
         OR performance_facts.source_reference ILIKE '%' || sqlc.narg(search) || '%')
    AND activity_in_scope(performance_facts.activity_id, sqlc.arg(sees_everything)::boolean,
        sqlc.arg(viewer_id)::text, sqlc.arg(views_plans)::boolean, sqlc.arg(viewer_team_ids)::text[])
ORDER BY performance_facts.fact_date DESC, performance_facts.id COLLATE "C"
LIMIT sqlc.arg(max_rows);

-- name: ReleaseHeldFactsForActivity :execrows
UPDATE performance_facts
SET held = false, version = version + 1, updated_at = sqlc.arg(updated_at),
    updated_by_id = sqlc.arg(updated_by_id), updated_by_name = sqlc.arg(updated_by_name)
WHERE activity_id = sqlc.arg(activity_id) AND held;

-- name: ListPerformanceImports :many
SELECT sqlc.embed(performance_imports)
FROM performance_imports
WHERE (sqlc.narg(id)::text IS NULL OR id = sqlc.narg(id))
    AND (sqlc.narg(source_id)::text IS NULL OR source_id = sqlc.narg(source_id))
ORDER BY created_at DESC, id COLLATE "C"
LIMIT sqlc.arg(max_imports);

-- name: PutPerformanceImport :exec
INSERT INTO performance_imports (
    id, source_id, file_name, origin, status, data_through, rows_read, rows_matched, rows_held,
    rows_unattributed, rows_rejected, rows_unmatched,
    version, created_at, created_by_id, created_by_name, updated_at, updated_by_id, updated_by_name
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19)
ON CONFLICT (id) DO UPDATE SET
    status = excluded.status, version = excluded.version, updated_at = excluded.updated_at,
    updated_by_id = excluded.updated_by_id, updated_by_name = excluded.updated_by_name;

-- name: ListPerformanceReviewRows :many
SELECT sqlc.embed(performance_review_rows), s.name AS source_name
FROM performance_review_rows
JOIN performance_sources s ON s.id = performance_review_rows.source_id
WHERE (sqlc.narg(status)::text IS NULL OR performance_review_rows.status = sqlc.narg(status))
    AND (sqlc.narg(status_not)::text IS NULL OR performance_review_rows.status <> sqlc.narg(status_not))
    AND (sqlc.narg(source_id)::text IS NULL OR performance_review_rows.source_id = sqlc.narg(source_id))
    AND (sqlc.narg(id)::text IS NULL OR performance_review_rows.id = sqlc.narg(id))
    AND (sqlc.narg(import_id)::text IS NULL OR performance_review_rows.import_id = sqlc.narg(import_id))
    AND (sqlc.narg(reason)::text IS NULL OR performance_review_rows.reason = sqlc.narg(reason))
    AND (sqlc.narg(key_text)::text IS NULL
         OR EXISTS (SELECT 1 FROM json_each_text(performance_review_rows.raw) kv
                    WHERE kv.value ILIKE '%' || sqlc.narg(key_text) || '%'))
    AND (sqlc.narg(since)::timestamptz IS NULL OR performance_review_rows.created_at >= sqlc.narg(since))
    AND (sqlc.narg(until)::timestamptz IS NULL OR performance_review_rows.created_at <= sqlc.narg(until))
ORDER BY performance_review_rows.created_at DESC, performance_review_rows.id COLLATE "C"
LIMIT sqlc.arg(max_rows);

-- name: ListMediaCodesForActivity :many
SELECT sqlc.embed(media_codes)
FROM media_codes
WHERE activity_id = sqlc.arg(activity_id)
    AND activity_in_scope(activity_id, sqlc.arg(sees_everything)::boolean, sqlc.arg(viewer_id)::text,
        sqlc.arg(views_plans)::boolean, sqlc.arg(viewer_team_ids)::text[])
ORDER BY code;

-- Creates the configuration row on first use, so a fresh database does not need a seeded singleton, and
-- raises the counter in the same statement, so two codes can never take the same number (FR-042).
-- name: TakeNextMediaCode :one
INSERT INTO performance_configuration (
    id, media_code_next, code_attributes,
    version, created_at, created_by_id, created_by_name, updated_at, updated_by_id, updated_by_name
) VALUES ('performance', sqlc.arg(start_at)::bigint + 1, '[]',
    1, sqlc.arg(updated_at), sqlc.arg(updated_by_id), sqlc.arg(updated_by_name),
    sqlc.arg(updated_at), sqlc.arg(updated_by_id), sqlc.arg(updated_by_name))
ON CONFLICT (id) DO UPDATE SET
    media_code_next = performance_configuration.media_code_next + 1,
    version = performance_configuration.version + 1,
    updated_at = excluded.updated_at, updated_by_id = excluded.updated_by_id,
    updated_by_name = excluded.updated_by_name
RETURNING (performance_configuration.media_code_next - 1)::bigint AS code;

-- name: PutPerformanceFact :exec
INSERT INTO performance_facts (
    id, source_id, import_id, source_reference, activity_id, matched_on, matched_key, fact_date, stage, web_metric,
    count_value, amount_minor, currency, held, replaced_by_id,
    version, created_at, created_by_id, created_by_name, updated_at, updated_by_id, updated_by_name
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22)
ON CONFLICT (id) DO UPDATE SET
    activity_id = excluded.activity_id, matched_on = excluded.matched_on, held = excluded.held,
    replaced_by_id = excluded.replaced_by_id, version = excluded.version, updated_at = excluded.updated_at,
    updated_by_id = excluded.updated_by_id, updated_by_name = excluded.updated_by_name;

-- name: PutPerformanceReviewRow :exec
INSERT INTO performance_review_rows (
    id, import_id, source_id, raw, parsed, reason, faults, candidates, status, comment,
    version, created_at, created_by_id, created_by_name, updated_at, updated_by_id, updated_by_name
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
ON CONFLICT (id) DO UPDATE SET
    parsed = excluded.parsed, reason = excluded.reason, faults = excluded.faults,
    candidates = excluded.candidates,
    status = excluded.status, comment = excluded.comment, version = excluded.version,
    updated_at = excluded.updated_at, updated_by_id = excluded.updated_by_id,
    updated_by_name = excluded.updated_by_name;

-- name: CurrentFactsByReferences :many
SELECT sqlc.embed(performance_facts)
FROM performance_facts
WHERE source_id = sqlc.arg(source_id) AND source_reference = ANY(sqlc.arg(source_references)::text[])
    AND replaced_by_id IS NULL
ORDER BY source_reference COLLATE "C", stage COLLATE "C", web_metric COLLATE "C";

-- name: PutMediaCode :exec
INSERT INTO media_codes (
    id, code, activity_id, attributes, landing_page, status,
    version, created_at, created_by_id, created_by_name, updated_at, updated_by_id, updated_by_name
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
ON CONFLICT (id) DO UPDATE SET
    attributes = excluded.attributes, landing_page = excluded.landing_page, status = excluded.status,
    version = excluded.version, updated_at = excluded.updated_at,
    updated_by_id = excluded.updated_by_id, updated_by_name = excluded.updated_by_name;

-- name: PutIntegrationID :exec
INSERT INTO activity_integration_ids (
    id, activity_id, source_id, value,
    version, created_at, created_by_id, created_by_name, updated_at, updated_by_id, updated_by_name
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
ON CONFLICT (id) DO UPDATE SET
    value = excluded.value, version = excluded.version, updated_at = excluded.updated_at,
    updated_by_id = excluded.updated_by_id, updated_by_name = excluded.updated_by_name;

-- The four queries below serve the recycle bin, not a person: history collects an activity's children
-- before deleting it, so they carry no scope filter.
-- name: FactIDsOfActivities :many
SELECT id FROM performance_facts WHERE activity_id = ANY (sqlc.arg(activity_ids)::text[]) ORDER BY id;

-- name: MediaCodeIDsOfActivities :many
SELECT id FROM media_codes WHERE activity_id = ANY (sqlc.arg(activity_ids)::text[]) ORDER BY id;

-- name: AliasIDsOfActivities :many
SELECT id FROM activity_external_name_aliases
WHERE activity_id = ANY (sqlc.arg(activity_ids)::text[]) ORDER BY id;

-- name: IntegrationIDsOfActivities :many
SELECT id FROM activity_integration_ids
WHERE activity_id = ANY (sqlc.arg(activity_ids)::text[]) ORDER BY id;

-- name: GetPerformanceFact :one
SELECT sqlc.embed(performance_facts) FROM performance_facts WHERE id = $1;

-- name: GetMediaCode :one
SELECT sqlc.embed(media_codes) FROM media_codes WHERE id = $1;

-- name: GetAlias :one
SELECT sqlc.embed(activity_external_name_aliases) FROM activity_external_name_aliases WHERE id = $1;

-- name: GetIntegrationID :one
SELECT sqlc.embed(activity_integration_ids) FROM activity_integration_ids WHERE id = $1;

-- name: DeletePerformanceFacts :exec
DELETE FROM performance_facts WHERE id = ANY (sqlc.arg(ids)::text[]);

-- name: DeleteMediaCodes :exec
DELETE FROM media_codes WHERE id = ANY (sqlc.arg(ids)::text[]);

-- name: DeleteAliases :exec
DELETE FROM activity_external_name_aliases WHERE id = ANY (sqlc.arg(ids)::text[]);

-- name: DeleteIntegrationIDs :exec
DELETE FROM activity_integration_ids WHERE id = ANY (sqlc.arg(ids)::text[]);

-- name: PutAlias :exec
INSERT INTO activity_external_name_aliases (
    id, activity_id, name,
    version, created_at, created_by_id, created_by_name, updated_at, updated_by_id, updated_by_name
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
ON CONFLICT (id) DO UPDATE SET
    name = excluded.name, version = excluded.version, updated_at = excluded.updated_at,
    updated_by_id = excluded.updated_by_id, updated_by_name = excluded.updated_by_name;

-- FR-035: external names and aliases share one namespace, so a key never resolves to two activities.
-- name: ExternalNameHolders :many
SELECT a.id AS activity_id
FROM activities a
WHERE a.external_name <> '' AND lower(a.external_name COLLATE "C") = lower(sqlc.arg(name)::text COLLATE "C")
UNION
SELECT n.activity_id
FROM activity_external_name_aliases n
WHERE lower(n.name COLLATE "C") = lower(sqlc.arg(name)::text COLLATE "C");

-- name: PutPerformanceImportBreakdown :exec
INSERT INTO performance_import_breakdown (
    import_id, stage, web_metric, rows_read, rows_matched, rows_held,
    rows_unattributed, rows_rejected, rows_unmatched
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
ON CONFLICT (import_id, stage, web_metric) DO UPDATE SET
    rows_read = excluded.rows_read, rows_matched = excluded.rows_matched,
    rows_held = excluded.rows_held, rows_unattributed = excluded.rows_unattributed,
    rows_rejected = excluded.rows_rejected, rows_unmatched = excluded.rows_unmatched;

-- name: DeletePerformanceImportBreakdown :exec
DELETE FROM performance_import_breakdown WHERE import_id = sqlc.arg(import_id);

-- name: ListPerformanceImportBreakdown :many
SELECT * FROM performance_import_breakdown
WHERE import_id = ANY(sqlc.arg(import_ids)::text[])
ORDER BY import_id COLLATE "C", stage COLLATE "C", web_metric COLLATE "C";

-- name: ListCurrencyCodes :many
SELECT code FROM currencies ORDER BY code;

-- name: ListFactsOfImport :many
SELECT sqlc.embed(performance_facts)
FROM performance_facts
WHERE import_id = sqlc.arg(import_id)
ORDER BY id COLLATE "C";

-- name: FactsReplacedBy :many
SELECT sqlc.embed(performance_facts)
FROM performance_facts
WHERE replaced_by_id = ANY(sqlc.arg(replacer_ids)::text[])
ORDER BY id COLLATE "C";

-- name: ActivityKeyFacts :one
SELECT a.external_name, coalesce(a.mql IS NOT NULL OR a.revenue_minor IS NOT NULL, false)::boolean AS goal_entered
FROM activities a
WHERE a.id = sqlc.arg(activity_id)
    AND activity_in_scope(a.id, sqlc.arg(sees_everything)::boolean, sqlc.arg(viewer_id)::text,
        sqlc.arg(views_plans)::boolean, sqlc.arg(viewer_team_ids)::text[]);

-- name: FactTotalsBySource :many
SELECT f.source_id, s.name AS source_name, count(*) AS fact_count, max(f.fact_date)::date AS last_fact_date
FROM performance_facts f
JOIN performance_sources s ON s.id = f.source_id
WHERE f.activity_id = sqlc.arg(activity_id)::text AND f.replaced_by_id IS NULL
GROUP BY f.source_id, s.name
ORDER BY s.name COLLATE "C", f.source_id;

-- name: ListAliasesOfActivity :many
SELECT * FROM activity_external_name_aliases
WHERE activity_id = sqlc.arg(activity_id)
ORDER BY name COLLATE "C";

-- name: ListIntegrationIDsOfActivity :many
SELECT * FROM activity_integration_ids
WHERE activity_id = sqlc.arg(activity_id)
ORDER BY source_id COLLATE "C";

-- name: DeleteIntegrationID :execrows
DELETE FROM activity_integration_ids
WHERE activity_id = sqlc.arg(activity_id) AND source_id = sqlc.arg(source_id);

-- name: DeleteAlias :execrows
DELETE FROM activity_external_name_aliases
WHERE id = sqlc.arg(id) AND activity_id = sqlc.arg(activity_id);

-- name: HeldActivityIDsOfSource :many
SELECT DISTINCT activity_id::text
FROM performance_facts
WHERE source_id = sqlc.arg(source_id) AND held AND replaced_by_id IS NULL AND activity_id IS NOT NULL
ORDER BY activity_id::text;

-- name: ActivityFactTotals :many
SELECT stage, web_metric, date_trunc('month', fact_date)::date AS month, coalesce(currency, '') AS currency,
    coalesce(sum(count_value), 0)::bigint AS counts, coalesce(sum(amount_minor), 0)::bigint AS amount_minor,
    count(*) AS fact_count
FROM performance_facts
WHERE activity_id = sqlc.arg(activity_id)::text AND NOT held AND replaced_by_id IS NULL
    AND fact_date BETWEEN make_date(sqlc.arg(year)::int, 1, 1) AND make_date(sqlc.arg(year)::int, 12, 31)
    AND activity_in_scope(activity_id, sqlc.arg(sees_everything)::boolean, sqlc.arg(viewer_id)::text,
        sqlc.arg(views_plans)::boolean, sqlc.arg(viewer_team_ids)::text[])
GROUP BY stage, web_metric, month, currency
ORDER BY stage COLLATE "C", web_metric COLLATE "C", month, currency COLLATE "C";

-- name: LatestImportStatuses :many
SELECT DISTINCT ON (source_id) source_id, status
FROM performance_imports
WHERE source_id = ANY(sqlc.arg(source_ids)::text[])
ORDER BY source_id, created_at DESC, id COLLATE "C" DESC;

-- name: MediaCodeUsed :one
SELECT EXISTS (
    SELECT 1 FROM performance_facts
    WHERE matched_on = 'mediaCode' AND matched_key = sqlc.arg(code)::text
)::boolean AS used;

-- name: GetPerformanceConfiguration :one
SELECT * FROM performance_configuration WHERE id = 'performance';

-- name: PutPerformanceConfiguration :exec
INSERT INTO performance_configuration (
    id, media_code_next, code_attributes,
    version, created_at, created_by_id, created_by_name, updated_at, updated_by_id, updated_by_name
) VALUES ('performance', $1, $2, $3, $4, $5, $6, $7, $8, $9)
ON CONFLICT (id) DO UPDATE SET
    media_code_next = excluded.media_code_next, code_attributes = excluded.code_attributes,
    version = excluded.version, updated_at = excluded.updated_at,
    updated_by_id = excluded.updated_by_id, updated_by_name = excluded.updated_by_name;

-- name: ListCodeActivityTypes :many
SELECT activity_type_id FROM performance_code_activity_types ORDER BY activity_type_id COLLATE "C";

-- name: ClearCodeActivityTypes :exec
DELETE FROM performance_code_activity_types;

-- name: AddCodeActivityType :exec
INSERT INTO performance_code_activity_types (activity_type_id) VALUES ($1) ON CONFLICT DO NOTHING;
