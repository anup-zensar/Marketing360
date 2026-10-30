# Feature Specification: Approvals Report

**Feature Branch**: `feature/041-approvals-report`

**Created**: 2026-10-30

**Status**: Draft

**Input**: User description: "Approvals report of the Measure module (wave A, lane C, issue #34): the status mix
of approval rounds by record type (activities, purchase requests, modified purchase requests), the average days
to a decision, and an approvals grid (record, approver, level, submitted, decided, outcome, amount, comments)
with approver, status and record-type filters, search, paging and export. Read-only; plugs into 039's report
frame."

Cites: FA §E3.6; client 018 US1, US4 and FR-001–FR-024 (Approvals Dashboard); 017 (average days to a decision,
D-342); 031 (record types); 036 FR-016 (the same counts on the planning dashboards);
[measure-metrics.md](../../docs/product/measure-metrics.md) §1; the report frame of 039.

## Overview

Approvers' managers and planners can see approvals one by one on the Approvals page and the Monitoring tab. This
report is the governance view across **everything in the plans a person may see**: how many approval rounds are
in each status, how long decisions take, overall and per approver, and a searchable, exportable grid of every
single approval step. It lives under **Reports › Approvals** and uses the report frame of feature 039 — the
filter panel, Period, Show, display currency, saved views, table view and export of every widget — without
changing any of it. The report never approves, declines or recalls anything; that stays in the Approvals queue.

Hive9's Approvals Dashboard also covers plans, campaigns, initiatives, line items, accrual requests and custom
objects. SP&A One's approval engine approves three record types only (031), so the report offers those three.

## Clarifications

None yet. `/speckit-clarify` records its answers here as 041-C1, 041-C2 … (never in the decision log).

## User Scenarios & Testing *(mandatory)*

### User Story 1 - See approval workload and cycle time (Priority: P1)

A planning lead opens **Reports › Approvals** and sees, for one record type at a time, how many approval rounds
are in each status and each status's share of the total, and the average number of days from submission to a
decision, with the number of decisions behind it and a list per approver, slowest first. The record type is
chosen on the status widget (default **Activity**) and offers only the types that have rounds in the plans in
view. The report's filters (year, plans, Shell attributes, owner, team, approver) narrow every figure.

**Why this priority**: it answers the question the report exists for — is approval a bottleneck, and where.
The Approvals page shows one person's queue; nothing else shows workload and cycle time across plans.

**Independent Test**: Sign in as Alex Morgan (Administrator, all markets), open Reports › Approvals for the
demo year: the status widget lists each status with its count and share, the average widget shows days to one
decimal with its count, and the three numbers agree with the same plans' Approvals monitoring and with 036's
Approvals panel.

**Acceptance Scenarios**:

1. **Given** approval rounds in the plans in view, **When** the report opens, **Then** the status widget shows
   the record type **Activity** with the count of rounds in each current status and each status's share as a
   whole per cent, adding up to the total shown, and a table view lists the same figures.
2. **Given** the status widget, **When** the person chooses **Purchase request** or **Modified purchase
   request**, **Then** only that type's rounds are counted and the status words are that type's own.
3. **Given** completed decisions (approved or declined steps), **When** the report opens, **Then** the average
   widget shows the mean calendar days from submission to decision to one decimal with the number of decisions,
   and a list of approvers each with their own average and decisions, slowest first; pending, recalled,
   replaced, skipped and automatically approved steps are not counted (017).
4. **Given** the same plans, **Then** the average equals the one on Approvals › Monitoring and on 036's
   Approvals panel, because all three use 017's rule.
5. **Given** the filter panel, **When** the person narrows the year, plans, attributes, owner, team or approver
   and applies the change, **Then** both widgets and the grid recount for the narrowed set and the filters-applied
   summary names them.
6. **Given** no decision was made yet in the plans in view, **Then** the average shows "—" with the words "No
   decisions yet", never 0.
7. **Given** the display currency is changed, **Then** every amount on the page is shown in it with the currency
   named, at the current rate.

---

### User Story 2 - Find, inspect and export approval records (Priority: P2)

A planner or finance advisor looks through the approvals grid: one row per approver step per submission, with
the record, its owner, the approver, level, dates, outcome, amount and comments. They search, sort, filter by
approver, status and record type, page through the rows, read a long comment in full and export every matching
row.

**Why this priority**: the widgets say how things are; the grid answers "which ones" — needed to chase a
slow approver or to prepare an audit, but the report is useful without it.

**Independent Test**: Open the grid as Alex Morgan, filter by approver Carlos Mendes and status Approved, search
for a record name, then export: the file holds every matching row (not only the visible page) in the grid's
columns and order, and none of the rows belongs to a plan outside Alex's scope.

**Acceptance Scenarios**:

1. **Given** the report, **When** the grid shows, **Then** each row is one approver step of one submission with
   these columns in order: record ID, record name, record type, approver, decided on, outcome, submitted on,
   level, purchase order number (purchase requests), amount, owner, comments.
2. **Given** the grid, **When** the person types in the search box, **Then** it narrows to rows whose text
   columns contain the words, and the total count updates.
3. **Given** the grid, **When** the person sorts by a column, **Then** all rows (not only the page) are ordered
   by it, with the same order on every page; **When** they choose a page size of 5, 10, 25, 50 or 100 (default
   50), **Then** the page and the total change accordingly.
4. **Given** the filters **Approver**, **Status** (the outcome of the step) and **Record type**, **When** any is
   applied, **Then** the grid shows only matching rows; the Status choices are those of the record type chosen,
   or all of them when none is.
5. **Given** a long comment, **When** the person opens it, **Then** it is shown in full on the page without
   leaving the report.
6. **Given** a row, **Then** the record's name links to the record where the person may see it; nothing on the
   report approves, declines, recalls or delegates.
7. **Given** the grid or a widget, **When** the person exports as CSV or Excel, **Then** the file holds every
   matching row, the visible columns in their order, the display currency named, no helper columns, and
   nothing is stored.
8. **Given** a step with no decision yet, **Then** its **Decided on** and **Outcome** show "—" and the outcome
   reads as waiting, never as a date or a zero.

---

### Report states *(read-only: states, not create/edit/delete)*

These scenarios cover each state a widget or the grid can render; each has its own test.

1. **Data**: rounds exist in the plans in view → the widgets and grid show them (US1, US2).
2. **No plans in scope**: the person may see no plan of the chosen year → the page says "No plans you may see
   for <year>" and offers to change the year; no figures and no grid.
3. **No approvals yet**: plans exist but have no approval rounds → each widget says "No approvals yet" and what
   creates them ("Submit an activity or a purchase request for approval"); the grid is empty with the same words.
4. **No access to the report**: a person without `reports.view` cannot open the report (it is absent from
   Reports and the address answers as 039 defines); a person with it sees only plans in their scope, and a record
   outside it answers as missing.
5. **No financial access**: where the person may not see financial data of a plan (`finance.view`, D-301), the
   **Amount** column and amounts show "—" for that plan's rows, with the count of plans the figure leaves out;
   counts, statuses, dates and the average still show.
6. **Filters exclude everything**: the filters match no round → each widget and the grid say the filters exclude
   everything and offer a one-click reset.
7. **Loading and failure**: while loading each widget shows a placeholder of its own shape; when a figure cannot
   be loaded the widget names the problem and offers Try again, without blocking the other widgets.

### Edge Cases

- A purchase request sent back and resubmitted, or an activity whose approval restarted: every round and step
  counts once in its own status; a round replaced by a new submission shows as **Replaced by a new submission**
  and is left out of the average (031).
- Rounds that skipped a level or approved automatically (017): they appear as rows with that outcome, are counted
  in the status mix and are left out of the average.
- An approver who left or whose role changed: the row keeps the name they had; the Approver filter still lists
  them while rows exist.
- A decision made by a backup approver, a delegate or "on behalf" (017): the row names who decided, and the
  average counts it for that person.
- A plan with two teams (D-237): its rounds count once in totals and in each team's view, as 036 does.
- Dates are calendar dates (`YYYY-MM-DD`); two decisions on the day of submission count as 0.0 days, a decision
  the next day as 1.0.
- The grid holds many thousands of rows: paging, sorting and search stay quick and the export still holds every
  matching row.
- Filters in the address that name a record type, status or approver that does not exist: the report shows the
  "filters exclude everything" state and the reset, not an error.

## Requirements *(mandatory)*

### Functional Requirements

**Frame and access**

- **FR-001**: The report MUST appear as **Approvals** under Reports and use feature 039's report frame
  (filter panel, Period, Show, display currency, saved views, "as of" line, table view and export of every
  widget) without redefining any of it (client 018 FR-001, FR-004, FR-005, FR-008, FR-010).
- **FR-002**: The report MUST be read-only: it MUST NOT approve, decline, recall, delegate or change any record;
  the only thing it saves is a person's saved view through 039 (client 018 FR-002, FR-023).
- **FR-003**: The report MUST need `reports.view`, export MUST need `data.export`, and every figure and row MUST
  come only from plans in the signed-in person's scope; a record outside it MUST answer as missing. This
  replaces Hive9's "recorded for later, not enforced" (client 018 FR-011): SP&A One enforces roles and scopes
  (014).
- **FR-004**: Filters MUST follow 039's rules: they live in the page address, apply on **Update**, show a
  filters-applied summary and a one-click reset, and Back keeps them (client 018 FR-003, FR-004).
- **FR-005**: The Year filter MUST limit the report to rounds of records in plans of the selected years (client
  018 FR-016); there is no Period selector on this report because nothing is bucketed by period (client 018
  FR-013).

**Approval status widget** *(US1)*

- **FR-006**: The **Current approval status** widget MUST show, for the chosen record type, the count of rounds
  in each current status and each status's share as a whole per cent, plus the total, with a table view (client
  018 FR-012, FR-018; FA §E3.6.1).
- **FR-007**: The widget MUST have its own **Record type** selector offering **Activity** (default), **Purchase
  request** and **Modified purchase request**, listing only types that have approval rounds in the plans in
  view, and using each type's own status words (client 018 FR-015, FR-017; 031 FR-003).
- **FR-008**: The widget MUST also show how many rounds wait for a decision and since when the oldest waits,
  worked out as in 036 FR-016.

**Average widget** *(US1)*

- **FR-009**: The **Average days to a decision** widget MUST show the mean calendar days from submission to
  decision over completed steps (approved or declined), to one decimal, with the number of decisions, using
  017's rule exactly as the Approvals monitoring and 036's Approvals panel work it out (client 018 FR-019;
  measure-metrics §1).
- **FR-010**: The widget MUST list each approver with their own average and number of decisions, slowest first;
  pending, recalled, replaced, skipped and automatically approved steps MUST NOT be counted.
- **FR-011**: With no decision in scope the average MUST show "—" and say why, never zero (client 018 FR-006,
  FR-007).

**Approvals grid** *(US2)*

- **FR-012**: The **Approvals** grid MUST show one row per approver step per submission, over all three record
  types by default, with the columns of US2 scenario 1 (client 018 FR-020, FR-024; FA §E3.6.3).
- **FR-013**: The grid MUST offer a record-type filter of its own, independent of the widget's selector, listing
  the types present in scope (client 018 FR-017).
- **FR-014**: The grid MUST filter by **Approver** and **Status**, where Status offers each record type's own
  vocabulary and not one reduced list (client 018 FR-014, FR-015).
- **FR-015**: The grid MUST offer free-text search over its text columns, sorting by any column across all
  rows, page sizes 5, 10, 25, 50 and 100 (default 50) and the total count (client 018 FR-021).
- **FR-016**: A long comment MUST be readable in full without leaving the report (client 018 FR-022).
- **FR-017**: Each row's record name MUST link to the record where the person may open it; the report offers no
  action on a round (client 018 FR-023).

**Figures, money and states**

- **FR-018**: Amounts MUST be exact decimals with their currency in the display currency at the current rate
  (D-303), the amount of an activity round being the activity's planned cost and that of a purchase-request round
  the request's amount as 032 and 036 work it out (D-342); where the person may not see financial data of a plan,
  the amount MUST show "—" with the count of plans it leaves out (D-301, D-343).
- **FR-019**: A value that does not exist MUST show "—", never zero (client 018 FR-006).
- **FR-020**: Every empty widget and the grid MUST name the cause — no plans you may see, no approvals yet, you
  cannot see financial data, the filters exclude everything with a reset (client 018 FR-007; build guide §6).
- **FR-021**: Every figure MUST be computed on the server; the web app only draws what the API provides
  (constitution IV), and a total MUST equal the sum of its parts in the table view and in the export.
- **FR-022**: The report MUST render quickly and without a statement per row at the 028 volumes, with the
  benchmark recorded in the plan.

**Export**

- **FR-023**: The grid and each widget MUST export to CSV or Excel through 039's export: every matching row, the
  visible columns in their order, the display currency named, no helper columns; nothing is stored (client 018
  FR-010).
- **FR-024**: The report MUST NOT show Hive9's Vendor, accrual-request or custom-object filters, plan, campaign,
  initiative or line-item approvals, or the PO and attribute-group-name columns of other object types: SP&A One's
  engine does not approve them (031).

### Key Entities *(include if feature involves data)*

- **Approval round**: one run of a workflow over one record, with its record type, status, submitted date and
  outcome (008, 017, 031). Read, never changed.
- **Approval step**: an approver (or set of approvers) at one level of a round, with the decision date, outcome
  and comments; the grid's row.
- **Record type**: Activity, Purchase request or Modified purchase request (031).
- **Status mix**: the count and share of rounds per current status for one record type.
- **Average days to a decision**: the mean of decision date minus submitted date over completed steps (017).
- **Report widget**: a block of the report frame with a chart or card, a table view and an export (039).

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A planning lead sees the status mix and the average days to a decision for the demo year within
  one page view, with no filter set, and finds the slowest approver in under 30 seconds.
- **SC-002**: For the same plans, the waiting count, the oldest-waiting date and the average days are identical
  on this report, on Approvals › Monitoring and on 036's Approvals panel in 100% of checks.
- **SC-003**: A person finds one specific approval step by search, approver or status in under 1 minute.
- **SC-004**: The export holds 100% of the rows matching the filters and search — not only the visible page —
  and 0 rows from plans outside the person's scope.
- **SC-005**: With the demo data, 100% of the report's states (data, no plans, no approvals, no financial
  access, filters exclude everything, failure) render a message naming their cause; no figure is ever shown as
  zero when the real value is missing.
- **SC-006**: The report and its grid open and respond within 2 seconds at the 028 volumes and pass the
  accessibility checks (WCAG 2.2 AA, keyboard-only use of filters, selectors, grid and export).

## Assumptions

- Feature 039's report frame, filter model, saved views and export are available on `measure-module`; 041 plugs
  into them and never redefines them. Lane A's 039 spec was not yet pushed when this spec was written, so any
  detail of the frame named here follows the build guide and is checked against 039's contracts at planning.
- The three record types and their status words are those of 031; the status words of a type are read, not
  redefined here (see 041-Cn to come if `/speckit-clarify` finds a gap).
- "Decision" means an approved or declined step; calendar days between submission and decision, as 017's
  Monitoring average works it out. Working days are not used.
- The amount of an activity round is the activity's planned cost; of a purchase-request round, 032's amount.
- Hive9's accrual, plan, campaign, initiative, line-item and custom-object approvals are not built: SP&A One does
  not approve them (031; constitution 4.8.0).
- Showing when GSAP data was last received (client 018 FR-009) is not needed: no figure here comes from GSAP.
- No approval figure uses the simulated performance feed; this report is wave A only.
- Out of scope: approving, declining or recalling from the report, sharing or scheduling by e-mail, a new
  migration, and any change to `openapi.yaml`, the shared hub files or the decision log (the integrator applies
  those from this feature's `contracts/api.md` and pull request).
