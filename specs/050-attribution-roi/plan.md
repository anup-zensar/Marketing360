# Implementation Plan: Attribution and ROI

**Branch**: `feature/050-attribution-roi` | **Date**: 2026-11-01 | **Spec**: [spec.md](spec.md)

## Summary

050 adds the designated attribution model as a versioned, logged setting (Settings › Attribution, like 037's
Overspend control) and the crediting rules of the five models and ROI as plain functions in
`backend/internal/measure/attribution`, which 048 and 049 call. Reading 046's journeys and showing the figures on 048
and 049 come when 046 is merged and those reports are built (open point 050-Q1); the functions take journeys as
input so they are tested now.

## Technical Context

**Language/Version**: Go (pinned), TypeScript 5.9 / React 19 · **Storage**: one row in `attribution_settings`
(migration 00034, reserved for 050) · **Testing**: unit tests on fakes, contract tests, component tests; e2e in the
end-of-day batch · **Constraints**: money in minor units; figures defined once (measure-metrics §2)

## Constitution Check

| Principle | Check | Result |
|---|---|---|
| I Spec first | 050-C1, 050-C2 answered; follow-ups listed | Pass |
| IV Server is the source of truth | Credits, residuals and ROI worked out once in Go | Pass |
| V Simplicity | One package of pure functions plus 037's setting pattern; no new dependency | Pass |
| VI Meaningful tests | Every model's weights, the residual rule, single touch, late touches, unmatched touches, ROI cases, setting version and permission | Pass |
| VII Security and audit | Change needs Manage settings in the route table and the service; versioned; logged | Pass |
| IX Experience | Confirmation names the consequence; nothing greyed out without a reason | Pass |

## Project Structure

```text
backend/internal/measure/attribution/   model.go · credit.go · roi.go · store.go · store_postgres.go ·
                                        service.go · handler.go · kinds.go · attributiontest/ · *_test.go
backend/internal/platform/database/migrations/00034_attribution_settings.sql
backend/internal/platform/database/queries/measure_attribution.sql
backend/internal/app/measure_attribution.go
frontend/src/features/settings/attribution/   AttributionSettingsPage.tsx (+ test)
frontend/src/features/measure/shared/RoiValue.tsx (+ test)
```

## Business rules and the tests that prove them

| Rule | Test |
|---|---|
| The five models' weights (FR-003) | `TestCredit_FiveModels_050_FR003` |
| Residual to the largest weight, newest on ties; credited + unattributed = revenue (FR-005, SC-001) | `TestCredit_ResidualExact_FR005` |
| No journey → 100% to its activity; late touches get nothing; unmatched → Unattributed (FR-004, FR-006) | `TestCredit_SingleTouchLateAndUnmatched_FR004_FR006` |
| ROI, its basis and "—" (FR-007, FR-008) | `TestROI_FR007` |
| Setting: seed, change needs Manage settings, versioned, logged (FR-001, FR-002) | `TestSettings_FR001_FR002` |
| Contract of GET and PUT | `TestAttributionContract` |
