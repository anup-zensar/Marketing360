# Data model: Projections

Migration **00033** (reserved for 047). The migrations are the truth; this file describes them. The
projection itself is derived at read time and stores nothing (047-C8).

## models — three new columns (extends feature 002's table, 00006)

| Column | Type | Rules |
|---|---|---|
| `inq_to_mql_velocity_months` | integer | NOT NULL, CHECK 0–24, backfilled 0 (same-month conversion) |
| `mql_to_deal_velocity_months` | integer | NOT NULL, CHECK 0–24, backfilled 0 |
| `allowed_goal_stages` | text[] | NOT NULL, CHECK non-empty and each element IN ('inq','mql','revenue'), backfilled `{inq,mql,revenue}` |

No existing column, index or constraint changes; the model's record rules (version, audit columns,
case-insensitive unique name) stand as feature 002 built them.

## activity_categories — new (047-C7)

| Column | Type | Rules |
|---|---|---|
| `id` | text | PK, CHECK `^ACAT-[0-9]+$` |
| `name` | text | NOT NULL, 1–80; unique ignoring case: `activity_categories_name_key` on `lower(name COLLATE "C")` |
| `is_default` | boolean | NOT NULL; partial unique index `activity_categories_default_key` ON (is_default) WHERE is_default — at most one default |
| `active` | boolean | NOT NULL |
| `version` + created/updated at/by_id/by_name | | the standard record columns (record.Meta) |

Deletes go through the recycle bin (history.MoveToBin with a registered kind), never a plain DELETE.

## activity_category_types — new (047-C7)

| Column | Type | Rules |
|---|---|---|
| `activity_type_id` | text | **PK** — an activity type belongs to at most one category, by key |
| `category_id` | text | NOT NULL REFERENCES activity_categories (id) ON DELETE CASCADE |

The CASCADE serves the bin flow: binning a category snapshots its assignments into the bin record, the
delete releases the types, and restore re-inserts the snapshot (same pattern as 046's bin children). A
type never named here is *Uncategorised* — a report bucket, not a row.

## Projection (derived, not stored)

Inputs: the activity's goal facts (goal stage values, SetLaterStagesMyself, plan year — 046's GoalFacts,
extended with StartDate/EndDate/Ongoing), the model's rates, deal size, velocities, and today from the
injected clock. Output per stage: twelve months ("2026-01" …), each month actual-only, projected-only or
both (the month containing today), plus projected revenue and the model stamp (name, parameters'
updatedAt). No identity, no audit entry, no table.
