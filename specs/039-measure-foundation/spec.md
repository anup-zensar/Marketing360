# Feature Specification: Measure Foundation — the Reports Area and the Report Frame

**Feature Branch**: `feature/039-measure-foundation`

**Created**: 2026-10-30

**Status**: Draft

**Input**: User description: "Measure foundation (lane A, issue #32): the Reports area and everything every report
shares — report frame, filter panel (year, plans, Shell attributes per level with Not filled in, owner, activity
type, status, team), periods (months, quarters), display currency, saved report views, table view and CSV/Excel
export of every widget, the permission to see reports." (feature map row 039, D-366–D-369)

**References**: Constitution 4.8 — Current Phase & Scope (the Measure module, wave A), Principles IV (the server
works out every figure), VII (scope isolation), IX, X, F3 (state in the address), F6 (every chart has a table) ·
[measure-module-build.md](../../docs/sdd/measure-module-build.md) §1, §4, §6, §7 · decisions D-366 (scope), D-367
(who may open reports), D-368 (how a report behaves) · [measure-metrics.md](../../docs/product/measure-metrics.md)
· design system §2.8 Report pages, §8 Words on screen · Hive9 Measure summary §2–§4, §8 (FA §E2.4, §E3.0,
§E3.9.0, §E8.2) · Shell's client spec 016 (shared report frame: FR-005–FR-031, FR-074–FR-081; US2, US6, US7, US8) ·
features 010 (All work filters, saved views, export: FR-004–FR-009, FR-029–FR-034), 014 (permissions, scopes),
016 (Shell-attribute filters FR-039, plan type), 036 (filters in the address FR-006, FR-007; charts D-345), 038
FR-012 (unknown values fall back).

## Clarifications

### Session 2026-10-30

Answered by the product owner for lane A. Recorded here as `039-Cn`; the integrator moves them into the decision
log at the end of wave A (build guide §4.5).

- **039-C1** Q: When does a report follow a change in the filter panel? → A: on **Apply**, as Hive9's Update and
  client 016 FR-017. Ticking several values (brands, countries, cost centres) then refreshes the report once, not
  on every tick. The year, Period, Currency, Show, the chart or table choice and removing a chip still apply at
  once (FR-013).
- **039-C2** Q: Does a saved view belong to one report or to every report? → A: to **one report**, with its own
  default star per report. Reports use different controls (Period, Show), so a view shared by every report could
  carry choices a report does not have (FR-024, FR-027).
- **039-C3** Q: Where does Reports sit beside the 036 dashboards? → A: its **own sidebar item** right after
  Dashboards; the Executive and Regional dashboards are unchanged, so nothing built and tested in 036 moves
  (FR-034; answers Q-031).

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Open a report and narrow it to what I care about (Priority: P1)

Priya, a marketing team leader for Lubricants Europe, opens **Reports** in the sidebar. The Reports page lists the
reports she may open, each with one line saying what it answers. She opens one. The page shows the report's title,
the year and when its figures were worked out, and on the left a filter panel: **Year**, **Plans**, **Plan type**,
**Team**, **Owner**, **Activity type**, **Activity status**, and one tab per planning level holding that level's
Shell attributes, each with its values and **Not filled in**. Above the report a line says what is in view —
"In view: 1 plan · 6 campaigns · 14 initiatives · 38 activities · 72 line items". She ticks one brand and one
country, presses **Apply**, and the line, the chips above the report and every figure follow. She copies the
address and sends it to a colleague, who sees the same report with the same filters.

**Why this priority**: every Measure report (040–050) is built inside this frame; without it no report can be
filtered, scoped or shared, and Hive9's two different filter paradigms are its most confusing trait (FA §E2.4).

**Independent Test**: sign in as people with different scopes, open the Reports page and a report, change the
filters and compare the "In view" counts with All work's counts for the same year and filters.

**Acceptance Scenarios**:

1. **Given** Priya holds `reports.view` for Lubricants Europe only, **When** she opens Reports, **Then** the
   sidebar shows **Reports**, the page lists the reports of this release with their one-line purpose, and every
   report not yet available shows "Coming soon" without a link.
2. **Given** a person without `reports.view` in any plan, **When** they look at the sidebar or type the Reports
   address, **Then** the sidebar has no Reports item and the address shows the standard "no access" page naming
   the missing permission.
3. **Given** Priya opens a report for the first time, **When** the page loads, **Then** the year is the current
   year, every plan of that year in her scope is selected, every value of every group is ticked including **Not
   filled in**, the period is **Months**, the currency is her display currency, and the "In view" line counts
   exactly the records All work shows for the same year with no filters.
4. **Given** the filter panel, **When** Priya opens the **Activity** tab, **Then** it lists one group per active
   choice or yes/no Shell attribute of the activity level, each with "(selected / all)", **Select all**, **Clear
   all**, its values and **Not filled in**, and a search box when the group has more than ten values.
5. **Given** a Shell attribute with the same label at two levels, **When** the panel is shown, **Then** it
   appears as two separate groups, one in each level's tab.
6. **Given** Priya changes filters, **When** she has not pressed **Apply** yet, **Then** the report keeps its
   figures, the panel says "Changes not applied yet", and **Apply** shows the new figures, chips and "In view"
   line together.
7. **Given** filters are applied, **When** Priya presses **Back**, opens a bookmark or another person opens the
   copied address, **Then** the same year, filters, period and currency are shown — for the other person only
   within their own scope.
8. **Given** an address that names a plan, a value or a currency that no longer exists or that the reader may
   not see, **When** it is opened, **Then** the report applies everything else and says "1 filter value is no
   longer available and was left out", without naming an out-of-scope record.
9. **Given** applied filters, **When** Priya removes one chip or presses **Reset filters**, **Then** that filter
   (or every filter but the year) returns to its default and the report follows at once.
10. **Given** a report with time widgets, **When** Priya switches **Period** from Months to Quarters or chooses
    another **Currency**, **Then** every figure follows at once without **Apply**, the address keeps the choice,
    and every amount shows its currency.

---

### User Story 2 - Keep the views I use every month (Priority: P1)

Every month Priya looks at the same report for her two brands. She saves the current filters as "Helix and
Pennzoil", marks it as her default with the star, and from then on the report opens on that view. When she changes
a filter the Saved views menu says "Modified" and offers **Update view** or **Save as new view**. She renames one
view and deletes another, and **Undo** brings it back.

**Why this priority**: Hive9 users run the same reports every period; without saved views each run starts from
scratch (FA §E3.0.4, client 016 US8).

**Independent Test**: save, star, update, rename, delete and undo views on one report as one person, and check
another person never sees them.

**Acceptance Scenarios**:

1. **Given** applied filters, **When** Priya chooses **Save current filters** and types a name of at most 60
   characters, **Then** the view is saved for her and this report, listed in the Saved views menu and marked
   active; a name she already uses for this report (ignoring letter case) asks "Replace the view …?" first.
2. **Given** a saved view, **When** Priya stars it, **Then** opening the report from the sidebar or the Reports
   page — with no filters in the address — applies it, and the star moves if she stars another view.
3. **Given** the active view, **When** Priya changes a filter, the period, the currency or Show, **Then** the menu
   shows "Modified" with **Update view** and **Save as new view**; **Update view** asks before overwriting.
4. **Given** a view, **When** Priya renames it, **Then** the same name rules apply; **When** she deletes it, **Then**
   it disappears at once and the toast offers **Undo**, which restores it with its star.
5. **Given** a view that names a plan or value that no longer exists or is no longer in Priya's scope, **When** she
   opens it, **Then** the rest applies and the report says what was left out.
6. **Given** Carlos and Priya, **When** Carlos opens the Saved views menu, **Then** he never sees Priya's views,
   and an address that names one of her views opens with Carlos's own default instead.

---

### User Story 3 - See every chart as a table and take the figures away (Priority: P1)

Under a chart Priya chooses **Table** and sees every value the chart draws, with the period columns and a total.
She exports that widget as Excel for a slide, and exports the whole report with **Export**, which gives one sheet
per widget and an "About this export" sheet stating the report, year, filters, period, currency, who exported it
and when.

**Why this priority**: finance and marketing leads take figures into their own packs every month; a table view is
also how every chart is accessible (constitution F6).

**Independent Test**: on any report widget switch between chart and table, export the widget as CSV and Excel and
the report as Excel, and compare every exported value with the screen.

**Acceptance Scenarios**:

1. **Given** a chart widget, **When** Priya chooses **Table**, **Then** a table shows every value the chart draws
   with the same labels and a totals row where the widget totals, and **Chart** switches back; the choice is kept
   in the address.
2. **Given** a widget, **When** Priya chooses its export as CSV or Excel, **Then** a file named "SP&A One –
   <report> – <widget> – <year> – <date>" downloads with the widget's table, money as numbers with the currency in
   the column heading, and a first block (CSV) or sheet (Excel) stating the report, year, filters in words, period,
   currency, exporter and date and time.
3. **Given** a report, **When** Priya chooses **Export**, **Then** one Excel file "SP&A One – <report> – <year> –
   <date>" downloads with one sheet per widget and the "About this export" sheet.
4. **Given** a person without `data.export`, **When** they look at the export actions,
   **Then** the actions are shown unavailable with "Exporting isn't part of your role" and the reason in a tip.
5. **Given** any export, **When** it is compared with the screen, **Then** every value equals the value shown,
   figures the person may not see are "—" in both, and nothing about the export is stored, logged or sent.

---

### User Story 4 - Understand why a report is empty or incomplete (Priority: P2)

When a report shows nothing, Priya is told why in plain words and what to change: no plans she may see for that
year, filters that exclude everything, no data yet, or figures she is not allowed to see.

**Why this priority**: Hive9's ambiguous empty screens are its commonest support call (FA §E8.2 #4).

**Independent Test**: open a report in each empty or partial situation and read the message.

**Acceptance Scenarios**:

1. **Given** a year with no plans in Priya's scope, **When** the report opens, **Then** it says "No plans you may
   see for 2027" with the years that have plans, and shows no zeros as if they were figures.
2. **Given** filters that match no record, **When** they are applied, **Then** every widget is replaced by "Your
   filters exclude everything" with **Reset filters**.
3. **Given** a person who may view plans but not see financial data in some of them, **When** a money figure is
   shown, **Then** it covers only the plans where they may, says "Covers 1 of 2 plans", and is "—" with "You can't
   see financial data" when it covers none.
4. **Given** one widget fails to load, **When** the report is shown, **Then** that widget alone says "This part of
   the report couldn't be loaded" with **Try again**, and every other widget is shown.

### Edge Cases

- A plan shared by two teams counts once in "In view" and once per team in a Team filter.
- A value a business-unit condition hides on a record counts as **Not filled in** for that record (016 FR-039).
- A % split value matches when its share is above 0 (016 FR-039).
- An inactive Shell attribute value is listed, labelled "(inactive)", only while some record in view still uses it.
- A group with no value ticked would match nothing, so **Apply** stays off and the panel says why next to it:
  "Tick at least one value in Owner."
- The year changes the plan list; plans chosen in the old year are dropped, and while plans are chosen the panel
  says so under **Year** before the change ("Another year clears the 2 plans you chose.").
- A past year has no "in progress" period: every period is actual; a future year has no actual period.
- A deleted record in the recycle bin is not in view (as All work).
- Two tabs open on different filters keep their own state; only saved views are shared across tabs. If both
  change the same saved view, the second change is refused with "This view was changed in another tab" (FR-028).
- A saved view whose report is no longer offered to the person stays theirs and returns when access returns.

## Requirements *(mandatory)*

### Functional Requirements

**Reports area and access**

- **FR-001**: The sidebar MUST offer **Reports** to every person who holds `reports.view` in at least one plan. It
  opens the Reports page, which lists the reports of the Measure module in a fixed order — Plan budget, Approvals,
  Purchase requests, Plan details, GSAP link, User access, and later Summary, Revenue, Funnel stages — each with its
  name and a one-line purpose; a report not yet available in the release shows "Coming soon" without a link, and a
  report the person may not open (FR-003) is not listed.
- **FR-002**: A new scoped permission `reports.view` ("See reports", refusal noun "Seeing reports") MUST be added to
  014 Table A in the Plans and planning group. It MUST be granted to every role, built-in or not, that holds
  `plans.view`, so nobody's access changes until an administrator removes it in Settings › Access (D-367).
- **FR-003**: Every report page MUST need `reports.view` for the plans it covers; a report with a stronger rule
  states it (the User access report needs `access.manage`, D-367). Every figure, count, list entry and filter
  choice MUST cover only plans in the person's scope (014); a plan outside it, its records and its people MUST
  never be named or counted (404-style invisibility, constitution VII).
- **FR-004**: Money figures MUST cover only the plans where the person holds `finance.view` (D-301, D-343); a money
  figure that covers fewer plans than are in view MUST say "Covers N of M plans", and one that covers none MUST be
  "—" with "You can't see financial data".
- **FR-005**: Reports MUST be read-only: opening, filtering, switching views and exporting change no record and
  write nothing; the only things a person keeps are their saved views (FR-024–FR-029).

**Report frame**

- **FR-006**: Every report page MUST show the breadcrumb Reports › <report>, the report's title, a meta row with
  the year and "Figures as of <date> <time>", **Saved views** and **Export** (secondary); there is no primary
  action (design system §2.8).
- **FR-007**: Every report MUST show an **In view** line counting the plans, campaigns, initiatives, activities and
  line items that match the applied filters, by the matching rule of FR-012, so it equals All work's counts for the
  same year and filters.

**Filter panel**

- **FR-008**: The filter panel MUST offer, in this order: **Year** (one plan year; the years with plans in the
  person's scope plus the current year; default the current year), **Plans** (several; the year's plans in scope;
  default all), **Plan type** (Operational, Strategic, Cross-market; default all), **Team** (several; the teams of
  the plans in the year, **My teams** first; default all), **Owner** (several; people who own a record in the
  year's plans in scope; default all), **Activity type** (several; default all), **Activity status** (Created,
  Submitted, Declined, Approved, In progress, Complete; default all) and one tab per level — **Plan**,
  **Campaign**, **Initiative**, **Activity**, **Line item** — holding that level's Shell attributes (FR-009).
- **FR-009**: Each active choice or yes/no Shell attribute MUST be one group in its level's tab, identified by its
  level and attribute and never merged with another by label. A group lists its active values, the inactive
  values still used by a record in view labelled "(inactive)", and **Not filled in**; text, long-text and number
  attributes are not offered.
- **FR-010**: Every group with choices MUST show "(selected / all)" counting **Not filled in**, **Select all** and
  **Clear all**, and a search box when it has more than ten values; each level tab MUST offer **Select all** and
  **Clear all** for its groups.
- **FR-011**: The Plans list MUST follow the year; plans of the old year are dropped and the panel says how many.
  Owner and Activity type MUST list only choices present in the selected plans.
- **FR-012**: A record MUST match when it meets every filter tested on its level, with All work's rules (010
  FR-005, 016 FR-039): a filter tested on lower levels is met when a record below matches; a % split value matches
  when its share is above 0; a value hidden by its business-unit condition counts as **Not filled in**; a group
  with every choice ticked does not filter. Records in the recycle bin never match.
- **FR-013**: Changes in the filter panel MUST take effect on **Apply** (039-C1); until then the report keeps its
  figures and the panel says "Changes not applied yet". The year, Period, Currency, Show, the chart or table
  choice and removing a chip apply at once.
- **FR-014**: The applied filters MUST appear above the widgets as removable chips naming each restricted filter
  ("Brand: Helix, Pennzoil"); removing a chip applies at once; **Reset filters** returns every filter but the year
  to its default.
- **FR-015**: The year, filters, period, currency, Show choice and each widget's chart or table choice MUST live in
  the page address, so Back, a bookmark and a shared link show the same report (constitution F3, D-368); nothing is
  kept between visits except saved views. A value in the address that no longer exists, is inactive or is outside
  the reader's scope MUST be left out with "N filter values are no longer available and were left out", never
  naming it and never breaking the page (038 FR-012).

**Controls above the widgets**

- **FR-016**: Reports with time widgets MUST offer **Period**: Months (default; January to December of the plan
  year) or Quarters (Q1–Q4); the change applies at once.
- **FR-017**: Reports that rank records MUST offer **Show**: Plan, Campaign, Initiative or Activity; the change
  applies at once.
- **FR-018**: Every report MUST offer **Currency**: the active currencies, default the person's display currency.
  Every amount MUST be worked out in US dollars and converted once into the chosen currency at the current rate
  (D-303), and shown with its currency.
- **FR-019**: Time widgets MUST place a **Today** marker: periods before the current one show actual figures, the
  current period is labelled "in progress", later periods show planned or projected figures; a past year shows
  only actual periods and a future year none.

**Widgets**

- **FR-020**: Every widget MUST have a title and an ⓘ tip that names each figure it shows and its definition
  (measure-metrics.md); a value that cannot be worked out MUST be "—" with the reason in the tip, never 0.
- **FR-021**: Every chart MUST offer **Chart** and **Table**; the table lists every value the chart draws with the
  same labels and a totals row where the widget totals. Charts MUST draw bars and lines in the chart colours (never
  yellow or red for a series), with a legend in words and a text summary for assistive technology (D-345).
- **FR-022**: Every widget MUST offer export as CSV and Excel of its table, and every report **Export** as one
  Excel file with one sheet per widget. Every file MUST state, before the data (CSV) or on an "About this export"
  sheet (Excel), the report, year, filters in words, period, currency, exporter and date and time; money is
  exported as numbers with the currency in the column heading; file names are "SP&A One – <report> [– <widget>] –
  <year> – <date>". Exports need `data.export`, as the dashboards' exports do (D-343), hold exactly the values
  shown, and are never stored, logged or sent.
- **FR-023**: Each widget MUST load, fail and retry on its own: a skeleton while loading, and on failure "This part
  of the report couldn't be loaded" with **Try again**, without blanking the other widgets.

**Saved views**

- **FR-024**: A person MUST be able to save a report's whole address — its applied filters, period, currency, Show
  and each widget's chart or table choice — under a name of at most 60 characters, unique for that person and that report ignoring letter case; a view belongs
  to one report (039-C2); saving under a name in use asks before replacing it.
- **FR-025**: Saved views MUST be personal: nobody else sees them, and an address naming another person's view
  opens the reader's own default instead.
- **FR-026**: The Saved views menu MUST list the person's views of the report, mark the active one, and show
  "Modified" with **Update view** and **Save as new view** when the state differs from it; **Update view** asks
  before overwriting.
- **FR-027**: A person MUST be able to star one view per report as their default; opening the report with no
  filters in the address applies it; starring another moves the star; unstarring returns to FR-030's defaults.
- **FR-028**: A person MUST be able to rename and delete their views; deleting removes the view at once and the
  toast offers **Undo**, which restores it with its star. Every save, change, star and delete of a view MUST be
  written to the activity log with who and when (constitution VII), seen by the view's owner and by people who
  may see the whole activity log; a change made from a page that shows an older version of the view MUST be
  refused with "This view was changed in another tab — reload it to see the latest" instead of overwriting.
- **FR-029**: A view that names something that no longer exists or is outside the person's scope MUST apply the
  rest and say what was left out.

**Defaults, empty states and words**

- **FR-030**: With no filters in the address and no default view, a report MUST open on: the current year, every
  plan of the year in scope, every value ticked including **Not filled in**, Months, the person's display currency.
- **FR-031**: A report MUST say why it is empty: "No plans you may see for <year>" with the years that have plans;
  "Your filters exclude everything" with **Reset filters**; a widget-specific "No data yet" naming what would create
  it; and never show zeros as if they were figures.
- **FR-032**: Report pages MUST use Shell's words and the replacements of design system §8 and the build guide §7
  (Reports, Summary cards, Not filled in, Period, Saved views, Show); nothing on screen says Measure, Tactic,
  Program or HUD.
- **FR-033**: The Reports page and every report MUST meet WCAG 2.2 AA: the filter panel, menus and chart/table
  switches work by keyboard, Escape closes a menu or dialog without change, focus returns to the control that
  opened it; below 1024 px wide the panel collapses into **Filters (n)**.
- **FR-034**: **Reports** MUST be its own sidebar item, directly after **Dashboards**; the 036 Executive and
  Regional dashboards stay where they are, unchanged (039-C3, answers Q-031).

### Key Entities *(include if feature involves data)*

- **Report**: one entry of the Reports page — name, one-line purpose, whether it is available in this release,
  the permission it needs beyond `reports.view`, and which controls it uses (Period, Show).
- **Report filter**: the state a report is shown with — year, plans, plan types, teams, owners, activity types,
  activity statuses, the chosen values of each Shell-attribute group, period, currency, Show and each widget's
  chart or table choice. It lives in the page address; it is never stored except inside a saved view.
- **Saved report view**: a person's named report filter for one report — owner, report, name, the filter, whether
  it is their default, its version and who changed it when; created, renamed, updated and deleted only by its
  owner, each change in the activity log, with Undo on delete. A personal preference, not a business record, so
  a delete does not go to the recycle bin.
- **Widget**: one part of a report — title, the figures it shows and their definitions, a table equivalent, and
  its export.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: For every seeded person, year and filter combination in the automated set, the "In view" counts equal
  All work's counts for the same year and filters — in 100% of comparisons.
- **SC-002**: A person whose scope is one market never sees a name, count, filter choice or figure of another
  market on the Reports page, a report or an exported file — in 100% of scope tests.
- **SC-003**: At the reference volume (feature 028: 50 plans, 5,000 activities, 20,000 line items) a report's frame
  — filter panel choices and the "In view" line — is shown within 2 seconds of opening or applying.
- **SC-004**: A first-time user narrows a report to one market and one brand and saves it as their default view in
  at most 8 steps and under one minute, without help (design system §10).
- **SC-005**: Opening any report address reproduces the same year, filters, period, currency and Show — 100% of
  round-trip tests — and no address, however edited, breaks the page.
- **SC-006**: Every exported file opens in Excel or a text editor without a repair prompt, and every value in it
  equals the screen.
- **SC-007**: The Reports page and the report frame pass automated accessibility checks for WCAG 2.2 AA with no
  serious or critical finding, and every chart has a table equivalent.

## Assumptions

- A report covers one **plan year** (D-368); Hive9's multi-year selection and overlap semantics are not adopted,
  because plans are one year and multi-year linking is out of scope.
- Hive9's **Date sets** (from–to ranges on activity dates) and **Show hierarchies** filters are not offered: the
  year, activity status and Show cover the reports of wave A; a later report may add them through its own spec.
- **Activity category** (Hive9's Tactic Category) is a wave B concept (047) and is not a filter here.
- Amounts convert at the **current rate** (D-303); "Actual at monthly FX" stays open (Q-023).
- One permission `reports.view` opens every report; per-report rights as Hive9's Measure tab are not built (Q-028).
- Report sharing by e-mail and scheduled reports are out of scope (constitution 4.8); the address is the link.
- "Today" is the application's calendar date, as on every other page; timestamps are shown in the person's format.
- Saved views follow the lifecycle of All work's saved views (010 FR-029–FR-031) without sharing (D-368).
- The first reports (040–045) arrive in later features of the same wave; until then the Reports page lists them
  as "Coming soon", and the frame is exercised on their pages and in the component gallery.
