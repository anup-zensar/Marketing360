# API contract: Approvals Report (041)

For lane A's **wave A contract** (`backend/api/openapi.yaml`), following 039's convention: `GET
/api/v1/reports/<path>` returns the report's widgets and the report's `measure.Report.Sheets` feeds the shared
export, so this report adds **no export operation** of its own. Errors use the one error envelope; every
operation answers 401 when signed out and 403 `forbidden` when the route rule refuses.

## Operation

| Operation | Rule (`access/routes_measure.go`) | Answers |
|---|---|---|
| `GET /api/v1/reports/approvals` | `need(authz.ReportsView)` | 200 `ApprovalsReport`; 400 problems only for a malformed frame filter, `sort`, `order`, `page` or `pageSize` |

Not an error (041-C10): a `statusType`, `recordType`, `outcome` or `approver` that does not exist, or filters
that match nothing — the answer is 200 with empty widgets and the grid's `total` 0 (spec edge cases; the page
shows "Your filters exclude everything"). A report outside the person's reach answers as 039 defines (403 from the rule). Exports go through
`GET /api/v1/reports/approvals/export` of 039 (`need(authz.DataExport)`), widgets `status`, `average`, `approvals`.

## Query

039's report filter (`year`, `plans`, `planTypes`, `teams`, `owners`, `activityTypes`, `activityStatuses`,
attributes, `currency`, `table`) plus:

| Parameter | Meaning | Default |
|---|---|---|
| `statusType` | Record type of the status widget: `activity`, `purchaseRequest`, `modifiedPurchaseRequest` | `activity` |
| `recordType` | Grid record-type filter (same values), repeatable | all |
| `approver` | Grid approver filter only; the widgets are not narrowed by it (041-C9): person IDs, repeatable | all |
| `outcome` | Grid status filter: `waiting`, `approved`, `declined`, `recalled`, `replaced`, `autoApproved`, `notNeeded`, repeatable | all |
| `q` | Free text over the grid's text columns | — |
| `sort`, `order` | Column key (`itemId`, `itemName`, `recordType`, `approver`, `decidedOn`, `outcome`, `submittedOn`, `level`, `purchaseOrder`, `amount`, `owner`, `comments`) · `asc` or `desc` | `submittedOn`, `desc` (ties by `itemId`, then `level`) |
| `page`, `pageSize` | 1-based · 5, 10, 25, 50 or 100 | 1 · 50 |

`period` and `show` are accepted and ignored: nothing is bucketed.

## Schemas

```yaml
ApprovalsReport:
  statusMix: ApprovalStatusMix
  average: ApprovalAverage
  grid: ApprovalGrid
  financeCovers: { plans: int, of: int }   # plans whose amounts the person sees, of the plans with rows
  # purchase requests are in view through their line items (041-C8)

ApprovalStatusMix:
  recordType: string                       # the one shown
  recordTypes: string[]                    # types that have rounds in the plans in view (FR-007)
  total: int
  waiting: { count: int, oldestSince: date-time | null }   # the selected type's; the three types sum to 036's (041-C10)
  items: { outcome: waiting|approved|declined|recalled|replaced|autoApproved, count: int, sharePercent: int }[]

ApprovalAverage:
  days: string | null                      # "2.5"; null → "—", "No decisions yet" (017, one decimal)
  decisions: int                           # rounds behind the figure
  approvers: { person: PersonRef, days: string, decisions: int }[]   # slowest first (041-C7)

ApprovalGrid:
  items: ApprovalRow[]
  total: int
  page: int
  pageSize: int
  choices:
    recordTypes: string[]
    outcomes: string[]                     # the seven words of 041-C1 and 041-C2
    approvers: PersonRef[]                 # people with a row in the plans in view

ApprovalRow:
  itemId: string                           # activity or purchase request ID
  itemName: string
  recordType: activity|purchaseRequest|modifiedPurchaseRequest
  canOpen: boolean                         # false → the name is plain text (041-C5)
  approver: PersonRef | null               # null for an automatic level
  decidedOn: date | null                   # calendar date, UTC
  outcome: waiting|approved|declined|recalled|replaced|autoApproved|notNeeded
  submittedOn: date
  level: int
  levelName: string
  purchaseOrderNumber: string | null       # purchase requests only
  amount: Money | null                     # display currency; null → "—" (no finance.view)
  owner: PersonRef
  comments: string                         # in full; the grid truncates and the page expands
PersonRef: { id: string, name: string }
```

`Money` is the existing `{amount, currency}` with the amount as an exact decimal string, in the display currency
at the current rate (D-303). Dates are `YYYY-MM-DD`, timestamps UTC.

## Sheets for the shared export

| Widget key | Sheet | Columns |
|---|---|---|
| `status` | Current approval status | outcome, count, share % (all three record types, one block each) |
| `average` | Average days to a decision | approver, days, decisions; an "All approvers" first row |
| `approvals` | Approvals | the grid's columns in order, **every** matching row (no paging), currency named in the amount heading |

`Sheets` takes the same query as the operation and returns the three sheets in this order.

## Not changed by this feature

`openapi.yaml`, `routes_measure.go`, `app/measure.go` and `internal/approval` — lane A applies the operation and
the route rule from this file; the report's wiring is the body of `registerApprovals` in
`app/measure_approvals.go`.
