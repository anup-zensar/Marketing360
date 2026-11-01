# Data model: Performance Data from a Simulated Feed

Three migrations are reserved for this feature: **00030–00032**. The build guide §2 names the four headline
tables (`performance_sources`, `performance_imports`, `performance_facts`, `media_codes`); the seven
supporting tables below carry the mappings, the per-measure breakdown, the review queue, the two matching
keys that live on an activity and the media-code configuration — **eleven tables in all**. They are grouped
into the reserved migrations rather than taking further numbers another lane may hold. 00030 also widens
the `roles` permissions CHECK with D-367's two keys (`performance.import`, `performance.link`).

Every table embeds `record.Meta` (version, created/updated) where its records are edited by people, registers
a `history.Kind`, and writes one activity-log entry per change inside the same transaction. An optional
text column is `NOT NULL` with `''` meaning *not set* — the schema's empty-string sentinel — so CHECKs
and partial indexes compare with `<> ''` instead of juggling NULLs.

## 00030 — sources, mappings and the permission keys

### `performance_sources`

| Column | Type | Notes |
|---|---|---|
| `id` | text, PK | |
| `name`, `description` | text | name unique |
| `owner_id` | text → `people` | |
| `category` | text | web analytics, marketing automation, CRM, lead progression |
| `status` | text | `active`, `inactive` |
| `delivery_route` | text | `upload`, `simulated` — `simulated` is what enables a pull (046-C13) |
| `cadence` | text | drives the *Stale* threshold in FR-005 |
| `integration_id` | text | the source's own id; `''` when it has none, unique `WHERE integration_id <> ''` |

Freshness (FR-005) is **derived**, never stored: last successful import, its data-through date, the cadence
and the resulting *Current · Stale · Failed · Inactive*. Storing it would go stale on its own.

### `performance_source_mappings`

| Column | Type | Notes |
|---|---|---|
| `id` | text, PK | |
| `source_id` | text → `performance_sources`, cascade | |
| `column_name` | text | the file's column heading, unique per source |
| `meaning` | text | `key`, `reference`, `factDate`, `stage`, `revenue`, `currency`, `webMetric`, `journey` |
| `key_type` | text | `activityId`, `mediaCode`, `integrationId`, `externalName`; `''` unless a `key` |
| `stage` / `web_metric` | text | `''` unless meaning matches; sessions, visits, engagement touches (046-C11) |
| `model_id` | text, nullable → `models` | set exactly when meaning is `stage` or `revenue` |

CHECKs tie each shaping column to its meaning, `key_type` included (046-C23), and a partial unique index on
(`source_id`, `key_type`) `WHERE meaning = 'key'` allows at most one key column per type per source.

**Not a constraint**: FR-003 (at most one *active* source per model stage, and per model revenue) spans
this table and `performance_sources.status`, and an index predicate may not contain a subquery. A plain
unique index would wrongly refuse a mapping whose only rival is inactive, which FR-004 allows. The rule is
enforced in the service and recorded in **ADR-0021**; 00030 carries two plain non-unique indexes that only
make that check a single indexed read.

## 00031 — imports, facts and the review queue

### `performance_imports`

| Column | Type | Notes |
|---|---|---|
| `id` | text, PK | |
| `source_id` | text → `performance_sources` | |
| `file_name` | text | the uploader and the time ride on `record.Meta` (`created_by`, `created_at`) |
| `origin` | text | `upload` or `simulated` — both travel the same path |
| `status` | text | `completed`, `failed`, `undone` (046-C17) |
| `data_through` | date, nullable | the newest fact date a completed load delivered |
| `rows_read` and the five counts | integer | matched, held, unattributed, rejected, unmatched |

**Constraint**: `rows_read = matched + held + unattributed + rejected + unmatched` as a CHECK — FR-022's
reconciliation is an invariant (constitution B5), applied to **completed** loads only (046-C17): a load
that fell over records how far it got. A file refused before any row is read — personal data, an unknown
or inactive source, an unreadable file — is not a load and leaves no import record at all (046-C17).

### `performance_import_breakdown`

The same five counts per measure (046-C18, 046-C21). A row carrying two measures counts in both groups,
so the groups deliberately do not sum to the import's totals.

| Column | Type | Notes |
|---|---|---|
| `import_id` | text → `performance_imports` | PK together with the measure columns |
| `stage` / `web_metric` | text | exactly one is non-empty — the measure the counts describe |
| `rows_read` and the five counts | integer | the reconciliation CHECK holds on every row here |

### `performance_facts`

| Column | Type | Notes |
|---|---|---|
| `id` | text, PK | |
| `source_id`, `import_id` | text → above | |
| `source_reference` | text | the source's own row id |
| `activity_id` | text → `activities`, nullable | null only while unattributed |
| `matched_on` | text | `activityId`, `mediaCode`, `integrationId`, `externalName`; `''` while unattributed |
| `matched_key` | text | the value that matched; `''` exactly when `matched_on` is `''` |
| `fact_date` | date | the fiscal year comes from this, never the activity (FR-015) |
| `stage` / `web_metric` | text | exactly one of the two is non-empty |
| `count_value` | bigint, nullable | non-negative, whole — a fraction is rejected at load |
| `amount_minor`, `currency` | bigint / text | integer minor units, never a float |
| `held` | boolean | released on approval (research R4); indexed, every figure filters on it |
| `replaced_by_id` | text, nullable | the audit trail an undo walks back |

**Constraint**: a partial unique index on (`source_id`, `source_reference`, `stage`, `web_metric`)
`WHERE replaced_by_id IS NULL` — the measure joins the identity (046-C21, correcting research R3's pair)
because one file row holding an INQ column and an MQL column produces one fact for each, under the same
source reference.

### `performance_review_rows`

| Column | Type | Notes |
|---|---|---|
| `id` | text, PK | |
| `import_id`, `source_id` | text → above | |
| `raw` | json | the row exactly as read — `json`, not `jsonb`, so the text survives byte for byte |
| `parsed` | json | the values after mapping, so a correction can be compared with the original |
| `reason` | text | free text, 1–200 characters |
| `faults` | json | array naming each fault in the row |
| `candidates` | json, nullable | the activities named when keys disagree (FR-019) |
| `status` | text | `open`, `resolved`, `dismissed` |
| `comment` | text | `''` unless dismissed — a CHECK requires it exactly when `dismissed` (FR-024) |

## 00032 — media codes, the two activity-level keys and the configuration

### `media_codes`

| Column | Type | Notes |
|---|---|---|
| `id` | text, PK | |
| `code` | bigint, unique | system-generated, never reused, from a configurable start (FR-042) |
| `activity_id` | text → `activities` | never a line item or any other level (FR-040) |
| `attributes` | json | the configured catalogue's values |
| `landing_page` | text | `''` when not set |
| `status` | text | `active`, `inactive` — deactivated rather than deleted once used |

### `performance_configuration`

One row (`id = 'performance'`), like the other configuration tables: `media_code_next` (bigint) and
`code_attributes` (json array) carry the generator's next number and the attribute catalogue shown in
Settings › Performance data configuration (046-C12).

### `performance_code_activity_types`

Which activity types require media codes (FR-038, FR-040; 046-C12): one `activity_type_id` per row,
→ `activity_types` and the primary key. Media-code configuration lives with this feature, not on the
activity type, because activity types belong to feature 002, which this lane does not edit.

### `activity_external_name_aliases`

The activity's **current** external name stays `activities.external_name`, which feature 006 owns and a
person edits. This feature stores only the superseded names, so rows already in flight still match after a
rename (046-C14). It never writes the activity's own column.

| Column | Type | Notes |
|---|---|---|
| `id` | text, PK | the row carries `record.Meta` like any edited record |
| `activity_id` | text → `activities` | |
| `name` | text | unique across aliases, and checked against current external names before it is kept (FR-035) |

### `activity_integration_ids`

| Column | Type | Notes |
|---|---|---|
| `id` | text, PK | |
| `activity_id`, `source_id` | text → above | **unique together** — at most one per source (FR-037) |
| `value` | text | unique within the source |

## Matching keys, in resolution order

| Order | Key | Lives in |
|---|---|---|
| 1 | Activity ID | `activities.id` |
| 2 | Media code | `media_codes.code` |
| 3 | Integration ID | `activity_integration_ids.value`, scoped to the source |
| 4 | External name | `activities.external_name` (feature 006, read only) **or** `activity_external_name_aliases.name` |

All four are resolved set-wise per batch (research R2). A key resolving to more than one activity is
*ambiguous*; two keys resolving to different activities *disagree*; neither is ever guessed (FR-019).
