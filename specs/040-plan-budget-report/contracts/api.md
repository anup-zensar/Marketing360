# API contract: Plan Budget Report (040)

| Operation | Rule | Answers |
|---|---|---|
| `GET /api/v1/reports/plan-budget` | `need(authz.ReportsView)` | 200 `PlanBudgetReport`; 400 for a malformed frame filter |
| `GET /api/v1/reports/plan-budget/ranked` | `need(authz.ReportsView)` | 200 `RankedList`; 400 for a malformed filter or ranked parameter |
| `GET /api/v1/reports/plan-budget/export` | 039's shared export | widgets `cards`, `periods`, `ranked` |

## Query

Both take 039's frame parameters (`year`, `plan`, `planType`, `team`, `owner`, `activityType`, `activityStatus`,
`attr`, `period`, `show`, `currency`). `ranked` also takes `sort` (budget, planned, committed, actual, leftToPlan,
unspent), `order` (desc, asc), `group` (none, plan, owner, `attr:<id>`), `size` (10, 25, 50), `page` (≥ 1), `q`
(≤ 100 characters) and `drill` (a record ID).

## Schemas

- `PlanBudgetFigures`: `budget`, `planned`, `committed`, `actual`, `leftToPlan`, `unspent` — `Money` or null;
  `overAllocated` boolean.
- `PlanBudgetCards`: the figures plus `budgetUsed`, `budgetSpent`, `committedShare` (integer per cent or null) and
  `notYetCommitted` (`Money` or null).
- `PlanBudgetPeriod`: `key`, `state` (actual, inProgress, planned), `budget`, `planned`, `actual` (null after today),
  `budgetToDate`, `plannedToDate`, `actualToDate`.
- `PlanBudgetReport`: `cards`, `period` (month, quarter), `periods[]`, `notes` (`notPhased`, `noDates`, `otherYears`,
  `actualOtherYears`: `Money` or null).
- `RankedEntry`: `id`, `level` (plan, campaign, initiative, activity, lineItem), `name`, `planId`, `figures`,
  `children[]` (`RankedEntry` without children).
- `RankedGroup`: `key`, `label`, `totals` (`PlanBudgetFigures`), `entries[]`.
- `RankedList`: `level`, `sort`, `order`, `group`, `page`, `size`, `total`, `trail[]` (`id`, `level`, `name`),
  `groups[]`, `drillLost` (true when the drilled record is no longer in view).
