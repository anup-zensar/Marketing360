# Tasks: Purchase Request Report

**Input**: [plan.md](plan.md), [spec.md](spec.md), [contracts/api.md](contracts/api.md). Tests first (constitution VI);
only the end-to-end test uses a real database. Lane rules as 041: own folders only, plus the contract, route rule
and `notYetServed` line that lane A delegated (T013).

## Phase 1: Contract and skeleton

- [ ] T001 `backend/api/openapi.yaml`: `getPurchaseRequestsReport` and its schemas and parameters from `contracts/api.md`; `routes_measure.go` rule `need(authz.ReportsView)`; the `notYetServed` line; `pnpm run api:types`
- [ ] T002 `purchasereport/model.go` (API types), `ports.go` (`Transactor`, `Scoper`, `Purchases`, `Deps`), `purchasereporttest/` fakes with a fixture: requests in all six statuses, two vendors and one without, owners, lines in several months and one without a delivery date, PO lines with and without invoices, a plan without `finance.view`

## Phase 2: User Story 1 — commitment by vendor and owner by delivery month (P1)

- [ ] T003 [P] [US1] Tests: `TestStatusMix_042_FR007`, `TestCommitment_042_FR009_C1`, `TestMonths_042_C2`, `TestTopTenAndOther_042_FR011`, `TestNoVendor_042_FR012`, `TestChartsAddUpToTheCard_042_FR013`, `TestNoFinanceView_042_FR020`, `TestFilters_042_FR004_C6`
- [ ] T004 [P] [US1] Tests: contract `TestPurchaseRequestsReport_Contract_042`, `TestPurchaseRequestsReport_EqualsDashboard_042_SC002`
- [ ] T005 [US1] `query.go`, `months.go`, `charts.go`, `service.go`, `handler.go`, `report.go`; `app/measure_purchasereport.go` adapter; delete the `notYetServed` line
- [ ] T006 [P] [US1] Web tests: `CardsRow`, `StatusWidget`, `CommitmentChart` per state (data, loading, no purchase requests, nothing committed, no financial access, error), `schemas.test.ts`
- [ ] T007 [US1] Web: `schemas.ts`, `api.ts`, `CardsRow.tsx`, `StatusWidget.tsx`, `CommitmentChart.tsx` (SVG bars, chart tokens), `PurchaseRequestsReport.tsx`
- [ ] T008 [US1] `frontend/e2e/measure-purchase-requests.spec.ts` US1 with axe
- [ ] T009 [US1] Checkpoint: quality checks, comment budget, file length; commit

## Phase 3: User Story 2 — the hierarchical grid (P2)

- [ ] T010 [P] [US2] Tests: `TestGridHierarchy_042_FR014`, `TestGridSearch_042_FR017`, `TestGridSortAndPaging_042_FR017`, `TestInvoiced_042_FR019_C5`, `TestParseQuery_042_FR016`
- [ ] T011 [US2] `rows.go`, `grid.go`, the grid part of `service.go`; adapter invoices and workflows
- [ ] T012 [P] [US2] Web tests: `RequestsGrid` (vendor › request › line › line item, open and close, column chooser, search, sort, paging, states)
- [ ] T013 [US2] Web: `RequestsGrid.tsx`, `ColumnChooser.tsx`, address parameters
- [ ] T014 [US2] Checkpoint; commit

## Phase 4: User Story 3 — export (P2)

- [ ] T015 [P] [US3] Test `TestSheets_042_FR026_FR027`
- [ ] T016 [US3] `sheets.go`; e2e export scenario
- [ ] T017 [US3] Checkpoint; commit

## Phase 5: Polish

- [ ] T018 `BenchmarkPurchaseRequestsReport` (needs PostgreSQL to run)
- [ ] T019 Comment budget and file length of every file touched
- [ ] T020 Contract problems found while building, written in the pull request for lane A
- [ ] T021 `/speckit-converge`; spec status Built; merge `measure-module`; pull request with "042 Built"

## Coverage of the requirements

| Requirement | Tasks |
|---|---|
| FR-001 to FR-005 | T001, T005, T007 |
| FR-006 to FR-013 | T003, T005, T007 |
| FR-014 to FR-019 | T010, T011, T013 |
| FR-020 to FR-025 | T003, T005, T018 |
| FR-026 to FR-028 | T015, T016 |
