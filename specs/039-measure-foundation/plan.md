# Implementation Plan: Measure Foundation — the Reports Area and the Report Frame

**Branch**: `feature/039-measure-foundation` (from `measure-module`) | **Date**: 2026-10-30 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `specs/039-measure-foundation/spec.md`

**Design artifacts**: [research.md](research.md) (R1–R10) · [data-model.md](data-model.md) ·
[contracts/api.md](contracts/api.md) · [contracts/ui.md](contracts/ui.md) · [quickstart.md](quickstart.md) ·
[measure-module-build.md](../../docs/sdd/measure-module-build.md)

## Summary

1. **A `measure` package** (Go) holds what every report shares: the report filter (All work's parameters plus
   period, show, currency, table; R1, R8), the catalogue and the **registry** that reports plug into without
   editing the hub (R6), the **frame** (choices, In view, finance coverage, left out), saved report views (R4,
   R5) and the export (R7). It reads plans through ports adapted in `internal/app` (D-196), as 036 does.
2. **One filter engine.** A new exported `allwork.Match` (new file `allwork/match.go`) runs All work's own
   matching on the year's nodes, so a report and All work with the same filters cover the same records (R1,
   SC-001). The nodes are narrowed to the `reports.view` scope (R2).
3. **Access.** A new scoped permission `reports.view`, granted by migration 00028 to every role holding
   `plans.view` and by the seed to the built-in ones (R3, D-367); `access/routes_measure.go` holds every rule of
   wave A; `app/measure.go` registers one stub per report from its own `app/measure_<report>.go` (D-369).
4. **Web app** (`features/measure`): the Reports page, the shared frame (`shared/`: ReportFrame, FilterPanel,
   FilterChips, InViewLine, controls, ReportWidget, ReportTable, SavedViewsMenu, ExportMenu, address ↔ filter with
   Zod) and a stub page per report; Reports in the sidebar after Dashboards (039-C3).

## Technical Context

**Language/Version**: Go 1.27 (pinned in `go.mod`); TypeScript 5.9 strict + React 19

**Primary Dependencies**: none new — `platform/sheet` for the files; the existing web stack (TanStack Query, Zod,
Radix, lucide-react, CSS Modules)

**Storage**: PostgreSQL; migration **00028**: table `report_views` (data-model §1), `roles_permissions_check`
widened and `reports.view` granted (data-model §4). One new scoped store read: `measure.Store.ListViews(person,
report)` — person-isolation test on PostgreSQL. Plans come from the existing scoped reads `plan.Store.ListInScope`
and All work's `Source.Year` (both already scope-tested).

**Testing**: Go unit tests of every rule on fakes (filter parsing and canonical form, `allwork.Match` against All
work's own counts, frame choices, scope narrowing, finance coverage, left out, views, export, catalogue,
registry); handler and contract tests on `apitest` for the seven operations with scope tests per persona;
Vitest + Testing Library + MSW + vitest-axe for every frame and widget state, fast-check for the address round
trip; Playwright + axe for US1–US3 (`e2e/measure.spec.ts`)

**Target Platform**: developers' machines (127.0.0.1); GitHub Actions on pull requests into `measure-module`

**Project Type**: Web application (Go API + React web app)

**Performance Goals**: SC-003 — frame within 2 s at the 028 volume; `BenchmarkReportsFrame`
(`internal/platform/apibench`); statements per frame fixed: year, plans in scope, choices, attributes, currency
(R9)

**Constraints**: constitution VII scope isolation; IV the server works out every figure; F3 state in the
address; F6 every chart has a table; X palette (chart tokens, never yellow or red for a series); V file and
function metrics; VIII comment budget; build guide §4 ownership (no lane edits the hub after 039)

**Scale/Scope**: 7 operations; 1 new Go package (~12 files), 1 new file in `allwork`, small additions to
`authz`, `ids`, `access`, `app`, `seed`; 1 new web feature (~30 files) and its stubs; 1 migration; 1 e2e spec

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Check | Result |
|---|---|---|
| I Spec first | Spec clarified (039-C1–C3); FR IDs in test names; every FR has a task | Pass |
| II Shell language | Reports, Summary cards, Not filled in, Period, Saved views, Show; no Measure/Tactic/HUD on screen (FR-032) | Pass |
| III Documented design | plan, research R1–R10 (threat model R10), data model, contracts, quickstart | Pass |
| IV Server is the source of truth | Matching, counts, coverage, choices and files from the API; the web app formats and keeps the address | Pass |
| V Simplicity | Reuses All work's engine and `platform/sheet`; no new dependency; one table | Pass |
| VI Meaningful tests | Unit per rule, contract per operation, scope per persona, one PostgreSQL person-isolation test, state tests per widget, e2e for P1 | Pass |
| VII Security | Scope from the actor; route rules fail closed; left-out values never named; views by person; allowlists; files never stored | Pass |
| VIII Clean construction | Comment budget per file; one concept per file | Pass |
| IX Experience | One title; Apply with "not applied yet"; Back keeps state; no primary button on a read-only page; Undo on delete; every empty state says why | Pass |
| X Palette | Tokens only; chart tokens of D-345 | Pass |
| B1–B9 | Gin only in `handler.go`; service without `net/http`; ctx first; contract first; tests on fakes | Pass |
| F1–F10 | Query hooks in `api.ts`; Zod for the address; no chart library; tests by role; axe | Pass |

## Project Structure

### Documentation (this feature)

```text
specs/039-measure-foundation/
├── spec.md · plan.md · research.md · data-model.md · quickstart.md
├── contracts/api.md · contracts/ui.md
├── checklists/requirements.md
└── tasks.md
```

### Source Code (repository root)

```text
backend/
├── api/openapi.yaml                       # wave A contract: 039's 7 operations + 040–045 (step 1)
├── internal/measure/                      # NEW, owned by 039
│   ├── model.go          # ReportFilter, ReportFrame, choices, catalogue entries, views (API types)
│   ├── filter.go         # ParseFilter (allwork.ParseQuery + period/show/currency/table), canonical query
│   ├── catalogue.go      # the nine reports in order (data-model §3); Report interface; Registry
│   ├── ports.go          # Transactor; Planning (year, choices, attributes, currency), Plans (reports scope)
│   ├── frame.go          # Frame: narrow to reports scope, allwork.Match, In view, finance, left out, choices
│   ├── choices.go        # filter choices from the year's nodes and Settings
│   ├── store.go · store_postgres.go       # report_views
│   ├── views.go          # list, create, update, delete (010's lifecycle, per report)
│   ├── export.go         # workbook/CSV: report sheets + In view + About this export; file names
│   ├── handler.go
│   └── *_test.go, store_fake_test.go, measuretest/
├── internal/allwork/match.go              # NEW exported Match (R1)
├── internal/platform/authz/catalogue.go   # ReportsView
├── internal/platform/ids/ids.go           # PrefixReportView "RV"
├── internal/platform/database/migrations/00028_reports.sql
├── internal/platform/database/queries/measure_views.sql
├── internal/access/routes_measure.go      # every wave A rule; routes.go gains measureRoutes()
├── internal/app/history_access.go         # reportView entries: the owner and activityLog.viewAll (R5)
├── internal/app/measure.go                # wiring hub: registry + one call per report
├── internal/app/measure_{planbudget,approvals,purchasereport,plandetails,gsaplink,useraccess}.go  # stubs
├── internal/app/contract_routes_test.go   # notYetServed: wave A groups, one comment line per feature
├── internal/platform/apibench/reports_test.go   # BenchmarkReportsFrame
└── seed/access.go · seed/seed.go          # built-in roles get reports.view; the call to seed/measure.go (D's)

frontend/
├── src/features/measure/
│   ├── ReportsPage.tsx · index.ts         # catalogue; exports every report page
│   ├── shared/                            # contracts/ui.md "Shared parts"
│   └── {plan-budget,approvals,purchase-requests,plan-details,gsap-link,user-access,summary,revenue,funnel}/  # stubs
├── src/app/{navigation.ts,pages.tsx,router.tsx}   # Reports item and every /reports route
└── e2e/measure.spec.ts
```

**Structure Decision**: web application; one business package and one feature folder, laid out as constitution B2
and F2; the hub files are written once here so the other lanes never edit them (build guide §4).

## Business rules and the tests that prove them

| Rule | Where | Test |
|---|---|---|
| Filter parsing, defaults, allowlists, canonical form (FR-008, FR-015, FR-030) | `measure/filter.go` | `TestParseFilter_FR008_FR030`, `TestCanonicalQuery_FR024` |
| Same records as All work for the same filters (FR-007, FR-012, SC-001) | `allwork/match.go` | `TestMatch_EqualsAllWorkCounts_039_FR012` (every filter kind, % split, business-unit condition, line-item attribute, Not filled in) |
| In view counts; line items of matching activities (FR-007) | `measure/frame.go` | `TestFrame_InView_FR007` |
| Only plans in the reports scope; nothing out of scope named (FR-003, SC-002) | `measure/frame.go` + adapter | `TestFrame_OnlyReportsScope_FR003`, contract `TestReportsFrame_ScopePerPersona_SC002` |
| Finance coverage "Covers N of M" (FR-004) | `measure/frame.go` | `TestFrame_FinanceCoverage_FR004` |
| Left out counted, never named (FR-015) | `measure/frame.go` | `TestFrame_LeftOutCountedNotNamed_FR015` |
| Choices: year list, plans of the year, owners and types of the selected plans, attribute groups per level, inactive values in use, no text attributes (FR-008–FR-011) | `measure/choices.go` | `TestChoices_FR008_FR011` |
| Catalogue order, availability, extra rules (FR-001, FR-003) | `measure/catalogue.go` | `TestCatalogue_NineInOrder_FR001`, `TestCatalogue_UserAccessWithAccessManage_FR003`, `TestCatalogue_RegisteredReportAvailable_FR001` |
| `reports.view` in Table A; built-in roles that view plans hold it (FR-002) | `authz`, `seed/access.go` | `TestCatalogue_ReportsView_039_FR002`, `TestSeedRoles_ReportsView_039_FR002` |
| Saved views: name rules, per report, default per report, update, delete, owner only (FR-024–FR-029) | `measure/views.go` | `TestViews_FR024_FR029` |
| Views versioned (409 on a stale version) and every change logged; entries seen by the owner and `activityLog.viewAll` (FR-028, R5) | `measure/views.go`, `app/history_access.go` | `TestViews_VersionAndLog_FR028`, `TestHistoryAccess_ReportView_039_FR028` |
| Today marker: actual, in progress, planned per period of the plan year (FR-019) | `shared/periods.ts` | `periods.test.ts` |
| Views limited to their person in SQL (FR-025) | `measure/store_postgres.go` | `TestPostgresStoreReportViewsPersonIsolation_039_FR025` |
| Export: sheets, About this export, CSV needs a widget, file names, values as screen (FR-022) | `measure/export.go` | `TestExport_FR022`, contract `TestReportsExport_FR022` |
| Route rules for every wave A operation, fail closed | `access/routes_measure.go` | existing route-table tests |
| Address ↔ filter round trip; malformed values dropped (FR-015, SC-005) | `shared/report-address.ts` | `report-address.test.ts` (fast-check) |
| Apply and "not applied yet"; chips; reset; controls at once (FR-013, FR-014, FR-016–FR-018) | `shared/FilterPanel`, `FilterChips` | `FilterPanel.test.tsx`, `ReportFrame.test.tsx` |
| Widget states: loading, data, table, error with Try again, "—", no financial access (FR-020–FR-023, FR-031) | `shared/ReportWidget` | `ReportWidget.test.tsx` |
| Saved views menu: modified, update, save as new, star, rename, delete with Undo (FR-024–FR-029) | `shared/SavedViewsMenu` | `SavedViewsMenu.test.tsx` |
| Reports page: coming soon, no access (FR-001) | `ReportsPage` | `ReportsPage.test.tsx` |
| US1–US3 end to end with axe (SC-004–SC-007) | — | `e2e/measure.spec.ts` |
| Frame within 2 s at the 028 volume (SC-003) | — | `BenchmarkReportsFrame` |

## Complexity Tracking

| Violation | Why needed | Simpler alternative rejected because |
|---|---|---|
| Saved report views are deleted outright with Undo instead of through `history.MoveToBin` (CLAUDE.md "every new record type"; R5). They are versioned and every change is logged, as constitution VII requires | Constitution VII sends **business records** to the recycle bin; a saved view is a personal preference that only its owner may see or restore | A recycle-bin entry for a private bookmark gives a second restore path beside Undo that nobody else may use |
| 039 changes files outside its package: `allwork/match.go` (new file), `authz`, `ids`, `access/routes.go`, `app/history_access.go` (one case for `reportView`), `seed` | One filter engine (R1), the permission (R3) and the log rule (R5) live where the rest of their kind lives | Copies inside `measure` would drift from All work, from 014's permission catalogue and from the activity log's visibility rules |
