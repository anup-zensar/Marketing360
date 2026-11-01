-- name: ReportFactTotals :many
-- The counted facts of the activities in view, for one plan year (049-C1). One row per activity, stage,
-- web metric, month and currency, so a caller can roll up to any level: an entry's total is the sum of
-- its activities'. Held, unmatched and replaced facts are excluded, exactly as 046's own totals are.
SELECT activity_id::text AS activity_id, stage, web_metric,
    date_trunc('month', fact_date)::date AS month, coalesce(currency, '') AS currency,
    coalesce(sum(count_value), 0)::bigint AS counts, coalesce(sum(amount_minor), 0)::bigint AS amount_minor,
    count(*) AS fact_count
FROM performance_facts
WHERE activity_id = ANY(sqlc.arg(activity_ids)::text[]) AND NOT held AND replaced_by_id IS NULL
    AND fact_date BETWEEN make_date(sqlc.arg(year)::int, 1, 1) AND make_date(sqlc.arg(year)::int, 12, 31)
    AND activity_in_scope(activity_id, sqlc.arg(sees_everything)::boolean, sqlc.arg(viewer_id)::text,
        sqlc.arg(views_plans)::boolean, sqlc.arg(viewer_team_ids)::text[])
GROUP BY activity_id, stage, web_metric, month, currency
ORDER BY activity_id COLLATE "C", stage COLLATE "C", web_metric COLLATE "C", month, currency COLLATE "C";
