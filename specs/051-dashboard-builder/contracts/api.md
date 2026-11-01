# API Contract: Dashboard Builder

`backend/api/openapi.yaml` holds the paths and schemas (tags **Custom dashboards**, **Datasets**, **Data sources**;
schemas `CustomDashboard*`, `Widget*`, `ChartPalette`, `Decimal`, `SharePercent`, `Dataset*`, `DataSource*`,
`DataField*`, `SharingUpdate`). This page lists the operations with their access rule and the errors each answers.
Every operation is under `/api/v1`, answers the one error envelope, and is listed in `notYetServed`
(`internal/app/contract_routes_test.go`) until its lane serves it.

## Operations

| Operation | operationId | Rule (`access/routes_*.go`) | Service checks | Answers |
|---|---|---|---|---|
| `GET /custom-dashboards` | `listCustomDashboards` | `need(dashboards.view)` | visible only | 200 `CustomDashboardList` |
| `POST /custom-dashboards` | `createCustomDashboard` | `need(dashboards.build)` | name, filters, widgets (data model §1–§4) | 201 `CustomDashboard`; 400, 409 `name_taken`, 413 |
| `GET /custom-dashboards/{dashboardId}` | `getCustomDashboard` | `need(dashboards.view)` | visible | 200; 404 |
| `PATCH /custom-dashboards/{dashboardId}` | `updateCustomDashboard` | `need(dashboards.build)` | visible (404), owner (403), version | 200; 400, 403, 404, 409 `version_conflict` / `name_taken`, 413 |
| `DELETE /custom-dashboards/{dashboardId}` | `deleteCustomDashboard` | `need(dashboards.build)` | visible, owner | 200 `DeleteResult`; 403, 404 |
| `POST /custom-dashboards/{dashboardId}/copy` | `copyCustomDashboard` | `need(dashboards.build)` | visible | 201 `CustomDashboard`; 400, 404, 409 |
| `PUT /custom-dashboards/{dashboardId}/sharing` | `shareCustomDashboard` | `need(dashboards.build)` | owner, teams allowed, version | 200; 400, 403, 404, 409 |
| `GET /datasets` | `listDatasets` | `need(dashboards.build)` | visible only | 200 `DatasetList` |
| `POST /datasets/preview` | `previewDataset` | `need(dashboards.build)` | file readable, limits | 200 `DatasetPreview`; 400, 413 |
| `POST /datasets` | `createDataset` | `need(dashboards.build)` | file, columns, every cell, 50 per person | 201 `Dataset`; 400, 409 `name_taken`, 413 |
| `GET /datasets/{datasetId}` | `getDataset` | `need(dashboards.build)` | visible | 200; 404 |
| `PATCH /datasets/{datasetId}` | `updateDataset` | `need(dashboards.build)` | owner, version | 200; 400, 403, 404, 409 |
| `DELETE /datasets/{datasetId}` | `deleteDataset` | `need(dashboards.build)` | owner | 200 `DeleteResult`; 403, 404 |
| `GET /datasets/{datasetId}/rows` | `listDatasetRows` | `need(dashboards.build)` | visible; `offset` 0–10,000, `limit` 1–200 | 200 `DatasetRowPage`; 400, 404 |
| `PUT /datasets/{datasetId}/sharing` | `shareDataset` | `need(dashboards.build)` | owner, teams allowed, version | 200; 400, 403, 404, 409 |
| `GET /data-sources` | `listDataSources` | `customWith(dashboards.view)` | view **or** build; each source's own permission | 200 `DataSourceList`; 403 |
| `POST /data-sources/query` | `queryWidgetData` | `customWith(dashboards.view)` | view or build; source permission (403); dataset visible (404); query rules (400) | 200 `WidgetData`; 400, 403, 404 |

Paths of 401 (no session), and 403 from the route table (missing permission), as everywhere.

## Body limits (038 FR-054)

`POST /datasets/preview` and `POST /datasets`: 7 MB (a 5 MB file in base64 plus the columns); `POST` and
`PATCH /custom-dashboards…`: 512 KB (30 widgets); the others: the default.

## Validation messages (examples, every problem at once)

| Field | Message |
|---|---|
| `name` | "Enter a name." · "Use at most 80 characters." |
| `widgets[3].layout` | "This widget overlaps 'Planned cost by type'." · "This widget goes past the grid's right edge." |
| `widgets[1].query.dimension` | "Choose what to group by." |
| `widgets[1].query.measures` | "A pie chart shows one measure." |
| `widgets[0].query.measures[0].field` | "Sum needs a number or money field." |
| `widgets[2].query.filters[0].values` | "Enter a date as YYYY-MM-DD." |
| `widgets[4].display.kpi.target` | "Enter the target." |
| `teamIds` | "You can share only with your teams." |
| `content` | "The file is larger than 5 MB." · "The file has more than 10,000 rows." · "SP&A One can read .xlsx and CSV (UTF-8) files." |
| `columns[2]` | "Row 14, Leads: 'n/a' is not a number." · "… and 37 more problems in this column." |
| `query.source` | "This data source doesn't exist." |

## Widget data request (example)

```json
{
  "type": "bar",
  "query": {
    "source": "activities",
    "dimension": { "field": "activityType", "dateGroup": null },
    "series": null,
    "measures": [{ "aggregation": "sum", "field": "plannedCost", "label": "" }],
    "filters": [{ "field": "status", "operator": "ne", "values": ["Declined"] }],
    "sort": { "by": "measure", "measureIndex": 0, "direction": "desc" },
    "limit": 10,
    "groupRest": true
  },
  "filters": { "dateFrom": "2026-07-01", "dateTo": "2026-12-31", "planIds": [], "teamIds": [] },
  "comparison": null
}
```

Answer (shortened):

```json
{
  "source": "activities",
  "columns": [
    { "key": "dimension", "label": "Activity type", "kind": "text", "dateGroup": null },
    { "key": "m0", "label": "Sum of Planned cost", "kind": "money", "dateGroup": null }
  ],
  "rows": [
    { "dimension": { "value": "Event", "label": "Event", "kind": "value" }, "series": null,
      "values": [{ "number": null, "money": { "amount": "240000.00", "currency": "USD" }, "date": null, "share": "41.2" }] }
  ],
  "totals": [{ "number": null, "money": { "amount": "582500.00", "currency": "USD" }, "date": null, "share": null }],
  "groupCount": 6,
  "notes": [],
  "filtersApplied": { "date": true, "plan": true, "team": true },
  "comparison": null
}
```

## Contract tests (one per operation, `apitest` + `libopenapi-validator`)

`customdashboard/handler_test.go` (B1), `dataset/handler_test.go` (B2), `widgetdata/handler_test.go` (B3), each in
`package <name>_test`, on `apitest.DemoDB`: every operation's 2xx answer validated, plus 400 (all problems at once),
403 (Viewer for build operations; non-owner), 404 (another person's private record), 409 (version), named
`Test<Operation>_<Case>_FR0xx`.
