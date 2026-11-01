# Tasks: Revenue and Funnel-Stage Reports

**Spec**: [spec.md](./spec.md) · **Plan**: [plan.md](./plan.md). Tests come first in every story; `[P]`
marks tasks that can run in parallel.

## Phase 1 — Wave B contract and shared setup

- [ ] T001 Write the wave B contract's report half in `backend/api/openapi.yaml`: the four operations,
  schemas and parameters per [contracts/api.md](./contracts/api.md); the four route rules in
  `backend/internal/access/routes_measure.go`; delete this group's lines from
  `contract_routes_test.go`'s `notYetServed`. The one step this lane touches those files (plan R6)
- [ ] T002 [P] `pnpm run api:types`; contract test scaffolding for the four operations
- [ ] T003 [P] `queries/measure_revenue.sql`: `ReportFactTotals`, per-activity rows filtered by
  `activity_in_scope` and the plan year (research R2); `sqlc generate && sqlc vet`
- [ ] T004 [P] `measure/revenue/ports.go`: `Transactor`, `Scoper`, `Goals` (activities' goal facts + their
  plan's model), `Categories`; the `Goals` adapter in `backend/internal/app/measure_revenue.go` (research R3)

## Phase 2 — US1 Revenue against goal (P1)

- [ ] T010 Tests first: SC-001's hand-worked table for one plan, one campaign, one activity — goal, actual,
  projected, the split at today (FR-004, FR-022, FR-023)
- [ ] T011 `facts.go`: the roll-up from `ReportFactTotals` into per-activity month series (FR-021)
- [ ] T012 `figures.go`: project per activity via `projection.Project` and sum (FR-022, 049-C3); Performance
  = (actual + projected) ÷ goal against the activities' goals (FR-023, 049-C16)
- [ ] T013 `periods.go`: months, and months summed into quarters with the same in-progress rule (FR-003,
  049-C4) — tests under a fixed clock
- [ ] T014 `ranked.go` / `ranked_query.go` / `ranked_groups.go`: 040's rules with Revenue's figures (research
  R5) — two-level drill, trail, groups with totals, paging across groups, sortable by actual/projected
  revenue (FR-005, FR-006, FR-012)
- [ ] T015 `actualCost`/`roi` always "—" with the attribution-not-configured reason (FR-013, 049-C6) — test
  pinning the reason text
- [ ] T016 `cards.go`: the fixed Revenue card set — goal, actual, projected, performance (FR-010, 049-C9)
- [ ] T017 `report.go` + `handler.go`: `KeyRevenue` registration, `GET /reports/revenue`,
  `GET /reports/revenue/ranked`, info tip text (FR-014); wire `measure_revenue.go`
- [ ] T018 [P] `sheets.go`: export widgets cards/periods/ranked for Revenue
- [ ] T019 [P] Frontend `RevenueReport.tsx`: compose `ReportFrame` + cards + `ReportTable` (periods) +
  `RankedList`, mirroring `PlanBudgetReport.tsx`'s shape; `api.ts` for the report and ranked reads
- [ ] T020 Empty-state reasons (FR-008, 049-C20): no goals, unapproved activities, no data loaded — backend
  states which, frontend shows it
- [ ] T021 Gate: lint, tests, budgets, isolation test for the scoped read (FR-002, SC-004)

## Phase 3 — US2 Funnel stage against goal (P1)

- [ ] T030 Tests first: stage selector lists the union of allowed stages (FR-015); conversion = stage
  actual ÷ previous stage actual, "—" on zero denominator (FR-017); uncovered-plan count (FR-018)
- [ ] T031 Stage selector and `stage` parameter, default first allowed stage (FR-015, 049-C11); deals
  excluded from the selector (049-C12)
- [ ] T032 Conversion figure reusing 046's panel rule (FR-017, 049-C13)
- [ ] T033 `cards.go`: the Funnel card set — goal, actual, projected, performance for the selected stage,
  plus conversion (FR-010 equivalent, 049-C9)
- [ ] T034 `report.go` + `handler.go`: `KeyFunnel` registration, `GET /reports/funnel`,
  `GET /reports/funnel/ranked`; wire `measure_funnel.go`
- [ ] T035 [P] Frontend `FunnelReport.tsx`: stage selector + cards + series + ranked list
- [ ] T036 Gate: lint, tests, budgets

## Phase 4 — US3 Activity category grouping (P2)

- [ ] T040 Tests first: `category` parameter narrows cards/series/ranked (FR-019); Uncategorised loses no
  activity, All categories counts each once (FR-020); saved views keep the parameter (049-C10)
- [ ] T041 Category parameter on both reports, read via `settings.ListActivityCategories` (047)
- [ ] T042 [P] Frontend category selector on both report pages
- [ ] T043 Gate: lint, tests, budgets

## Phase 5 — Cross-cutting and close-out

- [ ] T050 [P] SC-003: a test pinning the statement count constant whatever the number of activities in view
- [ ] T051 [P] Cross-folder asks recorded for feature 039's lane: the export helper's hardcoded sort keys
  (049-C19), Expand all/Collapse all (FR-068, 049-C19), Compare (FR-067, 049-C7)
- [ ] T052 [P] `frontend/e2e/measure-revenue.spec.ts`: US1 and US2 against a seeded database, axe included
  (Constitution VI, the gap 047's converge found)
- [ ] T053 Check every FR against a passing test named with it; comment budgets; file sizes
- [ ] T054 Run `/speckit-analyze`; fix every CRITICAL before done
- [ ] T055 Run `/speckit-converge`; record remaining or "clean"
- [ ] T056 Open the PR into `measure-module` (note: branch stacked on 047's; retarget once PRs #62/#75
  merge), with the feature-map status change and the asks from T051
