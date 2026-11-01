# UI Contract: Dashboard Builder

## Routes (`frontend/src/app/dashboard-routes.tsx`, all lazy with `features/custom-dashboards`)

| Path | Page | Title (page header and tab) | Gate |
|---|---|---|---|
| `/dashboards` | `DashboardListPage` (replaces the redirect) | Dashboards | `plans.view` or `dashboards.view` |
| `/dashboards/executive`, `/dashboards/regional` | 036's `DashboardsPage` (unchanged) | Executive dashboard, Regional dashboard | `plans.view` |
| `/dashboards/:dashboardId` (`DB-…`) | `DashboardViewerPage` | the dashboard's name | `dashboards.view` |
| `/builder` | `DashboardBuilderPage` (new) | New dashboard | `dashboards.build` |
| `/builder/:dashboardId` | `DashboardBuilderPage` (edit) | Edit '<name>' | `dashboards.build` |
| `/datasets` | `DatasetsPage` | Datasets | `dashboards.build` |
| `/datasets/:datasetId` | `DatasetPage` | the dataset's name | `dashboards.build` |

`/dashboards/executive` and `/dashboards/regional` are declared before `/dashboards/:dashboardId`. Viewer filters in
the address: `from`, `to` (YYYY-MM-DD), `plan` and `team` (repeatable IDs), parsed with Zod; invalid values fall back
to the dashboard's saved filters.

## Sidebar (`frontend/src/app/navigation.ts`)

- **Dashboards** (`LayoutDashboard`): shown with `plans.view` **or** `dashboards.view` (`NavItem.permission` widened
  to `permissions?: Permission[]`, any of).
- **Dashboard builder** (`ChartNoAxesCombined` from lucide-react), `/builder`, `dashboards.build`, right after
  Dashboards; current also on `/datasets` (`alsoCurrentOn`).

## Feature folder `frontend/src/features/custom-dashboards/`

| Path | Lane | Contents |
|---|---|---|
| `index.ts` | F3 | exports `DashboardListPage`, `DashboardViewerPage`, `DashboardBuilderPage`, `DatasetsPage`, `DatasetPage` |
| `model.ts`, `keys.ts`, `widget-types.ts` | F0 | type aliases of the generated contract, query keys, the widget-type table (label, icon, group, needs, default size) |
| `charts/` | F1 | `EChart`, `buildOption` (one module per type family), `theme.ts`, `WidgetView`, `WidgetTable`, `WidgetStates`, `describe.ts`, `png.ts`, `format.ts`, `api.ts` (`useDataSources`, `useWidgetData`) |
| `builder/` | F2 | `DashboardBuilderPage`, `BuilderHeader`, `WidgetPalette`, `Canvas`, `CanvasWidget`, `grid.ts` (placement, collision, moves), `builder-reducer.ts` (state, undo/redo), `Inspector`, `DataTab`, `StyleTab`, `LayoutTab`, `DashboardDetails`, `validate.ts`, `api.ts` (create, update, copy) |
| `sharing/` | F3 | `SharingDialog` (dashboards and datasets) |
| `viewer/` | F3 | `DashboardListPage`, `DashboardCard`, `DashboardViewerPage`, `FiltersBar`, `ViewerGrid`, `filters-address.ts`, `api.ts` (list, get, delete, copy, share) |
| `datasets/` | F3 | `DatasetsPage`, `DatasetPage`, `UploadDialog`, `ColumnTypesForm`, `RowsTable`, `api.ts` |

Other features import only through `index.ts`; the sub-folders import each other's public files directly (charts'
`index.ts`, `sharing/SharingDialog`).

## Words on screen

| Place | Words |
|---|---|
| List sections | Built-in · My dashboards · Shared with me |
| Built-in cards | **Executive dashboard** — "Budget, planned cost and spend of every market you may see" · **Regional dashboard** — "The same for one market, with its plans and campaigns" |
| List actions | **New dashboard** (primary) · Datasets · card menu: Open, Edit, Make a copy, Delete |
| Empty list | "No dashboards yet" — "Build one from your plans or your own data." with **New dashboard** (builders) · "Nothing shared with you yet." |
| Palette groups | Charts: Bar, Line, Area, Bar and line, Pie, Donut, Scatter, Funnel, Gauge, Heatmap, Treemap, Radar · Figures and tables: Key figure, Table · Text: Text note |
| Inspector tabs | Data · Style · Layout; with nothing selected: Dashboard (Name, Description, Default filters) |
| Data tab | Source (groups "SP&A One data", "Uploaded data", link **Upload data**) · Group by · Group dates by (Day, Week, Month, Quarter, Year) · Break down by · Measures (**Add measure**; "Count of records", or field + Count, Count distinct, Sum, Average, Minimum, Maximum; Label) · Filters (**Add filter**; is, is not, is one of, contains, greater than, at least, less than, at most, between, is empty, is not empty) · Sort by (the group / a measure; Ascending, Descending) · Show top · Group the rest as Other |
| Style tab | Title · Subtitle · Colours (Categorical, Planning, Levels, Sequential) · Legend (Show, Position) · Data labels · Axis titles (Horizontal, Vertical, Right) · Gridlines · Smooth lines · Stacking (None, Stacked, 100%) · Orientation (Vertical, Horizontal) · Number format (Number, Money, Percent, Compact) · Decimals · Zoom slider · Reference line (Value, Label) · Bars or line / Left or right axis (per measure) · Minimum, Maximum (gauge) · Compare with (Nothing, Target, Previous period), Target, Higher is better · Columns, Totals row (table) · Note (text) |
| Layout tab | Column (1–12) · Row · Width (columns) · Height (rows) |
| Widget menu | Duplicate · Delete (builder) · Show as table / Show as chart · Download as PNG (viewer) |
| Builder header | name · Undo · Redo · Sharing · View · **Save** |
| Announcements | "Bar chart added." · "Moved to column 3, row 2." · "Resized to 6 columns by 4 rows." · "Can't move there: another widget is in the way." · "Undone: moved 'Planned cost by type'." |
| Toasts | "Dashboard saved" · "Dashboard 'X' deleted" + **Undo** · "Copy made" · "Sharing saved" · "Dataset 'X' saved" |
| Delete dashboard | "Delete dashboard 'X'?" — "It's shared with Lubricants Europe; they will no longer see it. It moves to the recycle bin; you can restore it." **Cancel** · **Delete dashboard** |
| Delete dataset | "Delete dataset 'X'?" — "2 of your dashboards use it; their widgets will show 'This data is no longer available'. …" **Delete dataset** |
| Version conflict | "This dashboard was changed elsewhere since you opened it. Your changes are still here." **Reload** · **Keep editing** |
| Sharing dialog | "Share 'X'" — team checkboxes — "People in these teams can view this dashboard. Each of them sees only the data they may see." **Save** · **Cancel** |
| Upload dialog | step 1 "Choose a file" (CSV or Excel .xlsx, up to 5 MB, 10,000 rows, 50 columns) · step 2 "Check the columns" (Name, Type, Currency, samples; first 20 rows) · **Save dataset** |

## States (each rendered and tested)

| View | States |
|---|---|
| Dashboard list | loading (skeleton cards), error + Retry, empty per section, data |
| Viewer page | loading, error + Retry, not found ("Dashboard not found" — also for private), no access (`NoAccess`), data |
| Each widget (`WidgetView`) | loading (reserved space, `aria-busy`), error + Retry, no data ("No data to show" + hint), no access ("You don't have access to this data"), gone ("This data is no longer available", owner), incomplete (builder: "Choose what to measure"), chart, table view, key figure, text note |
| Builder | loading, not found, view-only (not owner: "Only Priya Raman can edit this dashboard." + **Make a copy**), narrow window ("The builder needs a wider window"), editing, saving, version conflict, validation summary |
| Datasets | list loading / error / empty / data; upload: choosing, checking, problems, columns form, saving; dataset page loading / not found / data |

## Design tokens (design system §4.7, `frontend/src/ui/tokens.css`)

New: `--chart-cat-1` … `--chart-cat-8` and `--chart-seq-1` … `--chart-seq-5`, light and dark (research R3);
existing: `--chart-budget`, `--chart-planned`, `--chart-committed`, `--chart-actual`, `--level-*`, `--ink`,
`--ink-2`, `--hairline`, `--surface`. Grid: 12 columns, row height 80 px, gap `--space-4` (16 px); the selected
widget has a 2 px `--focus-ring` outline and the 4 px yellow "you are here" bar on its top edge; drop targets show a
dashed `--control-border` rectangle. The canvas background is `--page-bg`, widgets are `Card`s on `--surface`.

## Query keys (`keys.ts`)

`['custom-dashboards']`, `['custom-dashboards', id]`, `['datasets']`, `['datasets', id]`,
`['datasets', id, 'rows', offset]`, `['data-sources']`, `['widget-data', source, hashOf(request)]`. Saving or
deleting a dashboard invalidates `['custom-dashboards']`; a dataset change invalidates `['datasets']` and
`['data-sources']`; **Refresh** invalidates `['widget-data']`.
