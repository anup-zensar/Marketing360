# Implementation Plan: Projections

**Branch**: `feature/047-projections` · **Spec**: [spec.md](./spec.md) · **Depends**: 002 (models), 046
(facts, panel, Figures port). Research in [research.md](./research.md), schema in
[data-model.md](./data-model.md), operations in [contracts/api.md](./contracts/api.md).

## Summary

Three deliverables: (1) the model gains its two velocities and allowed goal stages in Settings; (2) a pure
projection engine turns an activity's goal, its model and today into monthly projected INQ, MQL and
revenue, served on 046's activity panel in its reserved fields; (3) activity categories as a Settings
record for 048/049 to group by. No report page (047-C11), no new permission keys, migration 00033.

## Technical Context

- **Language**: Go backend (settings + new `internal/projection` + performance panel), React/TS frontend
  (settings/models form extended, new settings/activity-categories, shared Projection parts).
- **Existing code this builds on**: `internal/settings` models (00006; rates in hundredths of a per cent
  via `money.Percent`, deal size in minor units, `settings.Waterfall`); `internal/activity` goals
  (`INQ`, `SetLaterStagesMyself`, `MQL`, `Revenue`, `StartDate`/`EndDate`/`Ongoing`, `goalsOf`);
  `internal/performance` panel (`StageFigure`/`RevenueFigure` with `projected` reserved, `GoalFacts`
  port, `StageTotals` months) and `internal/app/performance.go` (`performanceFigures` adapter);
  039's frame sums months into quarters, so months are the served grain (R1).
- **Clock**: the injected clock only; today is a UTC calendar date (047-C9).
- **Money**: `money.Percent.OfWhole` half-up for stage arithmetic, `money.Money` USD for revenue — no
  floats anywhere (constitution 9).

## Constitution Check

| Rule | How this feature keeps it |
|---|---|
| Spec first | Everything above traces to spec FRs and 047-C1–C11 |
| ≤300 lines, ≤5% comments, metrics | The engine splits phase/shift/revenue into small pure files |
| Handler→service→store; Gin in handlers only | Categories follow the settings package's existing shape |
| Server is the source of truth | Reports call `internal/projection` in Go; the web app renders |
| Contract first | contracts/api.md → openapi.yaml at the wave-B step (lane B's own) |
| Scope and fail-closed routes | Panel unchanged (scoped); categories: signedIn reads, settings.manage writes, rules in a new `routes_activity_categories.go` |
| Tests | SC-001 table drives the engine; property test for phasing (SC-005); each record rule per outcome; scope-isolation untouched (no new scoped read) |
| Money/dates | Integer minor units, `date.Date`, UTC |
| Records | Categories: kind, bin, one log entry per change, version_conflict |

## Project Structure

### Documentation (this feature)

```text
specs/047-projections/
├── spec.md · plan.md · research.md · data-model.md
├── contracts/api.md
└── tasks.md
```

### Source code

```text
backend/internal/projection/            # pure: no Gin, no SQL, no clock (R3, FR-015)
│   ├── model.go                        # Inputs, Series, MonthFigure, Stamp
│   ├── phase.go                        # day-exact phasing of the goal stage (FR-006)
│   ├── project.go                      # shift by velocity, rates half-up, revenue (FR-007/008/009)
│   └── split.go                        # actual/projected split at today (FR-010)
├── backend/internal/settings/          # 047's declared outside-folder touch (build guide §4.1)
│   ├── model.go + service_models.go    # velocities + allowed stages, validation, impact (FR-001–004)
│   ├── service_activity_categories.go · activity_categories_read.go · views_activity_categories.go
│   ├── handler_activity_categories.go · kinds_categories.go · store_postgres_categories.go
│   └── store (postgres + fakes) + queries/settings_categories.sql
├── backend/internal/performance/       # ports.go gains dates; panel fills projected (FR-016)
├── backend/internal/access/routes_activity_categories.go
├── backend/internal/platform/database/migrations/00033_projections.sql
├── frontend/src/features/settings/models/        # velocity + stages fields, example extended
├── frontend/src/features/settings/activity-categories/
└── frontend/src/features/performance/activity/   # the panel's projected column and model note
```

## Rules, and the tests that prove them

| # | Rule | Proof |
|---|---|---|
| R1 | Velocity 0–24 whole months, every fault at once (FR-001); the worked example shows the lag (FR-004) | settings model validation test; `WaterfallExample.test.tsx` |
| R2 | Allowed stages non-empty subset; narrowing warns via impact (FR-002/003) | impact + validation tests |
| R3 | Phasing is day-exact across the activity's months; evergreen/undated runs to year end (FR-006) | engine table + property test |
| R4 | Later stage = previous × rate half-up, shifted by velocity; drop past December (FR-007/008) | SC-001 table |
| R5 | Override activities project typed goals, no rates (FR-009) | engine test |
| R6 | Split at today; the in-progress month carries both (FR-010) | split test with fixed clock |
| R7 | Derived, never stored; stamp = model name + updatedAt (FR-011/012) | panel test: edit rate, re-read |
| R8 | No goal → no projection, reason noGoal (FR-013) | panel test |
| R9 | Actuals are 046's counted facts (FR-014) | panel test over the 046 fixture |
| R10 | Category record rules: unique ci name, one default, bin+restore with assignments (FR-018–021) | service tests per outcome |
| R11 | A type in at most one category; assignment moves it (FR-019) | service test |

## Queries, and their scope filter

Categories and model columns are organisation-wide Settings records (no scope, same as models today —
046-C15's logic). The panel's one scoped read (`ActivityTotals`) is unchanged and keeps its isolation
test; the engine adds **no queries** (SC-004): dates ride on the existing GoalFacts resolution.

## Permissions

| Action | Permission |
|---|---|
| Read models, categories, the panel | signed in (panel scope applied in the service, as 046 built it) |
| Edit model velocity/stages; manage categories | `settings.manage` (Table A unchanged, FR-023) |

## Contract

contracts/api.md: three additive model fields, five activity-categories operations, the panel's projected
fields and stamp. Applied to openapi.yaml at the wave-B contract step (lane B writes it), then
`pnpm run api:types`.

## Risks reported, not worked around

- The projection default is adopted per 047-C4 while 016 FR-036 stays "open with the vendor" — SC-001's
  table is the artefact to re-sign if the vendor answer differs.
- Enforcement of allowed stages at goal entry is a cross-folder ask to the Plan module (FR-005).
- 050 will swap the actual series' source (attribution); the engine's input shape keeps that a one-line
  adapter change (FR-014).

## Complexity Tracking

Nothing outside the constitution's acceptable bands is planned; any justified exception lands in the PR
description per GO_STANDARDS §14.
