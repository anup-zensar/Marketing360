# Tasks: Plan Details Report

**Input**: Design documents from `specs/043-plan-details-report/` — [plan.md](plan.md), [spec.md](spec.md),
[research.md](research.md), [contracts/api.md](contracts/api.md), [quickstart.md](quickstart.md); the operation and
schemas as merged in `backend/api/openapi.yaml` (PR #57).

**Tests**: required by constitution VI — a test per rule and per widget state, the equality test with All work's
Breakdown, a contract test, persona scope tests, an end-to-end test for US1. Write each test first and see it fail.

**Lane rules** (build guide §4): only issue #36's "Owns" paths — `measure/plandetails`, `app/measure_plandetails.go`,
`features/measure/plan-details/`, `e2e/measure-plan-details.spec.ts` — plus removing 043's line from `notYetServed`.
Commit once per finished phase or story; pull `measure-module` and push after each.

**Waits on 039** (lane A, Yash): a task marked **⏸ 039 Tnnn** starts only when that 039 task is on
`measure-module` or pushed on `feature/039-measure-hub`; until then it is not built here (no copy of another lane's
work).

## Phase 1: Setup

- [X] T001 Branch `feature/043-plan-details-report` has `measure-module` with 039's hub (PR #56) and the wave A
  contract (PR #57) merged; the backend builds and `go test -short ./internal/measure/...` passes

## Phase 2: Foundational (blocking for every story)

- [X] T002 [P] `backend/internal/measure/plandetails/model.go`: the API types exactly as `openapi.yaml`'s
  `PlanDetailsReport`, `PlanDetailsCards`, `PlanDetailsAttributeRow`, `PlanDetailsWidget`, `PlanDetailsValue`
  (JSON names as the schema; nullable fields as pointers; `plannedCost` as `money` API amounts as other packages
  answer them), plus the level and measure constants
- [X] T003 [P] `backend/internal/measure/plandetails/ports.go`: `Transactor` (Read); `Scoped` with
  `Matched(ctx, actor.Actor, measure.Filter, columns []string) (Matching, error)` and `Attributes(ctx)`;
  `Matching{Nodes []allwork.Node; Matched map[string]bool (IDs of the matching records of every level);
  ActivityTypes []allwork.Option; Currency allwork.Currency; Year int}` (the plan year, for start periods) — 043's own matched set, so nothing here waits for
  `allwork.Matched` (039 T022)
- [X] T004 `backend/internal/measure/plandetails/year_test.go`: an in-memory year builder for the tests (plans with
  Business unit options, campaigns, initiatives, activities with planned cost, dates, status, type, goals,
  `Missing`, `Attrs`; line items) and a fake `Scoped` returning it with a matched set and a currency

**Checkpoint**: the package compiles; every story can be written test first.

## Phase 3: User Story 1 — Activity spread and completeness (P1) 🎯 MVP

**Goal**: the Activity level: cards, built-in widgets and one three-widget row per active activity attribute.
**Independent test**: spec US1 on the demo data; hand counts; Not filled in opens All work filtered.

### Tests (first)

- [X] T005 [P] [US1] `attribution_test.go`: `TestAttribution_Split_FR011` (10,000.00 split 60/40 → 6,000.00 and
  4,000.00; 3 shares of 33.33 % → half up, the rest on the last value), `TestAttribution_CountsInEach_FR011`,
  `TestAttribution_ZeroShare_FR011` (a 0 % share neither counts the record nor gets an amount),
  `TestAttribution_DoesNotApply_FR012` (hidden by the Business unit condition → doesNotApply, not none)
- [X] T006 [P] [US1] `profile_test.go`: `TestRows_ActiveInDisplayOrder_FR008`, `TestRows_InactiveOption_FR008`
  ("(inactive)", `inactive` true), `TestRows_WidgetsByType_FR009_FR010` (three widgets for single, multiple,
  % split and yes/no; filledIn only for text, long text, number), `TestRows_NotFilledIn_FR012` (US1 scenario 3:
  GB01 12, CN04 10, none 6, 79 %), `TestFilledInPercent_LeavesOutDoesNotApply_FR012` (US1 scenario 11: 15 · 5 ·
  8 → 75 %), `TestRows_NoAttributesSetUp_FR016` (`attributesSetUp` false)
- [X] T007 [P] [US1] `builtin_test.go`: `TestBuiltIn_Activity_FR006` (by type count and planned cost, by
  status with every status that has activities), `TestStartPeriods_MonthsQuarters_FR006` (period month or quarter
  of the plan year, `noDates`, `outsideYear`, adding up to the count), `TestPlannedCostEntered_FR006` (above zero
  → Yes; US1 scenario 1: 26 · 2), `TestGoalEntered_FR006` (any of INQ, MQL, revenue goal above zero)
- [X] T008 [P] [US1] `level_test.go`: `TestCards_FR005` (records, planned cost, all filled in = `Missing` 0,
  per cent half up, null with no record), `TestLevel_OnlyMatchedRecords_FR003` (an unmatched or out-of-scope
  node never counts or is named), `TestEmpty_FR016` (no matching record: cards 0, `builtIn` empty, rows with
  total 0, no state of its own)
- [X] T009 [P] [US1] `widget_test.go`: `TestValues_LargestFirstThenNoneAndDoesNotApply_FR009`,
  `TestOther_AfterTenLargest_FR017` (`other` sums the rest; `values` still holds every value),
  `TestPlannedCost_InReportCurrency_FR013` (summed in USD, converted once at the currency's rate; CNY plan, EUR
  report), `TestKeys_ReportWidgetPattern` (`attr12Count` … match `^[a-z][A-Za-z0-9]{0,39}$`),
  `TestValues_LinkableAndFilterParam_FR014` (`none` linkable, `doesNotApply`, `other`, `total` not)
- [X] T010 [US1] `breakdown_test.go`: `TestAttribution_EqualsAllWorkBreakdown_SC002` — `allwork.NewService` over
  an in-memory `allwork.Source` with the same year; for every active choice and yes/no **activity** attribute,
  each value's planned cost equals `Service.Breakdown` by that attribute to the cent, and none + doesNotApply
  equals Breakdown's Not filled in

### Implementation

- [X] T011 [US1] `attribution.go`: 016 FR-041 for one record and one amount — values carried, % split half up
  with the remainder on the last value, several values without shares each get the whole amount (counts in each),
  0 % shares dropped; Does not apply from `Attribute.Condition` and the plan's `BusinessUnitIDs` before the value
  is read (research R2, R3); about 40 lines
- [X] T012 [US1] `widget.go`: values ordered largest first (count, or planned cost for plannedCost widgets), then
  `none` and `doesNotApply`; `other` after the ten largest; `total` counting each record once; filled in per cent
  half up; keys as T009; USD sums converted once to `Matching.Currency`
- [X] T013 [US1] `profile.go`: one row per active attribute of the level in Settings' order; the planned cost
  attributed **per matching activity** (the activity's own value at the activity level), so the amounts equal the
  Breakdown's rounding (T010); records counted once per value
- [X] T014 [US1] `builtin.go`: the Activity section of T007 with labels from `Matching.ActivityTypes` and the
  status words of domain model §2.3
- [X] T015 [US1] `level.go` and `service.go`: `Report(ctx, a, f)` in one `Read` — `Scoped.Matched` with the
  columns `completeness` and `inq`, the level `f.Show` (default `activity`), cards, built-in widgets, rows
- [X] T016 [US1] `sheets.go`: `measure.Report` (`Key` planDetails, `Available`, `Sheets` one `WidgetSheet` per
  widget of `f.Show` in screen order); `TestSheets_EveryValueAndTotal_FR017`, `TestSheets_KeysOfLevel_FR001`
- [ ] T017 [US1] **⏸ 039 T023** `handler.go`: `GET /api/v1/reports/plan-details` parsing the filter with 039's
  `measure.ParseFilter`; `TestHandler_ShowDefaultsToActivity_FR001`, 400 for a malformed filter
- [ ] T018 [US1] **⏸ 039 T022, T025** `backend/internal/app/measure_plandetails.go`: `registerPlanDetails`
  registers the report instead of the stub and mounts the handler; the `Scoped` adapter over the frame's step
  (or `measure.Planning.Year`, `measure.Plans.InReportsScope` and `allwork.Match`); remove
  `GET /api/v1/reports/plan-details` from `notYetServed`; contract tests in `plandetails/handler_test.go`:
  `TestPlanDetails_Contract`, `TestPlanDetails_NoReportsView_FR003`, `TestReport_OnlyPlansInScope_FR003` (persona
  with one of two plans), `TestReport_NoFinanceView_ShowsPlannedCost_FR004`
- [ ] T019 [US1] **⏸ 039 T027–T030, T046** `frontend/src/features/measure/plan-details/`: `api.ts`,
  `schemas.ts`, `PlanDetailsReport.tsx` (replaces 039's stub, inside `ReportFrame`), `PlanDetailsCards.tsx`,
  `BuiltInSection.tsx`, `AttributeRow.tsx`, `ValueWidget.tsx`, `FilledInWidget.tsx` (036's SVG bars in
  `ReportWidget`, table through `ReportTable`), `allWorkLink.ts` (`attr=<attributeId>:<key>` plus the frame's
  filters)
- [ ] T020 [US1] **⏸ 039 T027–T030, T046** `PlanDetailsReport.test.tsx` (one test per state: data, no plans,
  no data yet, no attributes set up, filters exclude everything), `AttributeRow.test.tsx`, `allWorkLink.test.ts`,
  vitest-axe in each
- [ ] T021 [US1] **⏸ 039 T031, #DEMO** `frontend/e2e/measure-plan-details.spec.ts`: spec US1 on the demo data
  with axe

**Checkpoint**: US1 works end to end on the Activity level.

## Phase 4: User Story 2 — Plans, campaigns and initiatives (P2)

**Goal**: the same layout for the other three levels. **Independent test**: spec US2; Campaign by Cluster equals
All work's Breakdown by Cluster.

- [ ] T022 [P] [US2] Tests: `TestBuiltIn_Containers_FR007` (plan status; approval progress and timing),
  `TestRows_OtherLevels_FR008`, `TestCards_PlannedCostFromMatchingActivities_FR005`, extend T010 to every level
  (`TestAttribution_EqualsAllWorkBreakdown_SC002` for plan, campaign and initiative attributes), US2 scenario 7
  (one of two plans in scope)
- [ ] T023 [US2] `level.go`, `builtin.go`, `profile.go`: records of the plan, campaign and initiative levels; a
  container's planned cost is its matching activities' (domain model §4.2), attributed per activity through the
  container's value
- [ ] T024 [US2] **⏸ 039 T029–T030** web: the level through 039's `ShowControl` (address `show`), "No <level>
  attributes are set up in Settings"; tests per level state

## Phase 5: User Story 3 — Export a widget (P2)

- [ ] T025 [P] [US3] `sheets_test.go`: per measure the columns Value, Records (named for the level), Planned cost
  (<currency>), Share; every value including those `other` gathers; a total row; empty widget → header and total
- [ ] T026 [US3] `sheets.go` completed to T025
- [ ] T027 [US3] **⏸ 039 T044–T046** e2e in `measure-plan-details.spec.ts`: widget export CSV and Excel equal the
  table; the report's Export holds one sheet per widget of the level

## Phase 6: Polish

- [ ] T028 `BenchmarkPlanDetails_SC001` in `plandetails/bench_test.go` at the 028 volume (research R5); result
  written in plan.md
- [ ] T029 Comment budget and file length of every file of this feature (constitution V, VIII);
  `scripts/check-changed.ps1` green
- [ ] T030 quickstart.md by hand on the demo data; `/speckit-converge`; spec status Built; pull request into
  `measure-module` with the feature-map status change

## Dependencies

- T001 → T002–T004 → US1 (T005–T016 need nothing outside lane D) → T017 (039 T023) → T018 (039 T022, T025) →
  T019–T021 (039 web, #DEMO).
- US2 needs US1's T011–T015; US3 needs T016. T024 and T027 wait on 039's web and export.
- The tests of each story come before its implementation.

## Parallel examples

- Phase 2: T002 and T003 together.
- US1 tests: T005, T006, T007, T008, T009 together (different files), then T010.

## Implementation strategy

MVP = US1's backend rules (T002–T016), then its handler, wiring and page as 039 lands. US2 and US3 extend the same
package. Each story is committed and pushed when its tests pass.
