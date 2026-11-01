# API contract: Purchase Requests Report (042)

Wave A contract of lane C for `backend/api/openapi.yaml`, following 039's convention: `GET /api/v1/reports/<path>`
returns the report's widgets and the report's `measure.Report.Sheets` feeds the shared export, so this report adds
**no export operation** of its own. One error envelope; 401 signed out; 403 `forbidden` when the route rule refuses.

## Operation

| Operation | Rule (`access/routes_measure.go`) | Answers |
|---|---|---|
| `GET /api/v1/reports/purchase-requests` | `need(authz.ReportsView)` | 200 `PurchaseRequestsReport`; 400 problems only for a malformed frame filter, `sort`, `order`, `page`, `pageSize` or `columns` |

Not an error (042-C6): a `vendor`, `requestOwner` or `status` that matches nothing — the answer is 200 with empty
figures. Exports go through `GET /api/v1/reports/purchase-requests/export` of 039 (`need(authz.DataExport)`),
widgets `cards`, `status`, `byVendor`, `byOwner`, `grid`.

## Query

039's report filter (`year`, `plan`, `planType`, `team`, `owner`, `activityType`, `activityStatus`, `attr`,
`currency`, `table`) plus:

| Parameter | Meaning | Default |
|---|---|---|
| `vendor` | Vendor name as entered; `none` is "Not filled in"; repeatable | all |
| `requestOwner` | Person ID of the purchase request's owner (the frame's `owner` is an activity owner); repeatable | all |
| `status` | `created`, `submitted`, `inReview`, `declined`, `approved`, `closed`; repeatable | all |
| `q` | Free text over the grid's text columns | — |
| `sort`, `order` | `name`, `status`, `amount`, `invoiced`, `created`, `delivery` · `asc` or `desc` | `name`, `asc` |
| `page`, `pageSize` | 1-based · 5, 10, 25, 50 or 100 — vendors, the top level | 1 · 25 |
| `columns` | Comma list of the grid's columns in order: `name`, `id`, `status`, `amount`, `owner`, `created`, `delivery`, `sent`, `poNumber`, `poLine`, `poLineStatus`, `invoiced`, `workflow` | all, in this order |

`period` and `show` are accepted and ignored.

## Schemas

```yaml
PurchaseRequestsReport:
  cards: { purchaseRequests: int, lines: int, vendors: int, committed: Money|null, invoiced: Money|null }
  statusMix:
    total: int
    totalAmount: Money|null
    waitingForGsap: int                       # approved requests sent and waiting for GSAP (036)
    items: { status, count: int, amount: Money|null, sharePercent: int }[]   # the six statuses, in order
  byVendor: CommitmentChart
  byOwner: CommitmentChart
  grid: PurchaseRequestGrid
  financeCovers: { plans: int, of: int }      # plans whose amounts the person sees, of the plans with requests

CommitmentChart:
  months: string[]                            # "2026-03" … each month between the first and last, then "none"
  chartSeries: Series[]                       # the ten largest, then "Other"
  allSeries: Series[]                         # every vendor or owner, largest first (table view, export)
  totals: (Money|null)[]                      # per month
  committed: Money|null                       # the sum; equals the card
Series: { key: string, name: string, values: (Money|null)[] }   # one value per month

PurchaseRequestGrid:
  vendors: VendorNode[]                       # one page of vendors, each with everything below it
  total: int                                  # vendors matching
  page: int
  pageSize: int
  columns: string[]                           # the columns shown, in order
  choices: { vendors: string[], owners: PersonRef[], statuses: string[] }

VendorNode: { vendor: string, label: string, purchaseRequests: int, lines: int, committed: Money|null,
              invoiced: Money|null, requests: RequestNode[] }
RequestNode: { id, name, status, amount: Money|null, invoiced: Money|null, owner: PersonRef, createdOn: date,
               sentOn: date|null, deliveryDate: date|null, poNumber: string|null, workflow: string|null,
               gsapNumber: string|null, waitingForGsap: bool, lines: LineNode[] }
LineNode: { id, lineItem: { id, name }, activity: string|null, amount: Money|null, deliveryDate: date|null,
            sentOn: date|null, poNumber: string|null, poLine: string|null, poLineStatus: string|null,
            invoiced: Money|null }
PersonRef: { id, name }
```

`Money` is the existing `{amount, currency}` in the display currency at the current rate (D-303); `null` shows "—".
Dates are `YYYY-MM-DD`.

## Sheets for the shared export

| Widget key | Sheet | Columns |
|---|---|---|
| `cards` | Summary | card, value |
| `status` | Purchase requests by status | status, count, amount, share % |
| `byVendor` | Commitment by vendor | vendor, then one column per month, total (every vendor) |
| `byOwner` | Commitment by owner | owner, then one column per month, total (every owner) |
| `grid` | Purchase requests | one row per line: vendor, purchase request, line item, then the chosen columns |

## Not changed by this feature

`internal/purchaserequest`, the shared hub files. The route rule, the contract and the `notYetServed` line are added
with the feature itself (lane A delegated T013 for 041 and 042 to lane C).
