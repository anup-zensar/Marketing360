# Feature Specification: Purchase Request Report

**Feature Branch**: `feature/042-purchase-request-report`

**Created**: 2026-11-01

**Status**: Draft

**Input**: User description: "Purchase request report of the Measure module (wave A, lane C, issue #35): counts
and amounts of purchase requests by status, commitment by vendor and by owner per delivery month (Approved and
Closed lines by delivery date), invoiced per line, and a grid vendor › purchase request › line › line item with
column choice, paging and export. Read-only; plugs into 039's report frame. Vendor delivery reliability is not
built (Q-030)."

Cites: FA §E3.7; client 018 US2 and FR-025–FR-037 (PR Dashboard; US5 is not built, Q-030);
032 (purchase request statuses, lines and amounts, D-335), 033 (GSAP PR and PO), 015 (invoices),
036 (the same counts on the planning dashboards); [measure-metrics.md](../../docs/product/measure-metrics.md) §1;
the report frame of 039.

## Overview

Procurement and planning leads can see purchase requests one by one on the Purchase requests page. This report is
the **forward commitment view** across everything in the plans a person may see: how many purchase requests sit in
each status and for how much, what Shell has committed to each vendor and each owner in each month of delivery,
what has been invoiced, and a searchable, exportable grid that opens from a vendor down to the line item a line
pays for. It answers "what have we committed to receive, and when", which is what the accrual conversation needs.

It lives under **Reports › Purchase requests** and uses the report frame of feature 039 — the filter panel,
display currency, saved views, "as of" line, table view and export of every widget — without changing any of it.
The report never creates, edits, sends, approves or closes a purchase request; that stays on the Purchase
requests pages.

Hive9's PR Dashboard also shows goods receipts (GRSP) and a vendor performance percentage. The simulated GSAP
sends no receipts (033, Q-030), so this report shows commitment, invoiced and paid only and does not offer a
GRSP card, a GRSP column or a vendor performance chart.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - See what is committed to each vendor and owner, by delivery month (Priority: P1)

A planning or procurement lead opens Reports › Purchase requests and sees summary cards (purchase requests, lines,
vendors, committed, invoiced), the number and amount of purchase requests in each of the six statuses and how many
approved ones still wait for GSAP, and two charts of committed value by calendar month of delivery: one stacked by
vendor, one by purchase request owner. Each chart has a table view.

**Why this priority**: it is the question the report exists for — how much will arrive when, from whom, and on
whose request — and nothing else in SP&A One answers it across plans.

**Independent Test**: Sign in as Alex Morgan (Administrator, all markets), open Reports › Purchase requests for
the demo year with no filter: the committed total equals the sum of the monthly values of both charts, the status
counts equal 036's Purchase requests panel for the same plans, and the largest vendor's monthly values match its
lines' delivery dates and amounts on the Purchase requests pages.

**Acceptance Scenarios**:

1. **Given** purchase requests in the plans in view, **When** the report opens, **Then** five cards show the
   number of purchase requests, the number of lines, the number of distinct vendors, the committed amount and the
   invoiced amount, each in the display currency with its currency named where it is an amount.
2. **Given** the status widget, **Then** it lists the six statuses of 032 — Created, Submitted, In review,
   Declined, Approved, Closed — each with its count, amount and share of the count in whole per cent, a total, the
   number of Approved purchase requests waiting for GSAP, and a table view; the counts add up to the total and the
   amounts to the amount of all purchase requests in view.
3. **Given** the commitment chart by vendor, **When** it is drawn, **Then** each bar is one calendar month of
   delivery date and stacks the committed value of the Approved and Closed lines due that month by vendor; the
   ten largest vendors are named and the rest are grouped as **Other**, while the table view and the export list
   every vendor.
4. **Given** the commitment chart by owner, **Then** it shows the same monthly values grouped by the purchase
   request's owner, with the same ten-and-Other rule.
5. **Given** a Closed purchase request, **Then** its lines count at the amount they were closed at (what GSAP
   invoiced, 032) and an Approved one at its current line amounts; Created, Submitted, In review and Declined
   requests are in the status widget and the counts but add nothing to commitment.
6. **Given** a purchase request without a vendor, **Then** its commitment is shown under **Not filled in**, never
   dropped.
7. **Given** the filter panel's year, plans, teams, owners, activity types, activity statuses and attributes,
   **When** any is applied, **Then** the cards, the status widget and both charts narrow together: a purchase
   request is in view when one of its line items is (as 036 counts it).
8. **Given** the report's own filters **Vendor**, **Owner of the purchase request** and **Status**, **When** any
   is applied, **Then** every widget and the grid narrow by it.
9. **Given** the display currency is changed, **Then** every amount is shown in it at the current rate, with the
   currency named.

---

### User Story 2 - Open the commitment from a vendor down to the line item (Priority: P2)

A lead who sees a large commitment wants to know which purchase requests and lines it comes from. A grid below
the charts groups rows **Vendor › Purchase request › Line › Line item**; each level opens to the next, and the
columns can be chosen.

**Why this priority**: the charts say how much; the grid says exactly where it comes from and what each line has
had invoiced — needed to reconcile with the vendor and with GSAP.

**Independent Test**: As Alex Morgan open the grid, expand the largest vendor, then one purchase request, and read
its lines with their delivery dates, purchase order numbers and invoiced amounts; hide two columns, reload the
page, and the same columns are still hidden; search for a purchase request name and only its vendor remains.

**Acceptance Scenarios**:

1. **Given** the grid, **Then** the top level lists one row per vendor with its purchase request count, line
   count, committed and invoiced amounts; opening a vendor lists its purchase requests, opening a purchase
   request lists its lines, and opening a line shows the line item it pays for.
2. **Given** a purchase request row, **Then** its columns include name, ID, status, amount, owner, created date,
   sent date and approval workflow; a line row shows its delivery date, purchase order number and line number,
   purchase order line status and invoiced amount (columns in the order of client 018 FR-037 less actions and the
   GRSP amount).
3. **Given** the column chooser, **When** the person shows, hides or reorders columns, **Then** the grid follows
   at once; the choice is kept in the page address, so Back, a link and a saved view keep it, and **Reset
   columns** restores the default order.
4. **Given** the grid, **When** the person types in the search box, **Then** it narrows to vendors, purchase
   requests, lines and line items containing the words, keeping the path to each match, and the total updates.
5. **Given** the grid, **When** the person sorts by a column, **Then** the rows of each level are ordered by it
   across all pages, with the same order on every page; **When** they choose a page size of 5, 10, 25, 50 or 100
   (default 25), **Then** the page and the total change accordingly; paging counts vendors, the top level.
6. **Given** a row, **Then** the name links to the record when the person may open it and is plain text when not;
   nothing in the grid creates, edits, sends or closes anything.
7. **Given** a purchase request routed to GSAP, **Then** its GSAP PR number and each line's purchase order number
   and date are those of the simulated GSAP (033); before GSAP answers they show "—".
8. **Given** a line, **Then** its invoiced amount is the sum of the invoices linked to its purchase order line
   (015, 033), in the display currency; a line with no invoices shows **0.00**, and a line with no purchase order
   yet shows "—".

---

### User Story 3 - Export the report (Priority: P2)

A lead who has the picture exports each widget or the grid, to check it in a spreadsheet or attach it to an
accrual discussion.

**Why this priority**: the grid and charts are for looking; the export is how the figures reach the people who
reconcile with GSAP.

**Independent Test**: As Alex Morgan filter by one vendor and export the grid as Excel: the file holds every
purchase request, line and line item of that vendor and no row of a plan outside Alex's scope; export the
commitment chart as CSV and its monthly totals equal the table view.

**Acceptance Scenarios**:

1. **Given** the report, **When** the person exports a widget as CSV or Excel through the frame's menu, **Then**
   the file holds exactly the widget's table view: every vendor and owner (not only the ten named in the chart),
   the display currency named, no helper columns.
2. **Given** the grid, **When** it is exported, **Then** the file holds every matching line, one row each, with
   the vendor, purchase request, line and line item they sit under and the visible columns in their order — not
   only the page or the opened rows.
3. **Given** a person without `data.export`, **Then** the export is not offered, as everywhere in Reports.
4. **Given** a figure the person may not see (no financial access), **Then** the file shows "—" for it, never a
   zero.

---

### Report states *(read-only: states, not create/edit/delete)*

These scenarios cover each state a widget or the grid can render; each has its own test.

1. **Data**: purchase requests exist in the plans in view → cards, widgets, charts and grid show them.
2. **No plans in scope**: the person may see no plan of the chosen year → the page says "No plans you may see for
   <year>" and offers to change the year; no figures.
3. **No purchase requests yet**: plans exist but no purchase request has a line item in them → each widget says
   "No purchase requests yet" and what creates them ("Raise a purchase request from an approved activity's line
   items"); the grid is empty with the same words.
4. **Purchase requests not switched on**: where the organisation has not switched purchase support on (032), the
   page says so and shows no figures.
5. **No commitment yet**: purchase requests exist but none is Approved or Closed → the charts say "Nothing is
   committed yet — commitment starts when a purchase request is approved"; the status widget and cards still show.
6. **No financial access**: where the person may not see financial data of a plan (`finance.view`, D-301),
   amounts show "—" for that plan with the count of plans each figure leaves out; counts, statuses and dates
   still show.
7. **Filters exclude everything**: the filters match no purchase request → each widget and the grid say the
   filters exclude everything and offer a one-click reset.
8. **No GSAP data yet**: GSAP numbers and invoices are absent → those columns show "—" and the page says when
   GSAP data was last received (client 018 FR-009).
9. **Loading and failure**: while loading each widget shows a placeholder of its own shape; when a figure cannot
   be loaded the widget names the problem and offers Try again, without blocking the others.

### Edge Cases

- A purchase request with lines on line items of two plans: it is in view when any of its line items is, and its
  lines count once each at their own delivery month (as 036 counts it).
- A purchase request changed after approval (032): commitment follows the amounts GSAP holds for its approved
  version, as the Purchase requests page shows them, so the report never disagrees with the page.
- A line whose delivery date is blank or in a year other than the report's: a blank date is shown under **No
  delivery date** in the charts' table view and the export, never dropped; the monthly axis covers every month
  that has commitment.
- A purchase request in a currency other than the display currency: each line converts at the current rate
  (D-303); a line whose currency has no rate counts at 1 and is flagged, as 032 does.
- A vendor named slightly differently in two purchase requests ("Branded Item" and "Branded Items") is two
  vendors: the report groups by the name as entered and does not merge them.
- More than ten vendors or owners: ten are named by committed value, the rest are one **Other** series; ties
  break by name.
- A purchase request whose owner left: the row keeps the name they had; the Owner filter still lists them while
  requests exist.
- Dates on screen are calendar dates (`YYYY-MM-DD`); months follow the calendar, not a fiscal calendar.
- Many thousands of lines: the cards, charts, paging, sorting and search stay quick and the export still holds
  every matching line.
- Filters in the address that name a vendor, owner or status that does not exist: the report answers with empty
  figures and the "filters exclude everything" state and its reset, not an error.

## Requirements *(mandatory)*

### Functional Requirements

**Frame and access**

- **FR-001**: The report MUST appear as **Purchase requests** under Reports and use feature 039's report frame
  (filter panel, display currency, saved views, "as of" line, table view and export of every widget) without
  redefining any of it; it offers no Period or Show control because the charts are bucketed by calendar month of
  delivery (client 018 FR-001, FR-004, FR-005, FR-008, FR-010, FR-025).
- **FR-002**: The report MUST be read-only: it MUST NOT create, edit, send, approve, close or delete a purchase
  request; the only thing it saves is a person's saved view through 039 (client 018 FR-002).
- **FR-003**: The report MUST need `reports.view`, export MUST need `data.export`, and every figure and row MUST
  come only from purchase requests the person may see in plans in their scope; a record outside it MUST answer
  as missing. This replaces Hive9's "recorded for later, not enforced" (client 018 FR-011): SP&A One enforces
  roles and scopes (014).
- **FR-004**: The report's own filters MUST be **Vendor**, **Owner of the purchase request** and **Status** (the
  six statuses of 032), each narrowing every widget and the grid, together with the frame's filters; the frame's
  year, plans, teams, owners, activity types, activity statuses and attributes narrow purchase requests through
  their line items (client 018 FR-025, FR-027).
- **FR-005**: A purchase request MUST be in view when one of its line items is in view, as feature 036 counts it,
  and MUST be counted once whatever the number of its lines in view.

**Cards and status**

- **FR-006**: Five cards MUST show the number of purchase requests, the number of lines, the number of distinct
  vendors, the committed amount and the invoiced amount for the purchase requests in view (client 018 FR-028;
  there is no GRSP card, Q-030).
- **FR-007**: The status widget MUST list the six statuses of 032 with their count, amount and share of the count
  in whole per cent, a total equal to the sum of its parts, and the number of Approved purchase requests waiting
  for GSAP, as 036's Purchase requests panel counts them (client 018 FR-029; 032, 036).
- **FR-008**: The status widget's counts and amounts MUST equal 036's Purchase requests panel for the same plans.

**Commitment charts**

- **FR-009**: **Committed** MUST be the amount of the lines of Approved and Closed purchase requests, each line
  counted in the calendar month of its delivery date, as the Purchase requests page shows them (D-335); a Closed
  purchase request counts at its closed amounts (032).
- **FR-010**: The report MUST offer two charts of committed value per calendar month of delivery: by vendor and by
  the purchase request's owner (client 018 FR-031, FR-032).
- **FR-011**: A chart with more than ten vendors or owners MUST name the ten with the largest committed value and
  group the rest as **Other**; its table view and its export MUST list all (client 018 FR-033).
- **FR-012**: A purchase request without a vendor MUST be grouped under **Not filled in**, and a line without a
  delivery date under **No delivery date**; neither is ever dropped (client 018 FR-026).
- **FR-013**: The monthly values of each chart MUST add up to the committed card, and the table view MUST list
  each month, each vendor or owner and a total.

**Purchase request grid** *(US2)*

- **FR-014**: The grid MUST group rows **Vendor › Purchase request › Line › Line item**, each level opening to the
  next, limited to the filters (client 018 FR-036).
- **FR-015**: The grid MUST offer the columns of client 018 FR-037 — name, ID, level, status, purchase request
  cost, owner, created date, delivery date, sent date, purchase order number, purchase order line number,
  purchase order line status, invoiced amount and approval workflow — and MUST NOT offer an actions column or a
  GRSP amount (read-only; Q-030).
- **FR-016**: The person MUST be able to show, hide and reorder columns and reset them; the choice MUST live in
  the page address so Back, a link and a saved view keep it.
- **FR-017**: The grid MUST offer free-text search over its text columns, sort by any column across all rows,
  page sizes 5, 10, 25, 50 and 100 (default 25) counting vendors, and the total count.
- **FR-018**: Each row's record name MUST link to the record when the person may open it and MUST be plain text
  when not; the report offers no action on a purchase request.
- **FR-019**: A line's purchase order number, line number, status and date, and a purchase request's GSAP number,
  MUST be those of the simulated GSAP (033); a line's invoiced amount MUST be the sum of the invoices linked to
  its purchase order line (015) (client 018 FR-037).

**Figures, money and states**

- **FR-020**: Amounts MUST be exact decimals with their currency in the display currency at the current rate
  (D-303); where the person may not see financial data of a plan, the amount MUST show "—" with the count of
  plans it leaves out (D-301, D-343).
- **FR-021**: A value that does not exist MUST show "—", never zero (client 018 FR-006); an amount that is truly
  nothing (no invoices yet on a line with a purchase order) is shown as 0.00.
- **FR-022**: Every empty widget and the grid MUST name the cause — no plans you may see, purchase requests not
  switched on, no purchase requests yet, nothing committed yet, you cannot see financial data, no GSAP data yet,
  the filters exclude everything with a reset (client 018 FR-007; build guide §6).
- **FR-023**: Every figure MUST be computed on the server; the web app only draws what the API provides
  (constitution IV), and a total MUST equal the sum of its parts in the table view and in the export.
- **FR-024**: The report MUST show when GSAP data was last received wherever a figure comes from the simulated
  GSAP (client 018 FR-009).
- **FR-025**: The report MUST respond within the time of SC-006 and without a statement per row at the 028
  volumes, with the benchmark recorded in the plan.

**Export**

- **FR-026**: Each widget MUST export to CSV or Excel through 039's export with the content of its table view
  (every vendor and owner, the display currency named, no helper columns); nothing is stored (client 018 FR-010).
- **FR-027**: The grid MUST export every matching line as one row with the vendor, purchase request, line and
  line item above it and the visible columns in their order, whatever is opened or paged.
- **FR-028**: The report MUST NOT show Hive9's goods receipt (GRSP) amount, vendor performance percentage,
  per-person card choice, accrual filters or the OA vendor, outline agreement and contract holder attribute
  groups that only the China pack uses; SP&A One does not hold the data or has no such attribute (Q-030, 033,
  constitution 4.8.0).

### Key Entities *(include if feature involves data)*

- **Purchase request**: a request to buy against line items, with a vendor, owner, status, approval workflow and
  GSAP number (032, 033). Read, never changed.
- **Purchase request line**: one line item's share of a purchase request, with its amount, delivery date, purchase
  order line and invoiced amount (032, 033, 015).
- **Status**: Created, Submitted, In review, Declined, Approved or Closed (032).
- **Committed**: the amount of the Approved and Closed lines, in the month of delivery (D-335).
- **Vendor**: the name entered on the purchase request; **Not filled in** where blank.
- **Report widget**: a block of the report frame with a card, chart or grid, a table view and an export (039).

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A procurement lead sees what is committed to each vendor and owner by delivery month for the demo
  year within one page view, with no filter set, and names the month with the largest commitment in under 30
  seconds.
- **SC-002**: For the same plans, the status counts and amounts are identical on this report and on 036's
  Purchase requests panel in 100% of checks, and the monthly values of each chart add up to the committed card.
- **SC-003**: A person finds the lines of one vendor and one purchase request by search or by opening the grid in
  under 1 minute.
- **SC-004**: The export holds 100% of the lines matching the filters and search — not only the visible page or
  opened rows — and 0 rows from plans outside the person's scope.
- **SC-005**: With the demo data, 100% of the report's states (data, no plans, no purchase requests, not switched
  on, nothing committed, no financial access, filters exclude everything, no GSAP data, failure) render a
  message naming their cause; no figure is ever shown as zero when the real value is missing.
- **SC-006**: The report and its grid open and respond within 2 seconds at the 028 volumes and pass the
  accessibility checks (WCAG 2.2 AA, keyboard-only use of filters, charts' table views, grid, column chooser and
  export).

## Assumptions

- Feature 039's report frame, filter model, saved views and export are available on `measure-module`; 042 plugs
  into them and never redefines them; the report's own parameters ride in the address and in saved views, as
  041's do.
- Purchase request statuses, lines, amounts and the waiting-for-GSAP rule are those of 032 and 036; the simulated
  GSAP's numbers, purchase order lines and invoices are those of 033 and 015 (D-335, D-358); no new migration or
  store method is needed (read through the existing stores).
- **Commitment** counts Approved and Closed requests only, as the issue states; Hive9's chart counted every
  request in scope, so the report's committed total is deliberately smaller than "all request lines" and the
  status widget carries the rest.
- The charts' axis covers the calendar months that have commitment in the plans in view, not the fiscal calendar,
  and not only the report year, because delivery can fall after it; Year selects plans through the frame, not a
  delivery date range (Hive9's date range is not offered, FR-001).
- A purchase request is in view through its line items, as 036 counts it (041-C8 follows the same rule).
- Amounts are shown only for plans whose financial data the person may see (`finance.view`, D-301); counts,
  statuses and dates always show.
- Vendor delivery reliability and the GRSP amount are not built (Q-030); the Hive9 PR attribute groups that exist
  only in the China pack (OA vendor, sub category, outline agreement, contract holder) are not offered.
- Hive9's accrual filters and the AR Invoice Details screen are out of scope (client 018 US3, US6).
- Out of scope: creating or changing a purchase request, sending to GSAP, per-person card choice, sharing or
  scheduling by e-mail, a new migration, and any change to `openapi.yaml`, the shared hub files or the decision
  log (the integrator applies those from this feature's `contracts/api.md` and pull request).
