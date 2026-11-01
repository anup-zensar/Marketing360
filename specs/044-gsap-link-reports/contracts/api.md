# API Contract: GSAP Link Reports (044)

For the **wave A contract**, written into `backend/api/openapi.yaml` (tag **Reports**) by PR #57; `openapi.yaml`
is the reference where the two differ. It follows
039's convention (`specs/039-measure-foundation/contracts/api.md`, "The convention every report follows"):
`GET /api/v1/reports/<path>` with the report filter returns the report's widgets, and the report's
`measure.Report.Sheets` feeds 039's shared export, so 044 adds **no export operation of its own**. Errors use
the one envelope; 401 when signed out, 403 `forbidden` when the route rule refuses.

## Operations

| Operation | Query | Answer | Rule (`access/routes_measure.go`) |
|---|---|---|---|
| `GET /api/v1/reports/gsap-link` (`getGsapLinkReport`) | 039's report filter (data-model §2: `year`, `plan`, `planType`, `team`, `owner`, `activityType`, `activityStatus`, `attr`, `currency`; `period`, `show` and `table` ignored) + `tab` (`swbs` default, `waiting`, `orders`), `sort` (a column key, `-` for descending), `page` (from 1), `pageSize` (5–500, default 20) | 200 `GsapLinkReport`; 400 problems for a malformed filter, an unknown `tab` or `sort`, or `pageSize` out of range | `need(authz.ReportsView)` |

**Export** is 039's `GET /api/v1/reports/gsap-link/export` (`need(authz.DataExport)`): without `widget`, one
Excel sheet per widget — `swbs`, `swbsBands`, `waiting`, `orders` — then 039's In view and About this export;
with `widget`, that one as CSV or Excel. Every row of the tab in the current `sort`, never only a page.

Records come only from plans in the person's `reports.view` scope (039 research R2). Amounts cover only plans
where the person holds `finance.view` (D-237): `waiting` rows of other plans carry `amount: null`; `orders` lists
no row of other plans. The empty states are 039's, chosen by the web app from the frame's In view and this
response's `total` and `finance`.

## Schemas

```yaml
GsapLinkReport:
  required: [tab, currency, total, page, pageSize, sort, finance, swbs, waiting, orders]   # the last four null when not set
  properties:
    tab:      { enum: [swbs, waiting, orders] }
    currency: { type: string }            # the report filter's currency, ISO code (D-303)
    total:    { type: integer }           # rows of the tab across all pages
    page:     { type: integer }
    pageSize: { type: integer }
    sort:     { type: string }            # as applied, default "-waiting" (swbs, waiting), "lineItem" (orders)
    finance:  { $ref: FinanceCoverage, nullable: true }   # 039's { plans, of }; null on swbs
    swbs:     { $ref: GsapSwbsTab, nullable: true }       # set when tab = swbs
    waiting:  { $ref: GsapWaitingTab, nullable: true }    # set when tab = waiting
    orders:   { $ref: GsapOrdersTab, nullable: true }     # set when tab = orders

GsapPath:                                  # where the line item or request sits
  required: [planId, plan, campaign, initiative, activityId, activity]
  properties: { planId, plan, campaign, initiative, activityId, activity: { type: string } }  # "" for none

GsapSwbsTab:
  required: [waiting, overTurnaround, bands, rows]
  properties:
    waiting:        { type: integer }     # card Waiting
    overTurnaround: { type: integer }     # card Over 48 hours
    bands: { type: array, items: { key: { enum: [under48h, days2to5, days5to10, over10days] }, label, count } }
    rows:  { type: array, items: { $ref: GsapSwbsRow } }
GsapSwbsRow:
  required: [lineItemId, lineItem, path, owner, companyCode, approvedAt, waitingHours, overTurnaround]
  properties:
    lineItemId:     { type: string }      # L-1006
    lineItem:       { type: string }
    path:           { $ref: GsapPath }
    owner:          { type: string }      # the activity owner's name
    companyCode:    { type: string }      # the activity's Company code option label, "" when not filled in
    approvedAt:     { type: string, format: date-time }
    waitingHours:   { type: integer }     # ≥ 0; the web app shows days with hours under 48
    overTurnaround: { type: boolean }     # waitingHours > 48

GsapWaitingTab:
  required: [waiting, amount, overTurnaround, rows]
  properties:
    waiting:        { type: integer }
    amount:         { $ref: Money, nullable: true }       # plans with finance.view only; null when none
    overTurnaround: { type: integer }     # card Over 2 working days
    rows: { type: array, items: { $ref: GsapWaitingRow } }
GsapWaitingRow:
  required: [purchaseRequestId, name, vendor, owner, path, routedAt, routing, gsapPrNumber, amount, waitingWorkingDays, overTurnaround]
  properties:
    purchaseRequestId:  { type: string }  # PR-1001
    name:               { type: string }
    vendor:             { type: string }
    owner:              { type: string }
    path:               { $ref: GsapPath }  # the activity of its first line
    routedAt:           { type: string, format: date-time }   # the routing GSAP has not answered
    routing:            { enum: [original, modification] }
    gsapPrNumber:       { type: string }  # "" until GSAP answered once
    amount:             { $ref: Money, nullable: true }   # null without finance.view in its plan → "—"
    waitingWorkingDays: { type: integer }
    overTurnaround:     { type: boolean } # waitingWorkingDays > 2

GsapOrdersTab:
  required: [rows, totals]
  properties:
    rows:   { type: array, items: { $ref: GsapOrderRow } }
    totals: { poAmount: Money, invoiced: Money, paid: Money }   # every row, every page
GsapOrderRow:
  required: [lineItemId, lineItem, path, purchaseRequestId, gsapPrNumber, poNumber, poLines, poDate, vendor, poAmount, invoiced, paid]
  properties:
    lineItemId:        { type: string }
    lineItem:          { type: string }
    path:              { $ref: GsapPath }
    purchaseRequestId: { type: string }   # "" when the PO was not raised from a request in MAPP
    gsapPrNumber:      { type: string }
    poNumber:          { type: string }
    poLines:           { type: array, items: { type: string } }   # "10", "20"
    poDate:            { type: string, format: date }             # earliest posting date of its lines
    vendor:            { type: string }
    poAmount:          { $ref: Money }
    invoiced:          { $ref: Money }
    paid:              { $ref: Money }
```

Sort keys: swbs `lineItem`, `plan`, `owner`, `companyCode`, `approvedAt`, `waiting`; waiting `purchaseRequest`,
`vendor`, `owner`, `routedAt`, `amount`, `waiting`; orders `lineItem`, `plan`, `purchaseRequest`, `poNumber`,
`poDate`, `vendor`, `poAmount`, `invoiced`, `paid`. Ties break by ID.

## Export sheets (`measure.Report.Sheets`)

`Sheets(ctx, a, f measure.Filter) ([]measure.WidgetSheet, error)` returns `swbs`, `swbsBands`, `waiting` and
`orders` in that order, each with the columns of its table (amounts as numbers, currency in the heading, "—" as an
empty cell for a hidden amount) and, for `orders`, the totals row. The sort is each tab's default; 039's export
has no report parameters (research R4).

## Go interfaces (`internal/measure/gsaplink/ports.go`)

```go
type Transactor interface {
	Read(ctx context.Context, fn func(ctx context.Context) error) error
}

// Scoped is the frame's step (043 contracts/api.md): nodes in the reports.view scope, matched, and the currency.
type Scoped interface {
	Matched(ctx context.Context, a actor.Actor, f measure.Filter, columns []string) (Matching, error)
	FinancePlans(ctx context.Context, personID string) (map[string]bool, error) // allwork.Choices.FinancePlans
}

// Store is 044's own reads (research R2), keyed by IDs the scoped match returned.
type Store interface {
	ActivitiesApprovedAt(ctx context.Context, activityIDs []string) (map[string]time.Time, error)
	WaitingForGsap(ctx context.Context, lineItemIDs []string) ([]WaitingRequest, error)
	PurchaseOrderTransactions(ctx context.Context, lineItemIDs []string) ([]OrderTransaction, error)
}
```

Proposed to lane A with 043: export the frame's narrow-and-match step (`Scoped`) and say whether saved views keep
report parameters (`tab`, `sort`).
