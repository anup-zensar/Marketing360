# Research: Approvals Report (041)

Only the real choices. Everything else follows 039's report pattern and the build guide §5–§6.

## R1 — The average is 017's, per round; the per-approver list is per decision

**Decision**: the overall **Average days to a decision** is `approval.AverageDecisionDays(decided, seconds)` over
the rounds whose outcome is approved or declined and that no rule approved automatically, `seconds` being
`EndedAt − SubmittedAt`, exactly as Approvals › Monitoring (`monitoring.go`) and 036's `Approvals.count`
(`dashboard/panels.go`) do. The **per-approver list** (client 018 FR-019) gives each person the mean of
`decision.At − round.SubmittedAt` over the approved or declined actions they made, through the same function, with
their own count (041-C7).

**Rationale**: 017's rule is per round and elapsed time; SC-002 requires the three screens to agree. My spec's
first draft said "per step, calendar days", which no existing screen computes.

**Alternatives**: a per-step overall average (differs from Monitoring and 036 on multi-level rounds — rejected);
calendar-date differences (two rules for one figure — rejected).

## R2 — Rows come from the approval store's existing reads, through an adapter

**Decision**: `internal/app/measure_approvals.go` implements the report's `Rounds` port with
`stores.ApprovalStore.RoundsOfSubjects` and `ActionsOfSubjects` (one statement each, for the subject IDs of the
plans in view), as 036's dashboard reads rounds. No change to `internal/approval`, no new SQL, so
`queries/measure_approvals.sql` is not created.

**Rationale**: `approval.Service.RoundSummaries` gives round-level facts only (no levels, approvers or comments),
but the rounds carry `Levels[].Approvers` and `Action` carries level, actor, time, comment and capacity — all the
grid needs. The `approval` package is outside this lane's Owns list.

**Alternatives**: widen `RoundSummary` (edits another area — rejected); a new SQL join (a second source of truth
for who may see a round, and a PostgreSQL test per statement — rejected).

## R3 — Engine outcomes to the six words

| Engine | Shown as |
|---|---|
| `open` | Waiting |
| `approved` (not `AutoApproved`) | Approved |
| `approved` with `AutoApproved` | Approved automatically |
| `declined` | Declined |
| `recalled`, `cancelled` | Recalled |
| `superseded`, `restarted` | Replaced by a new submission |

**Rationale**: 041-C6; every round has exactly one word, so the mix adds up. The grid's step outcomes add **Not
needed** (041-C2); a level with `resolution` skipped is not a row (041-C3).

## R4 — Names, owners, amounts and links of the records

**Decision**: an activity's name, owner, planned cost, plan and open-ability come from the year's nodes
(`Planning.Year`, All work's source), which already carry the plan's scope and `finance.view` coverage; a
purchase request's come from the purchase request store's reads for the plans in view (name, owner, amount as
032/036 work it out, purchase order number). An amount of a plan without `finance.view` is `null` (shown "—").
A record outside the actor's plans is never read; a name is a link only when the person may open it (041-C5).

**Purchase requests in view** (041-C8): a purchase request is in view when one of its line items is, so the
subject IDs come from the year's activities plus `Purchases.Summaries(lineItemIDs of the nodes in view)` — the
same rule as 036's `board.subjectIDs` — and every filter of the frame narrows purchase requests through their line
items.

**Open**: `purchaserequest.Summaries` carries no name, owner or PO number. Which existing read to adapt is
settled in the first backend task; if none gives them in one statement, the integrator widens `Summary` (outside
Owns). Listed in plan.md for lane A.

## R5 — Report-specific parameters

**Decision (provisional)**: the approver, outcome, record type, search, sort, page and page size of the grid, and
the record type of the status widget, are query parameters of `GET /api/v1/reports/approvals`, parsed in
041's own handler after 039's `ParseFilter`. They live in the page address (constitution F3). Whether they also
enter a saved view depends on lane A's answer (plan.md, question 1).

**Alternatives**: extending 039's `ReportFilter` (edits the hub, forbidden to lanes — rejected unless lane A
chooses it).
