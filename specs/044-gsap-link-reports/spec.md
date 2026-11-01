# Feature Specification: GSAP Link Reports

**Feature Branch**: `feature/044-gsap-link-reports`

**Created**: 2026-11-01

**Status**: Built

**Input**: User description: "Three tables in one report page with tabs: SWBS requested and not received (line
item, plan, requested on, days waiting), purchase requests waiting for GSAP (approved and routed, no GSAP PR
number), and the purchase orders per line item (the PO ledger: line item, purchase request, GSAP PR number, PO
number and date, PO amount, invoiced, paid). Each with paging and export; financial columns need `finance.view`.
Stories: US1 SWBS not received (P1) · US2 waiting for GSAP (P2) · US3 purchase orders per line item (P2)."
(GitHub issue #37, `docs/sdd/measure-issues.md` §#044)

**References**: Constitution 4.8.0 — Current Phase & Scope (Measure wave A, D-366), Principles IV (the server
works out every figure), VII (scope isolation), F3 (state in the address), F6 (every chart has a table) ·
Hive9 functional analysis §E3.10.3 (PO Report columns), §F4.4 (SWBS Not Received and Pending PR in the Measure
catalogue) · Shell's client spec 019 US3, **FR-044** (SWBS Not Received), **FR-045** (Pending PR), FR-039 (PO
Report row), FR-013, FR-014 (table paging and export), FR-018, FR-020 (currency, empty causes) and the GSAP
workshop turnaround defaults (019 Assumptions) · `docs/product/measure-metrics.md` §1 (SWBS not received, Waiting
for GSAP, Purchase orders per line item, Invoiced per PR line) · features 007 (line items, SWBS request flag),
015 (financial transactions and their links), 030 (SWBS request status, FR-001, FR-002), 032 (approval and
routing of purchase requests, FR-030, FR-032), 033 (GSAP PR number and purchase order, FR-002 – FR-005), 039
(the report frame: filters, Currency, saved views, widget table and export; D-367, D-368) · D-237, D-301
(financial data needs `finance.view`).

The issue cites client FR-021 – FR-030; those are 019's dashboard engine and Plan Details. The GSAP link
dashboards are 019 FR-044 and FR-045 (and FR-039 for the PO row), cited above.

## Clarifications

Answers given while specifying are recorded here as `044-Cn` with their reasoning; the integrator moves them
into the decision log at the end of wave A (D-369).

### Session 2026-11-01

Settled with the recommended answer, as the developer asked for planned work to go ahead without stopping;
each stands until the developer or the integrator overrides it.

- **044-C1** Q: In MAPP the simulated GSAP requests and receives a code in the same run (030), so no line item is
  ever "requested, not received". What does the SWBS tab list? → A: Line items whose SWBS request status is
  **Ready to send to GSAP** (request on, activity approved, no code), waiting since their activity was approved.
  *Why*: client 019 FR-044 counts "line items flagged for an SWBS code on approved activities still without one"
  from the approval date; Ready to send is exactly that set (030 FR-001), and with live GSAP the same tab will
  list requests sent and not answered without a change of meaning.
- **044-C2** Q: Client FR-045 has two Pending PR groups, awaiting approval and approved without GSAP's answer.
  Which does 044 show? → A: Only **Waiting for GSAP**, the issue's goal. *Why*: approval rounds of purchase
  requests, with days to a decision, are the Approvals report (041, lane C); 044 is the GSAP link.
- **044-C3** Q: Which purchase requests are waiting for GSAP? → A: 032 FR-032's definition, the one the Purchase
  requests page's **Waiting for GSAP** view uses: Approved, routed, and GSAP has not answered (no purchase order,
  or a line whose amount or currency differs from its PO line). *Why*: the issue's "no GSAP PR number" misses an
  approved top-up waiting for its PO change (033 FR-007); one definition keeps the report equal to the page.
- **044-C4** Q: What does "financial columns need `finance.view`" cover? → A: On the SWBS tab, nothing (codes
  and dates are line-item data, 030). On the Waiting for GSAP tab, the **amount** column only. The **Purchase
  orders** tab as a whole, because every column of it is 015 financial data (D-237): it covers only plans where
  the person holds `finance.view`. *Why*: D-237 makes every Financials card, column and page of 015 need
  `finance.view`; 039's catalogue marks this report "money columns need `finance.view`".
- **044-C5** Q: How are turnaround times measured? → A: SWBS: hours since approval, flagged **Over 48 hours**
  beyond 48 (client default). Waiting for GSAP: working days (Monday to Friday) since routing, flagged **Over 2
  working days** beyond 2. Fixed in wave A, not a setting. *Why*: the GSAP workshop defaults (019 Assumptions);
  a setting would change Settings, which is not lane D's, and no holiday calendar exists in MAPP.
- **044-C6** Q: An activity can be Approved with no approval round (records loaded as approved). Since when does
  its line item wait? → A: Since the activity was last changed, which is no earlier than its approval. *Why*:
  every approval in MAPP ends a round (013, 017); only loaded records lack one, and the row must still be listed
  for SC-002, with a waiting time that never overstates.

## User Scenarios & Testing *(mandatory)*

The report is **read-only**: nothing on it changes a record and opening it writes nothing except what 039's
saved views write. Its scenarios are the states a tab can be in — data, empty, no access, no financial access,
filters exclude everything — not create, edit or delete. Sending SWBS requests and creating purchase orders stay
on the line item and purchase request pages (030, 033).

Every scenario runs inside 039's report frame. The frame's filter panel, Currency, saved views, "as of" line,
widget table view, widget export and empty-state wording are 039's and are not redefined here; this spec says
only what the GSAP link report puts inside the frame. Period and Show are not used: the report has no time
series and lists records of one kind per tab.

### User Story 1 - See line items still waiting for their SWBS code (Priority: P1)

A finance coordinator opens **Reports › GSAP link** for the plan year. The **SWBS not received** tab lists
every line item of an approved activity that asks for an SWBS code and has none yet, with how long it has
waited, so that they can run the simulated GSAP or chase GSAP before purchase requests are held up.

**Why this priority**: Without an SWBS code no purchase request can be raised for the line item (030); it is
the first break in the GSAP link and the one Shell's Superuser role watches (FA §F4.4).

**Independent Test**: On the demo data, open the tab as a planner with one team in scope; compare the rows with
the line items that show "Ready to send to GSAP" in All work for the same filters; check the waiting time and
the Over 48 hours count of one row by hand.

**Acceptance Scenarios**:

*Data*

1. **Given** "Stand build" (L-1006) of the approved activity Billboard winter wave, request on, no code,
   approved 3 days ago, **When** the tab opens, **Then** it shows a row with the line item, its path (plan ›
   campaign › initiative › activity), the activity owner, the company code, **Approved on** and **Waiting 3
   days** marked **Over 48 hours**.
2. **Given** 7 such line items, 2 approved within the last 48 hours, **Then** the cards show **Waiting 7** and
   **Over 48 hours 5**, and the chart **Waiting by time** shows Under 48 hours 2 · 2 – 5 days *n* · 5 – 10 days
   *n* · Over 10 days *n*, adding up to 7.
3. **Given** a line item whose code was received, or whose request is off, or whose activity is not approved,
   **Then** it is not listed (030 FR-001).
4. **Given** an activity approved, then In progress, **Then** its line items still waiting are listed (030 FR-001).
5. **Given** more than 20 rows, **Then** the table shows 20 per page with page sizes 5 to 500, sorts by any
   column (longest waiting first by default) and states "1–20 of 64".
6. **Given** the planner selects a row, **Then** the line item page opens, where the SWBS request card offers
   what 030 offers.

*Empty*

7. **Given** plans in scope with no line item waiting, **Then** the tab shows 039's "No data yet" state saying
   "No line item is waiting for its SWBS code", not zeros as figures.
8. **Given** no plans in scope for the chosen year, **Then** the tab shows 039's "No plans you may see for
   <year>" state.

*No access*

9. **Given** a person without "See reports", **Then** GSAP link is not offered and its address answers as 039
   answers a refused report.
10. **Given** a plan outside the person's scope, **Then** its line items are not listed, counted or exported.

*No financial access*

11. **Given** a person without "View financial data" in any plan, **Then** the SWBS tab is shown in full: it has
    no money column (044-C4).

*Filters exclude everything*

12. **Given** filters that match no activity (for example a team with no plans this year), **Then** the tab
    shows 039's "Your filters exclude everything" state with **Reset filters**, and the cards show 0.

---

### User Story 2 - See purchase requests waiting for GSAP (Priority: P2)

The coordinator switches to **Waiting for GSAP**. It lists every approved purchase request routed to the
requisitioner that GSAP has not answered yet, with how many working days it has waited since routing.

**Why this priority**: An approved request without a GSAP PR or purchase order is money the plan expects to be
committed and is not (client 019 FR-045); it follows the SWBS step, so P2.

**Independent Test**: On the demo data, compare the tab with the Purchase requests page's **Waiting for GSAP**
view for the same plans; check one row's working days by hand.

**Acceptance Scenarios**:

1. **Given** a purchase request approved and routed (Original) on Wednesday 14 Oct 2026 and today Monday 19 Oct
   2026, **Then** its row shows the request's ID and name, vendor, owner, path, **Routed on** 14 Oct 2026,
   routing **Original**, GSAP PR number empty, amount, and **Waiting 3 working days** marked **Over 2 working
   days**.
2. **Given** an ordered request whose approved top-up was routed (Modification) and not yet answered, **Then**
   it is listed with its GSAP PR number and routing **Modification**, waiting since that routing (044-C3).
3. **Given** a request routed again with **Resend PR form**, **Then** the waiting time still counts from the
   routing GSAP has not answered; resending does not restart it.
4. **Given** the cards, **Then** they show **Waiting** (count), **Amount waiting** in the report currency and
   **Over 2 working days** (count).
5. **Given** the request's purchase order is created (033), **Then** on the next load it is no longer listed.
6. **Given** a person without "View financial data" in one of two plans, **Then** that plan's requests are
   listed with the amount "—", the amount card covers only the other plan and says "Covers 1 of 2 plans" (039).
7. **Given** no request waiting, **Then** the tab shows "No data yet" with "No purchase request is waiting for
   GSAP"; **given** filters that match nothing, 039's "Your filters exclude everything".
8. **Given** the planner selects a row, **Then** the purchase request page opens.

---

### User Story 3 - See the purchase orders of each line item (Priority: P2)

The coordinator switches to **Purchase orders**. It lists, per line item, each purchase order GSAP sent for it:
the purchase request it came from (when raised in MAPP), the GSAP PR number, the PO number, its lines and date,
the PO amount, what has been invoiced and what has been paid.

**Why this priority**: It closes the loop from line item to PO and invoice (FA §E3.10.3) and is where finance
reconciles; the figures already exist on line item and Financials pages, so it gathers rather than enables.

**Independent Test**: On the demo data, compare the rows of one line item with its Financials section (015)
and its purchase request's PO fields (033).

**Acceptance Scenarios**:

1. **Given** "Venue hire" with PO SIM-4500001001 raised from purchase request PR-1001 (GSAP PR SIM-0010001001),
   **Then** a row shows the line item, its path, PR-1001, SIM-0010001001, SIM-4500001001, PO lines 10, the PO
   date, the vendor, the PO amount, invoiced and paid, each in the report currency.
2. **Given** a purchase order loaded from a GSAP extract and not raised from a purchase request in MAPP,
   **Then** its row shows the purchase request and GSAP PR number empty.
3. **Given** a purchase order with lines on two line items, **Then** it shows one row per line item, each with
   its own lines and amounts.
4. **Given** invoices and payments that refer to a PO, **Then** they add to that PO's row for the line item
   they are linked to (015 FR-007); a credit note subtracts; one that is not linked adds nowhere.
5. **Given** the table, **Then** a totals row sums the PO amount, invoiced and paid of every row on every page.
6. **Given** a person who holds "View financial data" in one of two plans, **Then** the tab lists only that
   plan's purchase orders and says "Covers 1 of 2 plans" (044-C4); **given** none, it shows 039's "You can't see
   financial data" state and lists nothing.
7. **Given** no purchase order for the line items in view, **Then** the tab shows "No data yet" with "No
   purchase order has been received for these line items"; **given** filters that match nothing, 039's "Your
   filters exclude everything".

---

### User Story 4 - Export a tab (Priority: P2)

The coordinator exports a tab's table as CSV or Excel to chase GSAP or reconcile offline, or the whole report as
one Excel file.

**Why this priority**: Shell works these lists in spreadsheets with GSAP (client 019 FR-014); the frame gives
the export, the report supplies the rows.

**Independent Test**: Export each tab as CSV and as Excel and compare each with the table, page by page.

**Acceptance Scenarios**:

1. **Given** a person holding "Export data", **When** they export a tab, **Then** the file holds every row of
   every page in the current sort, with the same columns and totals as the screen, amounts as numbers with the
   currency in the heading; the filters and the "as of" time are those 039's export states.
2. **Given** the report's **Export**, **Then** the Excel file holds one sheet per tab, then 039's In view and
   About this export sheets.
3. **Given** a person without "View financial data" in some plans, **Then** the files hold exactly what the
   screen shows: "—" for hidden amounts and no purchase order of those plans.
4. **Given** a person without "Export data", **Then** no export is offered (039).

---

### Edge Cases

- **Line item moved to another activity**: listed under its current activity and path, waiting since that
  activity's approval.
- **Activity approved, then sent back and approved again**: waiting counts from the latest approval.
- **Line item deleted (in the recycle bin)**: not listed on any tab; its purchase orders stay in 015 but are not
  listed until it is restored.
- **Purchase request Closed**: not waiting for GSAP; a Close routing is not waited on (032 FR-034). Its purchase
  orders stay on the Purchase orders tab.
- **Purchase request with lines on several line items**: one row on Waiting for GSAP; on Purchase orders, one
  row per line item.
- **PO in a currency other than the report's**: converted at the current rate (D-303), as everywhere else.
- **Weekend**: Saturday and Sunday are not working days; a request routed on Friday has waited 1 working day
  on Monday.
- **Today before the approval time** (clock skew in demo data): waiting shows 0, never negative.
- **No GSAP data loaded yet**: the Purchase orders tab shows "No data yet" with "No GSAP data loaded yet"
  (015's wording); the frame's "as of" line has no time.

## Requirements *(mandatory)*

### Functional Requirements

**Report and access**

- **FR-001**: The Reports area MUST offer **GSAP link** inside 039's report frame with three tabs, **SWBS not
  received · Waiting for GSAP · Purchase orders**, the first the default. The tab, each table's sort, page size
  and page MUST live in the page address. *(client 019 US3, FR-013; D-368; constitution F3)*
- **FR-002**: GSAP link MUST be read-only: nothing on it changes a record or runs the simulated GSAP. *(client 019
  FR-022; D-366)*
- **FR-003**: GSAP link MUST need "See reports" and MUST list only records of plans in the person's scope; a
  plan outside the scope MUST NOT be listed, counted or named anywhere, export included. *(D-367; constitution
  VII)*
- **FR-004**: Amounts MUST cover only plans where the person holds "View financial data": the Waiting for GSAP
  amount shows "—" for other plans; the Purchase orders tab lists only those plans; each says how many plans it
  covers. *(D-237, D-301; 044-C4)*

**SWBS not received**

- **FR-005**: The tab MUST list every live line item whose SWBS request status is **Ready to send to GSAP**
  (030 FR-001) and whose activity matches the frame's filters, with: line item (name and ID), path, activity
  owner, company code, approved on, and waiting time. *(client 019 FR-044; 044-C1)*
- **FR-006**: Waiting MUST count from the activity's latest approval to now, shown in days with hours under two
  days, never negative; a row beyond 48 hours MUST be marked **Over 48 hours**. *(client 019 FR-044; 044-C5)*
- **FR-007**: The tab MUST show the cards **Waiting** and **Over 48 hours** and the chart **Waiting by time**
  with the bands Under 48 hours, 2 – 5 days, 5 – 10 days and Over 10 days, adding up to the Waiting card.
  *(client 019 FR-044)*

**Waiting for GSAP**

- **FR-008**: The tab MUST list every purchase request that is waiting for GSAP by 032 FR-032 and whose
  activity matches the frame's filters, with: ID and name, vendor, owner, path, routed on, routing type
  (Original or Modification), GSAP PR number when it has one, amount, and waiting time. *(client 019 FR-045;
  044-C2, 044-C3)*
- **FR-009**: Waiting MUST count working days (Monday to Friday) from the routing GSAP has not answered; Resend
  PR form MUST NOT restart it; a row beyond 2 working days MUST be marked **Over 2 working days**. *(client 019
  FR-045; 044-C5)*
- **FR-010**: The tab MUST show the cards **Waiting**, **Amount waiting** and **Over 2 working days**. *(client
  019 FR-045)*

**Purchase orders**

- **FR-011**: The tab MUST list one row per line item and purchase order for the live line items of matching
  activities, with: line item, path, purchase request (when raised in MAPP), GSAP PR number, PO number, PO lines,
  PO date, vendor, PO amount, invoiced and paid. *(FA §E3.10.3; client 019 FR-039; measure-metrics §1)*
- **FR-012**: PO amount MUST be the sum of the PO's lines linked to the line item; invoiced and paid the sums of
  the invoices (credit notes negative) and payments that refer to that PO and are linked to that line item, by
  015's linking rules. Unlinked transactions MUST NOT count. *(015 FR-001, FR-007; measure-metrics §1 Invoiced
  per PR line)*
- **FR-013**: The table MUST end with a totals row over every row, not only the page. *(client 019 FR-012)*

**Common**

- **FR-014**: Every amount MUST be converted to the currency of 039's **Currency** control at the current rate,
  each with its currency. *(D-303; client 019 FR-018; 039 FR-018)*
- **FR-015**: Each table MUST page (5 to 500 rows, default 20) and sort by any column; the default sort is longest
  waiting first on the first two tabs and line item then PO number on the third. *(client 019 FR-013)*
- **FR-016**: Selecting a row MUST open its line item (SWBS, Purchase orders) or purchase request (Waiting for
  GSAP) page. *(client 019 FR-015)*
- **FR-017**: The report MUST use 039's filters, Currency, saved views, "as of" line, widget table view and export
  as 039 defines them, and MUST NOT add filters of its own. *(D-368)*
- **FR-018**: Each tab MUST render one of 039's states: data; "No plans you may see for <year>"; "No data yet",
  naming what is missing; "Your filters exclude everything" with **Reset filters**; and on Purchase orders "You
  can't see financial data". *(D-368; client 019 FR-020)*
- **FR-019**: Each tab's export, and each sheet of the report's export, MUST hold every row in the current sort
  with the screen's columns, hidden amounts and totals. *(client 019 FR-014; 039 FR-022)*

### Key Entities

- **Waiting line item** (calculated): a line item Ready to send to GSAP, with its activity's latest approval time
  and the waiting time.
- **Waiting purchase request** (calculated): a purchase request waiting for GSAP, with the routing GSAP has not
  answered and the working days since.
- **Purchase order of a line item** (calculated): one purchase order's lines linked to one line item, with the
  purchase request and GSAP PR number they came from and the invoices and payments that refer to them.
- Read from existing records only: line items and their SWBS request (007, 030), activities and their approvals
  (006, 011), purchase requests, their routings and GSAP answers (032, 033), financial transactions and their
  links (015), currencies and rates (016). No new stored data.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: For a person whose scope holds 50 plans, 5,000 activities, 20,000 line items and 20,000 purchase
  request lines, each tab shows its first page within 2 seconds, and each page or sort change within 2 seconds,
  for 95 % of requests (028 volumes; client 019 SC-002).
- **SC-002**: The SWBS tab lists exactly the line items All work shows as Ready to send to GSAP for the same
  filters, and Waiting for GSAP exactly the requests of the Purchase requests page's Waiting for GSAP view, for
  100 % of cases (client 019 SC-006).
- **SC-003**: Every Purchase orders row's PO amount, invoiced and paid equals the line item's Financials figures
  for that PO to the cent.
- **SC-004**: A coordinator finds the line items waiting longest for an SWBS code and the requests waiting longest
  for GSAP, and opens one of each, within 2 minutes on the first attempt.
- **SC-005**: Every export equals its table, row for row; no plan outside the person's scope, and no amount of a
  plan without "View financial data", appears in any figure or file.
- **SC-006**: Every tab passes WCAG 2.1 AA with no critical or serious finding, and the chart has a table view
  (client 019 SC-010).

## Assumptions

- **Year**: a record belongs to the year of its plan (D-368).
- **No rejection reason or request status column** on the SWBS tab: the simulated GSAP never rejects (030,
  ADR-0016) and every row has the same status; both come with live integration.
- **No e-mail group or PR type filter** (client FR-045): the frame's filters are the only filters (FR-017).
- **Paging and sort** are the report's own table controls; 039's widget table view has no paging (039
  contracts/ui.md).
- **Days, not hours, on Waiting for GSAP**: working days as the client's default states the turnaround.
- 039's frame, its route rules and its export are in place before 044's web part is built (build guide §3); names
  follow 039's contract as pushed on `feature/039-measure-hub` (checked 2026-11-01).
