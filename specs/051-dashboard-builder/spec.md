# Feature Specification: Dashboard Builder

**Feature Branch**: `feature/051-dashboard-builder`

**Created**: 2026-11-01

**Status**: Planned

**Input**: User description: "A no-code dashboard builder using Apache ECharts: users create and edit their own
dashboards, visualise all their SP&A One data or data they import, with as many customisation options as
reasonable. Viewing at /dashboards, creating at /builder, a new sidebar item. End-to-end, built the SDD way."
(feature map row 051, D-371)

**References**: Constitution 4.9 — Current Phase & Scope (051 in scope), Principles IV (the server works out
every figure), VII (scope isolation), IX (experience), X (palette), F1 and F6 (charts and their tables) ·
ADR-0020 (Apache ECharts), ADR-0018 (content security policy), ADR-0007 (activity log and recycle bin) ·
D-371–D-378, Q-028, Q-029 · features 014 (permissions, scopes, Table A), 036 (built-in dashboards, figures),
012 (checked file imports), 010 (All work's year data), 015 (financial transactions), 017 and 031 (approval
rounds), 032 (purchase requests), 003 (teams) · `docs/product/domain-model.md` §1–§4, §7 ·
`docs/product/design-system.md` §2.4, §4.7, §5, §6, §9.

## Clarifications

### Session 2026-11-01

The product owner answered the scoping questions up front (design brief of 2026-11-01); each answer is recorded
as a decision.

- Q: Where do people view and build dashboards? → A: one list at `/dashboards` with 036's Executive and
  Regional dashboards as built-in cards beside the person's own and those shared with them; a custom dashboard
  at `/dashboards/:dashboardId`; building at `/builder` and `/builder/:dashboardId`; a new sidebar item
  **Dashboard builder** (D-372).
- Q: Which data can a widget show? → A: the Plan-module data the person may see — plans, campaigns,
  initiatives, activities, line items, planned cost by month, financial transactions, approvals, purchase
  requests, global budgets — always limited to the viewer's scope, plus CSV and Excel files they upload as
  datasets (D-373, D-376).
- Q: Who sees a dashboard? → A: its owner; the owner may share it view-only with teams they belong to, whose
  members see it with their own scope's data. Datasets are shared the same way, and a widget whose dataset the
  viewer may not see says so and shows nothing of it (D-375).
- Q: Which permissions? → A: **See dashboards** (`dashboards.view`) and **Build dashboards** (`dashboards.build`),
  added to 014 Table A; who holds them by default is in D-374.
- Q: Do dashboard filters apply to datasets? → A: no (Q-028); each widget can filter its own data.
- Q: Can administrators see private dashboards? → A: no (Q-029); private means private.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Build and save a dashboard from SP&A One data (Priority: P1)

Priya, a planner for Lubricants Europe, chooses **Dashboard builder** in the sidebar. An empty canvas opens with a
palette of widgets on the left and the dashboard's details on the right. She names the dashboard "Lubricants
Europe overview", clicks **Bar** in the palette and a bar chart appears in the first free place on the grid. In
the inspector's **Data** tab she chooses the source **Activities**, groups by **Activity type**, measures **Sum of
Planned cost** and sorts by that measure, descending; the chart shows her activities' planned cost by type at once.
She adds a **Line** chart of **Planned cost by month** grouped by **Month**, broken down by **Campaign**; a **Key
figure** with the **Sum of Budget** of her plans; and a **Table** of campaigns with budget and planned cost. In the
**Style** tab she gives the bar chart horizontal bars, data labels and money format; she drags the line chart to
the right of the bar chart and makes it wider, then uses **Undo** once because she went too far. She chooses
**Save**: the dashboard is saved, the address becomes `/builder/DB-1002` and a toast says "Dashboard saved".

**Why this priority**: it is the feature — without building and saving there is nothing to view or share.

**Independent Test**: sign in as a planner, build a dashboard with at least one widget of each data-bearing type
from built-in sources, save, reload and compare each widget's figures with the record pages and All work for the
same plans.

**Acceptance Scenarios**:

1. **Given** Priya holds Build dashboards, **When** she opens `/builder`, **Then** she sees a canvas titled "New
   dashboard", the widget palette, the inspector showing the dashboard's details and a yellow **Save** button.
2. **Given** an empty canvas, **When** she clicks or drags a widget type from the palette, **Then** a widget of
   that type is placed in the first free place (or where she drops it) with a default size, selected, and the
   inspector shows its Data tab.
3. **Given** a bar chart widget, **When** she chooses source Activities, group by Activity type and Sum of Planned
   cost, **Then** the preview shows one bar per activity type of the activities she may see, each the sum of their
   planned costs in US dollars as their activity pages show them, worked out by SP&A One.
4. **Given** Priya may see only Lubricants Europe, **When** any widget is previewed, **Then** no record, name or
   amount of Mobility China counts in it.
5. **Given** a widget, **When** she changes any option in the Data, Style or Layout tab, **Then** the preview
   follows within a second without saving anything.
6. **Given** a selected widget on the canvas, **When** she drags it or its resize handle, or presses the arrow keys
   (move one cell) or Shift with the arrow keys (resize one cell), **Then** the widget moves or resizes on the
   12-column grid, never over another widget or past the grid's edges, and the new position is announced.
7. **Given** changes since the last save, **When** she chooses **Undo** (or Ctrl+Z / ⌘Z), **Then** the last change
   is reversed; **Redo** (Ctrl+Shift+Z / ⌘⇧Z, or Ctrl+Y) applies it again.
8. **Given** a dashboard without a name or with a widget missing what its type needs, **When** she chooses Save,
   **Then** nothing is saved and every problem is listed at once — in a summary with links and on each widget
   ("Choose what to measure") — and the first is focused.
9. **Given** valid changes, **When** she chooses Save, **Then** the dashboard and its widgets are stored, the
   activity log records it, and she stays in the builder with the saved dashboard.
10. **Given** unsaved changes, **When** she leaves the builder, **Then** she is asked "Discard changes?" with
    **Keep editing** (focused) and **Discard changes**.
11. **Given** the dashboard was saved elsewhere after she opened it, **When** she saves, **Then** nothing is
    overwritten; she is told the dashboard changed meanwhile, her changes stay on the canvas, and **Reload** shows
    the newer version.
12. **Given** a person without Build dashboards, **When** they open `/builder`, **Then** they see the no-access
    page and the sidebar shows no Dashboard builder item.

---

### User Story 2 - View dashboards at Dashboards, built-ins kept (Priority: P1)

Daniel opens **Dashboards**. The list shows **Built-in** — the Executive and Regional dashboards — then **My
dashboards** (his own) and **Shared with me**, where "Lubricants Europe overview" by Priya Raman appears. He opens
it: the widgets are laid out as Priya placed them, each worked out for the plans Daniel may see. He hovers a bar to
read its value, hides a series by clicking its legend entry, chooses **Show as table** on the line chart, and
**Download as PNG** on the bar chart. He narrows the dashboard with the filters bar to the second half of 2026; the
address keeps the filters, so Back and a bookmark show the same view. Priya sees an **Edit** button on her own
dashboard; Daniel does not.

**Why this priority**: dashboards are built to be read; the list is the one way in, and the built-in dashboards
must stay reachable.

**Independent Test**: sign in as people with different scopes, open the same shared dashboard and compare each
widget with the record pages for the plans each may see; open both built-in dashboards from the list.

**Acceptance Scenarios**:

1. **Given** a person who holds View plans and See dashboards, **When** they open `/dashboards`, **Then** they see
   the Executive and Regional cards under Built-in, their own dashboards under My dashboards and the dashboards
   shared with their teams under Shared with me, each card with its name, description, owner, last update and the
   teams it is shared with.
2. **Given** the Executive card, **When** it is chosen, **Then** `/dashboards/executive` opens as before (036);
   `/dashboards/executive` and `/dashboards/regional` typed or bookmarked keep working.
3. **Given** a custom dashboard the person may see, **When** they open `/dashboards/DB-…`, **Then** every widget
   is shown in its place with its title, its figures worked out for the viewer's own scope, and a text summary for
   screen readers.
4. **Given** a chart, **When** the person points at or focuses a data point, **Then** a tooltip shows its group,
   series and formatted value; clicking a legend entry hides or shows that series.
5. **Given** a chart widget, **When** the person chooses **Show as table**, **Then** a table lists the same rows
   and figures; **Show as chart** returns.
6. **Given** a chart widget, **When** the person chooses **Download as PNG**, **Then** an image of the chart named
   "SP&A One – <dashboard> – <widget title>.png" downloads.
7. **Given** the filters bar, **When** the person sets a date range, plans or teams, **Then** every widget whose
   source has that field is worked out again for it, widgets whose source does not have it say "Not filtered by
   date" (or plan, team), and the address keeps the filters; **Reset filters** returns to the dashboard's saved
   filters.
8. **Given** the dashboard's owner, **When** they view it, **Then** **Edit** opens `/builder/DB-…`; other viewers
   see no Edit.
9. **Given** a dashboard the person may not see (private to someone else, unknown, or deleted), **When** they
   open its address, **Then** they see "Dashboard not found" exactly as for a missing one.
10. **Given** a widget whose data takes time, fails or is empty, **When** the dashboard is shown, **Then** that
    widget alone shows loading in reserved space, an error with Retry, or "No data to show" with what to change;
    the other widgets are unaffected.
11. **Given** a person without View plans, **When** they open `/dashboards`, **Then** the Built-in section is not
    shown; a person without See dashboards sees only Built-in.
12. **Given** **Refresh**, **When** it is chosen, **Then** every widget asks for its data again.

---

### User Story 3 - Upload a dataset and chart it (Priority: P2)

Priya has a spreadsheet of event leads from an agency. On **Datasets** (linked from Dashboards and from the
builder's source picker as **Upload data**) she chooses **Upload data** and picks the .xlsx file. SP&A One reads
the first sheet and shows its columns with a proposed type each — Event (Text), Date (Date), Leads (Number), Cost
(Money) — and the first rows. She renames "Cost" to "Agency cost", sets its currency to EUR, names the dataset
"Event leads 2026" and saves. Back in the builder, she picks **Event leads 2026** under Uploaded data as a source
and charts leads by month.

**Why this priority**: markets hold data SP&A One does not; charting it beside plan data is what makes the builder
self-service, but building from SP&A One data comes first.

**Independent Test**: upload CSV and .xlsx files with each column type, correct types, save, chart the dataset,
and compare the figures with the spreadsheet's own sums; upload files that break each limit.

**Acceptance Scenarios**:

1. **Given** a CSV (comma, semicolon or tab separated, UTF-8) or .xlsx file within the limits, **When** Priya
   uploads it, **Then** SP&A One shows the file name, the sheet used (the first), the number of rows, each column
   with its name from the first row and a proposed type with sample values, and the first 20 rows — and stores
   nothing yet.
2. **Given** the check, **When** she changes column names or types and saves, **Then** every cell is checked
   against its column's type and either the dataset is stored or every problem is listed at once ("Row 14,
   Leads: 'n/a' is not a number"), with nothing stored.
3. **Given** a file over 5 MB, with more than 10,000 rows or 50 columns, not readable as CSV or .xlsx, or without
   a header row, **When** it is uploaded, **Then** it is refused with the reason and the limit.
4. **Given** a saved dataset, **When** it is chosen as a widget's source, **Then** its columns are the fields,
   grouped, measured and filtered like any built-in field, and money columns are shown in US dollars converted at
   the current rate.
5. **Given** the Datasets page, **When** Priya opens a dataset, **Then** she sees its details, columns, rows page by
   page, and the dashboards that use it that she may see; she may rename it, change its description and column
   names, share it or delete it.
6. **Given** a dataset used by widgets, **When** she deletes it, **Then** she is told how many of her dashboards use
   it and that those widgets will show "This data is no longer available"; it moves to the recycle bin with Undo.

---

### User Story 4 - Share a dashboard view-only with teams (Priority: P2)

Priya chooses **Sharing** in the builder and ticks **Lubricants Europe**, one of her teams. Daniel and Sara, its
members, now find the dashboard under Shared with me. Sara may see both markets, so the shared widgets show both
plans' figures to her; Daniel sees Lubricants Europe only. One widget charts Priya's dataset, which she has not
shared: Sara sees "You don't have access to this data" in that widget, and nothing of it, until Priya shares the
dataset too.

**Why this priority**: sharing turns a personal tool into a team one; the dashboard is useful to its owner first.

**Independent Test**: share a dashboard and a dataset with a team, sign in as members with different scopes and as
a non-member, and check what each sees.

**Acceptance Scenarios**:

1. **Given** the owner, **When** they open Sharing, **Then** they see the teams they belong to (and any team shared
   before they left it), ticked where shared, with the note "People in these teams can view this dashboard. Each
   of them sees only the data they may see."
2. **Given** a share saved, **When** a member of the team who holds See dashboards opens Dashboards, **Then** the
   dashboard is under Shared with me, view-only (no Edit, no Delete, no Sharing).
3. **Given** a member with a different scope, **When** they view a shared widget of a built-in source, **Then** its
   figures cover only the records they may see.
4. **Given** a widget of a dataset not shared with the viewer, **When** the dashboard is shown, **Then** the widget
   shows "You don't have access to this data" and neither the dataset's name, its columns nor any figure.
5. **Given** someone who is not the owner, **When** they try to change, share or delete the dashboard through the
   API, **Then** it is refused (403 when they may see it, 404 when they may not).
6. **Given** a viewer who holds Build dashboards, **When** they choose **Make a copy**, **Then** a dashboard "Copy of
   <name>" owned by them, shared with nobody, opens in the builder.
7. **Given** a share is removed or a member leaves the team, **When** they next open the dashboard, **Then** it
   answers "Dashboard not found".

---

### User Story 5 - Further customisation (Priority: P3)

Carlos wants a dashboard for leadership: he adds a reference line at the target on a bar chart, a gauge with a
minimum and maximum, a key figure compared with the previous period, a treemap of planned cost by campaign and
activity type, a radar of campaigns across measures, and text notes explaining the figures; he adds activity Shell
attributes (Cluster, Country / OU) as fields, chooses the **Planning** palette so budget and actual have their usual
colours, and sets the dashboard's default filters.

**Why this priority**: each option adds polish to dashboards people can already build, view and share.

**Independent Test**: set every display option of every widget type and check the preview, the saved dashboard and
the table view.

**Acceptance Scenarios**:

1. **Given** a bar, line, area or combo chart, **When** a reference line with a value and a label is set, **Then** a
   line at that value with its label is drawn across the plot.
2. **Given** a key figure with comparison Previous period and a dashboard date range, **When** it is shown, **Then**
   it shows the value, the value of the period of the same length just before, and the change in percent worked
   out by SP&A One, coloured and worded as better or worse by "Higher is better".
3. **Given** the Activities source, **When** fields are listed, **Then** the activity-level Shell attributes appear
   as text fields by their Settings names.
4. **Given** a dashboard's details, **When** default filters are saved, **Then** every viewer opens the dashboard
   with them until they change the filters bar.

---

### Edge Cases

- A widget whose source the viewer may not read at all (Financial transactions without See financial data, Global
  budgets without See global budgets): "You don't have access to this data", like a dataset not shared.
- Financial fields of built-in sources (Committed, Actual) where the viewer sees financial data in some plans only:
  they count those plans and the widget says "Financial figures cover 1 of 2 plans"; with none, those fields count
  as empty.
- A grouped field with several values (a plan with two teams): the record counts in each group and once in totals
  (036 FR-012); the widget says "A plan with several teams counts in each".
- Empty values: grouped under "Not set"; ignored by sum, average, minimum and maximum; counted by Count.
- More groups than the limit: the widget shows the top N by the sort and says "Showing 10 of 37"; with **Group the
  rest as Other** the rest is one "Other" group.
- Average of no values, a percentage of a total of zero, a comparison with zero: "—", never 0 or an error.
- An amount in a currency removed from Settings (dataset money column): left out of sums and said so (015 D-257).
- A plan, campaign or dataset deleted while a widget filters by it: the filter value shows "(deleted)" in the
  builder and matches nothing.
- A dataset deleted (or its owner's share removed): its widgets show "This data is no longer available" to the
  owner and "You don't have access to this data" to others; Undo or Restore brings them back.
- A team a dashboard or dataset is shared with is deleted: the share goes with it, so its members no longer see
  the dashboard through that team; restoring the team does not restore the share, and a dashboard or dataset
  restored from the recycle bin keeps only the shares of teams that still exist (D-379).
- A dashboard's owner made inactive: the dashboard stays visible to the teams it is shared with; nobody can edit it
  (Q-029).
- Two people saving the same dashboard: the second save is refused as changed meanwhile; nothing is lost on screen.
- The grid full at the bottom: new widgets go below the last row; the canvas grows (at most 200 rows, 30 widgets).
- A very long title, label or category name: shortened with "…" on the chart, in full in the tooltip and table.
- A spreadsheet with formulas: their saved results are read; merged cells, empty rows and columns without a name
  are reported; a CSV not in UTF-8 is refused with "Save the file as CSV UTF-8".
- Text notes and every title, label and cell are plain text: anything that looks like HTML or a formula is shown
  as written, never interpreted.
- A narrow screen (below 960 px): the viewer stacks widgets in one column in grid order; the builder says "The
  builder needs a wider window" and offers to view the dashboard.

## Requirements *(mandatory)*

### Functional Requirements

**Access and navigation**

- **FR-001**: Two permissions MUST be added to 014 Table A under a group **Dashboards**, both not scoped: **See
  dashboards** (`dashboards.view`, "See custom dashboards shared with them") and **Build dashboards**
  (`dashboards.build`, "Build dashboards and upload data"); the built-in roles MUST hold them as D-374 lists,
  stored roles included.
- **FR-002**: The sidebar MUST offer **Dashboard builder** (to holders of Build dashboards), opening `/builder`;
  **Dashboards** MUST stay and open `/dashboards`, shown to holders of View plans or See dashboards.
- **FR-003**: `/dashboards` MUST list **Built-in** (Executive and Regional, to holders of View plans), **My
  dashboards** and **Shared with me** (to holders of See dashboards), each by name; each card shows name,
  description, owner, last update, widget count and shared teams. **New dashboard** (to holders of Build
  dashboards) is the page's one primary action, and **Datasets** a secondary link.
- **FR-004**: `/dashboards/executive` and `/dashboards/regional` MUST keep working as feature 036 specifies.
- **FR-005**: A custom dashboard MUST be visible only to its owner and to members of the teams it is shared with,
  each holding See dashboards; to anyone else it MUST answer exactly like a missing dashboard (404).
- **FR-006**: Only the owner, holding Build dashboards, MAY edit, share, rename or delete a dashboard; a refused
  action on a dashboard the person can see answers 403 with the reason.
- **FR-007**: Every widget's data MUST be worked out for the person viewing, from the records they may see: plans
  in their View plans scope and everything in them; financial fields and transactions only of plans where they
  see financial data; global budgets only with See global budgets. Nothing outside it may count or be named.

**Dashboard record**

- **FR-008**: A dashboard MUST have a **name** (required, at most 80 characters, unique among the owner's
  dashboards ignoring case), a **description** (at most 500), an owner, the teams it is shared with, its default
  **filters** and its **widgets** (at most 30), with an ID `DB-…`, version, created and updated by and when.
- **FR-009**: Creating, changing, sharing, copying and deleting a dashboard MUST each write an activity-log entry
  naming the person and what changed (name, description, filters, sharing, widgets added, removed and changed).
- **FR-010**: Deleting a dashboard MUST move it to the recycle bin with Undo for at least 10 seconds, after a
  confirmation naming it and saying who it is shared with; its owner may restore it from the recycle bin.
- **FR-011**: **Make a copy** MUST create a dashboard owned by the person with the same widgets and filters, named
  "Copy of <name>" (shortened to fit 80, and numbered if taken), shared with nobody; it needs Build dashboards and
  sight of the original.
- **FR-012**: Saving MUST send the version the editor started from; a dashboard changed meanwhile MUST be refused
  (version conflict) without losing the editor's changes, and offer **Reload**.

**Widgets**

- **FR-013**: A widget MUST have a type, a title (at most 80), a subtitle (at most 120), a position on a 12-column
  grid — column 1–12, row 1–200, width 2–12 columns, height 2–12 rows, inside the grid — its data query and its
  display options. Widgets MUST NOT overlap.
- **FR-014**: The widget types MUST be: **Bar** (vertical or horizontal; grouped, stacked or 100 %), **Line**,
  **Area**, **Bar and line** (bars and lines on one or two value axes), **Pie**, **Donut**, **Scatter**, **Funnel**,
  **Gauge**, **Heatmap**, **Treemap**, **Radar**, **Key figure** (one value with an optional comparison),
  **Table** and **Text note** (plain text, at most 2,000 characters, no formatting).
- **FR-015**: Each type MUST need what it shows, checked on save and by the API (data model, widget rules): a
  group for bar, line, area, bar and line, pie, donut, scatter, funnel, heatmap, treemap and radar; a breakdown
  for heatmap; one measure for pie, donut, funnel, gauge, heatmap, treemap, radar and key figure; two or three for
  scatter (x, y, size); two to four for bar and line; one to four for bar, line and area (one when broken down);
  one to eight for table; no query for text note.

**Data query** (worked out by SP&A One; the web app aggregates nothing)

- **FR-016**: A query MUST read one **source**: a built-in source — Plans, Campaigns, Initiatives, Activities, Line
  items, Planned cost by month, Financial transactions, Approvals, Purchase requests, Global budgets — or a dataset
  the viewer may see. The **data catalogue** MUST list the sources the person may read with their fields, each
  field with its name, kind (text, number, money, date) and whether it can be grouped and measured.
- **FR-017**: A query MAY **group by** one field (the categories or x axis) and **break down by** a second (the
  series); a date field MUST be grouped by **day, week (ISO, from Monday), month, quarter or year**.
- **FR-018**: A query MUST have **measures** (up to 8), each **Count of records**, or a field with **Count**,
  **Count distinct**, **Sum**, **Average**, **Minimum** or **Maximum**; sum and average only of number and money
  fields, minimum and maximum also of dates. Amounts MUST be exact (never floating point) in US dollars;
  averages of numbers have at most 4 decimals and of money 2, rounded half up.
- **FR-019**: A query MAY have up to 10 **filters**, each a field, an operator — is, is not, is one of, contains,
  greater than, at least, less than, at most, between, is empty, is not empty — and the values the operator needs,
  typed by the field (text, number, money in USD, date); all filters must hold (and).
- **FR-020**: A query MUST be **sorted** by the group or by one measure, ascending or descending (default: dates and
  the group ascending, otherwise the first measure descending), and MAY be **limited** to the top N groups (1–500,
  default 50), optionally with **Group the rest as Other**; the answer MUST say how many groups there were.
- **FR-021**: The answer MUST give, per value, its **share** of its group (with a breakdown) or of the total
  (without), for counts and sums of non-negative values, so pie labels and 100 % bars use SP&A One's figures.
- **FR-022**: Dashboard **filters** — date range, plans, teams — MUST apply to every built-in source that has the
  field (each source names its date, plan and team field) and never to datasets; a widget not filtered by one says
  so.
- **FR-023**: A **key figure** MAY compare with a **target** (a number it is shown against, with the share of target
  reached) or the **previous period** (the same length just before the dashboard's date range); the API MUST work
  out the comparison value and the change in percent.

**Display options** (each optional with a default; chosen in the Style tab)

- **FR-024**: Colours MUST come from four palettes only — **Categorical**, **Planning**, **Levels**,
  **Sequential** — whose colours are design tokens (design system §4.7); never Primary Yellow or Secondary Red;
  patterns on bars, areas and slices so colour is never the only carrier.
- **FR-025**: Charts MUST offer, where the type draws them: **legend** (shown or not; top, bottom, left, right),
  **data labels**, **axis titles** (horizontal, vertical, right; at most 40 characters), **gridlines**, **smooth
  lines**, **stacking** (none, stacked, 100 %), **orientation** (vertical, horizontal), **number format** (number,
  money, percent, compact — "1.2M"), **decimals** (0–4), **zoom slider**, **reference line** (value and label),
  per measure of bar and line **bars or line** and **left or right axis**, gauge **minimum and maximum**, key
  figure **comparison**, **target** and **higher is better**, and for tables the **columns** shown in order and a
  **totals row**.

**Builder** (`/builder`, `/builder/:dashboardId`)

- **FR-026**: The builder MUST show a header with the dashboard's name, **Undo**, **Redo**, **Sharing** (saved
  dashboards), **View** and **Save** (the one yellow action); a left **palette** of widget types grouped Charts,
  Figures and tables, Text; a centre **canvas** (the 12-column grid with each widget's live preview); a right
  **inspector**: the dashboard's details (name, description, default filters) when no widget is selected, otherwise
  the widget's **Data**, **Style** and **Layout** tabs.
- **FR-027**: A widget type MUST be added by clicking it (first free place, default size) or by dragging it onto the
  canvas (where dropped, if free), and selected.
- **FR-028**: Widgets MUST be moved and resized with the pointer (drag the widget, drag the corner handle) and with
  the keyboard on the focused widget — arrow keys move one cell, Shift+arrow keys resize one cell — and by the
  Layout tab's column, row, width and height; never overlapping and never outside the grid; each change is
  announced to screen readers ("Moved to column 3, row 2").
- **FR-029**: A widget MUST offer **Duplicate** (placed in the next free place) and **Delete** (Delete key or menu,
  undoable in the editor).
- **FR-030**: **Undo** and **Redo** MUST reverse and repeat the editor's changes since it opened (at least 50 steps;
  keyboard Ctrl/⌘+Z, Ctrl/⌘+Shift+Z, Ctrl+Y); they change only the editor, not saved data.
- **FR-031**: The Data tab's fields MUST come from the data catalogue for the chosen source; changing the source
  keeps the options that still fit and clears the others, saying which.
- **FR-032**: The preview MUST ask SP&A One for the widget's data whenever its query or the dashboard filters
  change (waiting 400 ms for typing to settle) and show the widget's states as the viewer does.
- **FR-033**: **Save** MUST never be disabled; it checks everything at once — name, every widget's needs, layout —
  shows every problem in a summary with links and on each widget, focuses the first, and saves nothing while any
  remain. Enter in a dialog submits it; Escape closes it without saving.
- **FR-034**: Leaving the builder with unsaved changes MUST ask "Discard changes?" with **Keep editing** (focused)
  and **Discard changes**.
- **FR-035**: Opening `/builder/:dashboardId` for a dashboard the person does not own MUST answer like a missing
  dashboard when they cannot see it and show it view-only with **Make a copy** when they can.

**Viewer** (`/dashboards/:dashboardId`)

- **FR-036**: The viewer MUST show the dashboard's name, description, owner and sharing, the filters bar,
  **Refresh**, **Edit** (owner), **Make a copy** (holders of Build dashboards) and the widgets on the grid, stacked
  in grid order below 960 px.
- **FR-037**: The filters MUST live in the page address; the dashboard's saved filters apply when the address has
  none; an unknown value falls back to the saved filter and never breaks the page.
- **FR-038**: Every chart MUST have tooltips, legend toggling, **Show as table** (a real table of the same rows) and
  **Download as PNG**; every widget a text summary for assistive technology.
- **FR-039**: Each widget MUST render its own states: loading (reserved space), error with Retry, **No data to
  show** with what to change, **You don't have access to this data** (source or dataset not readable by the
  viewer), **This data is no longer available** (dataset deleted, to its owner) and the data; the page MUST
  render loading, error, not found and the dashboard.

**Datasets** (`/datasets`)

- **FR-040**: Holders of Build dashboards MUST be able to list their datasets and those shared with them, upload a
  dataset, open one, rename it, change its description and column names, share it and delete it.
- **FR-041**: Uploading MUST first **check** the file without storing it: CSV (UTF-8, comma, semicolon or tab) or
  .xlsx (first sheet), first row as column names; proposing for each column **Text**, **Number**, **Date** or
  **Money** with samples; and showing the first 20 rows and every problem.
- **FR-042**: Saving MUST store the dataset with the names, types and money currencies chosen, after checking every
  cell against its type and listing every problem at once; empty cells are empty values.
- **FR-043**: Limits MUST be enforced: 5 MB file, 10,000 rows, 50 columns, column names unique and at most 60
  characters, cells at most 500 characters, dataset names at most 80 characters and unique among the owner's,
  descriptions at most 500, at most 50 datasets per person.
- **FR-044**: A dataset MUST be visible only to its owner and to members of the teams it is shared with — on the
  Datasets page and as a builder source to those holding Build dashboards, as widget data to those holding See
  dashboards — and answer 404 to anyone else; only its owner changes, shares or deletes it; column types are fixed
  once saved.
- **FR-045**: Creating, changing, sharing and deleting a dataset MUST write the activity log; deleting moves it to
  the recycle bin with Undo, after a confirmation saying how many of the owner's dashboards use it.

**Sharing**

- **FR-046**: The owner MUST be able to share a dashboard or dataset view-only with any team they belong to, remove
  any shared team, and see who it is shared with; sharing changes are versioned like any change.

### Key Entities

- **Custom dashboard**: a named layout of widgets with default filters, owned by one person and shared view-only
  with teams (`DB-…`). Stored, versioned, logged, deleted to the recycle bin.
- **Widget**: one chart, key figure, table or text note on a dashboard's grid: type, title, subtitle, position,
  data query, display options. Part of its dashboard.
- **Data query**: source, group, breakdown, measures, filters, sort and limit of a widget; worked out by SP&A One.
- **Data source**: a built-in source of Plan-module records or a dataset, with its fields.
- **Dataset**: a table uploaded from a CSV or Excel file: name, description, owner, shared teams, typed columns and
  rows (`DS-…`). Stored, versioned, logged, deleted to the recycle bin.
- **Dashboard filters**: date range, plans and teams applied to every built-in widget.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A first-time planner builds and saves a dashboard with three widgets from SP&A One data in under 5
  minutes without help (design system §10).
- **SC-002**: For every seeded person, every widget of the demo dashboard equals the same figure on the record pages
  and All work for the plans that person may see — 100 % of automated comparisons.
- **SC-003**: A person never sees a figure, name or count of a plan outside their scope, or of a dataset not shared
  with them, in any widget, tooltip, table or image — 100 % of scope tests.
- **SC-004**: A dashboard of 12 widgets at the feature 028 reference volume shows every widget within 3 seconds, and
  one widget's preview updates within 1 second of a change.
- **SC-005**: A 10,000-row, 50-column file is checked within 5 seconds and saved within 10.
- **SC-006**: Every builder action is possible with the keyboard alone, and every page passes an automated WCAG 2.2
  AA check in both themes.
- **SC-007**: No page of the feature reports a content security policy violation (end-to-end check).

## Assumptions

- Dashboards hold no figures of their own: every figure is worked out when shown, so they are always current.
- Dashboard and widget data are not exported to PowerPoint or PDF (036 exports the built-in dashboards); a chart is
  downloaded as a PNG image only.
- Datasets are static: a new file is a new dataset (replacing a dataset's data is a later step).
- Calculated fields typed by people, joins between sources, drill-down to records, scheduled or e-mailed dashboards,
  public links and live data connectors are out of scope (constitution 4.9.0).
- Shell attributes appear as fields of the Activities source only (P3); other levels' attributes are a later step.
- The demo data gains one dashboard by Priya Raman shared with Lubricants Europe and one dataset (appended, no
  existing demo ID moves).
