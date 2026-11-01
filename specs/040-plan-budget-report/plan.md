# Implementation Plan: Plan Budget Report — Spend Pacing and the Ranked List

**Branch**: `feature/040-plan-budget-report` | **Date**: 2026-11-01 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `specs/040-plan-budget-report/spec.md`

## Summary

The Plan budget report adds three widgets to 039's frame: summary cards, a pacing chart (budget, planned cost and
actual by month or quarter, split at today) and a ranked list with drill to line items, group by, sort and paging.
Every figure is worked out on the server from 039's scope step (`measure.Service.Scope`: the year's records in the
reports.view scope, matched by the filters, with the resolved currency) and the payments by posting month that 036
already reads (`financials.SpendByMonth`). Nothing is stored; the export reuses 039's shared export through
`Report.Sheets`.

## Technical Context

**Language/Version**: Go (pinned in `go.mod`); TypeScript 5.9 / React 19
**Primary Dependencies**: Gin (handler only), existing `allwork`, `measure`, `financials`, `platform/money`,
`platform/sheet`; web: TanStack Query, existing `ReportFrame`, `ReportWidget`, `ReportTable`
**Storage**: none new — reads through existing stores; no migration, no SQL file
**Testing**: Go `testing` on fakes and `apitest` contract tests; Vitest + Testing Library + axe; Playwright e2e and
`apibench` run in one batch at the end (laptop limits)
**Target Platform**: local web app on 127.0.0.1
**Project Type**: web application (backend + frontend)
**Performance Goals**: report and ranked list within 2 s on the 028 volume (SC-002); one `Scope` read and one
`SpendByMonth` read per request
**Measured** (2026-11-02, `BenchmarkReportsFrame`, 028 volume, 3 runs each, with `fix/039-custom-plans`): report
1.44 s, ranked list by activity grouped by owner 1.41 s, 39 statements each (SC-002 met; 9.1 s and 8.5 s without it)
**Constraints**: figures defined once (measure-metrics §1); money exact decimals; files ≤ 300 lines; comments ≤ 5%
**Scale/Scope**: 2 endpoints, 1 package, 1 page with 3 widgets

## Constitution Check

| Principle | Check | Result |
|---|---|---|
| I Spec first | Spec clarified (040-C1–C3); FR IDs in test names | Pass |
| II Shell language | Budget, Planned cost, Committed, Actual, Left to plan, Over-allocated, Show, Group by | Pass |
| III Documented design | plan, research R1–R7, data model, contracts, quickstart | Pass |
| IV Server is the source of truth | Cards, periods, ranking, groups, totals and paging from the API; the web formats | Pass |
| V Simplicity | Reuses Scope, PlannedByActivity, SpendByMonth and the shared export; no new store or dependency | Pass |
| VI Meaningful tests | Unit per rule (budget once per branch, periods, today split, ranking, groups, totals), contract per operation, component per widget state | Pass |
| VII Security | Scope only from 039's Scope; finance figures only for plans the person may see; route rules fail closed | Pass |
| VIII Clean construction | One concept per file; comment budget | Pass |
| IX Experience | No primary button; Over-allocated in words; every chart with a table; drill with Back | Pass |
| X Palette | Chart tokens of D-345 (`--chart-budget`, `--chart-planned`, `--chart-actual`) | Pass |

## Project Structure

### Documentation (this feature)

```text
specs/040-plan-budget-report/
├── spec.md · plan.md · research.md · data-model.md · quickstart.md
├── contracts/api.md · contracts/ui.md
└── tasks.md
```

### Source Code (repository root)

```text
backend/internal/measure/planbudget/
├── model.go           # Report, Cards, Periods, Ranked, Entry, Group
├── ports.go           # Scoper (measure.Service), Spending (financials)
├── figures.go         # per-record figures over the records in view; budget once per branch
├── cards.go · periods.go
├── ranked.go · ranked_groups.go · ranked_query.go
├── service.go · handler.go · sheets.go (measure.Report for the export)
└── *_test.go
backend/internal/app/measure_planbudget.go     # registers the report and mounts its handler
backend/api/openapi.yaml                       # the two operations (wave A contract, 040's part)
frontend/src/features/measure/plan-budget/
├── PlanBudgetReport.tsx · api.ts
├── BudgetCards.tsx · PacingChart.tsx
└── *.test.tsx
frontend/src/features/measure/shared/RankedList.tsx (+ test)   # exported for 049
```

## Business rules and the tests that prove them

| Rule | Where | Test |
|---|---|---|
| Budget once per branch (FR-003, 040-C2) | `figures.go` | `TestBudget_OncePerBranch_040_FR003` |
| Cards as measure-metrics §1; "—" on a zero budget (FR-002) | `cards.go` | `TestCards_FR002` |
| Finance figures only for plans the person may see (FR-006) | `figures.go` | `TestCards_FinanceOnlyVisiblePlans_FR006` |
| Periods: phased budget, spread planned, actual by posting month, quarters, running totals (FR-004, FR-007) | `periods.go` | `TestPeriods_FR004_FR007` |
| Actual stops at today's period; notes for unphased, No dates, other years (FR-004, FR-005) | `periods.go` | `TestPeriods_TodayAndNotes_FR004_FR005` |
| Ranking, children one level down, line items under activities (FR-008, FR-009, 040-C1) | `ranked.go` | `TestRanked_LevelsAndSort_FR008_FR009` |
| Drill, trail, fallback when not in view (FR-010) | `ranked.go` | `TestRanked_Drill_FR010` |
| Over-allocated (FR-012) | `ranked.go` | in `TestRanked_LevelsAndSort_FR008_FR009` |
| Group by plan, owner, attribute with shares and Not filled in (FR-013) | `ranked_groups.go` | `TestRanked_Groups_FR013` |
| Paging, size, search (FR-014) | `ranked.go` | `TestRanked_PagesAndSearch_FR014` |
| Totals equal the cards (FR-015, SC-003) | `ranked.go` | `TestRanked_TotalsEqualCards_FR015` |
| Parameters parsed, allowlisted (FR-014) | `ranked_query.go` | `TestParseRanked_FR014` |
| Export sheets (FR-016) | `sheets.go` | `TestSheets_FR016` |
| Contract of both operations | `handler.go` | `TestPlanBudgetContract` |
| Widgets' states and drill on screen | web | `PlanBudgetReport.test.tsx`, `RankedList.test.tsx` |

## Complexity Tracking

None: no principle is broken.
