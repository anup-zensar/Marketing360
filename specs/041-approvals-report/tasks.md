# Tasks: Approvals Report

**Input**: Design documents from `specs/041-approvals-report/` — [plan.md](plan.md), [spec.md](spec.md),
[research.md](research.md), [contracts/api.md](contracts/api.md). No `data-model.md` (no new table), no
`quickstart.md` (plan.md stays short, build guide §8).

**Tests**: required by constitution VI — a unit test for every rule, a test for every acceptance scenario, a
contract test for the operation, a test for each state a widget renders, an end-to-end test for the P1 story
(US1). Only the end-to-end test uses a real database; this feature adds no store read, so no scope-isolation
test on PostgreSQL. Write each test first and see it fail.

**Lane rules** (build guide §4): this lane owns `backend/internal/measure/approvalreport/`,
`backend/internal/app/measure_approvals.go`, `frontend/src/features/measure/approvals/` and
`frontend/e2e/measure-approvals.spec.ts` (plus `specs/041-approvals-report/`). It never edits `openapi.yaml`,
`routes_measure.go`, `app/measure.go`, `app.go`, `navigation.ts`, `pages.tsx`, `router.tsx`,
`features/measure/index.ts`, `seed/seed.go`, `feature-map.md`, the decision log, `internal/approval` or
`purchaserequest`. Its only edit outside its folders is deleting its own lines from `notYetServed` in
`internal/app/contract_routes_test.go`. Implement **one user story per prompt**; run the checks of
`docs/sdd/quality-gates.md` on what changed after each story (`scripts/check-changed.ps1` is not in the repo
yet); commit once per finished story.

## Phase 1: Setup and gates (no code)

- [X] T001 Plan, research and `contracts/api.md` pushed to `origin/feature/041-approvals-report` for lane A's
  wave A contract (commit `feba42c4`)
- [X] T002 Record lane A's answers to plan.md questions 1–5 (report-specific parameters in address and saved
  views; matching helper and export location; file name `measure_approvals.go`; purchase request name/owner/PO
  read; benchmark file) as 041-C8… in `specs/041-approvals-report/spec.md` Clarifications and fix
  `contracts/api.md` if the parameters move
- [X] T003 Gate check, no work until all hold: 039's hub is merged into `measure-module` (build guide step 2,
  about hour 16) and the wave A contract is in `backend/api/openapi.yaml` (039 T013–T014); then `git fetch`,
  rebase `feature/041-approvals-report` on `measure-module`, `pnpm run api:types` in `frontend/`

## Phase 2: Foundational (blocking — after T003)

**Purpose**: the package skeleton, its ports and fakes, and the one read decision both stories depend on.

- [X] T004 Confirm which existing `purchaserequest` read gives name, owner, amount and PO number of the
  purchase requests on the plans in view in one statement (R4). Finding: `Store.OnLineItems(ctx, ViewScope,
  lineItemIDs)` returns whole purchase requests (name, owner, lines with PO numbers) in one scoped statement but
  no USD amount; `Service.Summaries` (built on it) gives the amount, status and plan. The adapter calls both —
  two fixed statements, no change to `purchaserequest` (research R4, plan.md question 4)
- [X] T005 [P] `backend/internal/measure/approvalreport/model.go`: the API types of `contracts/api.md`
  (`ApprovalsReport`, `ApprovalStatusMix`, `ApprovalAverage`, `ApprovalGrid`, `ApprovalRow`, `PersonRef`) and the
  outcome words (`waiting`, `approved`, `declined`, `recalled`, `replaced`, `autoApproved`, `notNeeded`) as
  constants; money as `money.Money`, never a float
- [X] T006 [P] `backend/internal/measure/approvalreport/ports.go`: the small consumer interfaces the report reads
  through — `Transactor` (`Read`), `Planning` (the year's nodes for plans in the reports scope, as 039's
  `measure.Planning`), `Rounds` (`RoundsOfSubjects`, `ActionsOfSubjects` for subject IDs), `Purchases` (R4), `Deps`
- [X] T007 [P] `backend/internal/measure/approvalreport/approvalreporttest/`: fakes of the ports on
  `platform/fake` (`fake.DB`, `fake.Table[T]`), with a fixture of rounds in every engine outcome, levels with
  one, any and all approvers, an automatic level, a skipped level and a plan without `finance.view`

## Phase 3: User Story 1 — See approval workload and cycle time (P1) 🎯 MVP

**Goal**: Reports › Approvals shows the status mix per record type and the average days to a decision, with the
per-approver list, narrowed by 039's filters. **Independent test**: as Alex Morgan on the demo year the waiting
count, the oldest waiting date and the average days equal Approvals › Monitoring and 036's Approvals panel for the
same plans.

### Tests first (write, see fail)

- [X] T008 [P] [US1] `TestOutcomeWord_041_FR006` in `approvalreport/outcomes_test.go`: table of every engine
  `Outcome` and `AutoApproved` value → one word (R3: `restarted`→replaced, `cancelled`→recalled)
- [X] T009 [P] [US1] `TestStatusMix_041_FR006_FR007` in `service_test.go` on the fakes: counts and whole-number
  shares add to the total; the selected record type only; only types that have rounds are offered (and no plan,
  campaign, initiative, line-item or accrual type, FR-024); default `activity`; a type with no rounds in view, or
  an unknown `statusType`, is answered with total 0 and no error (041-C10)
- [X] T010 [P] [US1] `TestWaiting_ByType_041_FR008` and `TestPurchaseRequestsInView_041_C8`: waiting count and
  oldest-waiting-since of the selected type; a purchase request is in view only through a line item in view (the
  frame's plan, owner, activity type, status and attribute filters narrow it); equality with 036 is proved at
  contract level in T015
- [X] T011 [P] [US1] `TestAverage_041_FR009`, `TestAverage_NoDecisions_041_FR011` in `average_test.go`:
  `approval.AverageDecisionDays` over person-decided rounds, auto-approved and open rounds left out, a round
  decided within the hour counts 0.0 days, `null` and never zero when none (R1); the rule is the one of Approvals
  › Monitoring, which covers rounds ended this month only, so equality is not asserted against it
- [X] T012 [P] [US1] `TestApproverAverages_041_FR010`: per-approver mean from the round's submission to that
  person's own decision, slowest first, the decider credited for backup, delegate and on-behalf decisions, waiting,
  recalled, replaced, not-needed and automatic steps not counted (041-C2, C7)
- [X] T013 [P] [US1] `TestScope_PerPersona_041_FR003` in `service_test.go`: a person sees only rounds of plans in
  their `reports.view` scope; a round of another plan is neither read nor counted; amounts `null` without
  `finance.view` and `financeCovers` says how many plans
- [X] T014 [P] [US1] `TestSheets_041_FR023` for the `status` and `average` widgets: sheet order, headings, the
  "All approvers" first row, currency named in headings
- [X] T015 [P] [US1] `TestApprovalsReport_Contract_041` in `handler_test.go` (`package approvalreport_test`, on
  `apitest`): 200 shape against `openapi.yaml`, 400 problems only for a malformed frame filter, `sort`, `order`,
  `page` or `pageSize`, 401 signed out, 403 without `reports.view`, 200 with empty widgets for an unknown
  `statusType`, `recordType`, `outcome` or `approver` (041-C10); plus `TestApprovalsReport_EqualsDashboard_041_SC002`
  on DemoDB: calls this endpoint and `GET /dashboards/executive`, and waiting summed over the three types, the
  oldest waiting date and the average equal 036's Approvals panel for the same plans
- [X] T016 [P] [US1] `frontend/src/features/measure/approvals/StatusWidget.test.tsx` and `AverageWidget.test.tsx`
  with MSW `reply()` and `vitest-axe`: one test per state — data, loading skeleton, "No approvals yet" with what
  creates them, "No decisions yet" with "—", "You can't see financial data" / "Covers N of M plans" where relevant,
  error with **Try again**; record-type selector offers only present types; table view lists the same figures
- [X] T017 [P] [US1] `ApprovalsReport.test.tsx`: page states from the frame — "No plans you may see for <year>",
  "Your filters exclude everything" with **Reset filters**, no access (the standard page), data; `schemas.test.ts`
  with fast-check: the report's own parameters round-trip through the address and malformed values are dropped

### Implementation

- [X] T018 [US1] `approvalreport/outcomes.go`: engine outcome → word (R3), `OutcomeWord(round)`
- [X] T019 [US1] `approvalreport/average.go`: overall and per-approver averages through
  `approval.AverageDecisionDays`, slowest first
- [X] T020 [US1] `approvalreport/service.go`: `Service.Report(ctx, a, f, q)` in one `Read`: the year's nodes for
  plans in scope, **one** `RoundsOfSubjects` and **one** `ActionsOfSubjects` for those subjects (activities and
  purchase requests), status mix, waiting, averages; no statement per row; `Service.Sheets` for `status` and
  `average`
- [X] T021 [US1] `approvalreport/handler.go`: `RegisterRoutes` for `GET /reports/approvals`; parse 039's filter
  once (`measure.ParseFilter`) and the report's own parameters; the service never reads the request; `web.Respond`
  and `web.RespondError`
- [X] T022 [US1] `backend/internal/app/measure_approvals.go`: the body of `registerApprovals` — the adapter on
  `stores.ApprovalStore`, `NewService`, `mount(w.signedIn, handler)`, register the `measure.Report` with
  `Available() == true`; delete this operation's line from `notYetServed` in `contract_routes_test.go`
- [X] T023 [P] [US1] `frontend/src/features/measure/approvals/schemas.ts` (Zod for `statusType` and the grid
  parameters) and `api.ts` (`useApprovalsReport(filter, params)`, key `['reports','approvals',query]`, typed by the
  generated contract)
- [X] T024 [P] [US1] `StatusWidget.tsx` (036's SVG bars with the four chart tokens, never a chart library; each
  count and share also as text; record-type selector; Chart · Table) and `AverageWidget.tsx` (card, approver list)
  inside `ReportWidget`
- [X] T025 [US1] `ApprovalsReport.tsx`: replace the stub with `ReportFrame` and the widgets; Period and Show are
  not offered (spec FR-005); the grid slot stays empty until US2
- [X] T026 [US1] (written; CI runs it, not yet run locally) `frontend/e2e/measure-approvals.spec.ts` US1 with axe: Alex opens Reports › Approvals, the
  figures equal Approvals › Monitoring, switching the record type changes the mix, a filter narrows both widgets,
  Sara (Viewer) sees the report without export, a person without `reports.view` gets the no-access page
- [ ] T027 [US1] Checkpoint: the checks of `docs/sdd/quality-gates.md` on the changed files, comment budget and
  file length of every file touched; commit `feat(041): approvals status mix and average (US1)`; stop

## Phase 4: User Story 2 — Find, inspect and export approval records (P2)

**Goal**: the grid of every approver step with filters, search, sort, paging, full comments and export.
**Independent test**: as Alex filter by approver Carlos Mendes and status Approved, search a record name, export:
the file holds every matching row and no row of a plan outside Alex's scope.

### Tests first (write, see fail)

- [X] T028 [P] [US2] `TestRows_041_FR012_C2_C3` in `rows_test.go`: no row of a plan, campaign, initiative or line
  item, no Vendor filter or accrual column (FR-024); one row per named approver per level; others
  read Not needed once one decided; an automatic level is one row with no approver; a skipped level no row;
  waiting rows show "—" for decided on
- [X] T029 [P] [US2] `TestGridQuery_041_FR013_FR015`, `TestUnknownValues_041_C10`,
  `TestApproverFilterGridOnly_041_C9`: record type, approver and outcome filters, an unknown value gives an
  empty page not an error, the approver filter leaves the widgets unchanged, search over text columns, sort by
  every column including `comments` across all rows with ties by item then level, every page size, total
- [X] T030 [P] [US2] `TestAmounts_041_FR018`, `TestCanOpen_041_FR017`: planned cost now and 032's amount in the
  display currency, changing the display currency re-expresses every amount with the currency named (US2 sc.9),
  `null` without `finance.view`; `canOpen` false when the person may not open the record
- [X] T031 [P] [US2] `TestSheets_041_FR023` extended for `approvals`: every matching row, no paging, the grid's
  columns in order; contract test extended for the grid parameters and 400 for an unknown `sort`
- [X] T032 [P] [US2] `ApprovalsGrid.test.tsx` with `vitest-axe`: columns in order, search, sort, outcome, approver
  and record-type filters, page size, total, a long comment opens in full, a name is plain text when `canOpen` is
  false, empty grid with the cause (no approvals / filters exclude everything with reset), loading, error with
  **Try again**

### Implementation

- [X] T033 [US2] `approvalreport/rows.go`: rows from rounds and actions (levels, approvers, decisions, capacity),
  names, owners, amounts, joined PO numbers (041-C11) and `canOpen` from the year's nodes, `Purchases` and
  `OnLineItems` (T004)
- [X] T034 [US2] `approvalreport/grid.go`: filters, search, sort and paging over all rows; page-size allowlist;
  `choices` from the people and types with rows
- [X] T035 [US2] `service.go`, `handler.go`, `app/measure_approvals.go`: wire the grid, the extra parameters and
  the `approvals` sheet into `Service.Report` and `Sheets` (still one statement per read)
- [X] T036 [P] [US2] `frontend/src/features/measure/approvals/ApprovalsGrid.tsx` (+ `GridFilters.tsx`,
  `CommentDialog.tsx` on Radix Dialog): the grid in `ReportWidget`, address-backed parameters, export through the
  frame's menu
- [X] T037 [US2] `ApprovalsReport.tsx`: place the grid below the widgets; extend `frontend/e2e/measure-approvals.spec.ts`
  with the export scenario (P2, one short test)
- [ ] T038 [US2] Checkpoint: the quality-gates checks on the changed files, comment budget and file length;
  commit `feat(041): approvals grid and export (US2)`; stop

## Phase 5: Polish and cross-cutting

- [ ] T039 `BenchmarkApprovalsReport` in `backend/internal/platform/apibench/measure_approvals_test.go` (location
  per lane A, T002) at the 028 volumes; record statements and time in `plan.md` § Benchmark
- [X] T040 Comment budget and file length of **every file this feature touched**, including the test files; split
  any file above 300 non-blank lines
- [ ] T041 Run the report by hand on the demo data (Alex, Daniel, Sara): every spec scenario and state; fix what
  differs
- [ ] T042 Contract problems found while building (`openapi.yaml`, route rule, shared parts) written in the pull
  request for lane A, not edited here (build guide §4.3)
- [ ] T043 `/speckit-converge`; spec status Built; rebase on `measure-module`; pull request into `measure-module`
  with the feature-map status change "041 Built" in its description (build guide §4.5), three checks green

## Dependencies

- T002 and T003 gate everything below; T001 is done.
- Phase 2 (T004–T007) before both stories; T005–T007 in parallel; T004 before T033.
- US1 before US2 (the grid sits in the same page and service); US2 does not change US1's behaviour.
- Within a story: tests first, then backend (outcomes → average → service → handler → adapter), then web, then
  e2e, then the checkpoint.
- US1's web tasks (T023–T025) need 039's `ReportFrame`, `ReportWidget`, `ReportTable` and `measure.ParseFilter`
  (039 US1 and US3); until they land, the backend (T018–T022) and its tests can be built against the ports.

## Parallel examples

- Phase 2: T005, T006, T007 together.
- US1 tests: T008–T017 at once (different files); backend T018–T022 and web T023–T025 in two worktrees from the
  fixed contract.
- US2 tests T028–T032 at once; T033–T035 backend and T036 web in two worktrees.

## Implementation strategy

MVP = Phases 1–3: the status mix and the average, which answer "is approval a bottleneck?". US2 adds the grid and
the export. One story per implementation prompt; commit at each checkpoint; stop and report after each.

## Requirement coverage

| FR / SC | Tasks |
|---|---|
| FR-001, FR-004, FR-005 | T010, T015, T017, T025, T026 |
| FR-002 | T015 (read-only operation), T026 |
| FR-003 | T013, T015, T026 |
| FR-006, FR-007 | T008, T009, T016, T018, T020, T024 |
| FR-008 | T010, T015, T020 |
| FR-009, FR-010, FR-011 | T011, T012, T019, T024 |
| FR-012, FR-013, FR-014 | T028, T029, T033, T034, T036 |
| FR-015, FR-016, FR-017 | T029, T030, T032, T034, T036 |
| FR-018, FR-019 | T013, T030, T033 |
| FR-020 | T016, T017, T032 |
| FR-021 | T014, T020, T031 |
| FR-022 | T039 |
| FR-023 | T014, T031, T035, T037 |
| FR-024 | T009, T028 (assertions that only the three record types and the columns of US2 are offered) |
| SC-001 | T026, T041 |
| SC-002 | T015 (contract-level equality with 036), T026 |
| SC-003 | T029, T032 |
| SC-004 | T013, T031, T037 |
| SC-005 | T016, T017, T032 |
| SC-006 | T016, T032, T026 (axe), T039 |

## Phase 6: Convergence

- [X] T044 Confirm with the purchase request (032) owners when a person may open a purchase request, and make `canOpen` follow it per request: today the adapter sets it for every request from the purchase-requests list permission alone, in `backend/internal/app/measure_approvals.go`, per FR-017 and 041-C5 (partial)
- [X] T045 Decide whether the status and average widgets need the "You can't see financial data" and "Covers N of M plans" states T016 listed, or record in the spec that the grid carries them because the widgets show no amounts, per FR-018 and FR-020 (partial)
