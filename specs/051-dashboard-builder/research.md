# Research: Dashboard Builder

Each section states the decision, why, and what was rejected.

## R1 — Chart library: Apache ECharts 6.1.0, modular, SVG renderer (ADR-0020)

**Decision**: `echarts@6.1.0` (Apache-2.0; published 2026-05-19 — `npm view echarts time`: 6.0.0 2025-07-30, 6.1.0
2026-05-19, nothing newer; well past pnpm's one-day `minimumReleaseAge`), dependencies `zrender@6.1.0`
(BSD-3-Clause) and `tslib@2.3.0` (0BSD), no install scripts. Imported from `echarts/core` with only the renderer,
components, features and charts used (list in ADR-0020); `echarts.init(el, null, {renderer: 'svg', locale: 'EN'})`
in our own `EChart` component (resize observer, theme re-render, dispose). Measured bundle (esbuild 0.25, minified /
gzip): core + SVG 315 / 108 KB; with the common components 529 / 181 KB; everything the builder registers 759 / 256
KB; whole library 1,141 / 382 KB. Per chart type on top of the components: bar +26 / +9 KB, line +30 / +11, pie
+16 / +6, scatter +14 / +5, funnel +9 / +3, gauge +11 / +4, heatmap with visualMap +45 / +14, treemap +40 / +14,
radar +13 / +4. Only the lazy custom-dashboards routes load it (F8).

**Why**: one library covers all twelve chart types with one option model; maintained by the Apache Software
Foundation (6.0 in July 2025, 6.1 in May 2026); SVG keeps charts sharp, inspectable in browser tests and drawable
as a PNG at any scale.

**Rejected**: hand-written SVG for twelve types; Recharts, Nivo, Victory (gaps: funnel, gauge, treemap, radar,
heatmap together); Chart.js (canvas only, plugins for treemap and funnel); Highcharts (commercial); Vega-Lite
(expression evaluation vs. CSP); `echarts-for-react` (imports all of ECharts); the canvas renderer (same size, no
benefit below 500 groups).

## R2 — Accessibility of charts

**Decision**: `aria: {enabled: true, label: {description}}` with our own description ("Bar chart, Sum of Planned
cost by Activity type, 6 bars; highest Event 240,000.00 USD"), built by `describeWidget` from the API's rows;
`aria.decal.show: true` so bars, areas and slices carry patterns (WCAG 1.4.1); **Show as table** renders a real
`<table>` with caption, column headers and the same formatted values (F6); legend entries are buttons in ECharts'
SVG and are also listed in words under the chart; the canvas grid is a list of focusable widget regions with
keyboard moves (R10). ECharts claims no WCAG conformance; the table and description carry the information.

**Rejected**: ECharts' generated aria description (wording not ours, lists every value of 500 points); colour
alone for series.

## R3 — Theming and palettes from tokens (constitution X)

**Decision**: `buildOption` takes a `ChartTheme` read once per render from computed styles of the chart element:
`--ink` (text), `--ink-2` (secondary text), `--hairline` (gridlines, axes), `--surface` (background, PNG fill) and
the palette tokens. Four palettes (design system §4.7, D-377):

| Palette | Tokens (light / dark) |
|---|---|
| Categorical | `--chart-cat-1`…`-8`: Shell dark blue `#003C88`/`#6F9BD1`, Shell teal `#008192`/`#3FB6C5`, Shell purple `#87217D`/`#C96BBF`, green `#008557`/`#3DBB8C`, magenta `#A3195B`/`#E06A9F`, blue `#336094`/`#8FB4E3`, slate `#5B6B7A`/`#8696A6`, grey `#6E6E6E`/`#8E99A4` |
| Planning | `--chart-budget`, `--chart-planned`, `--chart-committed`, `--chart-actual` (036), then Categorical 5–8 |
| Levels | `--level-plan`, `--level-campaign`, `--level-initiative`, `--level-activity`, `--level-line-item` |
| Sequential | `--chart-seq-1`…`-5`: `#C9D8EA`, `#93B1D6`, `#5E8AC1`, `#336094`, `#003C88` / dark `#1F3A5C`, `#2B5482`, `#4B78AA`, `#6F9BD1`, `#A9C6EA` |

Every value is already in Shell's palette or the design system's meaning colours (no new hue, no yellow, no red);
each Categorical colour is ≥ 3:1 on its surface. Heatmaps use a **piecewise** visualMap of 5 bins, each exactly one
sequential token, so no colour is interpolated (no tints), and the legend names each bin's range. More than 8
series repeat the palette with different decal patterns.

**Rejected**: an ECharts registered theme (a second source of colour); continuous visualMap (interpolated tints).

## R4 — Content security policy (ADR-0018)

**Decision**: checked in `echarts@6.1.0`/`zrender@6.1.0` sources: the SVG painter sets styles through
`element.style.cssText` (CSSOM, allowed) and presentation attributes (`fill`, `stroke`), never `style` attributes;
`new Function` exists only in the geo component (not imported); the HTML tooltip writes HTML strings with inline
styles (refused by `style-src 'self'`) unless its formatter returns DOM nodes, which are appended as they are, and
its arrow is another HTML string unless `confine: true`. So every tooltip's formatter builds elements with
`textContent` and CSS-module classes, and `confine: true`. The toolbox (data view writes HTML) is not used. The
e2e suite fails on any `securitypolicyviolation` already.

## R5 — Widgets stored as rows with JSON query and display, validated by the service

**Decision**: one row per widget (`custom_dashboard_widgets`) with the layout and type as checked columns, the
source as a column (indexed, for "used by" and the dataset delete warning) and `query` / `display` as JSONB objects
whose shape is the contract's `WidgetQuery` / `WidgetDisplay`. The service decodes them into Go structs with
`DisallowUnknownFields`, validates them (data model §2–§3) on every write, and re-validates the fields against the
catalogue when the widget's data is asked for (a dataset column may be renamed but never retyped, so stored
queries stay valid; a deleted dataset answers "no longer available").

**Why**: the query has 8 measures × 10 filters × values and 15 display options that differ by type — normalised
tables would be a dozen tables written and read together for one dashboard, with no query ever filtering on them.
Grid geometry and type, which do have invariants the database can hold, are columns with constraints.

**Rejected**: one JSON column for the whole dashboard (no constraints at all); fully normalised measures, filters
and options (no benefit, many joins).

## R6 — Datasets: relational columns, one JSON array per row

**Decision**: `datasets`, `dataset_columns` (kind, currency, label constraints), `dataset_shares`, and `dataset_rows`
with `cells jsonb` (an array, one value per column). Values are typed at upload: numbers as exact decimal text,
dates `YYYY-MM-DD`, money as minor units of the column's currency, text as text, `null` empty. Rows are inserted
with `pgx.CopyFrom` (one round trip for 10,000 rows) and read in one statement.

**Why**: columns vary per dataset; an entity-attribute-value table (one row per cell, up to 500,000 rows per
dataset) multiplies storage and reads by 50; a table per dataset (DDL at run time) breaks sqlc, goose and
constraints. Typing at upload keeps the query engine free of parsing and errors.

**Rejected**: EAV cells; dynamic tables; storing the uploaded file (re-parsing on every widget); `numeric` per
column (needs dynamic tables).

## R7 — Upload: check, then save, with the 012 reader

**Decision**: the file travels base64 in JSON (`{fileName, content}`), as 012's workbook import
(`ImportUpload`), within 5 MB before encoding (operation body limit 7 MB, 413 `too_large` above). `POST
/datasets/preview` reads it with `internal/platform/sheet` — `ReadCSV(data, maxRows)` (UTF-8, BOM, delimiter
detection) or `ReadXLSX(data, maxRows)` (first table; zip parts limited to 50 MB unpacked; dates from styles,
formulas by saved value) — with `maxRows = 10,001` (header + 10,000) and answers the proposal (data model §7);
nothing is stored. `POST /datasets` sends the file again with the names, kinds and currencies chosen and stores it
in one `Write`. Stateless: no half-uploaded files to clean up.

**Rejected**: multipart upload (no other operation uses it; the contract tests and openapi-fetch handle JSON);
storing the checked file between the calls (cleanup, expiry, a new table).

## R8 — Visibility and sharing

**Decision**: the person comes only from the actor; "visible" (owner, or a share to a team the person is a member
of now) is one SQL condition in every store read of dashboards and datasets, so a record outside answers 404 like a
missing one (constitution VII). Permissions are checked by the route table (`need`) and the service (owner for
changes, 403 `forbidden` when visible but not owned). Share targets: the owner's current teams
(`people.Store` memberships) plus teams already shared (removable). Each visible read has a PostgreSQL
scope-isolation test (a second person's dashboard and dataset, shared and not).

## R9 — Built-in sources through ports, over existing scoped reads

**Decision**: `widgetdata` defines one port per need (tasks.md, "Go ports") and `internal/app` adapts the owning
areas, as 036 does (`internal/app/dashboards.go`): All work's year data (`allwork.Source.Years`, `.Year` — plans in
`plans.view` scope with campaigns, initiatives, activities, line items, USD figures, spreads, GSAP facts,
completeness) for Plans, Campaigns, Initiatives, Activities, Line items and Planned cost by month (with
`allwork.PlannedByActivity`); `financials` for transactions of the in-scope line items whose finance the person
sees (the scoped `SpendOnLineItems` of 036 or a sibling read); `approval.Service.RoundSummaries` for Approvals;
`purchaserequest` scoped reads for Purchase requests; `budget.Service.YearFigures` for Global budgets. The years
read are those overlapping the dashboard's date range, or every year with plans in scope. Within one request the
year data is read once (`internal/platform/memo`).

**Rejected**: SQL aggregates per source (redefines planned cost, conversions and spreads — 036 R1 rejected the
same); calling record services per plan (N+1).

## R10 — The grid editor (hand-built)

**Decision**: CSS grid of 12 columns, 80 px rows, 16 px gaps; widgets positioned by `grid-column` / `grid-row`.
Pointer: drag a widget's header to move, the corner handle to resize, with pointer capture; the target cell is
computed from the pointer and the column width; a move or resize that would overlap or leave the grid snaps back to
the last valid cell (no pushing others). Keyboard on the focused widget (`role="group"`, `aria-roledescription=
"widget"`): arrows move one cell, Shift+arrows resize, Delete removes, Ctrl/⌘+D duplicates, Escape deselects; a
polite live region announces the result. The Layout tab gives the same with numbers (WCAG 2.5.7, an alternative to
dragging). Editor state is a reducer (`builderReducer`) with an undo stack of past states (50) and a redo stack;
every change is an action, so undo/redo and the unsaved-changes guard compare states, not events.

**Rejected**: react-grid-layout and gridstack (dependencies with their own styles and inline style attributes; no
keyboard model we control).

## R11 — Exact numbers

**Decision**: numbers are `widgetdata.Number` (int64 ten-thousandths) and money `money.Amount` (cents, USD after
conversion with `money.Convert` at the current rate, D-258); sums check overflow; averages round half up (4 places
for numbers, 2 for money); shares are percent with one decimal, half up, null when the total is zero or a value is
negative. The web app receives decimal strings and formats them; it converts to a JavaScript number only to place a
mark on the chart, never to show a figure.

## R12 — PNG download

**Decision**: `chart.getDataURL()` (SVG with the SVG renderer) → `Image` → canvas at device pixel ratio 2, filled
with `--surface` first → `toBlob('image/png')` → a download named "SP&A One – <dashboard> – <widget>.png" (file name
characters `/\:*?"<>|` replaced by "-"). `img-src 'self' data:` allows it; nothing is sent anywhere.

## R13 — Widget data requests

**Decision**: one `POST /data-sources/query` per widget with `{query, filters, comparison?}`; TanStack Query key
`['widget-data', source, hash(query, filters)]`, `staleTime` 60 s, so identical widgets share an answer; the builder
debounces 400 ms. The API caps work: 500 groups, 8 measures, 10 filters, 10,000 dataset rows. **Refresh**
invalidates `['widget-data']`.

**Rejected**: one call for the whole dashboard (one slow widget would hold up all; the builder previews one at a
time anyway).

## R14 — Permissions and their defaults

**Decision**: D-374. Constants `authz.DashboardsView`, `authz.DashboardsBuild`, group `GroupDashboards =
"Dashboards"` after Global budgets; seed roles in `seed/access.go`; migration 00028 adds them to stored built-in
roles. The Administrator holds every permission (`authz.All()`), so it gains both automatically in the seed.

## R15 — Threat and risk model (constitution VII)

| Threat | Control |
|---|---|
| Reading another market's data through a widget | Built-in sources read only through scoped reads (R9); scope from the actor; contract tests per persona; SC-003 |
| Reading a private dataset | Every dataset read uses the visible condition; 404 otherwise; widget answers "no access" without the name |
| Learning that a private dashboard exists | 404 like missing; `usedBy` lists only visible dashboards |
| Injection through field names or values | Fields only from the catalogue allowlist; no SQL built from a query (the engine works in Go on rows); values parsed by kind |
| Script or style injection through titles, notes, dataset cells | Plain text everywhere; React escapes; tooltips built with `textContent`; CSP |
| Zip bomb, huge files, huge queries | 5 MB body, 50 MB per unpacked part, 10,000 rows, 50 columns, limits on measures, filters, groups |
| Spreadsheet formula injection | Not applicable: datasets are never exported; formulas are read as their saved values |
| Denial of service by many widgets | 30 widgets per dashboard, per-request memo of year data, query caps |

## R16 — Numbering

**Decision**: D-378. Migrations 00028 and 00029 collide with `measure-module`'s `00028_reports.sql`; ADR-0020 with
the performance-feed ADR planned there. The branch merged second renumbers. ID prefixes `DB`, `DS` are unused on
both branches (`measure-module` adds `RV`).
