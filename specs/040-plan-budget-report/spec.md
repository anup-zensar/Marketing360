# Feature Specification: Plan Budget Report — Spend Pacing and the Ranked List

**Feature Branch**: `feature/040-plan-budget-report`

**Created**: 2026-11-01

**Status**: Draft

**Input**: User description: "Plan budget report (lane A, issue #40): budget, planned cost, committed and actual by
period split at today, summary cards, and the ranked list (show plan, campaign, initiative or activity; group by;
sort; order; page size; drill with a trail; a link to the record), exported for reuse by 049." (feature map row 040)

**References**: Constitution 4.8 — Principles IV (the server works out every figure), VII (scope), IX, X, F3 (state
in the address), F6 (every chart has a table) · [measure-module-build.md](../../docs/sdd/measure-module-build.md)
row 040 · [measure-metrics.md](../../docs/product/measure-metrics.md) §1 · feature 039 (the report frame: filters,
In view, saved views, export, widgets) · 036 figures (D-342) · 010 Budget & cost (D-266) · 016 phased budgets ·
Shell client spec 017 (US1, US3; FR-001 – FR-011) and 016's ranked list (FR-063 – FR-069) · Hive9 FA §E3.4, §E3.2.5.

## Clarifications

### Session 2026-11-01

Answered by the product owner for lane A; the integrator moves them to the decision log at the end of wave A.

- **040-C1** Q: Does the ranked list go below Activity? → A: yes, to **line items**, as client 017 US3 asks; the
  revenue reports (049) stop at Activity as 016 FR-063 says, because line items carry no revenue.
- **040-C2** Q: When the filters keep only part of a plan, which budget counts? → A: client 017 FR-005: each budget
  counts **once per branch** — a parent's budget when all its children are in view, otherwise the budgets of its
  children in view.
- **040-C3** Q: How does the ranked list open? → A: sorted by **planned cost, largest first**.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - See how spend is pacing against the budget (Priority: P1)

Priya opens Plan budget for 2026. Cards show the budget, planned cost, committed and actual of her plans, with budget
used, left to plan, unspent budget and budget spent. A chart shows budget, planned cost and actual by month: budget
and planned cost run to December, actual stops at the current month, and today's month is marked. She switches to
quarters and the chart follows.

**Why this priority**: marketing leads review pacing every month; it is the first question of client 017 (US1).

**Independent Test**: open the report for a year with phased budgets, spread line items and GSAP payments, and check
the cards and every period against the plan page and the dashboards of 036.

**Acceptance Scenarios**:

1. **Given** Priya's 2026 plans, **When** she opens Plan budget, **Then** the cards show Budget, Planned cost,
   Committed, Actual, Budget used, Left to plan, Unspent budget, Budget spent, Not yet committed and % committed,
   each equal to the same figure on the plan page or the 036 dashboard for the same plans.
2. **Given** today is in November 2026, **When** the chart shows months, **Then** budget and planned cost have a
   value for every month, actual has values up to November and none after, and November is marked as today.
3. **Given** a plan's budget is not phased, **When** the chart is shown, **Then** that budget is left out of the
   periods and a note says how much budget is not phased; the cards still count it.
4. **Given** planned cost without dates or in another year, **When** the chart is shown, **Then** a note names the
   amounts *No dates* and *Other years* instead of dropping them silently.
5. **Given** a plan whose financial data Priya may not see, **When** the report is shown, **Then** committed, actual
   and the figures made from them cover only the plans she may see, and 039's coverage note says so.
6. **Given** the chart, **When** she chooses Table, **Then** every period is a row with budget, planned cost and
   actual, their running totals, and a totals row.

---

### User Story 2 - Find which plans, campaigns and activities drive the money (Priority: P1)

Under the chart a ranked list shows her campaigns, largest planned cost first, each with its initiatives below. She
opens an initiative: it moves to the top with its activities below, and a trail shows Plan › Campaign › Initiative.
She opens an activity and sees its line items. Back returns one level. A link opens any record in the Plan
workspace.

**Why this priority**: knowing *where* money is planned and spent is client 017's US3 and Hive9's ranked list.

**Independent Test**: with two campaigns, their initiatives, activities and line items, rank by each figure, drill
to line items and back, and open a record.

**Acceptance Scenarios**:

1. **Given** Show is Campaign, **When** the list opens, **Then** campaigns are ranked by planned cost, largest
   first, each listing its initiatives with their figures, and the sorted column is marked.
2. **Given** the list shows campaigns, **When** she opens an initiative, **Then** it becomes the top entry with its
   activities below, the trail shows its plan and campaign, and the filter panel is unchanged.
3. **Given** an activity is the top entry, **When** the list is shown, **Then** its line items are listed below it
   with their planned cost, committed and actual.
4. **Given** a drilled list, **When** she presses Back or a trail step, **Then** the list returns to that level; the
   browser's Back does the same because the drill lives in the address.
5. **Given** any entry, **When** she follows its link, **Then** the record opens in the Plan workspace.
6. **Given** a campaign whose planned cost is above its budget, **When** it is listed, **Then** it is marked
   *Over-allocated* in words, not only by colour.

---

### User Story 3 - Group, sort and page the list (Priority: P2)

She groups the list by owner, then by a Shell attribute such as Brand, sorts by actual, low to high, and shows 25
entries a page.

**Why this priority**: comparing owners and brands is the next question once pacing is known (FR-066 – FR-068).

**Independent Test**: group by each choice, sort by each figure in both orders and page through, checking group
totals and that nothing is counted twice.

**Acceptance Scenarios**:

1. **Given** Group by Owner, **When** the list is shown, **Then** entries sit under their owner's heading with the
   group's totals, groups ordered by their total of the sorted figure.
2. **Given** Group by a Shell attribute with a % split, **When** the list is shown, **Then** an entry's figures
   count in each value's group by its share, and entries without a value sit under *Not filled in*.
3. **Given** Group by None and no drill, **When** she adds up the top entries of every page, **Then** planned cost,
   committed and actual equal the cards.
4. **Given** 60 campaigns and page size 25, **When** she pages, **Then** pages hold 25, 25 and 10 entries and the
   page lives in the address.
5. **Given** a name typed in the list's search, **When** the list is shown, **Then** only top entries whose name
   contains it are listed.

---

### User Story 4 - Take the report away (Priority: P2)

She exports the report: the file holds the cards, the periods and the whole ranked list (every page, both levels)
with 039's In view and About this export sheets.

**Why this priority**: the figures go into monthly packs (client 017 US6, export only).

**Independent Test**: export with filters, grouping and a drill, and compare the file with the screen.

**Acceptance Scenarios**:

1. **Given** the report on screen, **When** she exports, **Then** the file has a Cards sheet, a Periods sheet and
   a Ranked list sheet whose rows and totals equal the screen without paging.
2. **Given** a widget's own export, **When** she exports the ranked list as CSV, **Then** the file holds the list
   as grouped, sorted and drilled on screen.

### Edge Cases

- A year without plans in scope shows 039's "No plans you may see" state; a year with plans and no line items shows
  zero planned cost, never a blank.
- A budget of zero gives "—" for Budget used and Budget spent, never a division error.
- Planned cost above budget makes Left to plan negative: it is shown with *Over-allocated*.
- An activity whose line items exceed its planned cost: the activity's own planned cost counts (domain model §4.2),
  its line items as listed.
- A plan in a currency other than the display currency converts at the current rate as 036 does (D-342).
- A record in the recycle bin is never counted.
- Actual posted outside the plan's year is not in the periods; a note names the amount, as 036's dashboard does.
- Drilling to a record the filters no longer match (an address from a saved view) falls back to the top level with a
  note.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The report MUST live at `/reports/plan-budget` inside 039's frame (filters, In view, saved views,
  export) with the Period control (months, quarters); it MUST read only plans in the person's reports.view scope.
- **FR-002**: Summary cards MUST show Budget, Planned cost, Committed, Actual, Budget used, Left to plan, Unspent
  budget, Budget spent, Not yet committed and % committed, each as [measure-metrics.md](../../docs/product/measure-metrics.md)
  §1 defines it, in the frame's currency, with no second definition of any figure.
- **FR-003**: Budget MUST count once per branch (040-C2): a plan's budget when all its campaigns are in view,
  otherwise the budgets of its campaigns in view, and likewise for initiatives.
- **FR-004**: The chart MUST show Budget (phased), Planned cost (by line-item spread, an activity's unassigned balance
  evenly over its months) and Actual (by posting month) for each month or quarter of the year; actual MUST stop at
  today's period and today's period MUST be marked (039 FR-019).
- **FR-005**: Budget not phased, planned cost with *No dates* or in *Other years*, and actual posted outside the year
  MUST be named in notes under the chart with their amounts.
- **FR-006**: Committed, Actual and every figure made from them MUST cover only plans whose financial data the person
  may see; the rest show "—" when nothing remains, with 039's coverage note.
- **FR-007**: The chart MUST have a table view (039 FR-020) with each period's budget, planned cost, actual, their
  running totals and a totals row.
- **FR-008**: The ranked list MUST show the Show level's entries (plan, campaign, initiative or activity) ranked by
  the chosen figure and order, each listing its children one level down; an activity's children are its line items
  (040-C1).
- **FR-009**: The figures to rank by MUST be Budget, Planned cost, Committed, Actual, Left to plan and Unspent budget;
  the list MUST open sorted by Planned cost, largest first (040-C3), and mark the sorted column.
- **FR-010**: Opening a child MUST make it the top entry with its children below, with a trail of its ancestors and
  Back; drilling MUST NOT change the filter panel and MUST live in the address.
- **FR-011**: Every entry MUST link to its record in the Plan workspace.
- **FR-012**: An entry whose planned cost is above its budget MUST say *Over-allocated* in words.
- **FR-013**: Group by MUST offer None, Plan, Owner and each choice Shell attribute of the Show level or above;
  groups MUST carry totals and be ordered by their total of the sorted figure; a % split counts by share (D-270) and
  records without a value group under *Not filled in*.
- **FR-014**: The list MUST page its top entries (10, 25 or 50, default 10) and search them by name; page, size,
  search, group, sort, order and drill MUST live in the address; saved views keep all of them but the page (039).
- **FR-015**: With Group by None and no search or drill, the top entries' planned cost, committed and actual over all
  pages MUST equal the cards.
- **FR-016**: The export MUST hold a Cards sheet, a Periods sheet and a Ranked list sheet (every page, both levels,
  as grouped, sorted and drilled) besides 039's sheets; each widget MUST export on its own as CSV and Excel.
- **FR-017**: The ranked list MUST be built as a part 049 reuses with its own figures (revenue, stages).
- **FR-018**: The report is read-only; it MUST store nothing but the person's saved views (039).

### Key Entities

- **Card figures**: the ten figures of FR-002 for the records in view.
- **Period**: a month or quarter of the year with budget, planned cost, actual and their running totals, and whether
  it is past, today's or future.
- **Ranked entry**: a plan, campaign, initiative, activity or line item with its figures, its owner, its link, its
  children and, when grouped, its group.
- **Drill trail**: the ancestors of the top entry, from the plan down.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: For the same plans and year, every card equals the plan page's and the 036 dashboard's figure to the
  cent.
- **SC-002**: The report and its ranked list are shown within 2 seconds on the 028 volume.
- **SC-003**: With Group by None, the ranked list's totals equal the cards for every Show level (FR-015).
- **SC-004**: From the report, a person reaches the line items behind the largest campaign in at most three clicks.
- **SC-005**: An exported file's figures equal the screen for every widget.

## Assumptions

- Budgets are annual with optional monthly phasing (016); MAPP has no forecast series, so none is shown (client 017
  FR-011's open point stays open).
- Actual at monthly FX is out of scope: MAPP keeps current rates only (D-342), so amounts convert at the current rate
  as 036 does.
- *Use planned cost as actual cost* belongs to 050; this report always shows actual from GSAP payments.
- Finance budgets (client 017 US2, US4, US5) and sharing by e-mail are out of scope (constitution 4.8.0).
