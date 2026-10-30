# Data Model: Measure Foundation

039 adds **one table** and **one permission**. Everything else a report frame shows is worked out on read from
the records the Plan module already stores (D-041).

## 1. Saved report view — table `report_views` (migration 00028)

A person's named report filter for one report (spec FR-024–FR-029; research R4, R5).

| Column | Type | Rule |
|---|---|---|
| `id` | text | Primary key, `^RV-[0-9]+$`, from `ids.Generator` with the new prefix `PrefixReportView = "RV"` |
| `person_id` | text | Not null, references `people (id)`, deferrable (as `saved_views`) |
| `report` | text | Not null, one of the report keys of §3 (check constraint) |
| `name` | text | Not null, trimmed, 1–60 characters (check `name <> ''` and `char_length(name) <= 60`) |
| `is_default` | boolean | Not null |
| `query` | text | Not null, the canonical report-filter query string, at most 4,000 characters (check) |
| `created_at` | timestamptz | Not null |
| `updated_at` | timestamptz | Not null |

Indexes: `report_views_name_key` unique on `(person_id, report, lower(name COLLATE "C"))`;
`report_views_default_key` unique on `(person_id, report) WHERE is_default`; `report_views_person_key` on
`(person_id, report)` for the list.

**Rules** (service, each with its message as 010's views):

- Name trimmed, required ("Enter a name for the view."), at most 60 characters ("Use 60 characters or fewer."),
  unique for the person and the report ignoring case ("You already have a view called '%s'."). The web app asks
  before replacing (it then updates the existing view); the API refuses a duplicate with a field problem on
  `name`, all problems at once (`errs.Problems`, 400), as 010's views do.
- The query is parsed with the report filter's parser; a malformed query is refused ("The view can't be saved —
  reload the page and try again."); it is stored canonicalised (§2).
- Starring a view unstars the person's other view of the same report in the same write.
- Only the owner lists, reads, changes or deletes a view; another person's ID answers 404 like a missing one.
- No version, no activity-log entry, no recycle bin: a view is a personal preference like 010's saved views
  (research R5, plan Complexity Tracking). Delete is immediate; **Undo** re-creates it.

## 2. Report filter (not stored, except as a saved view's `query`)

The state a report is shown with (spec Key Entities), carried in the page address.

| Parameter | Meaning | Default |
|---|---|---|
| `year` | Plan year | current year |
| `plan` (several) | Plans | all plans of the year in scope |
| `planType` (several) | `operational`, `strategic`, `crossMarket` | all |
| `team` (several) | Teams | all |
| `owner` (several) | Owners, tested on every level | all |
| `activityType` (several) | Activity types | all |
| `activityStatus` (several) | `created`, `submitted`, `declined`, `approved`, `inProgress`, `complete` | all |
| `attr` (several) | `SA-…:value,value` — values are option IDs, `yes`, `no`, `none` (Not filled in) | none (no restriction) |
| `period` | `month`, `quarter` | `month` |
| `show` | `plan`, `campaign`, `initiative`, `activity` | per report |
| `currency` | ISO code of an active currency | the person's display currency |
| `table` (several) | widget keys shown as a table | none (charts) |

The filter parameters are All work's (research R1) and parsed by `allwork.ParseQuery`; `period`, `show`,
`currency` and `table` by `measure`. **Canonical form** (saved views, exports): parameters in the order above,
values sorted, defaults omitted.

## 3. Report catalogue (code, not stored)

| Order | Key | Path | Name | Wave | Extra rule |
|---|---|---|---|---|---|
| 1 | `planBudget` | `plan-budget` | Plan budget | A (040) | — |
| 2 | `approvals` | `approvals` | Approvals | A (041) | — |
| 3 | `purchaseRequests` | `purchase-requests` | Purchase requests | A (042) | — |
| 4 | `planDetails` | `plan-details` | Plan details | A (043) | — |
| 5 | `gsapLink` | `gsap-link` | GSAP link | A (044) | money columns need `finance.view` |
| 6 | `userAccess` | `user-access` | User access | A (045) | `access.manage` |
| 7 | `summary` | `summary` | Summary | B (048) | — |
| 8 | `revenue` | `revenue` | Revenue | B (049) | — |
| 9 | `funnel` | `funnel` | Funnel stages | B (049) | — |

Each entry has a one-line purpose (contracts/ui.md) and `available` from the registry (research R6).

## 4. Permission

`reports.view` — "See reports", group Plans and planning, scoped, refusal noun "Seeing reports" (014 Table A).
Migration 00028: `roles_permissions_check` gains `'reports.view'`; `UPDATE roles SET permissions =
array_append(permissions, 'reports.view'), version = version + 1, updated_at = now(), updated_by_id = 'system',
updated_by_name = 'SP&A One' WHERE 'plans.view' = ANY (permissions) AND NOT 'reports.view' = ANY (permissions)`.
The seed gives it to the built-in roles that hold `plans.view` (`seed/access.go`).

## 5. In view (worked out on read)

| Count | Rule |
|---|---|
| Plans, Campaigns, Initiatives, Activities | Records shown and matching by All work's engine (`allwork.Match`, research R1) |
| Line items | Live line items of the matching activities |
| Finance coverage | Matching plans whose financial data the person may see (`Choices.FinancePlans`), of all matching plans |
| Left out | How many filter values named something unknown, inactive or out of scope — counts only, never names |
