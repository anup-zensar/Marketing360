# UI contract: Plan Budget Report (040)

| Part | What it shows |
|---|---|
| `PlanBudgetReport` | 039's `ReportFrame` with Period and Show (default Campaign); then the cards, the pacing chart and the ranked list |
| `BudgetCards` | Ten summary cards; "—" for null; Left to plan negative says *Over-allocated* |
| `PacingChart` | A `ReportWidget` (`periods`): budget, planned cost and actual per period as bars, today's period marked, notes below; table: each period with running totals and a totals row |
| `RankedList` (shared, exported) | A `ReportWidget` (`ranked`): Sort by, Order, Group by, page size and search; group headings with totals; each top entry with its children; a child opens as the new top with a trail and Back; each name links to its record; the sorted column marked; paging |

Words: Budget, Planned cost, Committed, Actual, Budget used, Left to plan, Unspent budget, Budget spent, Not yet
committed, % committed, Over-allocated, Sort by, Order (High to low, Low to high), Group by (None, Plan, Owner,
attribute names), Back.
