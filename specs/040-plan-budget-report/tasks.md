# Tasks: Plan Budget Report (040)

**Input**: [spec.md](spec.md), [plan.md](plan.md), [research.md](research.md), [data-model.md](data-model.md),
[contracts/](contracts/). Heavy runs (e2e, `apibench`) wait for the end-of-day batch (laptop limits); every task
runs its own unit tests, lint and type-check.

## Phase 1: Contract

- [X] T001 `backend/api/openapi.yaml`: `GET /api/v1/reports/plan-budget` and `/ranked` with the schemas of
  contracts/api.md; rules `need(authz.ReportsView)` in `access/routes_measure.go`; `pnpm run api:types`

## Phase 2: User Story 1 — spend pacing (P1)

- [X] T002 [US1] Tests first in `backend/internal/measure/planbudget/`: `TestBudget_OncePerBranch_040_FR003`,
  `TestCards_FR002`, `TestCards_FinanceOnlyVisiblePlans_FR006`, `TestPeriods_FR004_FR007`,
  `TestPeriods_TodayAndNotes_FR004_FR005`
- [X] T003 [US1] `model.go`, `ports.go`, `figures.go`, `cards.go`, `periods.go`, `service.go` (`Report`)
- [X] T004 [US1] `handler.go` (`GET /reports/plan-budget`), `app/measure_planbudget.go` wiring, contract test
  `TestPlanBudgetContract`
- [X] T005 [US1] Web: `features/measure/plan-budget/api.ts`, `BudgetCards.tsx`, `PacingChart.tsx`,
  `PlanBudgetReport.tsx` with tests (cards, "—", Over-allocated, chart and table, notes, loading, error)

## Phase 3: User Story 2 — ranked list and drill (P1)

- [X] T006 [US2] Tests first: `TestRanked_LevelsAndSort_FR008_FR009`, `TestRanked_Drill_FR010`,
  `TestRanked_TotalsEqualCards_FR015`, `TestParseRanked_FR014`
- [X] T007 [US2] `ranked.go`, `ranked_query.go`, service `Ranked`, handler `GET /reports/plan-budget/ranked`;
  the report implements `measure.Parameterised`
- [X] T008 [US2] Web: `shared/RankedList.tsx` (exported from `features/measure` for 049) with drill, trail, Back,
  record links; tests

## Phase 4: User Story 3 — group, sort, page (P2)

- [X] T009 [US3] Tests first: `TestRanked_Groups_FR013`, `TestRanked_PagesAndSearch_FR014`
- [X] T010 [US3] `ranked_groups.go`, paging and search; web controls (Sort by, Order, Group by, page size, search)

## Phase 5: User Story 4 — export (P2)

- [X] T011 [US4] `sheets.go`: Cards, Periods and Ranked list sheets through 039's export; `TestSheets_FR016`

## Phase 6: Polish (end-of-day batch)

- [X] T012 `frontend/e2e/measure-plan-budget.spec.ts` (US1, US2) with axe
- [X] T013 Plan budget cases in `BenchmarkReportsFrame`; record in plan.md (SC-002)
- [ ] T014 Comment budget and length of every touched file; quickstart by hand; spec status Built
