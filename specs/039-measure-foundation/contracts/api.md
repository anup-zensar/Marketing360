# API contract: Measure Foundation (039)

The operations below go into `backend/api/openapi.yaml` with the **wave A contract** (build guide §3 step 1),
together with the report operations of 040–045 (their own `contracts/api.md`). Errors use the one error
envelope; every operation answers 401 when signed out and 403 `forbidden` when the route rule refuses.

## Operations

| Operation | Rule (`access/routes_measure.go`) | Answers |
|---|---|---|
| `GET /api/v1/reports` | `need(authz.ReportsView)` | 200 `ReportCatalogue` |
| `GET /api/v1/reports/frame` | `need(authz.ReportsView)` | 200 `ReportFrame`; 400 problems for a malformed filter; 404 unknown `report` |
| `GET /api/v1/reports/{report}/export` | `need(authz.DataExport)` | 200 file (`.xlsx` or `.csv`); 400 `format=csv` without `widget`, unknown `widget`; 404 unknown report. A report not built yet exports its **In view** and **About this export** sheets only |
| `GET /api/v1/reports/{report}/views` | `need(authz.ReportsView)` | 200 `ReportViewPage` (the person's views of the report) |
| `POST /api/v1/reports/{report}/views` | `need(authz.ReportsView)` | 201 `ReportView`; 400 problems (name, query) |
| `PATCH /api/v1/reports/{report}/views/{viewId}` | `need(authz.ReportsView)` | 200 `ReportView`; 400 problems; 404 another person's or missing view; 409 `version_conflict` for a stale `version` |
| `DELETE /api/v1/reports/{report}/views/{viewId}` | `need(authz.ReportsView)` | 200 `ReportView` (the deleted view, for Undo); 404 as above |

`{report}` is the path of data-model §3 (`plan-budget`, `approvals`, …). The frame takes `report` as the key
(`planBudget`, …) in its query, with the report filter of data-model §2.

## Schemas

```yaml
ReportCatalogue:
  items: ReportEntry[]              # data-model §3 order; only reports the person may open
ReportEntry:
  key: string                       # planBudget, approvals, …
  path: string                      # plan-budget, …
  name: string
  purpose: string
  available: boolean                # false → "Coming soon", no link

ReportFrame:
  report: string
  filter: ReportFilter              # as applied, defaults filled in
  choices: ReportChoices
  inView: { plans: int, campaigns: int, initiatives: int, activities: int, lineItems: int }
  finance: { plans: int, of: int }  # plans whose financial data the person sees, of the plans in view
  leftOut: int                      # filter values left out (unknown, inactive or out of scope); never named
  displayCurrency: string           # the person's
  generatedAt: date-time
  financeAsOf: date-time | null
ReportFilter:
  year: int
  plans: string[]; planTypes: string[]; teams: string[]; owners: string[]
  activityTypes: string[]; activityStatuses: string[]
  attributes: { attributeId: string, values: string[] }[]
  period: month | quarter
  show: plan | campaign | initiative | activity | null
  currency: string
ReportChoices:
  years: int[]
  plans: { id, name, planType, teamIds[] }[]
  planTypes: string[]
  teams: { id, name, mine: boolean }[]
  owners: { id, name }[]            # owners of a record in the selected plans
  activityTypes: { id, name }[]     # types used in the selected plans
  activityStatuses: string[]
  attributeGroups: { level, attributeId, label, values: { id, label, active }[] }[]   # Not filled in = "none"
  currencies: { code, name }[]

ReportView:
  id: string                        # RV-1001
  report: string                    # key
  name: string
  isDefault: boolean
  query: string                     # canonical filter query string
  version: int                      # record.Meta
  createdAt: date-time
  updatedAt: date-time
ReportViewPage: { items: ReportView[], total: int }
ReportViewCreate: { name: string, query: string, isDefault?: boolean }
ReportViewUpdate: { version: int, name?: string, query?: string, isDefault?: boolean }
```

Export query: the report filter plus `format` (`xlsx` default, `csv`) and `widget` (a key the report declares;
required for CSV). `Content-Disposition` names the file "SP&A One – <report> [– <widget>] – <year> – <date>"
(research R7).

`Permission` enum: `reports.view` is added.

## The convention every report follows (040–050)

- `GET /api/v1/reports/<path>` with the report filter returns the report's widgets — its own schema, defined in
  its feature's `contracts/api.md`; rule `need(authz.ReportsView)` plus the report's own (data-model §3).
- The report's widgets also come out of its `measure.Report.Sheets` (research R6), so the shared export endpoint
  serves them; a report adds no export operation of its own.
- Every operation of the wave is listed in `notYetServed` (`internal/app/contract_routes_test.go`) under a
  comment line per feature until its lane serves it.
