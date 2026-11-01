# Research: GSAP Link Reports (044)

Only the real choices are recorded here (build guide §8). Each states the decision, why, and what was rejected.
Code references are as of `measure-module` at `03c8f6ab` and `feature/039-measure-hub` at `f5c15d74`.

## R1 — The records in view come from the frame's match, not from a second filter

**Decision**: every tab starts from the year's nodes in the `reports.view` scope matched by `allwork.Match` with
039's `measure.Filter` — the frame's own steps (039 R1, R2), so the frame's **In view** and the report agree. The
SWBS tab is then the matched **line-item nodes** whose `Node.Swbs` is `readyToSend` (All work's own SWBS status,
`allwork/source.go`); Waiting for GSAP and Purchase orders take the matched live line-item IDs into 044's own
reads (R2, R3). As in 043, the step is reached through a `Scoped` port that lane A is asked to export once from
`measure` (043 research R4); until then `app/measure_gsaplink.go` composes it from `measure.Planning.Year`,
`measure.Plans.InReportsScope` and `allwork.Match`.

**Rationale**: SC-002 asks for exactly the line items All work shows as Ready to send; reading the same nodes makes
it true by construction, and the status rule of 030 FR-001 stays in one place. Filters on plan, team, owner,
activity type, status and Shell attributes apply to line items through their activity, as All work does.

**Rejected**: a SQL read of `line_items` with its own status rule (a second copy of 030 FR-001, its own scope rule
and its own filters — three ways to drift).

## R2 — One small query file for what the nodes do not carry

**Decision**: `queries/measure_gsaplink.sql` holds three reads, each keyed by IDs that came out of the scoped
match, one statement each, none per row:

1. `ListActivitiesApprovedAt(activity_ids)` — the latest `approval_rounds.ended_at` with `subject_kind =
   'activity'` and `outcome = 'approved'` per activity (013; 031 renamed the column `subject_id`).
2. `ListPurchaseRequestsWaitingForGsap(line_item_ids)` — purchase requests with a line on one of the line items
   and 032 FR-032's predicate (status `approved`, at least one routing, a line with no PO line or a PO value or
   currency different from the line), with ID, name, vendor, owner, plan, lines (line item, amount, currency),
   GSAP PR number and the time and type of the **latest routing of type `original` or `modification`** — the
   routing GSAP has not answered; `resend` and `close` are skipped (spec FR-009).
3. `ListPurchaseOrderTransactions(line_item_ids)` — the `financial_transactions` of type `purchaseOrder`,
   `invoice` and `payment` linked to the line items, with the purchase request line whose `po_number` and
   `line_item_id` match (for the PR ID and GSAP PR number).

**Rationale**: the nodes carry neither approval times, routings nor transactions. Reading by IDs keeps scope in
one place (R1) and the statements fixed in number (028). The waiting predicate is a copy of 032's SQL (four
lines). Which routing GSAP has not answered (the latest original or modification; Resend never restarts the wait)
is decided in the service from every routing the store returns, so it is unit-tested (review of #64).

**Rejected**: calling `purchaserequest`'s store (its list is scoped by `purchaseRequests.list`, not
`reports.view`, and pages by vendor; changing it is outside lane D's paths); `financials.Store.SpendOnLineItems`
(no invoices; extending it is outside lane D's paths).

## R3 — Purchase orders: rows per line item and PO, sums in Go

**Decision**: the service groups the transactions of R2.3 by line item and PO number: PO amount = its
`purchaseOrder` lines, invoiced = its `invoice` rows (credit notes are negative already, 015 FR-001), paid = its
`payment` rows. Each amount is converted to USD by the transaction's currency and then once to the report
currency at the current rate (D-303), as All work does. Only line items of plans in the person's `finance.view`
scope are read (`allwork.Choices.FinancePlans`, the frame's coverage); the tab states "Covers N of M plans".

**Rationale**: 015's link rules already put every linked invoice and payment on one line item (015 FR-007), so
grouping by (line item, PO number) gives measure-metrics' "Invoiced per PR line" without re-linking. Summing in
Go keeps one currency rule.

**Rejected**: grouping in SQL (amounts in several currencies cannot be summed there without the rate table).

## R4 — One operation, one tab per request; paging and sort in memory

**Decision**: `GET /api/v1/reports/gsap-link` takes 039's filter plus `tab`, `sort`, `page` and `pageSize` and
returns one tab: its cards, its chart (SWBS) and one page of rows with the total count and totals. All rows of
the tab are built, sorted and then paged in the service; `Sheets` uses the same function with every row.
The handler removes its four parameters before handing the query to 039's parser.

**Rationale**: one row builder for screen and export makes SC-005 true by construction; at the 028 volume the
largest tab (purchase orders, at most a few thousand rows of the person's scope) sorts in milliseconds next to the
year read. 039's convention is one `GET /reports/<path>` per report.

**Rejected**: three operations (three copies of the frame handling); SQL paging (the SWBS rows come from nodes,
not SQL, so sorting would be split across two places).

**Hand-off**: `tab`, `sort`, `page` and `pageSize` are report parameters 039's address and saved views do not
know. Lane A is asked whether the canonical query keeps report parameters; until then a saved view opens on the
first tab with the default sort, which the spec allows (FR-001 asks only for the address).

## R5 — Waiting times

**Decision**: worked out in `waiting.go` from `now` (the request's clock, injected): SWBS in hours from the
activity's latest approval, shown as days with hours under two days, flag above 48 hours, bands Under 48 hours, 2–5
days, 5–10 days, Over 10 days; purchase requests in working days (Monday to Friday, dates in UTC as the routing
is stored) from the unanswered routing, flag above 2. Never negative.

**Rationale**: spec FR-006, FR-009 and 044-C5; the server works out every figure (constitution IV).

## R6 — Benchmark inside the package

**Decision**: `BenchmarkGsapLink_SC001` in `gsaplink` runs each tab over an in-memory year at the 028 volume
(50 plans, 5,000 activities, 20,000 line items, 20,000 PR lines, 100,000 transactions) through fakes; the year
read is already measured by 036 and 028. Target: under 300 ms added, keeping SC-001's 2 s.

**Rationale**: `internal/platform/apibench` is not a lane D path (043 R5).
