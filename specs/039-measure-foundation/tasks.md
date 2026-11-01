# Tasks: Measure Foundation — the Reports Area and the Report Frame

**Input**: Design documents from `specs/039-measure-foundation/` — [plan.md](plan.md), [spec.md](spec.md),
[research.md](research.md), [data-model.md](data-model.md), [contracts/](contracts/), [quickstart.md](quickstart.md)

**Tests**: required by constitution VI — unit tests for every rule, a test for every acceptance scenario, a
contract test per operation, scope tests per persona, one PostgreSQL person-isolation test, a test per widget
state, an end-to-end test for each P1 story (US1–US3). Write each test first and see it fail.

**Lane rules** (build guide §4): 039 is the only feature that edits the hub files (`openapi.yaml`,
`access/routes.go`, `routes_measure.go`, `app/measure.go`, `navigation.ts`, `pages.tsx`, `router.tsx`,
`features/measure/index.ts`, `seed/seed.go`), plus `app/history_access.go` for the `reportView` log rule. Implement **one user story per prompt**; run the checks of
`docs/sdd/quality-gates.md` on the changed files after each story; commit once per finished phase or story.

## Phase 1: Setup (governance)

- [X] T001 Add `reports.view` to 014's Table A in `specs/014-access-control/spec.md` (group Plans and planning,
  after `data.export`: label "See reports", scoped "Yes", refusal noun "Seeing reports") with "(039, D-367)"

## Phase 2: Foundational — the backend hub first (blocking; merge into `measure-module` by hour 16)

**Purpose**: everything lanes C and D import or rely on: the permission, the `measure` types and registry, the
rules, the wiring hub with one stub per report, the migration, the routes and stub pages.

- [X] T002 Contract first: add 039's seven operations and the `Report*` schemas of `contracts/api.md` to
  `backend/api/openapi.yaml` (tag Reports), `reports.view` to the `Permission` enum; run `pnpm run api:types`
  in `frontend/`
- [X] T003 [P] Test first, then add `ReportsView Permission = "reports.view"` and its catalogue row
  `{ReportsView, GroupPlanning, "See reports", "Seeing reports", true}` after `DataExport` in
  `backend/internal/platform/authz/catalogue.go`; test `TestCatalogue_ReportsView_039_FR002` in
  `catalogue_test.go`
- [X] T004 [P] Add `PrefixReportView = "RV"` in `backend/internal/platform/ids/ids.go`
- [X] T005 Migration `backend/internal/platform/database/migrations/00028_reports.sql`: table `report_views`
  exactly as data-model §1 — `id` "`^RV-[0-9]+$`", `person_id` "references people (id), deferrable",
  `report` "one of the report keys of §3", `name` "`name <> ''` and `char_length(name) <= 60`", `is_default`
  not null, `query` "at most 4,000 characters", `version` "at least 1", `created_at`, `created_by_id`,
  `created_by_name`, `updated_at`, `updated_by_id`, `updated_by_name` (`record.Meta`); indexes `report_views_name_key`
  unique `(person_id, report, lower(name COLLATE "C"))`, `report_views_default_key` unique `(person_id, report)
  WHERE is_default`; `roles_permissions_check` widened with `'reports.view'`; grant
  to every role with `plans.view` (data-model §4); Down drops the table and restores the check. Queries in
  `backend/internal/platform/database/queries/measure_views.sql` (`ListReportViews` by person and report,
  `GetReportView`, `PutReportView` upsert, `DeleteReportView`; starring unsets the other default in the service); run
  `sqlc generate && sqlc vet`
- [X] T006 [P] Test first `TestSeedRoles_ReportsView_039_FR002` in `backend/seed/access_test.go`, then give
  `authz.ReportsView` to every built-in role that holds `PlansView` in `backend/seed/access.go`
- [X] T007 Test first `TestCatalogue_NineInOrder_FR001`, `TestCatalogue_UserAccessWithAccessManage_FR003` and `TestCatalogue_RegisteredReportAvailable_FR001` in `backend/internal/measure/catalogue_test.go` (nine reports
  in data-model §3 order with key, path, name, purpose of contracts/ui.md; a registered stub is not available;
  `userAccess` needs `access.manage`); then `backend/internal/measure/model.go` (API types of contracts/api.md),
  `catalogue.go` (the `Report` interface `Key`, `Available`, `Sheets`; `Registry` with `Register`, `Get`,
  `Catalogue(ctx)`; `Stub(key)`), `ports.go` (`Transactor`; `Planning`: `Years`, `Year`, `Choices`,
  `Attributes`, `Currency`, `RatesUsed`; `Plans`: `InReportsScope(ctx) ([]string, error)`; `Currencies`)
- [X] T008 Rules in `backend/internal/access/routes_measure.go` — the seven of contracts/api.md
  (`need(authz.ReportsView)` ×6, `need(authz.DataExport)` for the export) — and `measureRoutes()` joined in
  `backend/internal/access/routes.go`
- [X] T009 Wiring hub `backend/internal/app/measure.go`: builds the `Registry`, calls `registerPlanBudget(w)`,
  `registerApprovals(w)`, `registerPurchaseRequests(w)`, `registerPlanDetails(w)`, `registerGsapLink(w)`,
  `registerUserAccess(w)`, `registerSummary(w)`, `registerRevenue(w)`, `registerFunnel(w)`; one file each,
  `backend/internal/app/measure_{planbudget,approvals,purchasereport,plandetails,gsaplink,useraccess,summary,revenue,funnel}.go`,
  registering `measure.Stub(key)`; called once from `areas.go` (the adapters `Planning` over `allWorkSource` and
  `Plans` over `plan.Store.ListInScope(ctx, authz.PermissionScopeOf(ctx, authz.ReportsView))` come with the frame
  service in T025, so nothing unused lands before it)
- [X] T010 `notYetServed` in `backend/internal/app/contract_routes_test.go`: 039's seven operations under the
  comment line `// 039 Measure foundation`, removed story by story as they are served
- [X] T011 Web hub: **Reports** (icon `ChartColumnIncreasing`, permission `reports.view`) right after **Dashboards**
  in `frontend/src/app/navigation.ts` (039-C3, FR-034); `/reports` and the nine report routes of contracts/ui.md
  in `frontend/src/app/pages.tsx` and `router.tsx` (or a `measure-routes.tsx` beside `dashboard-routes.tsx` if
  `router.tsx` is at its line limit); `frontend/src/features/measure/index.ts` exporting `ReportsPage` and one
  stub page per report from `features/measure/<report>/<Name>Report.tsx`
- [X] T012 Checkpoint: `go test -race -short ./...`, `pnpm run check`; open the pull request of Phase 2 into
  `measure-module` and merge it, so lanes C and D build on the hub (build guide §3 step 2)

**Checkpoint**: the permission, migration, `measure` types, registry, rules, wiring stubs, routes and stub pages
exist on `measure-module`.

## Phase 2b: Wave A contract (integrator; when lanes C and D hand over their `contracts/api.md`)

- [ ] T013 Add the report operations of 040–045 to `backend/api/openapi.yaml` from their
  `specs/04x-*/contracts/api.md`, their rules to `backend/internal/access/routes_measure.go`, and each feature's
  operations to `notYetServed` under its own comment line in `contract_routes_test.go`; run `pnpm run api:types`
- [ ] T014 Pull request of the wave A contract into `measure-module`; tell the lanes to rebase (build guide §4.3)

## Phase 3: User Story 1 — Open a report and narrow it (P1) 🎯 MVP

**Goal**: the Reports page and the report frame with the filter panel, chips, In view, Period, Show and Currency,
all in the address. **Independent test**: open the Reports page and a report as different personas, filter, and
compare "In view" with All work's counts.

### Tests (first)

- [X] T015 [P] [US1] `TestMatch_EqualsAllWorkCounts_039_FR012` in `backend/internal/allwork/match_test.go`: for
  every filter kind (plan, plan type, team, owner, activity type, activity status, attribute value, Not filled
  in, % split share above 0, value hidden by a business-unit condition, line-item attribute) and a fixed-seed
  random mix, `Match` counts equal `Service.Query`'s summary on All work's fixture
- [X] T016 [P] [US1] `TestParseFilter_FR008_FR030` and `TestCanonicalQuery_FR024` in
  `backend/internal/measure/filter_test.go` (defaults; All work's parameters; `period` month|quarter, `show`
  plan|campaign|initiative|activity, `currency` ISO code, `table` widget keys; canonical order, sorted values,
  defaults omitted)
- [X] T017 [P] [US1] `backend/internal/measure/frame_test.go` on fakes: `TestFrame_InView_FR007` (line items of
  matching activities only), `TestFrame_OnlyReportsScope_FR003`, `TestFrame_FinanceCoverage_FR004`,
  `TestFrame_LeftOutCountedNotNamed_FR015`
- [X] T018 [P] [US1] `TestChoices_FR008_FR011` in `backend/internal/measure/choices_test.go` (years with plans in
  scope plus the current; plans of the year; owners and activity types of the selected plans; one group per
  active choice or yes/no attribute per level, never merged by label; inactive values only when used; no text,
  long-text or number attributes; Not filled in as `none`; plans of another year dropped and counted)
- [X] T019 [P] [US1] Contract tests in `backend/internal/measure/handler_test.go` (`package measure_test`,
  `apitest`): `GET /api/v1/reports` and `GET /api/v1/reports/frame` — `TestReportsFrame_ScopePerPersona_SC002`
  (every demo persona: no plan, owner, team or value of another market), 400 for a malformed filter, 404 for an
  unknown report, 403 without `reports.view`
- [X] T020 [P] [US1] `frontend/src/features/measure/shared/report-address.test.ts`: fast-check round trip of
  `ReportFilter` ↔ address; malformed values dropped; canonical query equals the API's
- [X] T021 [P] [US1] Component tests with vitest-axe: `FilterPanel.test.tsx` (order of FR-008, counters, search
  above ten values, Select all / Clear all per group and per tab, "Changes not applied yet", Apply, Escape closes
  without change, Filters (n) below 1024 px), `ReportFrame.test.tsx` (header and meta row, In view line, chips,
  remove a chip, Reset filters, Period/Show/Currency apply at once, left-out note),
  `ReportsPage.test.tsx` (order, purposes, Coming soon without a link; without `reports.view` no sidebar item
  and the standard no-access page — US1 scenario 2); focus returns to the control that opened a menu (FR-033)

### Implementation

- [X] T022 [US1] `backend/internal/allwork/match.go`: exported `Match(nodes []Node, q Query, attributes
  []Attribute, c Choices) Matched` running `knownValues`, `resolveAttrs`, `buildTree`, `applyFilters`; `Matched`
  holds the IDs per level (shown and matching), the live line items of matching activities and `LeftOut`
- [X] T023 [US1] `backend/internal/measure/filter.go`: `ParseFilter(url.Values)` over `allwork.ParseQuery` plus
  `period`, `show`, `currency`, `table`; `Canonical(Filter) string`
- [X] T024 [US1] `backend/internal/measure/choices.go`: the choices of contracts/api.md `ReportChoices`
- [X] T025 [US1] `backend/internal/measure/frame.go`: `Service.Frame(ctx, a, report, f)` in one `Read`: year
  nodes narrowed to `Plans.InReportsScope`, `allwork.Match`, In view, finance coverage from
  `Choices.FinancePlans`, left out as a count, choices, display currency, `generatedAt`, `financeAsOf`
- [X] T026 [US1] `backend/internal/measure/handler.go`: `GET /api/v1/reports` and `/reports/frame`; remove both
  from `notYetServed`
- [X] T027 [P] [US1] `frontend/src/features/measure/shared/report-address.ts` (Zod) and `shared/api.ts`
  (`useReportCatalogue`, `useReportFrame`; keys of contracts/ui.md)
- [X] T028 [US1] `frontend/src/features/measure/ReportsPage.tsx` (catalogue, purposes, Coming soon) with its
  CSS module
- [X] T029 [US1] `frontend/src/features/measure/shared/`: `FilterPanel.tsx` (+ `FilterGroup.tsx`,
  `LevelTabs.tsx`), `FilterChips.tsx`, `InViewLine.tsx`, `PeriodControl.tsx`, `ShowControl.tsx`,
  `CurrencyControl.tsx`, `ReportFrame.tsx`, `ReportFrame.module.css` — tokens only, sentence case, no primary
  button (design system §2.8), every file within 300 lines
- [X] T030 [US1] Stub pages `frontend/src/features/measure/<report>/<Name>Report.tsx` render `ReportFrame` with
  "Coming soon — this report arrives with feature 0NN"
- [ ] T031 [US1] `frontend/e2e/measure.spec.ts` US1 with axe: Reports after Dashboards; the Reports page; a report
  opens on the defaults with "In view" equal to All work's counts; tick a value, "Changes not applied yet",
  Apply; copy the address into a new page; an out-of-scope plan in the address gives the left-out note; chips,
  Reset filters; Period and Currency at once

**Checkpoint**: US1 works on every report page; commit "feat(039): Reports page and report frame with filters
in the address (US1, FR-001–FR-019)".

## Phase 4: User Story 2 — Keep the views I use every month (P1)

**Goal**: personal saved views per report. **Independent test**: save, star, update, rename, delete and undo
views as one person; another person never sees them.

### Tests (first)

- [X] T032 [P] [US2] `TestViews_FR024_FR029` in `backend/internal/measure/views_test.go` on fakes: name trimmed,
  "Enter a name for the view.", "Use 60 characters or fewer.", "You already have a view called '%s'." per person
  and report ignoring case; the same name allowed on another report; query validated ("The view can't be saved —
  reload the page and try again.") and canonicalised, at most 4,000 characters; starring unstars the other view
  of the same report; update; delete returns the deleted view; another person's ID answers not found; a view of a
  report the person may no longer open stays theirs (edge case); `TestViews_VersionAndLog_FR028`: a stale
  version refused with 409 `version_conflict` and "This view was changed in another tab — reload it to see the
  latest", one log entry of kind `reportView` per create, update, star and delete, in the same write
- [X] T033 [P] [US2] `TestPostgresStoreReportViewsPersonIsolation_039_FR025` in
  `backend/internal/measure/store_postgres_test.go` (`TEST_DATABASE_URL`; skipped with `-short`)
- [X] T034 [P] [US2] Contract tests of the four view operations in `backend/internal/measure/handler_test.go`
- [X] T035 [P] [US2] `frontend/src/features/measure/shared/SavedViewsMenu.test.tsx` with vitest-axe: list with
  star and active mark, Save current filters, replace-name confirmation, "Modified" with Update view (confirm)
  and Save as new view, star moves, rename, delete with the Undo toast re-creating the view, default applied when
  the address has no filters, left-out note, the "changed in another tab" message on 409, Escape and focus
  return (FR-033)

### Implementation

- [X] T036 [US2] `backend/internal/measure/store.go` (`Store`: `ListViews(person, report)`, `GetView`,
  `PutView`, `DeleteView`), `store_postgres.go` over the T005 queries, the fake in
  `store_fake_test.go` and `backend/internal/measure/measuretest/`; add the store to `internal/stores`; the view
  embeds `record.Meta`
- [X] T037 [US2] `backend/internal/measure/views.go`: list, create, update, delete as 010's `allwork/views.go`,
  per report, with `record.CheckVersion` and `Touch` on update and one `history.Log` per change in the same
  `Write`; register `history.Kind` "reportView" ("saved report view") in `backend/internal/app/measure.go`; in
  `backend/internal/app/history_access.go` show `reportView` entries to their actor and to `activityLog.viewAll`,
  test first `TestHistoryAccess_ReportView_039_FR028` in `history_access_test.go`
- [X] T038 [US2] Handler operations for views; remove them from `notYetServed`
- [X] T039 [US2] `frontend/src/features/measure/shared/api.ts` view hooks and mutations;
  `shared/SavedViewsMenu.tsx`; `ReportFrame` applies the default view when the address has no filters
- [ ] T040 [US2] `frontend/e2e/measure.spec.ts` US2: save, star, reopen on the default, modify and update,
  rename, delete and Undo; a second persona sees none of them

**Checkpoint**: commit "feat(039): personal saved views per report (US2, FR-024–FR-029)".

## Phase 5: User Story 3 — Every chart as a table, and the figures taken away (P1)

**Goal**: the widget frame with Chart/Table and its export; the whole-report export. **Independent test**: switch
a widget to its table, export it, export the report, compare every value with the screen.

### Tests (first)

- [X] T041 [P] [US3] `TestExport_FR022` in `backend/internal/measure/export_test.go`: Excel sheets in order — the
  report's sheets, **In view**, **About this export** (report, year, filters in words, period, currency,
  exporter, date and time, rates used); CSV needs a widget and carries the notes first; unknown widget refused;
  file names "SP&A One – <report> [– <widget>] – <year> – <date>"; a stub report exports In view and About only
- [X] T042 [P] [US3] `TestReportsExport_FR022` in `handler_test.go`: `.xlsx` and `.csv` answers with
  `Content-Disposition`, 400 for CSV without a widget, 403 without `data.export`
- [X] T043 [P] [US3] `frontend/src/features/measure/shared/ReportWidget.test.tsx`, `ReportTable.test.tsx`,
  `ExportMenu.test.tsx` with vitest-axe: title and ⓘ tip, Chart/Table switch kept in the address, table labels
  and totals row, widget export menu (CSV, Excel), skeleton, error with Try again without blanking others, "—"
  with its reason, "Covers N of M plans", export unavailable with "Exporting isn't part of your role";
  `periods.test.ts` for the Today marker (FR-019): in the current year periods before today's are actual, today's
  is "in progress", later ones planned; a past year only actual; a future year none; months and quarters

### Implementation

- [X] T044 [US3] `backend/internal/measure/export.go`: `Export(ctx, a, report, f, format, widget)` with
  `platform/sheet` (`WriteXLSX`, `WriteCSV`), In view and About sheets, `allwork.ContentDisposition` file names
- [X] T045 [US3] Handler `GET /api/v1/reports/{report}/export`; remove it from `notYetServed`
- [X] T046 [US3] `frontend/src/features/measure/shared/ReportWidget.tsx`, `ReportTable.tsx`, `ExportMenu.tsx`
  (whole report) using `src/api/download.ts`; chart tokens of D-345 only; `shared/periods.ts` (period state from
  the plan year and today) and the labelled **Today** marker every time chart of 040–050 draws (FR-019)
- [ ] T047 [P] [US3] Gallery entries for `ReportWidget` and `ReportTable` in
  `frontend/src/features/dev/ComponentGalleryPage.tsx` (a sample bar chart with its table)
- [ ] T048 [US3] `frontend/e2e/measure.spec.ts` US3: Export from a report page downloads an `.xlsx` whose sheets
  include In view and About this export

**Checkpoint**: commit "feat(039): table view and export of every widget and report (US3, FR-020–FR-023)".

## Phase 6: User Story 4 — Why a report is empty or incomplete (P2)

- [ ] T049 [P] [US4] Tests first: `TestFrame_NoPlansInYear_FR031` in `frame_test.go` (years that have plans
  returned); `ReportFrame.test.tsx` states "No plans you may see for 2027" with the years, "Your filters exclude
  everything" with Reset filters, "You can't see financial data", one widget failing alone
- [ ] T050 [US4] Empty and partial states in `backend/internal/measure/frame.go` and
  `frontend/src/features/measure/shared/ReportFrame.tsx`

**Checkpoint**: commit "feat(039): empty states that say why (US4, FR-031)".

## Phase 7: Polish and cross-cutting

- [ ] T051 `BenchmarkReportsFrame` in `backend/internal/platform/apibench/reports_test.go` (demo data and the 028
  volume); record the results in `plan.md` (SC-003, research R9)
- [ ] T052 Comment budget and file length of **every file this feature touched**, including the hub and Plan
  files (`routes.go`, `catalogue.go`, `ids.go`, `seed/access.go`, `navigation.ts`, `pages.tsx`, `router.tsx`,
  `contract_routes_test.go`): whole-file 5% budget (CLAUDE.md 3), at most 300 non-blank lines
- [ ] T053 Run `specs/039-measure-foundation/quickstart.md` by hand on the demo data; fix what differs
- [ ] T054 `/speckit-converge`; spec status Built; pull request into `measure-module` with the feature-map status
  change "039 Built" in its description (build guide §4.5)

## Dependencies

- T001–T012 before every story; T012 (hub merged) before lanes C and D start their backends.
- T013–T014 need the lanes' `contracts/api.md`; they do not block 039's stories.
- US1 (T015–T031) before US2 and US3 (both live inside `ReportFrame`); US4 after US1.
- Within a story: tests first, then backend (allwork → measure → handler), then web, then e2e.

## Parallel examples

- Phase 2: T003, T004 and T006 together; T005 after T004.
- US1 tests: T015–T021 at once (different files); backend T022–T026 and web T027–T030 in two worktrees from the
  fixed contract (T002).
- US2 tests T032–T035 at once; US3 tests T041–T043 at once.

## Implementation strategy

MVP = Phase 2 + US1: the hub merged early for the other lanes, then a working frame on every report page. US2
and US3 complete the P1 set; US4 and the polish phase finish the feature. One story per implementation prompt;
commit at each checkpoint.

## Requirement coverage

| FR / SC | Tasks |
|---|---|
| FR-001 | T007, T011, T021, T028, T031 |
| FR-002 | T001, T003, T005, T006 |
| FR-003 | T007, T008, T017, T019 |
| FR-004 | T017, T043, T049 |
| FR-005 | T019 (read-only operations), T025 |
| FR-006, FR-007 | T015, T017, T021, T022, T025, T029, T031 |
| FR-008–FR-012 | T015, T016, T018, T021, T022–T024, T029 |
| FR-013, FR-014 | T021, T029, T031 |
| FR-015 | T016, T017, T020, T027, T031 |
| FR-016–FR-018 | T016, T021, T029 |
| FR-019 | T043, T046 |
| FR-020–FR-023 | T041–T048 |
| FR-024–FR-029 | T005, T032–T040 (FR-028 version and log: T005, T032, T036, T037) |
| FR-030 | T016, T021 |
| FR-031 | T049, T050 |
| FR-032, FR-033 | T021, T029, T043 (axe), T052 |
| FR-034 | T011, T031 |
| SC-001 | T015, T031 |
| SC-002 | T019 |
| SC-003 | T051 |
| SC-004, SC-005 | T020, T031, T040 |
| SC-006 | T041, T048 |
| SC-007 | T021, T035, T043, T031 |
