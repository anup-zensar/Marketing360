# API contract: Projections

Every operation below goes into `backend/api/openapi.yaml` by lane B at the wave-B contract step, with
`pnpm run api:types` after. All paths under `/api/v1`. Reads marked *signed in* apply scope in the service
(FR-023); writes need `settings.manage`.

## Model (extended — feature 002's operations, additive fields)

The `Model` schema gains `inqToMqlVelocityMonths` (integer 0–24), `mqlToDealVelocityMonths` (integer
0–24) and `allowedGoalStages` (non-empty array of `inq`·`mql`·`revenue`). `ModelCreate`/`ModelUpdate`
accept them optionally — defaults 0, 0 and all three — so existing clients keep working. No path changes;
`listModels`/`createModel`/`getModel`/`updateModel`/`checkModelImpact`/`setModelStatus` stand. The impact
answer gains the goals stranded by a narrowed `allowedGoalStages` (FR-003).

## Activity categories (new, Settings)

| Operation | Path | Permission |
|---|---|---|
| listActivityCategories | GET `/settings/activity-categories` | signed in |
| createActivityCategory | POST `/settings/activity-categories` | settings.manage |
| getActivityCategory | GET `/settings/activity-categories/{categoryId}` | signed in |
| updateActivityCategory | PATCH `/settings/activity-categories/{categoryId}` | settings.manage |
| deleteActivityCategory | DELETE `/settings/activity-categories/{categoryId}` | settings.manage (recycle bin) |

A category **answers** with `name`, `isDefault`, `active`, `activityTypes` (`SettingRef[]`, sorted by
name), `usage` and the record envelope; create and update **take** `activityTypeIds` instead (each type is
in at most one category, so sending it here moves it). Validation
errors all at once; duplicate name ignoring case refused; version_conflict on stale saves; delete answers
the bin entry for Undo.

## Activity panel (extended — 046's getActivityPerformance, additive fields)

`StageFigure` gains `projected` — the **plan-year figure**: actual to the end of the last whole month
before today plus projected from today, so a closed year equals the actual and `actual` and `projected`
must never be added (FR-010) — plus `projectedReason` (`""` | `noGoal`, which `reason` cannot carry when
the stage has an actual) and `months` (month, `actual`, `projected`, `inProgress`). `RevenueFigure` gains
the same three. The panel gains `projection`: the model's name and `modelChangedAt` (047-C8, 020 FR-052).
046's rules (approval gate, scope, reasons vocabulary) are unchanged. Only the stages the panel shows —
INQ and MQL — carry a projected figure; deals stay internal to the engine.

`ActivityType` gains a required `category` (`SettingRef` or null): the category grouping that type, so the
reports and the Activity types page can read it without inverting the category list (FR-022).

## Not in this contract

No projection endpoint for reports: 048–050 call `internal/projection` in Go. No report routes (047-C11).
No new permission keys (FR-023).

## Cross-feature notes for the integrator

1. **Enforcing the allowed goal stages at goal entry** (FA §E7.5.1, FR-005, 047-C3) belongs to the Plan
   module's goal screens (features 004 and 006), not here. 047 stores the configuration, serves it on the
   model, and warns before a narrowing save strands an existing goal; refusing a *new* goal at a stage the
   plan's model does not allow is a one-line check on those forms, against `model.allowedGoalStages`.
   Nothing in 047 rewrites or invalidates a goal already set.
2. **The three new `Model` fields are additive and defaulted** (`inqToMqlVelocityMonths`,
   `mqlToDealVelocityMonths`, `allowedGoalStages`), so every existing client keeps working; migration 00033
   backfills zero lag and all three stages, which is exactly what those models meant before.
3. **The panel's `projected` fields and `projection` stamp** fill what 046's contract reserved. 048–050
   read the engine in Go (`internal/projection`), not over HTTP: there is no projection endpoint.
