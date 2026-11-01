# Research: Plan Budget Report

## R1 — Where the records come from

- **Decision**: one call to `measure.Service.Scope(ctx, a, f, nil)` inside the service's `Read`; the records in view
  are `Scope.Matched` (plans, campaigns, initiatives, activities, line items); finance visibility is
  `Scope.Choices.SeesFinance` and `FinancePlans`, as 039's coverage note uses.
- **Rationale**: the frame, the export and this report then cover the same records (039 FR-007).
- **Alternatives**: reading All work again — two definitions of "in view".

## R2 — Committed and actual

- **Decision**: `financials.SpendByMonth(ctx, a, lineItemIDs)` for the line items in view of plans whose finance the
  person may see; committed and actual are the months of the report's year; other years are a note.
- **Rationale**: the same reader and rule as 036's dashboard (D-342, D-357), so SC-001 holds by construction.
- **Alternatives**: `Node.Finance` — it holds every year, so it would not match the dashboard.

## R3 — Budget once per branch (040-C2)

- **Decision**: a container is *whole* when it and every container below it are in view. A whole container counts its
  own budget (or, without one, its children's); a container partly in view counts its in-view children's. Initiatives
  are the lowest budget level, so an initiative in view always counts its own budget.
- **Rationale**: client 017 FR-005, applied at the levels that carry budgets.
- **Alternatives**: always the plan's budget — misleading under filters (rejected in 040-C2).

## R4 — Periods and the today split

- **Decision**: budget per month from `PhasedBudgetUSD` of the containers counted by R3 (a counted container without
  phasing goes to the *not phased* note); planned from `allwork.PlannedByActivity`; actual from R2; quarters add three
  months; today's period from the injected clock; actual is empty after it.
- **Rationale**: the phasing of 016 and the spread of 010/036; no new definition.

## R5 — The ranked list's parameters

- **Decision**: level from the frame's `show` (default `campaign`); own parameters `sort` (budget, planned,
  committed, actual, leftToPlan, unspent; default planned), `order` (desc, asc), `group` (none, plan, owner,
  `attr:<id>`), `size` (10, 25, 50), `page`, `q`, `drill`; the report implements `measure.Parameterised` so saved
  views keep all but `page`.
- **Rationale**: 039's address rules; one place lists the names.

## R6 — Groups with % splits

- **Decision**: an entry with a split value appears under each value with its figures times the share (D-270);
  groups sort by their total of the sort figure; paging counts top entries after grouping.

## R7 — Reuse by 049

- **Decision**: the web's `RankedList` takes columns and rows from the API and knows nothing about money; the Go side
  keeps ranking, grouping and paging in plain functions over entries, so 049 can call them with its own figures.
