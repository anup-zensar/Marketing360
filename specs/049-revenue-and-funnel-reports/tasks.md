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
- [X] T036 Gate: lint, tests, budgets. Superseded by T021/T053's whole-package runs: 0 lint issues, full
  `-race` suite green, every file within budget.

## Phase 4 — US3 Activity category grouping (P2)

- [X] T040 Tests first: `category` parameter narrows cards/series/ranked (FR-019); Uncategorised loses no
  activity, All categories counts each once (FR-020); saved views keep the parameter (049-C10)
- [X] T041 Category parameter on both reports, read via the `Categories` port over `settings.ListActivityCategories`
- [X] T042 [P] Frontend category selector on both report pages (`CategorySelect.tsx`, shared from
  `revenue/`, imported by `funnel/`)
- [X] T043 Gate: lint, tests, budgets. Superseded by T021/T053's whole-package runs: 0 lint issues, full
  `-race` suite green, every file within budget.

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
- [X] T053 Check every FR against a passing test named with it; comment budgets; file sizes. Found and
  fixed: FR-014 was wholly unbuilt (no info tip on the Revenue goal card) — implemented and tested; real
  test gaps closed for FR-003 (a quarter sums its three months), FR-018 (a model that types only an INQ
  goal is uncovered for MQL even while another plan's model still offers it), FR-022 (two activities
  project separately then sum, not their INQ first), FR-023 on the Revenue side (`performanceOf` rounding,
  mirroring Funnel's own test), and FR-024 (a fact in a currency with no rate is excluded from the total,
  not zeroed into it). Retrofitted FR naming onto three tests that already proved a requirement without
  naming it (FR-015, FR-004 x2, FR-008). FR-025 (no new permission) is proven in `internal/access`, outside
  this lane's files, not re-tested here. Comment budgets: trimmed `measure_revenue.go` (9 → 7),
  `measure-revenue.spec.ts` (9 → 3) and `cards_funnel_internal_test.go` (2 → 1) to fit. File sizes: every
  file in scope was already well under 300 non-blank lines. Writing SC-002's test (the ranked list's
  ungrouped top-level totals must equal the report's own series totals) surfaced a real, pre-existing
  defect worse than any single FR gap: `revenueFiguresOfActivities` and `funnelFiguresOfActivities` never
  set `Projected` or `Performance` on any ranked entry at any level — both fields were silently nil/zero
  for every row of both reports, so sorting by "projected" (FR-006) sorted nothing, and FR-012/FR-017's
  promised per-entry figures never reached a reader. Fixed in `ranked_revenue.go`/`ranked_funnel.go` by
  computing `Projected` via the same actual-or-projected split the cards use (`projection.SplitAmounts`/
  `SplitCounts`), and deriving `Performance` once, after rollup, from `Projected ÷ Goal` — not summed
  per-activity, since percentages don't sum. Pinned with `TestRankedRevenue_TopLevelTotalsMatchTheSeriesTotals_SC002`,
  `TestRankedFunnel_EntriesCarryProjectedAndPerformance_FR012_FR017`, and two tightened e2e assertions
  that would have failed before the fix. Per-entry **Conversion** (FR-017) is a known remaining gap of the
  same shape — never set on any ranked entry — recorded as follow-up work, not fixed here, since a
  correct rollup needs each level's own previous-stage actual, not just its own figures summed.
- [X] T054 Run `/speckit-analyze`; fix every CRITICAL before done. No CRITICAL or HIGH findings. Two
  MEDIUM/LOW gaps found and closed in the same pass: SC-002 had no test
  (`TestRankedRevenue_TopLevelTotalsMatchTheSeriesTotals_SC002`, which caught the Projected/Performance
  defect above) and SC-006 was only half-tested (added a `ModelChangedAt` assertion to the existing FR-012
  test). T014's drill/group gap was reconfirmed as a disclosed, not hidden, gap.
- [ ] T057 [P] Set `Conversion` on every ranked Funnel entry, not just the report's own cards (FR-017):
  needs each level's own previous-stage actual rolled up alongside its own, not just summed from entries'
  own figures. Found while fixing T053/T054's Projected/Performance defect; same shape, left for its own
  pass rather than widening this one further.
- [X] T055 Run `/speckit-converge`; record remaining or "clean". Converged: no gap found that isn't
  already a traceable, open task (T014, T057 remain, both pre-existing and honestly disclosed).
- [X] T056 Open the PR into `measure-module` (note: branch stacked on 047's; retarget once PRs #62/#75
  merge), with the feature-map status change and the asks from T051. Opened as PR #82 against
  `feature/047-projections` (047's own PR #75's base chain); retarget to `measure-module` once #62/#75
  merge.
