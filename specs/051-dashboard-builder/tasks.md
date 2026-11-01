# Tasks: Dashboard Builder

**Input**: Design documents from `specs/051-dashboard-builder/` — [plan.md](plan.md), [spec.md](spec.md),
[research.md](research.md), [data-model.md](data-model.md), [contracts/](contracts/), [quickstart.md](quickstart.md)

**Status (2026-11-01)**: governance (constitution 4.9.0, D-371–D-378, ADR-0020), spec, plan and the API contract
are committed on `feature/051-dashboard-builder`; the 17 operations are in `notYetServed`. F0 (T003–T012) is done:
every operation has its route rule, and the lanes fork from the F0 commit. Until B1's migration 00028 widens
`roles_permissions_check`, `go run ./cmd/seed` on PostgreSQL refuses the new permissions in the seeded roles.

**Tests**: required by constitution VI — a unit test per rule, a test per acceptance scenario, a contract test per
operation, a scope-isolation test per scoped store read (real PostgreSQL, `TEST_DATABASE_URL`), a test per view
state, an end-to-end test per P1 story (US1, US2). Write each test first and see it fail. Test names carry the
requirement ID (`…_FR015`); source comments never do.

## How the work is split

**Format**: `- [ ] T### [P] [Lane] [Story] Description — path`. `[P]` = parallel within the lane.

1. **F0 — Foundation** runs first, on `feature/051-dashboard-builder` itself, in one commit: it adds everything two
   lanes would otherwise both edit (permissions, ID prefixes, route rules, wiring call sites, store-set fields,
   seed steps, the shared Go types and ports, the `echarts` dependency, the shared web-app types and stubs of every
   cross-lane export). After it, `go build ./...`, `go test -short ./...` and `pnpm run check` pass.
2. **Six lanes** then fork from that commit, each in its own worktree and branch
   (`feature/051-b1-custom-dashboards`, `-b2-datasets`, `-b3-widget-data`, `-f1-charts`, `-f2-builder`,
   `-f3-viewer-datasets`), touching only the files they own plus the additive edits listed below. Each lane keeps
   its package green on its own: backend lanes test against fakes and their own PostgreSQL store; frontend lanes
   against MSW answers typed from the contract and the F0 stubs.
3. **INT** merges the lanes in the order B2 → B1 → B3 → F1 → F3 → F2, wires what needs two lanes, removes the
   operations from `notYetServed`, writes the end-to-end tests and runs every check.

## Lane ownership

| Lane | Owns (creates and changes freely) | Additive edits allowed (append only; never reorder or rewrite others' lines) |
|---|---|---|
| **F0** | `backend/internal/widgetdata/{ports.go,table.go,number.go,number_test.go}`, `backend/internal/access/routes_custom_dashboards.go`, `routes_datasets.go`, `routes_data_sources.go`, `backend/internal/app/{custom_dashboards.go,datasets.go,widget_data.go}` (stubs), `backend/internal/stores/{custom_dashboards.go,datasets.go}` (empty interfaces), `backend/seed/{dashboards.go,datasets.go}` (stubs), `frontend/src/features/custom-dashboards/{model.ts,keys.ts,widget-types.ts,index.ts}`, stubs `charts/index.ts`, `sharing/SharingDialog.tsx`, `builder/DashboardBuilderPage.tsx`, `viewer/DashboardListPage.tsx`, `viewer/DashboardViewerPage.tsx`, `datasets/DatasetsPage.tsx`, `datasets/DatasetPage.tsx` | `authz/catalogue.go` + `catalogue_test.go`; `access/routes.go` (join list); `access/roles_test.go`, `access/service_test.go`, `access/handler_test.go`, `seed/access.go`, `seed/access_test.go` (permission counts); `ids/ids.go`; `app/areas.go` (three calls); `stores/stores.go` (Set fields); `seed/seed.go` (two steps); `internal/auth/handler_test.go`, `internal/people/handler_test.go` (the `/me` permissions); `frontend/package.json`, `pnpm-lock.yaml`; `frontend/src/features/people/access/access-words.ts` (+ test) and `src/testing/fixtures/access.ts` (Table A in the web app) |
| **B1** | `backend/internal/customdashboard/**`, `backend/internal/platform/database/migrations/00028_custom_dashboards.sql`, `queries/custom_dashboards.sql`, the sqlc output for it, `backend/internal/stores/custom_dashboards.go`, `backend/internal/app/custom_dashboards.go`, `backend/seed/dashboards.go` (+ test) | `stores/postgres.go`, `stores/storestest/*.go` (one field each); `sqlcdb` regenerated (`models.go`, `querier.go` — regenerate, never hand-edit); `docs/product/demo-users-and-data.md` (one paragraph) |
| **B2** | `backend/internal/dataset/**`, `migrations/00029_datasets.sql`, `queries/datasets.sql` and its sqlc output, `backend/internal/stores/datasets.go`, `backend/internal/app/datasets.go`, `backend/seed/datasets.go` (+ test), `frontend/e2e/fixtures/event-leads.xlsx`, `event-leads.csv` | as B1; the body limit of the two upload operations where 038 configures limits (one entry each) |
| **B3** | `backend/internal/widgetdata/**` except F0's three files (B3 may add methods to `number.go`), `backend/internal/app/widget_data.go`, `backend/internal/app/widgetdata_*.go`, `backend/internal/platform/apibench/widget_data_test.go` | `internal/financials/` (new file `transactions_on_line_items.go`, one `Store` method in `store.go`, `store_postgres.go`, the fake, one query in `queries/financials.sql`) if the transactions source needs it; `internal/allwork` exported helpers only if missing |
| **F1** | `frontend/src/features/custom-dashboards/charts/**` (replaces F0's stub `charts/index.ts`) | — |
| **F2** | `frontend/src/features/custom-dashboards/builder/**` (replaces F0's stub page) | — |
| **F3** | `frontend/src/features/custom-dashboards/{viewer,datasets,sharing}/**`, `index.ts`, `frontend/src/app/dashboard-routes.tsx`, `frontend/src/app/navigation.ts`, `frontend/src/app/Sidebar.tsx` (+ tests), `frontend/src/app/pages.tsx` | `frontend/src/session/access.ts` only if a helper is needed (no change to existing exports) |
| **INT** | `backend/internal/app/history_access_owned.go` (+ test), `frontend/e2e/dashboard-builder.spec.ts`, `frontend/e2e/custom-dashboards.spec.ts` | `internal/app/contract_routes_test.go` (empty the 051 entries of `notYetServed`), `internal/app/history_access.go` (call the owned check), `docs/sdd/feature-map.md` (status), this file (status) |

Nobody edits `backend/api/openapi.yaml` or `frontend/src/api/generated/` during the lanes: a contract change goes
back through the spec first, in one commit on the feature branch that every lane then merges.

## Lane interfaces (fixed by F0; code against these before the other lanes merge)

### Go — `internal/widgetdata` (F0 creates `ports.go`, `table.go`, `number.go`; B3 implements; B2 and the app adapters satisfy)

```go
package widgetdata

type FieldKind string

const (
	KindText   FieldKind = "text"
	KindNumber FieldKind = "number"
	KindMoney  FieldKind = "money"
	KindDate   FieldKind = "date"
)

// Number is an exact decimal of ten-thousandths: "12.5" is 125000.
type Number struct{ units int64 }

func ParseNumber(s string) (Number, error) // up to 15 integer digits and 4 decimals; "," thousands not accepted here;
                                            // ErrNumberFormat, or ErrNumberRange beyond ±922,337,203,685,477.5807 (int64 units)
func NumberOfUnits(units int64) Number
func (n Number) Units() int64
func (n Number) String() string            // shortest exact text: "12.5", "-3", "0.0001"

type Field struct {
	Key, Label     string
	Kind           FieldKind
	Multi, Finance bool // several values (a plan's teams); empty outside the finance scope
}

// Value is one cell; Empty wins over the other fields, and only the field of the column's kind is read.
type Value struct {
	Empty  bool
	Text   string
	Texts  []string    // a Multi field
	Number Number
	Money  money.Money // USD for built-in sources; the column's currency for datasets
	Date   date.Date
}

type Row []Value // in the order of the source's fields

type Prefilter struct {
	From, To         *date.Date
	PlanIDs, TeamIDs []string
}

type Rows struct {
	Items []Row
	Notes []string // plain sentences: "Financial figures cover 1 of 2 plans"
}

// SourceReader reads one built-in source for the actor, already limited to their scope and to the prefilter
// parts the source supports. internal/app implements one per built-in key.
type SourceReader interface {
	Rows(ctx context.Context, a actor.Actor, f Prefilter) (Rows, error)
}

type DatasetColumn struct {
	Key, Label string    // c1…c50
	Kind       FieldKind
	Currency   string    // money only
}

type DatasetInfo struct {
	ID, Name string
	Owner    authz.PersonRef
	Columns  []DatasetColumn
}

type Table struct {
	Info DatasetInfo
	Rows []Row
}

// DatasetTables is satisfied by *dataset.Service (B2). Table answers errs.NotFound for a dataset the actor may
// not see or that is deleted.
type DatasetTables interface {
	Visible(ctx context.Context, a actor.Actor) ([]DatasetInfo, error)
	Table(ctx context.Context, a actor.Actor, id string) (Table, error)
}

// Rates converts dataset money to USD at the current rate (settings.Service.Rates satisfies it).
type Rates interface {
	Rates(ctx context.Context) (money.Rates, error)
}

type Transactor interface {
	Read(ctx context.Context, fn func(ctx context.Context) error) error
}

type Deps struct {
	Sources  map[string]SourceReader // keys of data model §5
	Datasets DatasetTables
	Rates    Rates
}

// B3: func NewService(tx Transactor, deps Deps, now clock.Clock) *Service
//     func (s *Service) Catalogue(ctx context.Context, a actor.Actor) ([]Source, error)
//     func (s *Service) Query(ctx context.Context, a actor.Actor, r Request) (Data, error)
```

**B1 ↔ B3**: `customdashboard` validates a widget's query shape and options (data model §2–§3) without the
catalogue; fields are checked against the catalogue when data is asked for (B3). B1 exports
`customdashboard.DatasetSources(widgets []Widget) []string` for nothing else; no import between the two packages.

**B2 → B1 (INT)**: `dataset.Service.UsedBy` needs dashboards using a dataset: B1 exports
`(*customdashboard.Service).UsingDataset(ctx context.Context, a actor.Actor, datasetID string) ([]Ref, error)`
(visible dashboards only); B2 declares the port `dataset.Usage interface { UsingDataset(ctx, a, datasetID)
([]dataset.DashboardRef, error) }` and INT adapts one to the other in `internal/app/datasets.go`. Until then B2
answers `usedBy: []`.

**History (INT)**: B1 and B2 each export `OwnerOf(ctx context.Context, id string) (ownerID string, found bool, err
error)` on their PostgreSQL store; INT's `historyAccess` lets the owner (and holders of `activityLog.viewAll`) see
log entries of kinds `customDashboard` and `dataset`, and only the owner see and restore their bin entries.

### TypeScript — `frontend/src/features/custom-dashboards`

`model.ts` (F0):

```ts
import type { components } from '@/api/generated/schema';
type S = components['schemas'];
export type CustomDashboard = S['CustomDashboard'];
export type CustomDashboardSummary = S['CustomDashboardSummary'];
export type CustomDashboardFilters = S['CustomDashboardFilters'];
export type Widget = S['Widget'];
export type WidgetType = S['WidgetType'];
export type WidgetLayout = S['WidgetLayout'];
export type WidgetQuery = S['WidgetQuery'];
export type WidgetDisplay = S['WidgetDisplay'];
export type WidgetData = S['WidgetData'];
export type WidgetDataRequest = S['WidgetDataRequest'];
export type WidgetValue = S['WidgetValue'];
export type WidgetDataColumn = S['WidgetDataColumn'];
export type ChartPalette = S['ChartPalette'];
export type DataSource = S['DataSource'];
export type DataField = S['DataField'];
export type Dataset = S['Dataset'];
export type DatasetSummary = S['DatasetSummary'];
export type DatasetPreview = S['DatasetPreview'];
export type TeamRef = S['TeamRef'];
export const NO_FILTERS: CustomDashboardFilters; // {dateFrom: null, dateTo: null, planIds: [], teamIds: []}
export const GRID_COLUMNS = 12;
export const GRID_ROWS = 200;
export const MAX_WIDGETS = 30;
```

`keys.ts` (F0): `dashboardKeys.all`, `.one(id)`, `datasetKeys.all`, `.one(id)`, `.rows(id, offset)`,
`dataSourceKeys.all`, `widgetDataKeys.all`, `.one(request: WidgetDataRequest)` (`['widget-data', source,
stableHash(request)]`); `stableHash(value: unknown): string` is exported too (JSON with object keys sorted).

`widget-types.ts` (F0):

```ts
export type Need = 'required' | 'optional' | 'none';
export interface WidgetTypeInfo {
  type: WidgetType;
  label: string;                              // "Bar", "Bar and line", "Key figure", "Text note"
  group: 'charts' | 'figures' | 'text';
  icon: LucideIcon;
  size: { w: number; h: number };             // data model §2 default sizes
  dimension: Need;
  series: Need;
  measures: readonly [min: number, max: number]; // [0, 0] for text
}
export const WIDGET_TYPES: readonly WidgetTypeInfo[];
export function widgetTypeInfo(type: WidgetType): WidgetTypeInfo;
```

`charts/index.ts` (F1; F0 stubs each export with the same signature):

```ts
import type { EChartsCoreOption } from 'echarts/core';
export interface ChartHandle { toPng(): Promise<Blob> }
export interface EChartProps {
  option: EChartsCoreOption;
  description: string;                         // aria description (describeWidget)
  height: number;                              // px
  onReady?: (handle: ChartHandle) => void;
}
export function EChart(props: EChartProps): ReactElement;
export interface ChartTheme { ink: string; ink2: string; hairline: string; surface: string; palettes: Record<ChartPalette, string[]> }
export function useChartTheme(): ChartTheme;   // from tokens, follows the theme
export function buildOption(widget: Widget, data: WidgetData, theme: ChartTheme): EChartsCoreOption;
export function describeWidget(widget: Widget, data: WidgetData): string;
export function formatValue(value: WidgetValue, column: WidgetDataColumn, display: WidgetDisplay): string;
export function widgetDataRequest(widget: Widget, filters: CustomDashboardFilters | null): WidgetDataRequest | null; // null: text, or incomplete
export function useDataSources(): UseQueryResult<DataSource[], ApiError>;
export function useWidgetData(widget: Widget, filters: CustomDashboardFilters | null, options?: { debounceMs?: number }): UseQueryResult<WidgetData, ApiError>;
export interface WidgetViewProps {
  widget: Widget;
  filters: CustomDashboardFilters | null;
  mode: 'view' | 'edit';                       // edit: no menu, shows "incomplete" instead of fetching
  dashboardName: string;                       // PNG file name
  viewerOwnsDashboard: boolean;                // 404 of a dataset: "no longer available" vs "no access"
}
export function WidgetView(props: WidgetViewProps): ReactElement;   // every state of contracts/ui.md
export function WidgetTable(props: { widget: Widget; data: WidgetData }): ReactElement;
```

`sharing/SharingDialog.tsx` (F3; F0 stub):

```ts
export interface SharingDialogProps {
  open: boolean;
  recordLabel: string;                         // "dashboard" | "dataset"
  recordName: string;
  teams: TeamRef[];                            // shareableTeams
  selected: string[];                          // sharedWith ids
  saving: boolean;
  error: string | null;
  onSave: (teamIds: string[]) => void;
  onClose: () => void;
}
export function SharingDialog(props: SharingDialogProps): ReactElement;
```

Pages (each F0 stub exports a component with no props): `DashboardBuilderPage` (F2, reads `:dashboardId`),
`DashboardListPage`, `DashboardViewerPage`, `DatasetsPage`, `DatasetPage` (F3).

---

## Phase 1: Setup (done)

- [X] T001 [INT] Governance: constitution 4.9.0, CLAUDE.md, D-371–D-378, Q-028, Q-029, feature map, ADR-0020 —
  `.specify/memory/constitution.md`, `docs/sdd/*`, `docs/architecture/adr/0020-…`
- [X] T002 [INT] Spec, plan, research, data model, contracts, quickstart, tokens, product docs; the API contract and
  regenerated types; the 17 operations in `notYetServed` — `specs/051-dashboard-builder/`,
  `backend/api/openapi.yaml`, `frontend/src/api/generated/schema.d.ts`

## Phase 2: Foundation (F0 — blocks every lane; one commit on the feature branch)

- [x] T003 [F0] Permissions (FR-001, D-374): `authz.DashboardsView`, `authz.DashboardsBuild`, `GroupDashboards =
  "Dashboards"` between Global budgets and Administration, catalogue rows ("See custom dashboards shared with
  them" / "Seeing dashboards"; "Build dashboards and upload data" / "Building dashboards"; both not scoped) —
  `backend/internal/platform/authz/catalogue.go`; update the counts and groups in `catalogue_test.go`,
  `internal/access/roles_test.go`, `service_test.go`, `handler_test.go`
- [x] T004 [F0] Seed roles: `dashboards.view` to Global planner, Planner, Approver, Viewer and Requisitioner;
  `dashboards.build` to Global planner and Planner (Administrator holds `authz.All()`) — `backend/seed/access.go`,
  `seed/access_test.go`
- [x] T005 [F0] Route rules exactly as `contracts/api.md` — `backend/internal/access/routes_custom_dashboards.go`,
  `routes_datasets.go`, `routes_data_sources.go`, joined in `access/routes.go`; `TestEveryOperationHasAnAccessRule_SC001`
  passes again
- [x] T006 [F0] ID prefixes `PrefixDashboard = "DB"`, `PrefixDataset = "DS"` in `Prefixes` —
  `backend/internal/platform/ids/ids.go`
- [x] T007 [F0] Shared Go types and ports of "Lane interfaces", with `TestNumber_ParseAndString` —
  `backend/internal/widgetdata/{ports.go,table.go,number.go,number_test.go}`
- [x] T008 [F0] Wiring stubs: `func (w *wiring) customDashboardsArea() error`, `datasetsArea() error`,
  `widgetData()`, each returning nil / doing nothing, called from `areas()` (the two areas after
  `purchaseRequestsArea`, `widgetData` after `w.dashboards()`) — `backend/internal/app/{custom_dashboards.go,
  datasets.go,widget_data.go}`, `areas.go`; empty `CustomDashboardStore` and `DatasetStore` interfaces and a
  `Dashboards struct { CustomDashboards CustomDashboardStore; Datasets DatasetStore }` embedded in `stores.Set` —
  `backend/internal/stores/{custom_dashboards.go,datasets.go,stores.go}`; seed steps `d.addDatasets,
  d.addDashboards` after `d.addPurchaseRequests` with no-op bodies — `backend/seed/{datasets.go,dashboards.go,seed.go}`
- [x] T009 [F0] `cd frontend && pnpm add echarts@6.1.0` (exact pin; confirm `pnpm view echarts@6.1.0 time` is over
  a day old) — `frontend/package.json`, `pnpm-lock.yaml`
- [x] T010 [F0] `model.ts`, `keys.ts`, `widget-types.ts` as in "Lane interfaces", with `widget-types.test.ts`
  (needs match data model §2.3) — `frontend/src/features/custom-dashboards/`
- [x] T011 [F0] Stubs with the exact public signatures (placeholders rendering `null` or "Coming in this feature",
  hooks returning an idle query) — `charts/index.ts`, `sharing/SharingDialog.tsx`, `builder/DashboardBuilderPage.tsx`,
  `viewer/DashboardListPage.tsx`, `viewer/DashboardViewerPage.tsx`, `datasets/DatasetsPage.tsx`,
  `datasets/DatasetPage.tsx`, `index.ts`
- [x] T012 [F0] Run `go build ./... && go test -short ./...` and `pnpm run check`; commit
  `feat(051): foundation for the six lanes (FR-001)`

**Checkpoint**: the lanes fork from this commit.

## Phase 3: Lane B1 — `internal/customdashboard` (US1, US2, US4)

- [x] T013 [P] [B1] [US1] Migration and queries of data model §8 (00028, including the role update) —
  `migrations/00028_custom_dashboards.sql`, `queries/custom_dashboards.sql`; `sqlc generate && sqlc vet`
- [x] T014 [P] [B1] [US1] `model.go` (Dashboard, Widget, Layout, Query, Display, Filters, Summary, Create, Update,
  Copy, Sharing inputs; JSON decoding of query and display with unknown fields refused)
- [x] T015 [B1] [US1] Tests first: `TestValidateDashboard_NameAndDescription_FR008`,
  `TestValidateWidget_RulesByType_FR015` (every type of §2.3), `TestValidateQuery_FieldsShapesAndValues_FR016_FR020`
  (shape only; operators and value counts §2.2), `TestValidateDisplay_OptionsByType_FR024_FR025`,
  `TestLayout_InsideAndNoOverlap_FR013`, `TestValidate_EveryProblemAtOnce_FR033` — `validate_test.go`, `layout_test.go`
- [x] T016 [B1] [US1] Implement `validate.go`, `layout.go` (functions ≤ complexity 8: one function per rule group)
- [x] T017 [B1] [US1] `store.go` (Store: `ListVisible`, `GetVisible`, `GetOwned`, `Insert`, `Update`, `Delete`,
  `ReplaceWidgets`, `ReplaceShares`, `UsingDataset`, `OwnerOf`, `NameTaken`, `CountOf`), `store_postgres.go`,
  `store_fake_test.go`, `customdashboardtest/store.go`; `stores/custom_dashboards.go` interface; wire in
  `stores/postgres.go` and `stores/storestest`
- [x] T018 [B1] [US2] Scope isolation on PostgreSQL: `TestPostgresStoreVisibleDashboards_ScopeIsolation_FR005`
  (owner, shared team member, former member, other person; list and get) — `store_postgres_test.go`
- [x] T019 [B1] [US1] Tests first then `service.go`: `List`, `Get`, `Create`, `Update` (owner 403, invisible 404,
  version 409, name 409, widgets replaced), `Delete` (`history.MoveToBin`), access objects —
  `TestCreate_OwnedUnsharedLogged_FR008_FR009`, `TestUpdate_NotOwner403_Invisible404_FR006`,
  `TestUpdate_VersionConflict_FR012`, `TestDelete_ToRecycleBinAndUndo_FR010`
- [x] T020 [B1] [US4] Tests first then `service_share.go`: `Share` (owner's teams or already shared; log
  "Shared with …"), `Copy` ("Copy of …", 80 characters, numbered, unshared, any viewer with build) —
  `TestShare_OwnTeamsOnly_FR046`, `TestCopy_NameAndOwner_FR011`, `TestShare_FormerMemberLosesSight_US4`
- [x] T021 [B1] [US1] `kinds.go`: kind `customDashboard` (snapshot with widgets and shares, Describe for widgets
  and sharing, PrepareRestore renaming "(restored)") — `TestLog_DescribesWidgetChanges_FR009`, `TestRestore_RenamesWhenTaken_FR010`
- [x] T022 [B1] [US1] `handler.go` (Gin only here; body limit 512 KB) and contract tests for the 7 operations
  (`package customdashboard_test`, `apitest`) — `handler_test.go`
- [x] T023 [B1] [US2] Wiring `customDashboardsArea` (service, kind registration, handler) —
  `internal/app/custom_dashboards.go`; seed DB-1001 "Lubricants Europe overview" by Priya Raman, shared with
  Lubricants Europe: bar (activities by type, sum of planned cost), line (planned cost by month by campaign), key
  figure (sum of plan budgets), table (campaigns: budget, planned cost), text note, and a bar over `DS-1001` (leads by
  month) — `seed/dashboards.go`, `seed/dashboards_test.go`; demo story paragraph in `docs/product/demo-users-and-data.md`

## Phase 4: Lane B2 — `internal/dataset` (US3, US4)

- [ ] T024 [P] [B2] [US3] Migration and queries of data model §8 (00029) — `migrations/00029_datasets.sql`,
  `queries/datasets.sql`; rows inserted with `pgx.CopyFrom`
- [ ] T025 [B2] [US3] Tests first then `read.go` (base64 → `sheet.ReadCSV` / `sheet.ReadXLSX` first table,
  `maxRows` 10,001, header row, 50 columns, 5 MB, cells ≤ 500, empty or duplicate names) —
  `TestPreview_FormatsAndLimits_FR041_FR043` (CSV comma/semicolon/tab, BOM, .xlsx dates and formulas, non-UTF-8,
  too large, too many rows and columns, no header)
- [ ] T026 [B2] [US3] Tests first then `infer.go` and `cells.go` (data model §7: kinds proposed from the first 200
  cells, money by column name; parse numbers with separators, dates in three formats and Excel cells, money to minor
  units) — `TestInferKinds_FR041`, `TestParseCells_EveryKind_FR042` (fast-check round trips)
- [ ] T027 [B2] [US3] `model.go`, `store.go` (`ListVisible`, `GetVisible`, `Rows`, `RowsPage`, `Insert` with columns
  and rows, `Update`, `UpdateLabels`, `Delete`, `ReplaceShares`, `CountOfOwner`, `OwnerOf`), `store_postgres.go`,
  fakes (`store_fake_test.go`, `datasettest/store.go`), `stores/datasets.go`; wire in `stores/postgres.go`, `storestest`
- [ ] T028 [B2] [US4] `TestPostgresStoreVisibleDatasets_ScopeIsolation_FR044` (owner, shared member, other; list,
  get, rows) — `store_postgres_test.go`
- [ ] T029 [B2] [US3] Tests first then `service.go`, `service_upload.go` (`Preview`, `Create`: every cell problem at
  once, 50 per person, name unique, active currency), `service_share.go` (`Share`, owner only), `Update` (labels
  only, types fixed), `Delete` — `TestCreate_CellProblemsAllAtOnce_FR042`, `TestCreate_LimitsPerPerson_FR043`,
  `TestUpdate_ColumnKindsNeverChange_FR044`, `TestShare_OwnTeamsOnly_FR046`, `TestDelete_ToRecycleBin_FR045`
- [ ] T030 [B2] [US3] `tables.go`: `Visible` and `Table` satisfying `widgetdata.DatasetTables` (404 as `errs.NotFound`)
  — `TestTables_VisibleOnlyAndTyped_FR044`
- [ ] T031 [B2] [US3] `kinds.go` (kind `dataset`, snapshot with columns, shares, rows); `handler.go` (body limit 7
  MB on preview and create) and contract tests for the 8 operations — `handler_test.go`
- [ ] T032 [B2] [US3] Wiring `datasetsArea`; seed DS-1001 "Event leads 2026" by Priya Raman (12 rows: Event,
  Date, Leads, Agency cost EUR), not shared — `internal/app/datasets.go`, `seed/datasets.go` (+ test); fixtures
  `frontend/e2e/fixtures/event-leads.xlsx` and `.csv`

## Phase 5: Lane B3 — `internal/widgetdata` (US1, US2, US3, US5)

- [ ] T033 [P] [B3] [US1] `catalogue.go`: the ten sources of data model §5 (fields, kinds, multi, finance, date /
  plan / team filters, needed permission) — `TestCatalogue_SourcesByPermission_FR016` (Viewer, Planner, no finance,
  global budgets) with datasets from a fake `DatasetTables`
- [ ] T034 [B3] [US1] Tests first then `query.go` (request → validated query against the source's fields: groupable,
  aggregations allowed, operators by kind, values parsed, type rules §2.3; every problem at once) —
  `TestQueryValidation_AllProblems_FR019`
- [ ] T035 [P] [B3] [US1] Tests first then `filter.go` — `TestFilters_EveryOperator_FR019` (each kind, empty values,
  multi-valued fields)
- [ ] T036 [P] [B3] [US1] Tests first then `bucket.go`, `group.go` — `TestBuckets_DayWeekMonthQuarterYear_FR017` (ISO
  week across years), `TestGroup_MultiValuedCountsInEach_D373`, `TestGroup_EmptyIsNotSet`
- [ ] T037 [P] [B3] [US1] Tests first then `aggregate.go` (count, countDistinct, sum, avg, min, max; exact; half up;
  overflow → validation_failed; shares; totals) — `TestAggregate_EveryKind_FR018`, `TestShares_OfGroupOrTotal_FR021`
- [ ] T038 [P] [B3] [US1] Tests first then `order.go` (default sorts, ties by label, limit, Other, series cap 20,
  groupCount) — `TestOrder_LimitAndOther_FR020`
- [ ] T039 [P] [B3] [US5] Tests first then `compare.go` (target share; previous period of equal length; null on zero)
  — `TestComparison_TargetAndPreviousPeriod_FR023`
- [ ] T040 [B3] [US1] `service.go` (`Catalogue`, `Query`: permission view or build → 403; source permission → 403;
  dataset → 404; prefilter only where the source supports it, never datasets; notes; dataset money to USD with
  `Rates`) and `handler.go`; contract tests for the 2 operations on fakes of the readers — `handler_test.go`,
  `TestPrefilter_AppliedPerSource_FR022`
- [ ] T041 [B3] [US1] App adapters, one `SourceReader` per key, over All work's scoped year data (years overlapping
  the date range, else every year in scope; read once per request with `memo`), `allwork.PlannedByActivity`,
  financials (`SpendOnLineItems` or a new scoped `TransactionsOnLineItems` with
  `TestPostgresStoreTransactionsOnLineItems_ScopeIsolation_051_FR007`), `approval.Service.RoundSummaries`,
  purchase requests (`OnLineItems`), `budget.Service.YearFigures` — `internal/app/widgetdata_planning.go`,
  `widgetdata_money.go`, `widgetdata_workflow.go`; wiring `widgetData()` in `internal/app/widget_data.go` with a
  fake-free `DatasetTables` placeholder answering no datasets until INT
- [ ] T042 [B3] [US2] Persona contract tests on the demo data — `TestQuery_OnlyThePlansInScope_FR007` (Priya, Mei,
  Sara), `TestQuery_FinanceHiddenOutsideFinanceScope_FR007`, `TestQuery_EqualsTheRecordPages_SC002` — and
  `BenchmarkWidgetData` (SC-004) — `internal/app/widgetdata_test.go`, `internal/platform/apibench/widget_data_test.go`

## Phase 6: Lane F1 — `charts/` (US1, US2, US5)

- [ ] T043 [P] [F1] [US1] `theme.ts` (`useChartTheme`: tokens by `getComputedStyle`, re-read on theme change) and
  `format.ts` (`formatValue`: number, money with currency, percent, compact, decimals; date buckets in the person's
  date format) — `format.test.ts`
- [ ] T044 [F1] [US1] Tests first then `build-option.ts` dispatching to `build-axis.ts` (bar, line, area, combo,
  scatter, heatmap with piecewise 5-bin visualMap), `build-round.ts` (pie, donut, funnel, gauge), `build-tree.ts`
  (treemap, radar): palettes, decals, legend, labels, axis titles, gridlines, smooth, stack (100 % from `share`),
  orientation, zoom, reference line, measure styles, `aria`, `tooltip` with `confine: true` and a DOM formatter
  (`tooltip.ts`, `textContent` only) — `build-option.test.ts` (one block per type and option, FR-024, FR-025),
  `tooltip.test.ts` (no HTML strings)
- [ ] T045 [F1] [US2] `EChart.tsx` (`echarts/core`, registered parts of ADR-0020, SVG renderer, locale EN,
  ResizeObserver, dispose, `onReady` handle) and `png.ts` (SVG → canvas ×2 on `--surface` → PNG; file name) —
  `EChart.test.tsx` (init options, dispose; ECharts mocked), `png.test.ts`
- [ ] T046 [F1] [US2] `describe.ts` (`describeWidget`) and `WidgetTable.tsx` (caption, headers, formatted values,
  totals row, chosen columns) — `describe.test.ts`, `WidgetTable.test.tsx` (axe)
- [ ] T047 [F1] [US1] `api.ts`: `useDataSources`, `useWidgetData` (debounce, keys of `keys.ts`), `widgetDataRequest`
  — `api.test.ts` (MSW)
- [ ] T048 [F1] [US2] `WidgetView.tsx`, `WidgetStates.tsx`, `KeyFigure.tsx`, `TextNote.tsx`: every state of
  contracts/ui.md (loading, error + Retry, no data, no access for 403 and dataset 404, gone for the owner, incomplete
  in edit mode, chart, table view, key figure with comparison, text note), widget menu (Show as table, Download as
  PNG) and notes — `WidgetView.test.tsx` (one test per state, FR-039; axe)
- [ ] T049 [F1] Replace F0's stub `charts/index.ts` with the real exports; `pnpm run check`

## Phase 7: Lane F2 — `builder/` (US1, US4, US5)

- [ ] T050 [P] [F2] [US1] Tests first then `grid.ts` (first free place, placement of a drop, move and resize within
  bounds without overlap, keyboard steps) — `grid.test.ts` with fast-check (never overlaps, never outside)
- [ ] T051 [P] [F2] [US1] Tests first then `builder-reducer.ts` (load, rename, add, duplicate, delete, move, resize,
  set query, set display, set filters; undo 50 / redo; dirty flag) — `builder-reducer.test.ts` (FR-027–FR-030)
- [ ] T052 [F2] [US1] Tests first then `validate.ts` (the rules of data model §2.3 and §3 the API checks, mapped to
  field paths and widget messages) — `validate.test.ts` (FR-015, FR-033)
- [ ] T053 [F2] [US1] `Canvas.tsx`, `CanvasWidget.tsx` (pointer drag and resize with capture, keyboard arrows and
  Shift+arrows, Delete, Ctrl/⌘+D, Escape, live-region announcements, drop of palette items), `WidgetPalette.tsx`
  (click or drag) — `Canvas.test.tsx` (keyboard moves and announcements, FR-028; drag with pointer events)
- [ ] T054 [F2] [US1] `Inspector.tsx`, `DataTab.tsx` (fields from `useDataSources`, source change keeps what fits
  and says what was cleared, measures, filters, sort, top N, Other, Upload data link), `StyleTab.tsx` (options shown
  by type), `LayoutTab.tsx`, `DashboardDetails.tsx` (name, description, default filters); React Hook Form + Zod —
  `DataTab.test.tsx`, `StyleTab.test.tsx` (FR-031)
- [ ] T055 [F2] [US1] `BuilderHeader.tsx`, `DashboardBuilderPage.tsx`, `api.ts` (create, update, copy): Save (all
  problems at once, focus first, navigate to `/builder/DB-…` after create), Undo/Redo shortcuts, View, Sharing (opens
  `SharingDialog`), unsaved-changes guard (`useUnsavedChanges`), version conflict (Reload / Keep editing), view-only
  for non-owners with Make a copy, narrow-window note, live previews through `WidgetView mode="edit"` —
  `DashboardBuilderPage.test.tsx` (US1 scenarios 1–12, axe)

## Phase 8: Lane F3 — viewer, datasets, sharing, routes (US2, US3, US4)

- [ ] T056 [P] [F3] [US2] `viewer/api.ts` (list, get, delete with Undo, copy, share) and `filters-address.ts` (Zod;
  invalid values fall back) — `filters-address.test.ts`
- [ ] T057 [F3] [US2] `DashboardListPage.tsx`, `DashboardCard.tsx`: Built-in (plans.view), My dashboards, Shared with
  me, New dashboard, Datasets link, card menu; states — `DashboardListPage.test.tsx` (FR-003, each state, axe)
- [ ] T058 [F3] [US2] `DashboardViewerPage.tsx`, `FiltersBar.tsx`, `ViewerGrid.tsx` (grid, one column below 960 px),
  Refresh, Edit (owner), Make a copy, delete; not-found for private — `DashboardViewerPage.test.tsx` (FR-036–FR-039)
- [ ] T059 [P] [F3] [US4] `sharing/SharingDialog.tsx` (team checkboxes, note, Enter saves, Escape closes) —
  `SharingDialog.test.tsx` (FR-046, axe)
- [ ] T060 [F3] [US3] `datasets/api.ts`, `DatasetsPage.tsx`, `UploadDialog.tsx` (choose file → check → columns form
  with Name, Type, Currency, samples, first rows, problems → save), `ColumnTypesForm.tsx`, `DatasetPage.tsx`
  (details, columns rename, rows paged with `RowsTable.tsx`, used by, share, delete with the usage warning) —
  `UploadDialog.test.tsx`, `DatasetsPage.test.tsx`, `DatasetPage.test.tsx` (FR-040–FR-045, each state)
- [ ] T061 [F3] [US2] Routes of contracts/ui.md (`/dashboards` list instead of the redirect, executive and regional
  before `:dashboardId`, `/builder`, `/builder/:dashboardId`, `/datasets`, `/datasets/:datasetId`) with
  `RequirePermission`, lazy pages — `app/dashboard-routes.tsx`, `app/pages.tsx`; sidebar Dashboards (plans.view or
  dashboards.view) and Dashboard builder (dashboards.build) — `app/navigation.ts`, `app/Sidebar.tsx`,
  `Sidebar.test.tsx`, `router.test.tsx` (FR-002, FR-004)
- [ ] T062 [F3] Replace F0's stubs `index.ts`, viewer and datasets pages; `pnpm run check`

## Phase 9: Integration (INT)

- [ ] T063 [INT] Merge B2, B1, B3, F1, F3, F2 into `feature/051-dashboard-builder` in that order; resolve the additive
  edits; `sqlc generate` once more; `go build ./...`
- [ ] T064 [INT] [US3] Wire `dataset.Service` as `widgetdata.DatasetTables` and `customdashboard.Service.UsingDataset`
  as `dataset.Usage` — `internal/app/widget_data.go`, `internal/app/datasets.go`; `TestQuery_DatasetSharedOrNot_FR044`
  (Sara: 404 until shared; owner after delete: 404)
- [ ] T065 [INT] History access for owned kinds: log entries to the owner and `activityLog.viewAll`, bin entries and
  restore to the owner — `internal/app/history_access_owned.go`, a call from `history_access.go`,
  `history_access_owned_test.go` (FR-009, FR-010, FR-045)
- [ ] T066 [INT] Empty the 051 entries of `notYetServed`; all contract, access and matrix tests pass —
  `internal/app/contract_routes_test.go`
- [ ] T067 [INT] [US1] E2E: Priya builds a dashboard (palette click and drag, Data tab, keyboard move and resize, undo,
  validation summary, save, reload, figures equal the activity pages; guard on leaving; Viewer without the builder
  item) — `frontend/e2e/dashboard-builder.spec.ts` (axe, CSP guard, both themes)
- [ ] T068 [INT] [US2] E2E: Daniel opens Dashboards (built-ins open 036's pages), the shared dashboard, tooltip, legend
  toggle, Show as table, Download as PNG (file name and PNG signature), filters in the address with Back, no Edit;
  Mei gets "Dashboard not found" — `frontend/e2e/custom-dashboards.spec.ts`
- [ ] T069 [INT] Full checks: backend `golangci-lint fmt ./... && golangci-lint run ./...`, `sqlc generate && sqlc
  vet`, `go test -race ./...` with `TEST_DATABASE_URL`, `govulncheck ./...`; frontend `pnpm run check`,
  `pnpm run test:coverage` (80 % decision coverage of new code), `pnpm run build` (the custom-dashboards chunk holds
  ECharts; no other chunk does), `pnpm run test:e2e`; comment budget and 300-line limit of every touched file
- [ ] T070 [INT] Run `/speckit-analyze`; update `docs/sdd/feature-map.md` status to Built and this file's status;
  open the pull request against `develop` with the FR IDs

## Dependencies

- F0 (T003–T012) blocks every lane. Lanes B1, B2, B3, F1, F2, F3 run in parallel after it.
- Inside lanes: B1 T013–T014 → T015–T016 → T017–T018 → T019–T023; B2 T024 → T025–T026 → T027–T028 → T029–T032;
  B3 T033 → T034 → T035–T039 (parallel) → T040 → T041–T042; F1 T043 → T044 → T045–T048 → T049; F2 T050–T052 →
  T053–T054 → T055; F3 T056 → T057–T060 → T061–T062.
- INT T063 needs all lanes; T064–T066 before T067–T068; T069–T070 last.

## Requirement coverage

| Requirements | Tasks |
|---|---|
| FR-001, FR-002 | T003–T005, T061 |
| FR-003, FR-004 | T057, T061, T068 |
| FR-005, FR-006 | T018, T019, T058 |
| FR-007 | T041, T042 |
| FR-008–FR-012 | T015–T022, T055 |
| FR-013–FR-015 | T015, T016, T050, T052 |
| FR-016–FR-023 | T033–T042 |
| FR-024, FR-025 | T015, T044, T054 |
| FR-026–FR-035 | T050–T055, T067 |
| FR-036–FR-039 | T045–T048, T058, T068 |
| FR-040–FR-045 | T024–T032, T060, T064 |
| FR-046 | T020, T029, T059 |
| SC-001, SC-006, SC-007 | T067, T069 |
| SC-002, SC-003 | T042, T064 |
| SC-004, SC-005 | T042, T032 (`BenchmarkDatasetUpload` in T031's package tests) |
