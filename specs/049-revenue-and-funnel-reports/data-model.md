# Data model: Revenue and Funnel-Stage Reports

**No migration.** This feature stores nothing: every figure is worked out per request from 046's facts,
047's engine and the goals the `Goals` port supplies. What follows describes the one new query and the
shapes the service builds in memory.

## The one new query — `queries/measure_revenue.sql`

`ReportFactTotals`: the counted facts of the activities in view, for one plan year.

| Input | Meaning |
|---|---|
| `activity_ids text[]` | the activities in view, from the frame's scope |
| `year integer` | the plan year; a fact counts in the year of its own date (046 FR-015) |
| the four `activity_in_scope` arguments | the actor's scope, applied in SQL exactly as 046 applies it |

Returns one row per `(activity_id, stage, web_metric, month, currency)` with `counts`, `amount_minor` and
`fact_count`. Held, unmatched and replaced facts are excluded in the `WHERE`, as 046 does — the report
never re-decides what counts.

Per-activity rows, not pre-rolled: an entry at any level is the sum of its activities', so one read serves
the cards, the series and every level the ranked list shows, including a drill to a leaf.

## In-memory shapes (nothing persisted)

- **Figures** — per entry: goal, actual, projected, performance, and for the funnel the conversion from
  the previous stage. Money carries its currency; counts are whole numbers.
- **Entry · Group · Ranked** — the ranked list's wire shapes, mirroring 040's with this feature's figures
  (research R5): an entry's id, level, name and plan, its figures and its children; a group's key, label
  and totals; the page, total, drill trail and `drillLost` flag.
- **Period series** — one point per month or quarter: actual, projected and goal, with the period in
  progress carrying both actual and projected (047 FR-010, 049-C4).
- **Stamp** — the model's name and its last change, carried beside every projected figure (047 FR-012).

## What this feature reads, and from where

| Needs | From |
|---|---|
| The activities, plans and attributes in view, and the scope | 039's `Service.Scope` |
| Counted facts per activity and month | the new query above |
| Each activity's goals, dates and its plan's model | the `Goals` port (research R3) |
| The projected series | `projection.Project` per activity, summed (research R4) |
| Activity categories | `settings.ListActivityCategories` (047) |
