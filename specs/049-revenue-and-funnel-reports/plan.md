# Implementation Plan: Revenue and Funnel-Stage Reports

**Branch**: `feature/049-revenue-and-funnel-reports` · **Spec**: [spec.md](./spec.md) · **Depends**: 039
(frame, registry, export, shared ranked list), 040 (the ranked list's rules), 046 (facts), 047 (the
projection engine and activity categories). Decisions in [research.md](./research.md), figures in
[data-model.md](./data-model.md), operations in [contracts/api.md](./contracts/api.md).

## Summary

Two reports from one package: **Revenue** and **Funnel stages**. Each shows fixed summary cards, a series
of actual · projected · goal by month or quarter split at today, and 040's ranked list over the plans,
campaigns, initiatives or activities behind the figures. No migration: every figure is derived per request
from 046's facts and 047's engine. The new work is a set-based facts query and a goals port — the reports
themselves are composition.

## Technical Context

- **Backend**: `internal/measure/revenue` (both reports), `app/measure_revenue.go` and
  `app/measure_funnel.go` (the two stubs this lane owns), `queries/measure_revenue.sql`.
- **Reused as-is**: `measure.Filter`/`ParseFilter`, the registry and `measure.Report`/`Parameterised`,
  `Service.Scope` (the `Scoper` every report depends on), `Service.Export` through `Report.Sheets`,
  `projection.Project`/`SplitCounts`/`SplitAmounts`, `settings.ListActivityCategories`.
- **Copied from 040** (R5): the ranked list's rules, keeping its function names so the two can be diffed.
- **Frontend**: the two stubs `measure/revenue/RevenueReport.tsx` and `measure/funnel/FunnelReport.tsx`,
  which keep their exported names; `ReportFrame`, `RankedList`, `useRankedState`, `ReportWidget`,
  `ReportTable`, `periodStates` imported from 039's shared folder.
- **Money**: display currency, converted once at the current rate; facts without a rate excluded and
  counted (046-C7). **Clock**: injected; today is a UTC calendar date.

## Constitution Check

| Rule | How this feature keeps it |
|---|---|
| Spec first | Every section traces to an FR and to 049-C1–C20 |
| ≤300 lines, ≤5% comments, metrics | The package splits by concern as 040's does (model · ports · service · facts · figures · ranked · cards · sheets · handler) |
| Handler→service→store; Gin in handlers only | The filter is parsed in the handler; the service takes a parsed `Filter` and never sees a request |
| Server is the source of truth | Every figure is computed server-side; the web renders what it is given |
| Contract first | The wave B contract (R6), then `pnpm run api:types` |
| Scope and fail-closed routes | One scoped SQL read with an isolation test; the four route rules land with the contract |
| Tests | SC-001's hand-worked table; SC-003 pins the statement count; a test per FR; e2e for both P1 stories |
| Money/dates | Integer minor units, one conversion, `date.Date`, UTC |
| No new record type | Nothing is stored — no migration, no kind, no bin |

## Project Structure

```text
backend/internal/measure/revenue/
├── model.go          # Figures, Entry, Group, Ranked, the two reports' wire shapes
├── ports.go          # Transactor, Scoper, Goals (R3), Categories
├── service.go        # NewService, Report, Ranked — one Read per request
├── facts.go          # the set-based roll-up (R2) → per-activity month series
├── figures.go        # actual + projected + goal per entry, split at today (R4)
├── periods.go        # months, and months summed into quarters (049-C4)
├── ranked.go · ranked_query.go · ranked_groups.go   # 040's rules, our figures (R5)
├── cards.go          # the fixed card sets (049-C9)
├── sheets.go         # export widgets: cards · periods · ranked
├── report.go         # two measure.Report registrations, Params() for own parameters
└── handler.go        # GET /reports/{revenue,funnel} and …/ranked
backend/internal/app/measure_revenue.go · measure_funnel.go   # the stubs this lane replaces
backend/internal/platform/database/queries/measure_revenue.sql
frontend/src/features/measure/revenue/   · funnel/            # the two stubs' bodies
frontend/e2e/measure-revenue.spec.ts
```

## Rules, and the tests that prove them

| # | Rule | Proof |
|---|---|---|
| R1 | Actual is 046's counted facts by fact date; held, unmatched and replaced count nowhere (FR-021) | facts test over a seeded fixture |
| R2 | Projected runs per activity and sums; never from summed goals (FR-022, 049-C3) | figures test with two activities of different start dates |
| R3 | The series splits at today; the period in progress carries both and is counted once (FR-004) | periods test under a fixed clock |
| R4 | Quarters are the sum of their months, with the same in-progress rule (FR-003, 049-C4) | periods test |
| R5 | Performance = (actual + projected) ÷ goal, against the activities' goals (FR-023, 049-C16) | figures test |
| R6 | Conversion is a stage's actual over the previous stage's actual; "—" on a zero denominator (FR-017) | funnel test |
| R7 | ROI and actual cost are "—" with the attribution reason until 050 (FR-013, 049-C6) | report test + contract test |
| R8 | The stage selector is the union of the models' allowed stages; uncovered plans show "—" and are counted (FR-015, FR-018) | funnel test with two models |
| R9 | The category parameter narrows every part; *Uncategorised* loses no activity (FR-019, FR-020) | report test with two categories |
| R10 | The ranked list keeps 040's rules: two levels, drill with trail, groups with totals, paging across groups (FR-005) | ranked test, mirroring 040's |
| R11 | An empty report states which cause applies (FR-008, 049-C20) | report test per cause |
| R12 | Every read is scoped in SQL (FR-002) | isolation test on a real database |

## Queries, and their scope filter

One statement, `ReportFactTotals` in `queries/measure_revenue.sql`: the activities in view as an array,
filtered by `activity_in_scope(...)` and the plan year, grouped by activity, stage, web metric, month and
currency. Goals and models come through the `Goals` port, which reads what the plan and settings services
already expose — no second scoped read (R3).

## Permissions

No new permission. Both reports are read through the frame's rule, with the scope applied in the service,
exactly as every other report is.

## Contract

The **wave B contract** (R6): four operations in `openapi.yaml` — `/reports/revenue`,
`/reports/revenue/ranked`, `/reports/funnel`, `/reports/funnel/ranked` — their schemas, and the four route
rules in `routes_measure.go`, followed by `pnpm run api:types`. This is the one step at which this lane
touches those files, and the build plan assigns it to lane B.

## Risks reported, not worked around

- **Two copies of the ranking rules** (R5) can drift from 040's. Function names are kept identical so the
  two can be diffed, and SC-002 pins the behaviour that matters.
- **Two shared-file asks** for 039's lane (049-C19): the export helper's hardcoded sort keys, and Expand
  all / Collapse all. Until they land, a re-sorted ranked list exports in the default order; the spec and
  the PR say so.
- **Performance and conversion are shown here but defined by 048** (049-C13, 049-C15); if 048 words them
  differently, this feature follows it rather than the other way round.

## Complexity Tracking

Nothing outside the constitution's acceptable bands is planned. The ranked list's `compare` and `grouped`
are the complexity hot spots in 040's version; if either needs a justified exception it goes in the PR.
