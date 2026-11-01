# Tasks: GSAP Link Reports

**Input**: Design documents from `specs/044-gsap-link-reports/` — [plan.md](plan.md), [spec.md](spec.md),
[research.md](research.md), [contracts/api.md](contracts/api.md), [quickstart.md](quickstart.md); the operation and
schemas as merged in `backend/api/openapi.yaml` (PR #57); 039's frame, `measure.Service.Scope` and export
(PRs #58 – #60).

**Tests**: required by constitution VI — a test per rule and per tab state, a scope-isolation test per store read,
the equality tests with All work and with 032's Waiting for GSAP view, a contract test, an end-to-end test for US1.
Write each test first and see it fail.

**Lane rules** (build guide §4): only issue #37's "Owns" paths — `measure/gsaplink`, `app/measure_gsaplink.go`,
`queries/measure_gsaplink.sql` (and its generated `sqlcdb` file), `features/measure/gsap-link/`,
`e2e/measure-gsap-link.spec.ts` — plus removing 044's line from `notYetServed`. Commit and push after each story.

## Phase 1: Setup

- [X] T001 Branch `feature/044-gsap-link-reports` has `measure-module` with 039's hub, frame, export and `Scope` and
  the wave A contract merged; the backend builds

## Phase 2: Foundational

- [X] T002 [P] `backend/internal/measure/gsaplink/model.go`: the API types exactly as `openapi.yaml`'s
  `GsapLinkReport`, `GsapPath`, `GsapSwbsTab`, `GsapSwbsRow`, `GsapBand`, `GsapWaitingTab`, `GsapWaitingRow`,
  `GsapOrdersTab`, `GsapOrderRow`, `GsapOrderTotals`; tab and sort keys as constants
- [X] T003 [P] `backend/internal/measure/gsaplink/ports.go`: `Transactor`; `Scoped` (`Scope(ctx, a, f, columns)`
  returning 039's `measure.Scope`); `Store` with `ActivitiesApprovedAt(ctx, activityIDs)`,
  `WaitingForGsap(ctx, lineItemIDs)`, `PurchaseOrderTransactions(ctx, lineItemIDs)` and their row types
- [X] T004 `backend/internal/platform/database/queries/measure_gsaplink.sql` (research R2): `ListActivitiesApprovedAt`
  (latest `approval_rounds.ended_at` with `subject_kind = 'activity'` and `outcome = 'approved'` per activity),
  `ListPurchaseRequestsWaitingForGsap` (032 FR-032's predicate: status `approved`, a routing, a line without a PO
  line or with a PO value or currency different from the line; the time and type of the latest `original` or
  `modification` routing; lines on the given line items), `ListPurchaseOrderTransactions` (`purchaseOrder`,
  `invoice`, `payment` transactions linked to the given line items, with the purchase request line of the same PO
  number and line item); `go run github.com/sqlc-dev/sqlc/cmd/sqlc@v1.31.1 generate`
- [X] T005 `backend/internal/measure/gsaplink/store_postgres.go`: the three reads through sqlc; and
  `store_postgres_test.go` on the seeded demo data (`dbtest`): `TestStore_ScopeIsolation` (only the IDs asked for
  come back), `TestWaiting_EqualsPurchaseRequestsView_SC002` (the same requests as `purchaserequest`'s Waiting for
  GSAP list), `TestStore_LatestApprovalAndRouting`
- [X] T006 `backend/internal/measure/gsaplink/year_test.go`: an in-memory scope (nodes with line items and SWBS
  status, choices with finance plans) and fakes of `Scoped` and `Store` for the rule tests

## Phase 3: User Story 1 — SWBS not received (P1) 🎯 MVP

- [X] T007 [P] [US1] `waiting_test.go`: `TestWaitingHours_FR006` (hours since approval, never negative, over 48),
  `TestWorkingDays_FR009` (Monday to Friday, Friday to Monday is 1, same day 0)
- [X] T008 [P] [US1] `swbs_test.go`: `TestSwbs_ReadyToSendOnly_FR005` (received, not requested and waiting for
  approval left out; an activity In progress still in), `TestSwbs_Row_FR005` (path, owner, company code, approved
  on), `TestSwbsBands_FR007` (bands add up to Waiting; Over 48 hours card), `TestSwbs_EqualsAllWork_SC002`
- [X] T009 [P] [US1] `table_test.go`: `TestTable_SortAndPage_FR015` (every sort key both ways, ties by ID, default
  longest waiting first, page and page size, total over all pages)
- [X] T010 [US1] `waiting.go`, `swbs.go`, `table.go`: the rules of T007 – T009
- [X] T011 [US1] `service.go`: `Report(ctx, a, f, Table)` in one `Read` — `Scoped.Scope`, the tab, sort and page;
  `TestReport_Swbs_FR001_FR005`
- [X] T012 [US1] `handler.go`: `GET /api/v1/reports/gsap-link` with 039's filter and `tab`, `sort`, `page`,
  `pageSize` (allowlisted, 400 otherwise); `routes_test.go` on the handler, with `TestRoutes_ReadOnly_FR002` (only
  GET is mounted)
- [X] T013 [US1] `backend/internal/app/measure_gsaplink.go`: `registerGsapLink` registers the report and mounts the
  handler (adapters over `w.reports` and the store); remove `GET /api/v1/reports/gsap-link` from `notYetServed`;
  contract tests in `gsaplink/handler_test.go`: `TestGsapLink_Contract`, `TestGsapLink_NoReportsView_FR003`,
  `TestReport_OnlyPlansInScope_FR003`
- [ ] T014 [US1] `frontend/src/features/measure/gsap-link/`: `api.ts` (tab, sort, page in the address),
  `GsapLinkReport.tsx` (replaces 039's stub; inside `ReportFrame`), `GsapTabs.tsx`, `SwbsTab.tsx` (cards, the
  Waiting by time widget, the paged sortable table; a row opens its line item, FR-016), `PagedTable.tsx`,
  `gsap-text.ts`
- [ ] T015 [US1] Web tests with vitest-axe: `GsapLinkReport.test.tsx` (tab, sort and page in the address; the
  states of FR-018), `SwbsTab.test.tsx` (the chart's table view, SC-006; the row link)
- [ ] T016 [US1] `frontend/e2e/measure-gsap-link.spec.ts`: US1 on the demo data with axe

## Phase 4: User Story 2 — Waiting for GSAP (P2)

- [ ] T017 [P] [US2] `requests_test.go`: `TestWaiting_Row_FR008`, `TestWaiting_ResendDoesNotRestart_FR009`,
  `TestWaiting_Modification_FR008`, `TestWaiting_AmountOnlyWithFinanceView_FR004` (amount null and "Covers N of M"),
  `TestWaiting_Cards_FR010`
- [ ] T018 [US2] `requests.go` and the tab in `service.go`
- [ ] T019 [US2] Web: `WaitingTab.tsx` (a row opens its purchase request, FR-016) and its test; e2e US2 step in
  `measure-gsap-link.spec.ts`

## Phase 5: User Story 3 — Purchase orders per line item (P2)

- [ ] T020 [P] [US3] `orders_test.go`: `TestOrders_SumsPerLineItemAndPO_FR012` (credit notes negative, unlinked
  out), `TestOrders_PoOnTwoLineItems_FR011`, `TestOrders_TotalsAllPages_FR013`,
  `TestOrders_OnlyFinanceViewPlans_FR004`, `TestAmounts_InReportCurrency_FR014`; and in `store_postgres_test.go`
  `TestOrders_EqualLineItemFinancials_SC003` on the demo data
- [ ] T021 [US3] `orders.go` and the tab in `service.go`
- [ ] T022 [US3] Web: `OrdersTab.tsx` (totals row; "You can't see financial data"; a row opens its line item) and
  its test

## Phase 6: User Story 4 — Export (P2)

- [ ] T023 [P] [US4] `sheets_test.go`: `TestSheets_EveryRow_FR019` (the four widgets in order, every row, hidden
  amounts empty, totals), `TestSheets_EqualTables_SC005` (each sheet equals its tab's table over all pages)
- [ ] T024 [US4] `sheets.go`: `measure.Report` (`Key`, `Available`, `Sheets`)
- [ ] T025 [US4] e2e US4 step: a tab as CSV and Excel, the report's Export

## Phase 7: Polish

- [ ] T026 `BenchmarkGsapLink_SC001` in `gsaplink/bench_test.go` (research R6); result in plan.md
- [ ] T027 Comment budget and file length of every file of this feature; lint and format
- [ ] T028 `/speckit-converge`; spec status Built; pull request into `measure-module` with the feature-map status

## Dependencies

T001 → T002 – T006 → US1 (T007 – T016) → US2 → US3 → US4 → Polish. Tests before implementation in each story.
