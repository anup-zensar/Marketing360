# Tasks: Attribution and ROI (050)

**Input**: [spec.md](spec.md), [plan.md](plan.md). Heavy runs wait for the end-of-day batch.

## Phase 1: Crediting and ROI (US2, US3)

- [X] T001 [US2] Tests first: `TestCredit_FiveModels_050_FR003`, `TestCredit_ResidualExact_FR005`,
  `TestCredit_SingleTouchLateAndUnmatched_FR004_FR006`; [US3] `TestROI_FR007`
- [X] T002 `backend/internal/measure/attribution/model.go`, `credit.go`, `roi.go`

## Phase 2: The designated model (US1)

- [X] T003 Migration `00034_attribution_settings.sql`, `queries/measure_attribution.sql`, sqlc; store, PostgreSQL store,
  fake, store set
- [X] T004 Tests first `TestSettings_FR001_FR002`; `service.go`, `kinds.go` (logged, never removed), `handler.go`;
  `openapi.yaml`; route rules; `history_access` (Manage settings); wiring `app/measure_attribution.go`;
  `TestAttributionContract`; `pnpm run api:types`
- [X] T005 Web: `features/measure/attribution/AttributionSettingsPage.tsx` (in Measure, which lane A owns, not the
  Plan module's Settings feature) with the confirmation; Settings section and
  route; tests

## Phase 3: Labels for 048 and 049 (US3)

- [X] T006 `features/measure/shared/RoiValue.tsx`: ROI with its model and cost basis, "—" with the reason; tests

## Phase 4: After 046 merges (blocked: 050-Q1)

- [ ] T007 Read 046's revenue facts and journeys into `attribution.Conversion`; attributed revenue for 048 and 049
- [ ] T008 e2e `frontend/e2e/attribution.spec.ts`; restated figures after a model change (US4, SC-002)
