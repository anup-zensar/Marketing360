# Research: Measure Foundation — the Reports Area and the Report Frame

Each entry: **Decision**, **Rationale**, **Alternatives considered**. Code references are as of `measure-module`
at `84e257ef`.

## R1 — One filter engine: All work's

**Decision**: Reports match records with All work's engine. A new exported function in a new file
`allwork/match.go` — `Match(nodes []Node, q Query, attributes []Attribute, c Choices) Matched` — runs All work's
own steps (`knownValues`, `resolveAttrs`, `buildTree`, `applyFilters`) and returns the IDs of the matching
plans, campaigns, initiatives and activities (shown and matching, as `matchesOf` counts them), the live line items
of the matching activities, and what was left out. The report filter's parameters keep All work's names (`year`,
`plan`, `planType`, `team`, `owner`, `activityType`, `activityStatus`, `attr=SA-…:value,…`) and are parsed by
`allwork.ParseQuery`; the report adds only `period`, `show`, `currency` and `table`.

**Rationale**: FR-007 and FR-012 ask that a report and All work with the same filters cover the same records;
the only way to guarantee it is one implementation. All work's engine already has every rule the spec names —
lower-level filters met by a record below (010 FR-005), % split share above 0, values hidden by a business-unit
condition count as Not filled in (016 FR-039), line-item attributes tested on activities (D-113), unknown values
left out (`knownValues`, FR-031 of 010). A new file in `allwork` follows 036's `planned_by_month.go` precedent.

**Alternatives considered**: a second engine in `measure` (two rule sets drift; SC-001 would fail on the first
divergence); calling `allwork.Service.Query` (builds grid rows, columns and sorting a report never shows — wasted
work at the 028 volume, and its summary has no plan or line-item counts).

## R2 — Which plans a report covers

**Decision**: the year's nodes come from All work's source (`Source.Year`, plans in the `plans.view` scope, read in
SQL), then are narrowed to the plans in the `reports.view` scope, read by the existing scoped store read
`plan.Store.ListInScope(ctx, authz.PermissionScopeOf(ctx, authz.ReportsView))`. Money figures cover the plans of
`Choices.FinancePlans` (`finance.view`, D-301) and the frame states "Covers N of M plans".

**Rationale**: FR-003 asks for the `reports.view` scope; it can differ from the `plans.view` scope when a person
holds two roles with different scopes. Both reads are already scoped in SQL and covered by
`plan`'s and `allwork`'s scope-isolation tests; intersecting removes plans, never adds one, so no out-of-scope
record leaves the database (constitution VII, D-356).

**Alternatives considered**: a new year read scoped by `reports.view` in SQL (a second copy of All work's year
query for an intersection that removes, at most, a few plan IDs); using the `plans.view` scope only (wrong when an
administrator narrows `reports.view`).

## R3 — The permission

**Decision**: `reports.view` ("See reports", refusal noun "Seeing reports", scoped) joins `authz`'s catalogue in
the Plans and planning group after `data.export`. Migration 00028 widens `roles_permissions_check` and appends
`reports.view` to every role that holds `plans.view`; the seed gives it to the built-in roles that hold
`plans.view` (D-367). Route rules: `need(authz.ReportsView)` for the catalogue, the frame and saved views;
`need(authz.DataExport)` for exports (as 036's dashboard exports, D-343); each report adds its own rule.

**Rationale**: least privilege without changing anyone's access (FR-002); 032's `purchaseRequests.list`
(migration 00020) and 038's `finance.link` change (00027) are the pattern.

**Alternatives considered**: reusing `plans.view` (no way to withhold reports from a role, Q-028 would need a
migration later); one permission per report (Q-028 default says no).

## R4 — Saved report views

**Decision**: a table `report_views` (ID prefix `RV`): person, report key, name, default flag, the **address
query** (the canonical query string of the report filter), created and updated times. Unique name per person and
report ignoring case; one default per person and report (partial unique index). The service validates the query
with the report filter's parser before saving and stores it canonicalised (parameters sorted, defaults dropped).
Lifecycle as 010's saved views (`allwork/views.go`): personal, list, create, update (name, query, default), delete;
the web app's **Undo** re-creates the deleted view with its name, query and star.

**Rationale**: D-368 makes a saved view "a named address"; storing the query string keeps one format for the
address, the view and the export's filter description, and a view saved before a filter is added still parses.
A table of its own keeps All work's `saved_views` and its `ViewState` untouched (Plan-module code, other lanes).

**Alternatives considered**: a `report` column on `saved_views` (changes All work's table, unique indexes and
`ViewState`); structured JSON state (a second schema to keep in step with the address).

## R5 — Saved views: versioned and logged, deleted outright

**Decision** (revised after `/speckit-analyze` finding C1): a report view embeds `record.Meta` (version, created
and updated by and at); an update carries the version it was read at and a stale one is refused (409
`version_conflict`, "This view was changed in another tab — reload it to see the latest"); every create,
update, star and delete writes one `history.Log` entry inside the same `Write`, under a registered
`history.Kind` "reportView" ("saved report view"). The activity log shows those entries to the view's owner (the
entry's actor) and to holders of `activityLog.viewAll` — one case added to `app/history_access.go`, which
otherwise hides an unknown kind from everyone but `activityLog.viewAll` (fail closed). A delete is immediate and
does not go to the recycle bin; **Undo** re-creates the view (logged as created).

**Rationale**: constitution VII writes **every change** to the activity log with who, what and when, and new code
follows the constitution in full (Governance); 010's unlogged saved views predate constitution 2.0 and are not a
precedent for new code. The recycle bin is for **business records** (VII), and a saved view is a personal
preference, so deleting outright stays within the rule. The version also answers the spec's two-tab edge case.

**Alternatives considered**: no version and no log, as 010 (breaks VII); `history.MoveToBin` (a recycle-bin entry
for a private bookmark that nobody else may restore, and a second restore path beside Undo).

## R6 — The registry: how reports plug in without editing the hub

**Decision**: `measure` defines `type Report interface { Key() string; Available() bool; Sheets(ctx, a, f Filter)
([]sheet.Sheet, error) }` and a `Registry` filled by `app/measure.go`, which calls one function per report —
`registerPlanBudget(w)`, `registerApprovals(w)`, … — each living in its own `app/measure_<report>.go`. 039 writes
each function as registering a stub (`Available() == false`, no sheets). A lane replaces only its stub's body.
The catalogue (FR-001) lists the registry in its fixed order; the export endpoint asks the registry for the
report's sheets and adds the **About this export** and **In view** sheets itself.

**Rationale**: D-369 asks for registration points created once, so four lanes never edit the same file; one
export endpoint serves every report's whole-report and per-widget export (FR-022).

**Alternatives considered**: each report registering its own export endpoint (eight copies of the About sheet and
file naming); lanes adding lines to one shared list (merge conflicts every day).

## R7 — Export

**Decision**: `GET /api/v1/reports/{report}/export?format=xlsx|csv[&widget=…]` with the report filter. Excel: one
sheet per widget, then **In view** and **About this export** (report, year, filters in words, period, currency,
exporter, date and time, rates used — `Source.RatesUsed`). CSV: one widget (required), the notes before the rows
(`sheet.WriteCSV(notes, rows)`). File names "SP&A One – <report> [– <widget>] – <year> – <date>"; written with
`platform/sheet`, never stored; `Content-Disposition` as All work's (`allwork.ContentDisposition`).

**Rationale**: FR-022; `platform/sheet` (010, 012) already writes both formats with money as numbers.

**Alternatives considered**: client-side export from the screen's data (the server is the source of truth,
constitution IV; files would differ by browser).

## R8 — Frame response and the address

**Decision**: `GET /api/v1/reports/frame?report=…&<filter>` returns: the filter as applied (with defaults), the
choices (years, plans, plan types, teams with Mine, owners, activity types, activity statuses, attribute groups per
level with values and Not filled in, active currencies), the **In view** counts, finance coverage, what was left
out (counts only, never names), the person's display currency, `generatedAt` and `financeAsOf`. The web app keeps
the filter in the address, parsed with Zod (`features/measure/shared/report-address.ts`); a malformed value is
dropped before the request, an unknown or out-of-scope one is reported by the frame as left out (FR-015, 038
FR-012). Period, Show, currency and the chart/table choices apply at once; the panel's filters on **Apply**
(039-C1).

**Rationale**: one request gives everything every report page shares; reports fetch only their widgets.

**Alternatives considered**: choices in a separate endpoint (two round trips per Apply); filters in the person's
preferences (D-368: nothing persists silently).

## R9 — Statements per request and the benchmark

**Decision**: the frame reads, in one `Read` snapshot: the year (All work's `Source.Year`, 028-checked), the plans
in the `reports.view` scope (one statement), Choices, Attributes and Currency (one statement each). No statement
grows with the records. `BenchmarkReportsFrame` in `internal/platform/apibench` at the demo data and the 028
volume asserts SC-003 (2 s).

**Rationale**: constitution B5 and feature 028; 036's `BenchmarkDashboards` is the model.

## R10 — Threat and risk model

| Threat | Control |
|---|---|
| Reading another market's figures through filter values or a saved view | Plans come only from the actor's scopes (R2); left-out values are counted, never named (FR-015); views are filtered by person in SQL and a view ID of someone else answers 404 |
| Guessing report keys or view IDs | Report keys are an allowlist; `RV-` IDs checked against the owner; unknown key → 404 |
| Oversized or malicious query strings | Parameters through `allwork.ParseQuery` allowlists; saved query at most 4,000 characters; names at most 60 characters, trimmed |
| Spreadsheet formula injection in exported names | `platform/sheet` already writes text cells as strings and prefixes a CSV text starting with `=`, `+`, `-`, `@`, a tab or a carriage return with `'` (`csv_write.go`, `safeText`) |
| Export of figures the person may not see | Exports reuse the screen's View; finance coverage identical; files never stored or logged |
