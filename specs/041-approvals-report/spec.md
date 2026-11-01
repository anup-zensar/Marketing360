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

### Session 2026-10-30

- **041-C1** — Q: Does "status" in the status mix and the grid's Status filter mean the outcome of the approval
  round or step, or the status of the record being approved? → A: **The outcome of the round or step**, in one
  vocabulary for all three record types: Waiting, Approved, Declined, Recalled, Replaced by a new submission,
  Approved automatically. Reason: the report is about approval rounds, and its other figures (waiting, decided,
  average days) already work on outcomes; a record's own status (Created, Submitted, In review, Closed …) is shown
  by the plan and purchase-request pages. Applied to FR-006, FR-007, FR-014 and US1 scenario 2.

- **041-C2** — Q: When a level can be decided by any one of several approvers, or by all, does the grid show one
  row per level or one row per approver at that level? → A: **One row per named approver at each level** (client
  018 FR-020). Where any one may decide, the other approvers' rows read **Not needed** once someone has decided;
  **Not needed** is a seventh outcome word that exists only on grid rows, not in the status mix (which counts
  rounds) and not in the average. The person who actually decided is credited with the decision, including a
  backup, delegate or on-behalf decider. Applied to FR-010, FR-012, FR-014 and US2 scenarios 1, 4 and 8.

- **041-C3** — Q: Should a workflow level skipped because its condition did not apply appear in the grid, and
  under what outcome? → A: **No: skipped levels are not rows and are not counted.** Only levels with a named
  approver and levels the engine approved automatically (017) appear; an automatic level is one row with the
  approver shown as "—", the outcome **Approved automatically**, and the submission date as the decision date.
  Reason: a skipped level has no approver and no decision. Applied to the edge cases, US1 scenario 3 and FR-010.

- **041-C4** — Q: Should an activity round's amount be the activity's planned cost or its budget? → A: **The
  activity's planned cost as it is now**, in the display currency; a purchase-request or modified-purchase-request
  round shows the request's amount as 032 and 036 work it out. No snapshot at submission is kept. Reason: it is
  what the activity page and 036's Approvals panel show, so the report never disagrees with the record. Applied
  to FR-018 and the assumptions.

- **041-C5** — Q: Where does a row's record name link when the person may see the approval row but cannot open
  the record's page? → A: **Nowhere: the name is plain text, with no link, whenever the person may not open the
  record.** Reason: the report never offers a link that ends in "not found", and a record outside the person's
  scope still answers as missing. Applied to FR-017 and US2 scenario 6.

- **041-C6** — Q: How do the engine's outcomes `restarted` and `cancelled`, which 041-C1 has no word for, appear?
  → A: **`restarted` shows as Replaced by a new submission and `cancelled` as Recalled**, so the status mix always
  adds up to the total. Reason: a restart ends a round the same way a re-submission does (031), and a cancelled
  round was withdrawn before a decision, as a recall is. Applied to FR-006 and the assumptions. (Planning, from
  reading the engine's `Outcome` values; research R3.)
- **041-C7** — Q: Is the average days worked out per round or per approver step? → A: **The overall average is
  017's, per round**: the mean elapsed time from submission to the end of every round approved or declined by a
  person, to one decimal (`approval.AverageDecisionDays`), the same as Approvals › Monitoring and 036's panel, so
  the three agree (SC-002). **The per-approver list** gives each person the mean elapsed time from a round's
  submission to that person's own decision, over the decisions they made. Reason: 017's rule is per round; client
  018 FR-019 wants a per-approver list too, so both are shown with their own count. Applied to FR-009, FR-010,
  US1 scenario 3, the edge cases, the key entities and the assumptions. (Planning; research R1.)

- **041-C8** — Q: How do the frame's activity-level filters (type, status, attributes, owner) narrow purchase
  request rounds? → A: **A purchase request or modified purchase request is in view when one of its line items is
  in view**, as 036 does, so every filter of the frame reaches it through its line items. Applied to FR-005 and
  US1 scenario 5. (Analysis; research R4.)
- **041-C9** — Q: Does the approver filter narrow the widgets? → A: **No: it is a filter of the grid only.** The
  widgets stay the whole picture of the plans in view, and the per-approver list already gives each person's
  figures. Applied to US1's introduction and scenario 5 and to US2 scenario 4. (Analysis.)
- **041-C10** — Q: What does the waiting figure cover, and what do unknown values do? → A: **The waiting figure
  is the selected record type's**, and the three types add up to 036's waiting count; Approvals › Monitoring
  covers rounds ended this month only, so it is not the comparison for the totals (SC-002). **An unknown record
  type, outcome or approver, in the widget or the grid, answers 200 with empty figures**, never an error; only a
  malformed `sort`, `order`, `page`, `pageSize` or frame filter is a 400. Applied to FR-008, SC-002, the edge
  cases and `contracts/api.md`. (Analysis.)

## User Scenarios & Testing *(mandatory)*

### User Story 1 - See approval workload and cycle time (Priority: P1)

A planning lead opens **Reports › Approvals** and sees, for one record type at a time, how many approval rounds
are in each status and each status's share of the total, and the average number of days from submission to a
decision, with the number of decisions behind it and a list per approver, slowest first. The record type is
chosen on the status widget (default **Activity**) and offers only the types that have rounds in the plans in
view. The report's filters (year, plans, Shell attributes, owner, team) narrow every figure; the grid has an
approver filter of its own (041-C9).

**Why this priority**: it answers the question the report exists for — is approval a bottleneck, and where.
The Approvals page shows one person's queue; nothing else shows workload and cycle time across plans.

**Independent Test**: Sign in as Alex Morgan (Administrator, all markets), open Reports › Approvals for the
demo year: the status widget lists each status with its count and share, the average widget shows days to one
decimal with its count, and the waiting count (summed over the three record types), the oldest waiting date and
the average days equal 036's Approvals panel for the same plans.

**Acceptance Scenarios**:

1. **Given** approval rounds in the plans in view, **When** the report opens, **Then** the status widget shows
   the record type **Activity** with the count of rounds in each current status and each status's share as a
   whole per cent, adding up to the total shown, and a table view lists the same figures.
2. **Given** the status widget, **When** the person chooses **Purchase request** or **Modified purchase
   request**, **Then** only that type's rounds are counted, with the same outcome words for every type (041-C1).
3. **Given** completed decisions (approved or declined steps), **When** the report opens, **Then** the average
   widget shows the mean days from submission to the end of the round to one decimal with the number of
   decisions (041-C7), and a list of approvers each with their own average and decisions, slowest first; pending, recalled,
   replaced and automatically approved steps are not counted (017), and skipped levels are not rows (041-C3).
4. **Given** the same plans, **Then** the waiting count (summed over the three record types), the oldest-waiting
   date and the average equal 036's Approvals panel, which counts all rounds of the plans in view; Approvals ›
   Monitoring uses the same average rule for rounds ended this month only (041-C10).
5. **Given** the filter panel, **When** the person narrows the year, plans, attributes, owner or team and applies
   the change, **Then** both widgets and the grid recount for the narrowed set and the filters-applied summary
   names them; purchase requests narrow through their line items (041-C8).
6. **Given** no decision was made yet in the plans in view, **Then** the average shows "—" with the words "No
   decisions yet", never 0.

---

### User Story 2 - Find, inspect and export approval records (Priority: P2)

A planner or finance advisor looks through the approvals grid: one row per named approver at each level of each
submission, with the record, its owner, the approver, level, dates, outcome, amount and comments. They search, sort, filter by
approver, status and record type, page through the rows, read a long comment in full and export every matching
row.

**Why this priority**: the widgets say how things are; the grid answers "which ones" — needed to chase a
slow approver or to prepare an audit, but the report is useful without it.

**Independent Test**: Open the grid as Alex Morgan, filter by approver Carlos Mendes and status Approved, search
for a record name, then export: the file holds every matching row (not only the visible page) in the grid's
columns and order, and none of the rows belongs to a plan outside Alex's scope.

**Acceptance Scenarios**:

1. **Given** the report, **When** the grid shows, **Then** each row is one named approver at one level of one
   submission (041-C2) with these columns in order: record ID, record name, record type, approver, decided on, outcome, submitted on,
   level, purchase order number (purchase requests), amount, owner, comments.
2. **Given** the grid, **When** the person types in the search box, **Then** it narrows to rows whose text
   columns contain the words, and the total count updates.
3. **Given** the grid, **When** the person sorts by a column, **Then** all rows (not only the page) are ordered
   by it, with the same order on every page; **When** they choose a page size of 5, 10, 25, 50 or 100 (default
   50), **Then** the page and the total change accordingly.
4. **Given** the filters **Approver**, **Status** (the outcome of the step) and **Record type**, **When** any is
   applied, **Then** the grid shows only matching rows; the Status choices are the six outcome words of 041-C1
   plus **Not needed** (041-C2), whatever record type is chosen; the Approver filter narrows the grid only
   (041-C9).
5. **Given** a long comment, **When** the person opens it, **Then** it is shown in full on the page without
   leaving the report.
6. **Given** a row, **Then** the record's name links to the record when the person may open it and is plain text
   when not (041-C5); nothing on the report approves, declines, recalls or delegates.
7. **Given** the grid or a widget, **When** the person exports as CSV or Excel, **Then** the file holds every
   matching row, the visible columns in their order, the display currency named, no helper columns, and
   nothing is stored.
8. **Given** a step with no decision yet, **Then** its **Decided on** shows "—" and its **Outcome** reads
   **Waiting**, never a date or a zero; **Given** a level where any one approver may decide and one did,
   **Then** the other approvers' rows read **Not needed** with "—" for **Decided on** (041-C2).
9. **Given** the display currency is changed, **Then** every amount in the grid is shown in it with the currency
   named, at the current rate.

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
- Rounds that approved automatically (017): one row, approver "—", outcome **Approved automatically**; they are
  counted in the status mix and left out of the average. Levels skipped by a condition are not rows (041-C3).
- An approver who left or whose role changed: the row keeps the name they had; the Approver filter still lists
  them while rows exist.
- A decision made by a backup approver, a delegate or "on behalf" (017): the row names who decided, and the
  average counts it for that person.
- A plan with two teams (D-237): its rounds count once in totals and in each team's view, as 036 does.
- Dates on screen are calendar dates (`YYYY-MM-DD`); averages are elapsed days to one decimal (041-C7), so a
  round decided within the hour of its submission counts as 0.0 days.
- The grid holds many thousands of rows: paging, sorting and search stay quick and the export still holds every
  matching row.
- Filters in the address that name a record type (the status widget's or the grid's), outcome or approver that
  does not exist: the report answers with empty figures and shows the "filters exclude everything" state and the
  reset, not an error (041-C10).

## Requirements *(mandatory)*

### Functional Requirements

**Frame and access**

- **FR-001**: The report MUST appear as **Approvals** under Reports and use feature 039's report frame
  (filter panel, display currency, saved views, "as of" line, table view and export of every widget; Period and
  Show only where a report has time widgets or a ranked list, so not here, FR-005) without redefining any of it (client 018 FR-001, FR-004, FR-005, FR-008, FR-010).
- **FR-002**: The report MUST be read-only: it MUST NOT approve, decline, recall, delegate or change any record;
  the only thing it saves is a person's saved view through 039 (client 018 FR-002, FR-023).
- **FR-003**: The report MUST need `reports.view`, export MUST need `data.export`, and every figure and row MUST
  come only from plans in the signed-in person's scope; a record outside it MUST answer as missing. This
  replaces Hive9's "recorded for later, not enforced" (client 018 FR-011): SP&A One enforces roles and scopes
  (014).
- **FR-004**: Filters MUST follow 039's rules: they live in the page address, apply on **Apply** (039-C1), show a
  filters-applied summary and a one-click reset, and Back keeps them (client 018 FR-003, FR-004).
- **FR-005**: The Year filter MUST limit the report to rounds of records in plans of the selected years (client
  018 FR-016); there is no Period selector on this report because nothing is bucketed by period (client 018
  FR-013). A purchase request or modified purchase request is in view when one of its line items is, as in 036,
  so the frame's filters narrow it through its line items (041-C8).

**Approval status widget** *(US1)*

- **FR-006**: The **Current approval status** widget MUST show, for the chosen record type, the count of rounds
  in each current outcome — Waiting, Approved, Declined, Recalled, Replaced by a new submission, Approved
  automatically (041-C1, 041-C6) — and each one's share as a whole per cent, plus the total, with a table view (client
  018 FR-012, FR-018; FA §E3.6.1).
- **FR-007**: The widget MUST have its own **Record type** selector offering **Activity** (default), **Purchase
  request** and **Modified purchase request**, listing only types that have approval rounds in the plans in
  view; the outcome words are the same for every type (041-C1; client 018 FR-015 asked for each type's own
  status words, which this report does not use because it counts rounds, not records; client 018 FR-017; 031
  FR-003).
- **FR-008**: The widget MUST also show how many rounds of the selected record type wait for a decision and since
  when the oldest waits, worked out as in 036 FR-016, so the three types' counts add up to 036's (041-C10).

**Average widget** *(US1)*

- **FR-009**: The **Average days to a decision** widget MUST show the mean days from submission to
  the end of each round approved or declined by a person, to one decimal, with the number of decisions (041-C7),
  using 017's rule (`AverageDecisionDays`) as Approvals › Monitoring does for the current month and 036's
  Approvals panel does for all rounds of the plans in view, which this report follows (client 018 FR-019;
  measure-metrics §1).
- **FR-010**: The widget MUST list each approver with their own average — from each round's submission to that
  person's own decision — and number of decisions, slowest first (041-C7), crediting the person who actually decided (041-C2); waiting, recalled, replaced, not-needed and automatically
  approved steps MUST NOT be counted, and skipped levels are not rows at all (041-C3).
- **FR-011**: With no decision in scope the average MUST show "—" and say why, never zero (client 018 FR-006,
  FR-007).

**Approvals grid** *(US2)*

- **FR-012**: The **Approvals** grid MUST show one row per named approver at each level of each submission
  (041-C2), over all three record types by default, with the columns of US2 scenario 1 (client 018 FR-020, FR-024; FA §E3.6.3).
- **FR-013**: The grid MUST offer a record-type filter of its own, independent of the widget's selector, listing
  the types present in scope (client 018 FR-017).
- **FR-014**: The grid MUST filter by **Approver** and **Status**, where Status is the step's outcome and offers
  the six words of 041-C1 plus **Not needed** (041-C2) (client 018 FR-014; FR-015 is replaced by 041-C1).
- **FR-015**: The grid MUST offer free-text search over its text columns, sorting by any column across all
  rows, page sizes 5, 10, 25, 50 and 100 (default 50) and the total count (client 018 FR-021).
- **FR-016**: A long comment MUST be readable in full without leaving the report (client 018 FR-022).
- **FR-017**: Each row's record name MUST link to the record when the person may open it and MUST be plain text
  when not (041-C5); the report offers no action on a round (client 018 FR-023).

**Figures, money and states**

- **FR-018**: Amounts MUST be exact decimals with their currency in the display currency at the current rate
  (D-303), the amount of an activity round being the activity's planned cost as it is now and that of a purchase-request
  round the request's amount as 032 and 036 work it out (041-C4, D-342); where the person may not see financial data of a plan,
  the amount MUST show "—" with the count of plans it leaves out (D-301, D-343).
- **FR-019**: A value that does not exist MUST show "—", never zero (client 018 FR-006).
- **FR-020**: Every empty widget and the grid MUST name the cause — no plans you may see, no approvals yet, you
  cannot see financial data, the filters exclude everything with a reset (client 018 FR-007; build guide §6).
- **FR-021**: Every figure MUST be computed on the server; the web app only draws what the API provides
  (constitution IV), and a total MUST equal the sum of its parts in the table view and in the export.
- **FR-022**: The report MUST respond within the time of SC-006 and without a statement per row at the 028
  volumes, with the benchmark recorded in the plan.

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
- **Approval step**: one named approver at one level of a round, with the decision date, outcome (the six words
  of 041-C1 or Not needed) and comments; the grid's row (041-C2).
- **Record type**: Activity, Purchase request or Modified purchase request (031).
- **Status mix**: the count and share of rounds per current outcome (041-C1) for one record type.
- **Average days to a decision**: the mean elapsed time from submission to the end of the round over rounds
  approved or declined by a person (017); per approver, from submission to that person's own decision (041-C7).
- **Report widget**: a block of the report frame with a chart or card, a table view and an export (039).

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A planning lead sees the status mix and the average days to a decision for the demo year within
  one page view, with no filter set, and finds the slowest approver in under 30 seconds.
- **SC-002**: For the same plans, the waiting count (summed over the three record types), the oldest-waiting
  date and the average days are identical on this report and on 036's Approvals panel in 100% of checks; the
  average uses the rule of Approvals › Monitoring.
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
  into them and never redefines them. Lane A's 039 spec and contracts were read from `origin/feature/039-measure-hub`
  while planning; a later change of the frame is applied here and in `contracts/api.md`.
- The three record types are those of 031; the outcome words are those the approval engine records (008, 017,
  031), shown identically for every type (041-C1); `restarted` and `cancelled` rounds show as Replaced by a new
  submission and Recalled (041-C6).
- "Decision" means a round approved or declined by a person (017); days are elapsed days to one decimal, as
  017's Monitoring average works them out (041-C7). Working days are not used.
- The amount of an activity round is the activity's planned cost as it is now; of a purchase-request round,
  032's amount (041-C4).
- Hive9's accrual, plan, campaign, initiative, line-item and custom-object approvals are not built: SP&A One does
  not approve them (031; constitution 4.8.0).
- Showing when GSAP data was last received (client 018 FR-009) is not needed: no figure here comes from GSAP.
- No approval figure uses the simulated performance feed; this report is wave A only.
- Out of scope: approving, declining or recalling from the report, sharing or scheduling by e-mail, a new
  migration, and any change to `openapi.yaml`, the shared hub files or the decision log (the integrator applies
  those from this feature's `contracts/api.md` and pull request).
