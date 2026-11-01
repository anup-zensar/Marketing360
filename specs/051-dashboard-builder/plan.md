# Implementation Plan: Dashboard Builder

**Branch**: `feature/051-dashboard-builder` (from `develop`) | **Date**: 2026-11-01 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `specs/051-dashboard-builder/spec.md`

**Design artifacts**: [research.md](research.md) (R1–R16) · [data-model.md](data-model.md) ·
[contracts/api.md](contracts/api.md) · [contracts/ui.md](contracts/ui.md) · [quickstart.md](quickstart.md) ·
[ADR-0020](../../docs/architecture/adr/0020-apache-echarts-for-the-dashboard-builder.md) · [tasks.md](tasks.md) (six
lanes)

## Summary

1. **`internal/customdashboard`** (Go, new): the custom-dashboard record — name, description, owner, sharing,
   default filters and widgets (layout, type, query, display) — with validation of every widget rule (data model
   §2–§3), version checks, the activity log, delete to the recycle bin, Make a copy and sharing with the owner's
   teams. Store on PostgreSQL (migration 00028, sqlc), visible reads limited in SQL to the owner and the members of
   shared teams.
2. **`internal/dataset`** (Go, new): datasets uploaded as base64 CSV or .xlsx (`internal/platform/sheet` readers),
   checked first (`/datasets/preview`), typed columns proposed and corrected, every cell checked on save, stored as
   relational columns and JSON row arrays (migration 00029), shared like dashboards; it implements the
   `widgetdata.DatasetTables` port.
3. **`internal/widgetdata`** (Go, new, no store): the data catalogue (ten built-in sources and the visible datasets)
   and the query engine — filters, date buckets, groups, aggregations in exact decimals and cents, shares, totals,
   sort, top N and Other, key-figure comparison — over rows read through ports: built-in sources implemented in
   `internal/app` adapters over All work's scoped year data, financials, approvals, purchase requests and global
   budgets (as 036 does); datasets through `dataset.Service`.
4. **Web app** `features/custom-dashboards` (new): the Dashboards list (built-ins kept), the viewer with filters in
   the address, the builder (palette, hand-built 12-column grid with pointer and keyboard, inspector Data / Style /
   Layout, undo / redo, validation, unsaved-changes guard, version conflict), datasets and upload, sharing — charts
   drawn by Apache ECharts 6.1.0 with the SVG renderer (ADR-0020), each with a table view and PNG download.
5. **Access**: `dashboards.view` and `dashboards.build` in 014 Table A (D-374), route rules for 17 operations, two
   history kinds (`customDashboard`, `dataset`) whose log and bin entries the owner sees, ID prefixes `DB` and `DS`.

## Technical Context

**Language/Version**: Go (pinned in `go.mod`); TypeScript 5.9 strict + React 19

**Primary Dependencies**: new — `echarts@6.1.0` (Apache-2.0; with `zrender@6.1.0` BSD-3-Clause, `tslib@2.3.0` 0BSD;
published 2026-05-19; no install scripts; ADR-0020). Go: none new (standard library, pgx, sqlc, goose, Gin as
today; `internal/platform/sheet` readers reused)

**Storage**: PostgreSQL; migrations `00028_custom_dashboards.sql` (3 tables + the two permissions on the stored
built-in roles) and `00029_datasets.sql` (4 tables); sqlc queries `custom_dashboards.sql`, `datasets.sql`; every
visible read has a PostgreSQL scope-isolation test. Built-in sources add no table: they read through existing scoped
stores (`PlanStore.ListInScope` via All work, `financials.Store.SpendOnLineItems`, approval rounds of subjects,
purchase requests `OnLineItems`, global budgets as 036) — any new store read gets its scope-isolation test

**Testing**: Go unit tests of every rule (widget validation, layout collisions, sharing, copy naming, type inference,
cell parsing, every aggregation, bucket, filter, sort, limit, share, comparison, overflow); contract tests for the 17
operations on `apitest`; scope-isolation tests (dashboards, datasets, each new scoped read); Vitest + Testing Library
+ MSW + vitest-axe + fast-check for `buildOption` per type, grid moves (property tests: no overlap, inside grid),
the reducer (undo/redo), every state of every view; Playwright + axe for the two P1 stories
(`e2e/dashboard-builder.spec.ts`, `e2e/custom-dashboards.spec.ts`), with the existing CSP-violation guard

**Target Platform**: developers' machines (127.0.0.1); GitHub Actions

**Project Type**: Web application (Go API + React web app)

**Performance Goals**: SC-004 — 12 widgets within 3 s and a preview within 1 s at the 028 reference volume; the
year data read once per request (memo), widget queries in parallel from the browser, caps of R13; SC-005 — 10,000 ×
50 file checked within 5 s and saved within 10 (one `CopyFrom`). Benchmarks `BenchmarkWidgetData` and
`BenchmarkDatasetUpload` in `internal/platform/apibench`

**Constraints**: constitution IV (the web app aggregates nothing), VII (scope from the actor; 404 outside), X
(palettes from tokens), F6 (table view), F7 / ADR-0018 (no inline style attributes or HTML strings), F8 (ECharts
only in the lazy feature chunk), V (300 non-blank lines, function metrics), VIII (comment budget)

**Scale/Scope**: 17 operations; 3 Go business packages (~45 files), app adapters (~6 files); 1 web-app feature
(~60 files); 2 migrations; 2 e2e specs

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Check | Result |
|---|---|---|
| I Spec first | Spec with clarifications (D-371–D-376); FR IDs in test names; every FR has a task | Pass |
| II Shell language | Plan, Campaign, Initiative, Activity, Line item in source and field names; SP&A One on screen; new terms in the glossary | Pass |
| III Documented design | plan, research (R1–R16 with threat model R15), data model, contracts, quickstart, ADR-0020 | Pass |
| IV Server is the source of truth | Every group, sum, average, share and comparison by the API; the web app formats and places marks only | Pass |
| V Simplicity | One query engine for built-in sources and datasets; no grid or wrapper library; small ports; files split by responsibility | Pass |
| VI Meaningful tests | Unit tests per rule, contract test per operation, scope-isolation per scoped read, view-state tests, e2e per P1 story; ECharts itself not tested | Pass |
| VII Security | Scope from the actor; fail-closed rules; 404 for private records; allowlisted fields; plain text; upload limits; CSP | Pass |
| VIII Clean construction | Comment budget per touched file; no requirement IDs in comments | Pass |
| IX Experience | One yellow Save; Enter submits, Escape closes; Undo; unsaved-changes guard; all errors at once; consequences before deleting; keyboard grid | Pass |
| X Palette | Four palettes of tokens, never yellow or red; decals; text and lines from palette tokens | Pass |
| B1–B9 | Gin only in handlers; ctx first; actor to every service call; sqlc + goose; fakes for service tests | Pass |
| F1–F10 | ECharts by ADR; TanStack Query; React Hook Form + Zod for dialogs and the inspector forms; tests by role | Pass |

## Project Structure

### Documentation (this feature)

```text
specs/051-dashboard-builder/
├── spec.md · plan.md · research.md · data-model.md · quickstart.md · tasks.md
├── contracts/api.md · contracts/ui.md
└── checklists/requirements.md
```

### Source Code (repository root)

```text
backend/
├── api/openapi.yaml                                  # 17 paths, CustomDashboard*/Widget*/Dataset*/DataSource* (done)
├── internal/platform/database/migrations/00028_custom_dashboards.sql, 00029_datasets.sql
├── internal/platform/database/queries/custom_dashboards.sql, datasets.sql
├── internal/customdashboard/                         # B1
│   ├── model.go       # Dashboard, Widget, Layout, Query, Display, Filters, Summary, inputs
│   ├── validate.go    # name, filters, widgets: rules by type (§2.3), options by type (§3), values
│   ├── layout.go      # inside the grid, overlaps
│   ├── store.go       # Store, Transactor
│   ├── store_postgres.go, store_postgres_test.go (scope isolation)
│   ├── service.go     # List, Get, Create, Update, Delete
│   ├── service_share.go  # Share, Copy, copy names
│   ├── kinds.go       # history kind customDashboard
│   ├── handler.go, handler_test.go (contract)
│   ├── *_test.go, store_fake_test.go
│   └── customdashboardtest/store.go
├── internal/dataset/                                 # B2
│   ├── model.go, read.go (sheet → table), infer.go (types), cells.go (parse by kind)
│   ├── store.go, store_postgres.go, store_postgres_test.go
│   ├── service.go, service_upload.go, service_share.go, tables.go (widgetdata.DatasetTables)
│   ├── kinds.go, handler.go, handler_test.go, *_test.go, store_fake_test.go
│   └── datasettest/store.go
├── internal/widgetdata/                              # B3 (ports.go, table.go, number.go from F0)
│   ├── ports.go       # Transactor, BuiltInSource, DatasetTables, Rates (F0)
│   ├── table.go       # FieldKind, Field, Value, Row, Table, SourceInfo (F0)
│   ├── number.go      # exact decimal Number (F0)
│   ├── catalogue.go   # the ten built-in sources' fields and filters
│   ├── query.go       # request model and validation against the catalogue
│   ├── filter.go, bucket.go, group.go, aggregate.go, order.go, compare.go
│   ├── service.go     # Catalogue, Query
│   ├── handler.go, handler_test.go, *_test.go
├── internal/app/custom_dashboards.go, datasets.go, widget_data.go (wiring, F0 stubs)
├── internal/app/widgetdata_planning.go, widgetdata_money.go, widgetdata_workflow.go  # B3 adapters
├── internal/app/history_access_owned.go              # INT: log and bin entries of owned kinds
├── internal/access/routes_custom_dashboards.go, routes_datasets.go, routes_data_sources.go
├── internal/stores/custom_dashboards.go, datasets.go
└── seed/dashboards.go (B1), seed/datasets.go (B2)

frontend/
├── package.json, pnpm-lock.yaml                      # echarts 6.1.0 (F0)
├── src/features/custom-dashboards/                   # model.ts, keys.ts, widget-types.ts (F0); charts/ (F1);
│                                                     # builder/ (F2); viewer/, datasets/, sharing/, index.ts (F3)
├── src/app/dashboard-routes.tsx, pages.tsx, navigation.ts, Sidebar (F3)
├── src/ui/tokens.css                                 # chart palette tokens (done)
└── e2e/dashboard-builder.spec.ts, e2e/custom-dashboards.spec.ts (INT)
```

**Structure Decision**: web application; three new business packages laid out as constitution B2 and GO_STANDARDS §2;
one feature folder in the web app with sub-folders per lane (F2).

## Business rules and the tests that prove them

| Rule | Where | Test |
|---|---|---|
| Name required, ≤ 80, unique per owner; description ≤ 500 (FR-008) | `customdashboard/validate.go` | `TestValidateDashboard_NameAndDescription_FR008` |
| Widget rules per type; options per type; values typed (FR-013–FR-015, FR-024, FR-025) | `customdashboard/validate.go` | `TestValidateWidget_RulesByType_FR015` (table-driven, every type), `TestValidateDisplay_OptionsByType_FR025` |
| Grid bounds and no overlap (FR-013, FR-028) | `customdashboard/layout.go`; `builder/grid.ts` | `TestLayout_InsideAndNoOverlap_FR013`; `grid.test.ts` property test (fast-check) |
| Only owner changes; visible or 404 (FR-005, FR-006) | `customdashboard/service.go` | `TestUpdate_NotOwner403_Invisible404_FR006`; `TestPostgresStoreVisibleDashboards_ScopeIsolation_FR005` |
| Version conflict (FR-012) | service | `TestUpdate_VersionConflict_FR012` |
| Log and recycle bin, restore renames when taken (FR-009, FR-010) | service, `kinds.go` | `TestDelete_ToRecycleBinAndUndo_FR010`, `TestLog_DescribesWidgetChanges_FR009` |
| Copy name "Copy of …", numbered, unshared (FR-011) | `service_share.go` | `TestCopy_NameAndOwner_FR011` |
| Share with own teams only; removal of left teams (FR-046) | `service_share.go` | `TestShare_OwnTeamsOnly_FR046` |
| Upload check: formats, header, limits, inference (FR-041, FR-043) | `dataset/read.go`, `infer.go` | `TestPreview_FormatsAndLimits_FR041_FR043`, `TestInferKinds_FR041` (fast-check for numbers and dates) |
| Every cell against its kind, all problems (FR-042) | `dataset/cells.go`, `service_upload.go` | `TestCreate_CellProblemsAllAtOnce_FR042` |
| Dataset visibility, owner-only changes, types fixed (FR-044, FR-045) | dataset service, store | `TestPostgresStoreVisibleDatasets_ScopeIsolation_FR044`, `TestUpdate_ColumnKindsNeverChange_FR044` |
| Catalogue: sources by permission, datasets visible (FR-016) | `widgetdata/catalogue.go`, service | `TestCatalogue_SourcesByPermission_FR016` |
| Groups, date buckets (ISO weeks), multi-valued fields (FR-017) | `bucket.go`, `group.go` | `TestBuckets_DayWeekMonthQuarterYear_FR017`, `TestGroup_MultiValuedCountsInEach_D373` |
| Aggregations exact, rounding, overflow (FR-018) | `aggregate.go`, `number.go` | `TestAggregate_EveryKind_FR018`, `TestNumber_ParseAddRound` |
| Filters and operators (FR-019) | `filter.go`, `query.go` | `TestFilters_EveryOperator_FR019`, `TestQueryValidation_AllProblems_FR019` |
| Sort, limit, Other, group count (FR-020) | `order.go` | `TestOrder_LimitAndOther_FR020` |
| Shares and totals (FR-021) | `aggregate.go` | `TestShares_OfGroupOrTotal_FR021` |
| Dashboard filters only where the source has the field; never datasets (FR-022) | adapters, service | `TestPrefilter_AppliedPerSource_FR022` |
| Key-figure comparison (FR-023) | `compare.go` | `TestComparison_TargetAndPreviousPeriod_FR023` |
| Scope of built-in sources; finance only where visible (FR-007, SC-003) | `internal/app/widgetdata_*.go` | contract tests per persona `TestQuery_OnlyThePlansInScope_FR007`, `TestQuery_FinanceHiddenOutsideFinanceScope_FR007` |
| Widget data equals record pages (SC-002) | — | `TestQuery_EqualsTheRecordPages_SC002` (apitest, demo data) |
| Options per type, palettes, decals, tooltips as DOM, table view (FR-024, FR-025, FR-038) | `charts/build-*.ts`, `WidgetTable` | `build-option.test.ts` per type, `tooltip.test.ts`, `WidgetTable.test.tsx` |
| Builder: add, move, resize, keyboard, undo/redo, validation, guard, conflict (FR-026–FR-035) | `builder/*` | `builder-reducer.test.ts`, `Canvas.test.tsx`, `DashboardBuilderPage.test.tsx` |
| Viewer states, filters in address (FR-036–FR-039) | `viewer/*`, `charts/WidgetView` | `DashboardViewerPage.test.tsx`, `WidgetView.test.tsx` (each state) |
| P1 stories end to end (US1, US2) | — | `e2e/dashboard-builder.spec.ts`, `e2e/custom-dashboards.spec.ts` |

## Complexity Tracking

| Violation | Why needed | Simpler alternative rejected because |
|---|---|---|
| Widget `query` and `display` as JSONB validated by the service (research R5) | Their shape is the contract's and differs by widget type; nothing queries inside them | A dozen normalised tables written and read together for each dashboard |
| Dataset rows as one JSONB array per row (research R6) | Columns differ per dataset | EAV multiplies rows by 50; tables per dataset break sqlc, goose and constraints |
| Widget overlap checked by the service, not an exclusion constraint | Needs the `btree_gist` extension for one rule | — |
| Two chart approaches (036's SVG, the builder's ECharts) | Constitution 4.9.0 F1 records it; moving 036 is a separate change | Rewriting 036's charts now widens this feature |
| Large recycle-bin snapshot of a dataset (≤ 5 MB) | `history.MoveToBin` snapshots the record to restore it exactly | Soft delete bypasses the recycle bin of ADR-0007 |
