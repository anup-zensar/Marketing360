# Data Model: Dashboard Builder

Two stored records — **custom dashboard** (`DB-…`, package `customdashboard`) and **dataset** (`DS-…`, package
`dataset`) — and one computed answer, **widget data** (package `widgetdata`). The contract
(`backend/api/openapi.yaml`, schemas `CustomDashboard*`, `Widget*`, `Dataset*`, `DataSource*`) is the source of
truth for field names; this page gives the rules and the storage.

## 1. Custom dashboard

| Field | Type | Rule |
|---|---|---|
| `id` | text | `DB-` + number (`ids.PrefixDashboard = "DB"`), from the counters |
| `name` | text | required, trimmed, 1–80 characters, unique per owner ignoring case (`name_taken`) |
| `description` | text | 0–500 characters |
| `owner` | person | the creator (`actor`); never changes |
| `sharedWith` | teams | 0–20 teams; on change each must be a team the owner belongs to now, or one already shared |
| `filters` | DashboardFilters | the defaults viewers start from (§4) |
| `widgets` | Widget[] | 0–30, ids unique in the dashboard, no two overlapping (§2) |
| `version`, `createdAt/By`, `updatedAt/By` | `record.Meta` | version checked on every change (`version_conflict`) |

Answer-only fields: `access` (`edit`, `share`, `delete`, `copy` — `ActionAvailability`), `isOwner`,
`shareableTeams` (owner only: the teams they may tick), `widgetCount` in summaries.

**History kind** `customDashboard` (singular "dashboard", plural "dashboards", top-level). Snapshot data = the
dashboard row, its widgets and its shares. Log entries: created, changed (fields `name`, `description`, `filters`,
`widgets` — described as "4 widgets → 5 widgets; changed 'Planned cost by type'"), `sharing` ("Shared with
Lubricants Europe", "No longer shared with …"), copied ("Copied from 'X' (DB-1001)"), deleted, restored.
Restore of a name now taken renames to "<name> (restored)" (`PrepareRestore`).

## 2. Widget

| Field | Type | Rule |
|---|---|---|
| `id` | text | `w1`–`w999`, made by the web app, unique in the dashboard |
| `type` | WidgetType | `bar`, `line`, `area`, `combo`, `pie`, `donut`, `scatter`, `funnel`, `gauge`, `heatmap`, `treemap`, `radar`, `kpi`, `table`, `text` |
| `title` | text | 0–80 (an empty title shows the source and measure, e.g. "Sum of Planned cost by Activity type") |
| `subtitle` | text | 0–120 |
| `layout` | `{x, y, w, h}` | integers; `x` 0–11, `y` 0–199 (zero-based; the Layout tab shows `x+1`, `y+1`), `w` 2–12, `h` 2–12, `x+w ≤ 12`, `y+h ≤ 200`; rectangles of two widgets never share a cell |
| `query` | WidgetQuery \| null | null exactly when `type = text` |
| `display` | WidgetDisplay | options (§3); unknown options refused |
| `note` | text | 0–2,000 characters, plain text; used only by `text` (empty otherwise) |

Default sizes on add: charts 6×4, `kpi` 3×2, `table` 6×5, `text` 4×2, `gauge` 4×4.

### 2.1 WidgetQuery

| Field | Type | Rule |
|---|---|---|
| `source` | text | a built-in key (§5) or a dataset ID `DS-…` |
| `dimension` | `{field, dateGroup?}` \| null | the group (categories, x axis); `dateGroup` (`day`, `week`, `month`, `quarter`, `year`) required for a date field, refused otherwise; the field must be groupable |
| `series` | `{field, dateGroup?}` \| null | the breakdown; same rules; not the same field and bucket as `dimension` |
| `measures` | Measure[] 1–8 | `{aggregation, field?, label?}`; `count` without a field counts records; `countDistinct` any field; `sum`, `avg` number or money; `min`, `max` number, money or date; `label` 0–40 (default "Sum of Planned cost") |
| `filters` | Filter[] 0–10 | `{field, operator, values}` (§2.2) |
| `sort` | `{by, measureIndex?, direction}` | `by`: `dimension` or `measure` (`measureIndex` 0–7, below the number of measures); `direction` `asc` / `desc`; default §2.3 |
| `limit` | integer 1–500 | default 50; groups of `dimension` (with a breakdown, series beyond the 20 largest join "Other") |
| `groupRest` | boolean | default false: the groups after `limit` become one "Other" group |

### 2.2 Filter operators

| Operator | On screen | Values | Field kinds |
|---|---|---|---|
| `eq`, `ne` | is, is not | 1 | all |
| `in` | is one of | 1–50 | all |
| `contains` | contains | 1 (1–100 characters, ignoring case) | text |
| `gt`, `gte`, `lt`, `lte` | greater than, at least, less than, at most | 1 | number, money, date |
| `between` | between | 2 (from ≤ to, both included) | number, money, date |
| `isEmpty`, `notEmpty` | is empty, is not empty | 0 | all |

Values are strings typed by the field: numbers `^-?[0-9]{1,15}(\.[0-9]{1,4})?$`, money `^-?[0-9]{1,15}(\.[0-9]{1,2})?$`
in USD, dates `YYYY-MM-DD`, text 1–200 characters. Every problem is answered at once as `validation_failed` with
fields such as `widgets[2].query.filters[0].values`.

### 2.3 Widget rules (checked by the web app before Save and by the API)

| Type | `dimension` | `series` | Measures | Notes |
|---|---|---|---|---|
| `bar`, `line`, `area` | required | optional | 1–4; exactly 1 with a series | |
| `combo` | required | none | 2–4 | `display.measureStyles` one per measure |
| `pie`, `donut`, `funnel` | required | none | 1 | |
| `scatter` | required (each group is a point) | optional (colour) | 2–3 (x, y, size) | number or money measures |
| `gauge`, `kpi` | none | none | 1 | |
| `heatmap` | required | required | 1 | |
| `treemap` | required | optional (second level) | 1 | sum or count |
| `radar` | required (the spokes, at most 12 groups) | optional (one shape each, at most 8) | 1 | |
| `table` | optional | optional | 1–8 | |
| `text` | — | — | — | `query` null, `note` used |

Default sort: by the dimension ascending when it is a date or the type is `line`, `area`, `combo` or `heatmap`;
otherwise by measure 0 descending; `funnel` always measure 0 descending.

## 3. WidgetDisplay (all optional; defaults in brackets)

| Option | Values | Types |
|---|---|---|
| `palette` | `categorical`, `planning`, `levels`, `sequential` [`categorical`; `sequential` for `heatmap`] | all charts |
| `legend` | `{show: boolean [true], position: top \| bottom \| left \| right [bottom]}` | charts with series or slices |
| `dataLabels` | boolean [false; true for `pie`, `donut`, `funnel`] | charts |
| `axisTitles` | `{x, y, y2}` each 0–40 characters [""] | `bar`, `line`, `area`, `combo`, `scatter`, `heatmap` |
| `gridlines` | boolean [true] | axis charts |
| `smooth` | boolean [false] | `line`, `area`, `combo` |
| `stack` | `none`, `stacked`, `percent` [`none`; `stacked` for `area`] | `bar`, `line`, `area` with a series |
| `orientation` | `vertical`, `horizontal` [`vertical`] | `bar`, `funnel` |
| `numberFormat` | `number`, `money`, `percent`, `compact` [`money` for money measures, else `number`] | all with values |
| `decimals` | 0–4 [0 for counts, 2 for money] | all with values |
| `zoom` | boolean [false] | `bar`, `line`, `area`, `combo`, `scatter` |
| `referenceLine` | `{value: decimal, label: 0–40}` \| null [null] | `bar`, `line`, `area`, `combo` |
| `measureStyles` | `[{as: bar \| line, axis: left \| right}]`, one per measure [first bar left, others line right] | `combo` |
| `gauge` | `{min: decimal [0], max: decimal [auto: the value rounded up]}` with min < max | `gauge` |
| `kpi` | `{comparison: none \| target \| previousPeriod [none], target: decimal \| null, higherIsBetter: boolean [true]}`; `target` required with `target` | `kpi` |
| `table` | `{columns: string[] [all], totals: boolean [false]}` — column keys `dimension`, `series`, `m0`…`m7`, each once | `table` |

`percent` format multiplies nothing: it writes the value with "%" (for shares and percentages the API answers).
`compact` writes 1.2K, 3.4M, 5.6B with the chosen decimals.

## 4. Dashboard filters (`CustomDashboardFilters`)

`{dateFrom: date | null, dateTo: date | null, planIds: string[] 0–20, teamIds: string[] 0–20}`; `dateFrom ≤ dateTo`.
Stored as the dashboard's defaults; in the viewer, the address (`from`, `to`, `plan`, `team`) overrides them. Plans
and teams the viewer cannot see are ignored (never an error, never named). Applied by the built-in source's own
reads: the date range to its date field (the record's date falls in the range, both ends included), plans and
teams to its plan and team (a record whose plan has any of the teams). Never applied to datasets (Q-028).

## 5. Built-in sources (data catalogue)

Every source reads only the records of plans in the viewer's `plans.view` scope, as All work does (010, 014);
fields marked **F** are empty outside the plans where the viewer holds `finance.view`. Amounts are USD at the
current rate (D-258). "Multi" fields hold several values (a plan's teams). Date / plan / team say which dashboard
filters apply.

| Key | Label | Needs | Date filter on | Plan / team | Fields (key: kind) |
|---|---|---|---|---|---|
| `plans` | Plans | `plans.view` | — | yes / yes | `plan` text, `year` number, `status` text, `planType` text, `owner` text, `team` text multi, `currency` text, `budget` money, `plannedCost` money, `leftToPlan` money, `committed` money F, `actual` money F |
| `campaigns` | Campaigns | `plans.view` | `startDate` | yes / yes | `campaign`, `plan`, `year` number, `team` multi, `owner`, `status`, `timing`, `startDate` date, `endDate` date, `budget`, `plannedCost`, `leftToPlan`, `committed` F, `actual` F |
| `initiatives` | Initiatives | `plans.view` | `startDate` | yes / yes | `initiative`, `campaign`, `plan`, `year`, `team` multi, `owner`, `status`, `startDate`, `endDate`, `budget`, `plannedCost`, `leftToPlan` |
| `activities` | Activities | `plans.view` | `startDate` | yes / yes | `activity`, `initiative`, `campaign`, `plan`, `year`, `team` multi, `owner`, `activityType`, `status`, `complete` (Yes/No), `startDate`, `endDate`, `plannedCost`, `committed` F, `actual` F; P3: `attribute.<SA-id>` text for each activity-level Shell attribute |
| `lineItems` | Line items | `plans.view` | — | yes / yes | `lineItem`, `activity`, `campaign`, `plan`, `year`, `team` multi, `currency`, `financeBudgetItem`, `swbsStatus`, `complete`, `plannedCost`, `committed` F, `actual` F |
| `plannedByMonth` | Planned cost by month | `plans.view` | `month` | yes / yes | `month` date (first of month; empty = not yet spread), `plan`, `campaign`, `initiative`, `activity`, `activityType`, `team` multi, `plannedCost` money — All work's Budget & cost spread (010 FR-023) |
| `financialTransactions` | Financial transactions | `finance.view` | `postingDate` | yes / yes | `type` (Purchase order line, Invoice, Payment, Accrual), `postingDate` date, `amount` money, `vendor`, `documentNumber`, `lineItem`, `activity`, `campaign`, `plan`, `team` multi — only transactions linked to line items of plans where the viewer sees financial data |
| `approvals` | Approvals | `plans.view` | `submittedOn` | yes / yes | `record`, `recordType` (Activity, Purchase request, Modified purchase request), `status` (Waiting, Approved, Declined, Ended), `automatic` (Yes/No), `workflow`, `submittedOn` date, `decidedOn` date, `daysToDecide` number, `plan`, `team` multi |
| `purchaseRequests` | Purchase requests | `plans.view` | `createdOn` | yes / yes | `purchaseRequest`, `status`, `owner`, `vendor`, `amount` money, `createdOn` date, `sentToGsap` (Yes/No), `activity`, `campaign`, `plan`, `team` multi — the purchase requests the viewer may see (032) |
| `globalBudgets` | Global budgets | `globalBudgets.view` | — | no / no | `globalBudget`, `year` number, `owner`, `amount`, `allocated`, `unallocated`, `plannedCost` — as the Global budgets page (014) |

Every text field is groupable and can be counted; numbers and money can be summed, averaged and taken minimum and
maximum; dates groupable by bucket and minimum / maximum. A dataset's fields are its columns (`c1`…`c50`, labels
and kinds as saved; money columns converted to USD).

## 6. Widget data (answer of `POST /data-sources/query`)

- Request (`WidgetDataRequest`): `{type, query, filters: CustomDashboardFilters | null, comparison: {kind, target} | null}`;
  the type lets the API apply the widget rules of §2.3 before it reads anything.
- `columns[]`: `{key: dimension | series | m0…m7, label, kind: text | number | money | date, dateGroup?}`.
- `rows[]`: `{dimension: Key | null, series: Key | null, values: Value[]}` in the query's sort order, at most `limit`
  (+ "Other"); `Key = {value: string | null, label: string}` — `value` is the text, the decimal, or the bucket's
  first day (YYYY-MM-DD); `label` is "Not set" for empty, "Other" for the rest group, otherwise the value (the web
  app writes dates in the person's format); `Value = {number: decimal | null, money: Money | null, date: date |
  null, share: decimal | null}` — exactly one of number, money, date for the measure's kind; `share` in percent
  with one decimal (FR-021).
- `totals[]`: one Value per measure over every group (also those beyond the limit).
- `groupCount`: groups before the limit. `notes[]`: plain sentences ("Not filtered by date", "Financial figures
  cover 1 of 2 plans", "A plan with several teams counts in each", "Showing 10 of 37").
- `comparison` (kpi): `{kind: target | previousPeriod, value: Value, change: decimal | null, period?: {from, to}}`.
- Engine order: source rows (scope and dashboard filters applied by the reader) → widget filters → buckets → groups
  → aggregates → shares and totals → sort → limit / Other → comparison. Numbers are exact decimals with 4 places
  (int64 ten-thousandths), money integer cents; an overflow answers 400 `validation_failed` "The values are too
  large to add up". Ties sort by the group's label.

## 7. Dataset

| Field | Type | Rule |
|---|---|---|
| `id` | text | `DS-` + number (`ids.PrefixDataset = "DS"`) |
| `name` | text | required, 1–80, unique per owner ignoring case |
| `description` | text | 0–500 |
| `owner`, `sharedWith` | person, teams | as dashboards (§1) |
| `fileName`, `sheetName` | text | as uploaded (≤ 255; sheet "" for CSV) |
| `rowCount` | integer | 0–10,000 |
| `columns` | DatasetColumn[] 1–50 | `{key: c1…c50, position, label 1–60 unique ignoring case, kind: text \| number \| date \| money, currency: ISO code for money (an active currency), else null}` |
| `usedBy` | refs | dashboards the viewer may see whose widgets read it |

Cells, by kind: text ≤ 500 characters; number exact decimal with up to 4 places (thousands separators and a
decimal comma accepted on upload when unambiguous, "1 234,5"); date `YYYY-MM-DD`, `DD/MM/YYYY`, `D MMM YYYY` or an
Excel date cell; money as number with up to 2 places. Inference proposes Date when every non-empty cell of the first
200 is a date, Number when every one is a number, Money when a number column's name holds "cost", "amount",
"spend", "budget", "price" or a currency code, else Text. Empty cells are empty values in every kind.

**History kind** `dataset` (singular "dataset", plural "datasets", top-level). Snapshot data = the dataset row,
columns, shares and rows (bounded by the 5 MB file limit). Log: created ("Uploaded from 'leads.xlsx', 240 rows"),
changed (`name`, `description`, `columns` labels), `sharing`, deleted, restored.

## 8. Storage (PostgreSQL, goose, sqlc)

### Migration `00028_custom_dashboards.sql` (B1)

```sql
-- +goose Up
CREATE TABLE custom_dashboards (
    id              text        PRIMARY KEY CHECK (id ~ '^DB-[0-9]+$'),
    name            text        NOT NULL CHECK (char_length(name) BETWEEN 1 AND 80 AND name = btrim(name)),
    description     text        NOT NULL CHECK (char_length(description) <= 500),
    owner_id        text        NOT NULL REFERENCES people (id) DEFERRABLE INITIALLY DEFERRED,
    filter_date_from date,
    filter_date_to   date,
    filter_plan_ids text[]      NOT NULL CHECK (cardinality(filter_plan_ids) <= 20),
    filter_team_ids text[]      NOT NULL CHECK (cardinality(filter_team_ids) <= 20),
    version         integer     NOT NULL CHECK (version >= 1),
    created_at      timestamptz NOT NULL, created_by_id text NOT NULL, created_by_name text NOT NULL,
    updated_at      timestamptz NOT NULL, updated_by_id text NOT NULL, updated_by_name text NOT NULL,
    CHECK (filter_date_from IS NULL OR filter_date_to IS NULL OR filter_date_from <= filter_date_to)
);
CREATE UNIQUE INDEX custom_dashboards_owner_name_key ON custom_dashboards (owner_id, lower(name));

CREATE TABLE custom_dashboard_widgets (
    dashboard_id text    NOT NULL REFERENCES custom_dashboards (id) ON DELETE CASCADE,
    id           text    NOT NULL CHECK (id ~ '^w[1-9][0-9]{0,2}$'),
    position     integer NOT NULL CHECK (position BETWEEN 0 AND 29),
    type         text    NOT NULL CHECK (type IN ('bar','line','area','combo','pie','donut','scatter','funnel',
                                                  'gauge','heatmap','treemap','radar','kpi','table','text')),
    title        text    NOT NULL CHECK (char_length(title) <= 80),
    subtitle     text    NOT NULL CHECK (char_length(subtitle) <= 120),
    grid_x       integer NOT NULL CHECK (grid_x BETWEEN 0 AND 11),
    grid_y       integer NOT NULL CHECK (grid_y BETWEEN 0 AND 199),
    grid_w       integer NOT NULL CHECK (grid_w BETWEEN 2 AND 12),
    grid_h       integer NOT NULL CHECK (grid_h BETWEEN 2 AND 12),
    source       text    CHECK (source IS NULL OR source <> ''),
    query        jsonb   CHECK (query IS NULL OR jsonb_typeof(query) = 'object'),
    display      jsonb   NOT NULL CHECK (jsonb_typeof(display) = 'object'),
    note         text    NOT NULL CHECK (char_length(note) <= 2000),
    PRIMARY KEY (dashboard_id, id),
    UNIQUE (dashboard_id, position),
    CHECK (grid_x + grid_w <= 12 AND grid_y + grid_h <= 200),
    CHECK ((type = 'text') = (query IS NULL) AND (query IS NULL) = (source IS NULL))
);
CREATE INDEX custom_dashboard_widgets_source_idx ON custom_dashboard_widgets (source) WHERE source LIKE 'DS-%';

CREATE TABLE custom_dashboard_shares (
    dashboard_id text NOT NULL REFERENCES custom_dashboards (id) ON DELETE CASCADE,
    team_id      text NOT NULL REFERENCES teams (id) ON DELETE CASCADE DEFERRABLE INITIALLY DEFERRED,
    PRIMARY KEY (dashboard_id, team_id)
);
CREATE INDEX custom_dashboard_shares_team_idx ON custom_dashboard_shares (team_id);

-- Feature 051: the built-in roles see and build dashboards (D-374).
UPDATE roles SET permissions = array_append(permissions, 'dashboards.view'), version = version + 1,
    updated_at = now(), updated_by_id = 'system', updated_by_name = 'SP&A One'
WHERE built_in <> '' AND NOT 'dashboards.view' = ANY (permissions);
UPDATE roles SET permissions = array_append(permissions, 'dashboards.build'), version = version + 1,
    updated_at = now(), updated_by_id = 'system', updated_by_name = 'SP&A One'
WHERE built_in IN ('administrator', 'globalPlanner', 'planner') AND NOT 'dashboards.build' = ANY (permissions);
```

The overlap of widgets is checked by the service (an exclusion constraint would need the `btree_gist` extension
for one rule that the service already checks on every save). Query and display are JSON because their shape is the
contract's `WidgetQuery` and `WidgetDisplay`, validated by the service on every write (research R5).

### Migration `00029_datasets.sql` (B2)

```sql
-- +goose Up
CREATE TABLE datasets (
    id           text    PRIMARY KEY CHECK (id ~ '^DS-[0-9]+$'),
    name         text    NOT NULL CHECK (char_length(name) BETWEEN 1 AND 80 AND name = btrim(name)),
    description  text    NOT NULL CHECK (char_length(description) <= 500),
    owner_id     text    NOT NULL REFERENCES people (id) DEFERRABLE INITIALLY DEFERRED,
    file_name    text    NOT NULL CHECK (char_length(file_name) BETWEEN 1 AND 255),
    sheet_name   text    NOT NULL CHECK (char_length(sheet_name) <= 255),
    row_count    integer NOT NULL CHECK (row_count BETWEEN 0 AND 10000),
    version      integer NOT NULL CHECK (version >= 1),
    created_at   timestamptz NOT NULL, created_by_id text NOT NULL, created_by_name text NOT NULL,
    updated_at   timestamptz NOT NULL, updated_by_id text NOT NULL, updated_by_name text NOT NULL
);
CREATE UNIQUE INDEX datasets_owner_name_key ON datasets (owner_id, lower(name));

CREATE TABLE dataset_columns (
    dataset_id text    NOT NULL REFERENCES datasets (id) ON DELETE CASCADE,
    position   integer NOT NULL CHECK (position BETWEEN 0 AND 49),
    key        text    NOT NULL CHECK (key = 'c' || (position + 1)),
    label      text    NOT NULL CHECK (char_length(label) BETWEEN 1 AND 60),
    kind       text    NOT NULL CHECK (kind IN ('text', 'number', 'date', 'money')),
    currency   text    CHECK (currency ~ '^[A-Z]{3}$'),
    PRIMARY KEY (dataset_id, position),
    CHECK ((kind = 'money') = (currency IS NOT NULL))
);
CREATE UNIQUE INDEX dataset_columns_label_key ON dataset_columns (dataset_id, lower(label));

CREATE TABLE dataset_rows (
    dataset_id text    NOT NULL REFERENCES datasets (id) ON DELETE CASCADE,
    row_number integer NOT NULL CHECK (row_number BETWEEN 1 AND 10000),
    cells      jsonb   NOT NULL CHECK (jsonb_typeof(cells) = 'array' AND jsonb_array_length(cells) <= 50),
    PRIMARY KEY (dataset_id, row_number)
);

CREATE TABLE dataset_shares (
    dataset_id text NOT NULL REFERENCES datasets (id) ON DELETE CASCADE,
    team_id    text NOT NULL REFERENCES teams (id) ON DELETE CASCADE DEFERRABLE INITIALLY DEFERRED,
    PRIMARY KEY (dataset_id, team_id)
);
CREATE INDEX dataset_shares_team_idx ON dataset_shares (team_id);
```

A row's `cells` holds one JSON value per column in position order: `null` for empty; a string for text, number
(decimal text, exact) and date (YYYY-MM-DD); an integer of minor units for money. Rows are written with
`COPY`-style batched inserts (`pgx.CopyFrom`, one statement for the whole file) and read in one statement.

### sqlc queries

`queries/custom_dashboards.sql` (B1): `InsertCustomDashboard`, `UpdateCustomDashboard` (version-checked),
`DeleteCustomDashboard`, `GetCustomDashboardVisible` (owner or shared with a team the person is a member of; one
statement), `ListCustomDashboardsVisible`, `GetCustomDashboardForOwner`, `CustomDashboardWidgets` (`= ANY` of
dashboard ids), `InsertCustomDashboardWidgets` (copy), `DeleteCustomDashboardWidgets`, `CustomDashboardShares`,
`ReplaceCustomDashboardShares`, `DashboardsUsingDataset` (visible, by `source`), `CustomDashboardOwner` (history
access).

`queries/datasets.sql` (B2): `InsertDataset`, `UpdateDataset`, `DeleteDataset`, `GetDatasetVisible`,
`ListDatasetsVisible`, `CountDatasetsOfOwner`, `DatasetColumns`, `InsertDatasetColumns`, `UpdateDatasetColumnLabel`,
`DatasetRows` (`row_number` range for pages; all for queries), `InsertDatasetRows` (copy), `DatasetShares`,
`ReplaceDatasetShares`, `DatasetOwner`.

"Visible" in both is `owner_id = $person OR EXISTS (SELECT 1 FROM <x>_shares s JOIN team_memberships m ON
m.team_id = s.team_id WHERE s.<x>_id = id AND m.person_id = $person)`, with the person taken only from the actor;
each visible read has a scope-isolation test on PostgreSQL.
