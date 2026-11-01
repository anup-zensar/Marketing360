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

- [ ] T040 Tests first, per outcome: name 1–80 unique ci (FR-018), one default (FR-020), type moves between
  categories (FR-019), bin + restore with assignments and in-use warning (FR-021), version_conflict, one
  log entry per change, view-only refusals (FR-023)
- [ ] T041 `settings` category record: model/service/read/handler files, kind registered, bin via
  `history.MoveToBin`; store (postgres + fakes)
- [ ] T042 Route rules in `internal/access/routes_activity_categories.go` — signedIn reads,
  settings.manage writes, fail closed (FR-024)
- [ ] T043 [P] Contract: the five operations in openapi.yaml + api:types; contract tests
- [ ] T044 [P] Frontend `settings/activity-categories/`: list (template A, status filter, counts), side
  panel form with type assignment and default toggle, view-only notice; tests per state; the Settings hub
  card in `settings-sections.ts` (additive shared edit, as 046's)
- [ ] T045 Serve each type's category for the reports and the activity-type page (FR-022) with its test
- [ ] T046 Gate: lint, tests, budgets

## Phase 5 — Cross-cutting and close-out

- [X] T050 [P] Cross-folder ask recorded for the Plan module: enforce allowed stages at goal entry (FR-005)
  — written up in [contracts/api.md](./contracts/api.md) under "Cross-feature notes for the integrator"
- [ ] T051 Check every FR against a passing test named with it; comment budgets; file sizes
- [ ] T052 Run `/speckit-analyze`; fix every CRITICAL before done
- [ ] T053 Run `/speckit-converge`; record remaining or "clean"
- [ ] T054 Open the PR into `measure-module` (note: branch stacked on 046's; retarget/rebase once PR #62
  merges), with the feature-map status change and the FR-005 ask
