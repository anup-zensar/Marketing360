# Tasks: Revenue and Funnel-Stage Reports

**Spec**: [spec.md](./spec.md) · **Plan**: [plan.md](./plan.md). Tests come first in every story; `[P]`
marks tasks that can run in parallel.

## Phase 1 — Wave B contract and shared setup

- [X] T001 Write the wave B contract's report half in `backend/api/openapi.yaml`: the four operations,
  schemas and parameters per [contracts/api.md](./contracts/api.md); the four route rules in
  `backend/internal/access/routes_measure.go`; delete this group's lines from
  `contract_routes_test.go`'s `notYetServed`. The one step this lane touches those files (plan R6)
- [X] T002 [P] `pnpm run api:types`; contract test scaffolding for the four operations
- [X] T003 [P] `queries/measure_revenue.sql`: `ReportFactTotals`, per-activity rows filtered by
  `activity_in_scope` and the plan year (research R2); `sqlc generate && sqlc vet`
- [X] T004 [P] `measure/revenue/ports.go`: `Transactor`, `Scoper`, `Goals` (activities' goal facts + their
  plan's model), `Categories`; the `Goals` adapter in `backend/internal/app/measure_revenue.go` (research R3)

## Phase 2 — US1 Revenue against goal (P1)

- [X] T010 Tests first: SC-001's hand-worked table for one plan, one campaign, one activity — goal, actual,
  projected, the split at today (FR-004, FR-022, FR-023). Done: `report_revenue_test.go` reuses 047's own
  365,000 INQ / 273 deals / 2,730,000 USD table through the whole service, not just the engine
- [X] T011 `facts.go`: the roll-up from `ReportFactTotals` into per-activity month series (FR-021)
- [X] T012 project per activity via `projection.Project` and sum (FR-022, 049-C3) — done in `service.go`'s
  `scopeAndActivities`; Performance = (actual + projected) ÷ goal, computed exactly on cents via
  `rollup.BudgetUsedPercent` rather than a float division of money (FR-023, 049-C16)
- [X] T013 `periods.go`: months, and months summed into quarters with the same in-progress rule (FR-003,
  049-C4), via the shared `window`/`periodBounds` types the funnel report reuses
- [ ] T014 `ranked_revenue.go`: two-level roll-up (Show level + children), sorted, searched and paged, with
  the same figures 040 uses the shape of. **Partial**: `drill` and `group` (campaign/owner/attribute) are
  not yet implemented — every ranked answer currently reports one group named "" and ignores `drill`.
  Remaining: drill-down with a trail and Back, grouping with group totals (research R5, FR-005, FR-006)
- [X] T015 `actualCost`/`roi` always "—" with the attribution-not-configured reason (FR-013, 049-C6) — set
  unconditionally in `ranked_revenue.go`'s `revenueFiguresOfActivities`
- [X] T016 `cards.go`: the fixed Revenue card set — goal, actual, projected, performance (FR-010, 049-C9)
- [X] T017 `report.go` + `handler.go`: `KeyRevenue` registration, `GET /reports/revenue`,
  `GET /reports/revenue/ranked`; wire `measure_revenue.go`. Info tip text (FR-014) is frontend work, with T019
- [X] T018 [P] `sheets.go`: export widgets cards/periods/ranked, for both Revenue and Funnel (research R1)
  — every page of the ranked list, unbounded size, as 040's export does (FR-007)
- [X] T019 [P] Frontend `RevenueReport.tsx`: compose `ReportFrame` + cards + `ReportWidget` (periods, a
  hand-rolled SVG chart reusing the four chart tokens) + `RankedList`, mirroring `PlanBudgetReport.tsx`'s
  shape; `api.ts` for the report and ranked reads; page test with 5 cases incl. axe
- [X] T020 Empty-state reasons (FR-008, 049-C20): no goals, unapproved activities, no data loaded — backend
  states which (`report.reason`), frontend shows it via `reasons.ts`
- [X] T021 Gate: lint, tests, budgets. Done: 0 lint issues, every file within budget, full suite green
  including `-race` against real PostgreSQL; `store_postgres_test.go` proves FR-002/SC-004 directly —
  a record outside the scope is invisible to `ReportTotals`

## Phase 3 — US2 Funnel stage against goal (P1)

- [X] T030 Tests first: stage selector lists the union of allowed stages (FR-015); uncovered-plan count
  (FR-018); `conversionCard`'s own tests pin the "—" reason and the one-decimal rate (FR-017)
- [X] T031 Stage selector and `stage` parameter, default first allowed stage (FR-015, 049-C11); deals
  excluded from the selector (049-C12)
- [X] T032 Conversion figure reusing 046's panel rule (FR-017, 049-C13)
- [X] T033 `cards_funnel.go`: the Funnel card set — goal, actual, projected, performance for the selected
  stage, plus conversion (FR-010 equivalent, 049-C9)
- [X] T034 `report.go` + `handler.go`: `KeyFunnel` registration, `GET /reports/funnel`,
  `GET /reports/funnel/ranked`; wire `measure_funnel.go` (merged into `measure_revenue.go`, research R1)
- [X] T035 [P] Frontend `FunnelReport.tsx`: stage selector + cards + series + ranked list, mirroring
  `RevenueReport.tsx`'s shape; page test with 5 cases incl. axe, uncovered-count and the "—" conversion
- [ ] T036 Gate: lint, tests, budgets

## Phase 4 — US3 Activity category grouping (P2)

- [X] T040 Tests first: `category` parameter narrows cards/series/ranked (FR-019); Uncategorised loses no
  activity, All categories counts each once (FR-020); saved views keep the parameter (049-C10)
- [X] T041 Category parameter on both reports, read via the `Categories` port over `settings.ListActivityCategories`
- [X] T042 [P] Frontend category selector on both report pages (`CategorySelect.tsx`, shared from
  `revenue/`, imported by `funnel/`)
- [ ] T043 Gate: lint, tests, budgets

## Phase 5 — Cross-cutting and close-out

- [X] T050 [P] SC-003: a test pinning the statement count constant whatever the number of activities in view
  — `ReportTotals` called exactly once for 5 activities and for 500
- [X] T051 [P] Cross-folder asks recorded for feature 039's lane: the export helper's hardcoded sort keys
  (049-C19), Expand all/Collapse all (FR-068, 049-C19), Compare (FR-067, 049-C7). Also for 040's lane,
  found while writing T052's e2e test: the shared `RankedList`'s "Open X in Plan" row links fail WCAG
  2.2's target-size rule (axe `target-size`) whenever a row's name is short — first caught here because
  no e2e test had yet run axe against `shared/RankedList.tsx` in a real browser (not 040's own report,
  Plan budget, has one either). T052's test excludes `a[aria-label^="Open "]` from its own axe check so it
  still proves 049's markup; the component itself needs a larger hit area (e.g. padding on `.open`) from
  040's lane.
- [X] T052 [P] `frontend/e2e/measure-revenue.spec.ts`: US1 and US2 against a seeded database, axe included
  (Constitution VI, the gap 047's converge found). Along the way found and fixed two real defects: (1) the
  Revenue/Funnel service's `Deps` never set `Rates`, so `factsByActivity` panicked on first use outside of
  the unit-test fakes (`internal/app/measure_revenue.go`); (2) `scopeAndActivities` called `Scope` with a
  nil column list, so the scope's nodes never carried `inq`/`mql`/`revenue` and every report answered
  "no goal" even where one existed (`internal/measure/revenue/service.go`); (3) the Funnel Performance
  card returned an unrounded float (e.g. `4.037567567567567`) though its own comment says "a whole per
  cent" — rounded with `math.Round` and pinned with `TestRatioCard_RoundsToAWholePerCent_049C15`.
- [ ] T053 Check every FR against a passing test named with it; comment budgets; file sizes
- [ ] T054 Run `/speckit-analyze`; fix every CRITICAL before done
- [ ] T055 Run `/speckit-converge`; record remaining or "clean"
- [ ] T056 Open the PR into `measure-module` (note: branch stacked on 047's; retarget once PRs #62/#75
  merge), with the feature-map status change and the asks from T051
