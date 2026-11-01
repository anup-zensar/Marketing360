# Data Model: Plan Budget Report

Nothing is stored. Amounts are worked out in USD and converted once to the frame's currency (D-220).

## Figures of a record in view

| Field | Definition |
|---|---|
| budget | containers: their own budget (plans, campaigns, initiatives); none for activities and line items |
| planned | activity: its planned cost; container: its in-view activities'; line item: its own |
| committed, actual | line items in view of plans whose finance the person may see, months of the year (R2); containers and activities add up their line items' |
| leftToPlan | budget − planned, when the record has a budget |
| unspent | budget − actual, when the record has a budget and actual is shown |
| overAllocated | planned > budget |

## Cards

`budget` (R3), `planned`, `committed`, `actual`, `budgetUsed` = planned ÷ budget, `leftToPlan`, `unspent`,
`budgetSpent` = actual ÷ budget, `notYetCommitted` = planned − committed, `committedShare` = committed ÷ planned.
Per cents are whole numbers; a zero denominator gives null ("—"). Committed, actual and their figures are null when the
person may see no plan's finance.

## Periods

`period` (month | quarter), `items[]`: `key` ("2026-03" or "2026-Q1"), `state` (actual | inProgress | planned),
`budget`, `planned`, `actual` (null after today's period), `budgetToDate`, `plannedToDate`, `actualToDate`.
Notes: `notPhased`, `noDates`, `otherYears`, `actualOtherYears` (null when zero).

## Ranked list

`level` (plan | campaign | initiative | activity), `sort`, `order`, `group`, `page`, `size`, `total` (top entries
after search), `trail[]` (`{id, level, name}` from the plan down to the drilled record), `notes[]`.
`groups[]`: `key`, `label` ("" for no grouping), `totals` (figures), `entries[]`.
Entry: `id`, `level` (… | lineItem), `name`, `planId`, `figures`, `children[]` (one level down; none below a line
item).
