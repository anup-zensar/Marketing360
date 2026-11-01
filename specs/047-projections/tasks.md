# Tasks: Projections

**Spec**: [spec.md](./spec.md) · **Plan**: [plan.md](./plan.md). Tests come first in every story; `[P]`
marks tasks that can run in parallel. The migrations are the truth the documents describe.

## Phase 1 — Schema and shared setup

- [X] T001 Migration `00033_projections.sql`: the three `models` columns with their CHECKs and backfills,
  `activity_categories`, `activity_category_types`, the case-insensitive name index and the single-default
  partial unique index, per [data-model.md](./data-model.md); applies to an empty database
- [X] T002 Queries `settings_categories.sql` (list/get/put/delete, assignments) and the models queries
  extended with the new columns; `sqlc generate && sqlc vet` clean
- [X] T003 [P] Extend `performance.GoalFacts` with `StartDate`, `EndDate`, `Ongoing` and fill them in
  `internal/app/performance.go`'s adapter (R3) — 046's port, lane B's own file

## Phase 2 — US1 velocity and allowed stages on the model (P1)

- [X] T010 Tests first: velocity 0–24 whole months, every fault at once (FR-001); allowed stages non-empty
  subset, default all three (FR-002); stale version refused (FR-003); audit entry per change (FR-003)
- [X] T011 `settings.Model` gains `InqToMqlVelocityMonths`, `MqlToDealVelocityMonths`, `AllowedGoalStages`;
  parse/validate in `service_models.go`; store round-trip (postgres + both fakes)
- [X] T012 Impact check names goals stranded by a narrowed stage set (FR-003, 047-C3) with its test
- [X] T013 [P] Contract: the three additive `Model`/`ModelCreate`/`ModelUpdate` fields in openapi.yaml
  (lane B's wave-B step) + `pnpm run api:types`; contract test covers the round-trip (FR-004)
- [X] T014 [P] Frontend `ModelForm`: two velocity fields and the allowed-stages checkboxes, faults together;
  the live waterfall example shows the lag (FR-004); form + page tests per state
- [X] T015 Gate: lint clean, package tests green, comment/line budgets kept

## Phase 3 — US2 the projection engine and the panel (P1)

- [X] T020 `internal/projection` types: Inputs (goal facts + dates, model parameters, today, plan year),
  Series/MonthFigure, Stamp (FR-015)
- [X] T021 Tests first: SC-001's hand-calculated table (domain model §4.5 example, velocities 1 and 2,
  mid-year today) driving `Project` end to end (FR-006–FR-010)
- [X] T022 `phase.go`: day-exact phasing across the activity's months; evergreen/undated to year end
  (FR-006) with edge-case tests (start the 20th, start after today, year with no overlap)
- [X] T023 `project.go`: rate half-up per month, velocity shift, drop past December, revenue = deals × deal
  size (FR-007/008); override activities use typed goals (FR-009, 047-C6)
- [X] T024 `split.go`: actual months before today, projected from today, the in-progress month carries both
  (FR-010) under a fixed clock
- [X] T025 Property test: any goal/dates/velocities → series non-negative, inside the year, reconciling per
  SC-003 (SC-005)
- [X] T026 `performance` panel fills the reserved fields: projected per stage, projected revenue, months,
  stamp, reason noGoal (FR-012/013/016); derived-not-stored proven by edit-rate-and-re-read (FR-011, SC-002)
- [X] T027 Actuals feeding the split are 046's counted facts (FR-014) — test over the 046 fixture; no new
  queries (SC-004) pinned by the counter wrapper pattern from 046
- [X] T028 [P] Contract: panel's projected fields and `projection` stamp in openapi.yaml + api:types;
  contract test (FR-016)
- [X] T029 [P] Frontend `measure/shared/Projection…` parts + the activity panel showing projected beside
  goal and actual with the reasons; component tests per state (data, noGoal, not approved)
- [X] T030 Gate: lint, tests, budgets

## Phase 4 — US3 activity categories (P2)

- [X] T040 Tests first, per outcome: name 1–80 unique ci (FR-018), one default (FR-020), type moves between
  categories (FR-019), bin + restore with assignments and in-use warning (FR-021), version_conflict, one
  log entry per change, view-only refusals (FR-023)
- [X] T041 `settings` category record: model/service/read/handler files, kind registered, bin via
  `history.MoveToBin`; store (postgres + fakes)
- [X] T042 Route rules in `internal/access/routes_activity_categories.go` — signedIn reads,
  settings.manage writes, fail closed (FR-024)
- [X] T043 [P] Contract: the five operations in openapi.yaml + api:types; contract tests
- [X] T044 [P] Frontend `settings/activity-categories/`: list (template A, status filter, counts), side
  panel form with type assignment and default toggle, view-only notice; tests per state; the Settings hub
  card in `settings-sections.ts` (additive shared edit, as 046's)
- [X] T045 Serve each type's category for the reports and the activity-type page (FR-022) with its test
- [X] T046 Gate: lint, tests, budgets

## Phase 5 — Cross-cutting and close-out

- [X] T050 [P] Cross-folder ask recorded for the Plan module: enforce allowed stages at goal entry (FR-005)
  — written up in [contracts/api.md](./contracts/api.md) under "Cross-feature notes for the integrator"
- [X] T051 Check every FR against a passing test named with it; comment budgets; file sizes. Done: every FR
  has a named test except FR-005 (a cross-folder ask, nothing to test here) and FR-024, which the shared
  access-matrix test `TestAccessMatrix_EveryOperationRefusesWhomItsRuleExcludes_SC001` proves for every
  operation in the contract, the five new ones included. FR-022 was found unserved during the audit and
  implemented: an activity type now names its category, as it already names the models that allow it
- [X] T052 Run `/speckit-analyze`; fix every CRITICAL before done. Done 2026-11-02: two lenses
  (spec-vs-code, documents-vs-truth) agreed on three CRITICALs — FR-004's worked example ignored the
  velocities, the split marked January of a future plan year in progress, and FR-010's combined plan-year
  figure was never served — plus FR-013's reason being swallowed by an actual. All four fixed in code; the
  spec's own contradictions (US2's flat 100 a month, FR-012 against the stamp, SC-003's exact
  reconciliation) and the stale plan/research/contract wording fixed alongside
- [X] T053 Run `/speckit-converge`; record remaining or "clean". Done 2026-11-02: eight findings appended
  as Phase 6. The one CRITICAL is this feature's own — no end-to-end test exists for US1 and US2, which
  Constitution VI requires of every P1 story
- [ ] T054 Open the PR into `measure-module` (note: branch stacked on 046's; retarget/rebase once PR #62
  merges), with the feature-map status change and the FR-005 ask

## Phase 6: Convergence

- [X] T055 Write `frontend/e2e/projections.spec.ts` covering **US1 and US2**, this feature's P1 stories,
  with axe: an administrator sets a model's velocity and allowed stages and sees the lag in the worked
  example; an approved activity with a goal shows its projected figures, the months split at today and the
  model behind them. Per Constitution VI ("an end-to-end browser test for every P1 user story") (missing).
  Done 2026-11-02: both pass against the real API and a seeded database, axe included. The journey proved
  the stranded-goal warning end to end, in its own words: "1 plan set their goal at a stage this model
  would no longer allow. They keep it and it keeps projecting; new goals can't choose it." 
- [ ] T056 Extend `frontend/e2e/projections.spec.ts` to **US3**: create an activity category, assign a type
  that is already in another, mark it the default, delete it and undo. Needs the integrator's two route
  lines first (contracts/api.md, ask 1) per US3 (missing)
- [ ] T057 Note the activity types a restore takes back from other categories, as a released default is
  already noted, per FR-021 and 047-C14 (partial)
- [ ] T058 Convert or refuse a currency mismatch in `projection.SplitAmounts` instead of stamping the
  projected currency onto converted actuals, per FR-007 and 047-C6 (missing)
- [ ] T059 Decide whether a goal of exactly zero is a goal, record it as a clarification, and make
  `hasGoal` agree, per FR-013 (partial)
- [ ] T060 Show each activity type's category on the Activity types Settings page, per FR-022 (missing)
- [ ] T061 [integrator] Fold the goal-entry ask into `docs/sdd/decision-log.md` at wave end — a lane never
  edits it — per FR-005 (missing)
- [ ] T062 [lane D] Seed activity categories with the wave B demo data, so the list does not open empty,
  per plan: demo data (missing)
