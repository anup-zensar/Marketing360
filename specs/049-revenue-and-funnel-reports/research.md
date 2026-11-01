# Research: Revenue and Funnel-Stage Reports

Decisions behind plan.md. Sources: the 040 plan-budget report (the template), the 039 frame and registry,
046's facts, 047's engine, client 016 FR-054–FR-073 and measure-metrics §2.

## R1 — One package serving both reports

**Decision**: `backend/internal/measure/revenue` holds both the Revenue and the Funnel-stages report, as
the build plan allots. Each registers its own `measure.Report` (`KeyRevenue`, `KeyFunnel`) and its own
routes; everything beneath — the facts query, the projection, the roll-up, the ranked list — is shared.

**Why**: the two reports differ only in which measure they show. A second package would duplicate the
whole data path to vary one column set, and the build plan names one package and one wiring file.

**Alternatives rejected**: `measure/funnel` as its own package (duplication); one report with a mode
parameter (two catalogue entries already exist with their own names, purposes and saved views).

## R2 — The facts roll-up is new, and set-based

**Decision**: a new query file `queries/measure_revenue.sql` with one statement that takes the activities
in view as an array and returns `(activity_id, stage, web_metric, month, currency, counts, amount_minor,
fact_count)`, filtered by `activity_in_scope(...)` exactly as 046's `ActivityFactTotals` is, and bounded
by the plan year.

**Why**: 046's `ActivityTotals` reads one activity; these reports roll up hundreds. Calling it per
activity would make the statement count grow with the records, which the build guide forbids and SC-003
pins. Returning per-activity rows (not pre-rolled) lets one read serve every level the ranked list shows,
because an entry's total is the sum of its activities'.

**Alternatives rejected**: rolling up in SQL per level (four shapes of the same query, and the ranked
list's drill needs the leaves anyway); reusing 046's store (its interface is per-activity by design, and
widening it would make 046 serve a caller it was not written for).

## R3 — Goals and models come through a port of 049's own

**Decision**: `measure/revenue` declares a `Goals` port returning, for the activities in view, each
activity's goal facts and its plan's model as `projection.Model`. `app/measure_revenue.go` implements it
over the plan, activity and settings services — the same shape as 046's `Figures` port and its adapter.

**Why**: `allwork.Node` already carries each activity's INQ, MQL, revenue, dates and type, but not its
plan's model or its "set later stages myself" flag, and `allwork` belongs to feature 010 — widening it is
outside this lane. A port keeps the report package free of plan and settings, which `depguard` and the
report pattern both want.

## R4 — Project per activity, sum the series

**Decision**: run `projection.Project` once per activity in view and add the resulting month series; then
split at today with `projection.SplitCounts` / `SplitAmounts`, and sum months into quarters afterwards.

**Why**: activities differ in start date, model and velocity, so a projection from summed goals would
place conversions in months no activity reaches. The engine is pure arithmetic with no I/O, so the cost is
a loop over in-memory inputs, not a query per activity.

**Alternatives rejected**: projecting the roll-up (wrong by construction); caching projections (047-C8 —
derived at read time, never stored).

## R5 — Copy 040's ranked list in Go, reuse its React component

**Decision**: `measure/revenue` gets `ranked.go`, `ranked_query.go` and `ranked_groups.go` modelled on
040's, with a figures type of its own and its own sort keys; the web imports 039's generic `RankedList`,
`useRankedState` and `RankedControls` unchanged.

**Why**: 040's Go ranked list is package-private and welded to a money-only `Figures` type, so there is
nothing to import; its rules, however, are the specified behaviour (client 016 FR-063–FR-069) and must be
identical. The React component is already generic over its figures type, which is the half that was built
to be shared.

**Risk noted**: two copies of the ranking rules can drift. The copy keeps 040's function names so a reader
can diff them, and SC-002 pins the behaviour that matters (totals reconcile with the series).

## R6 — The wave B contract

**Decision**: lane B writes the wave B contract in `openapi.yaml` before 048–050 build against it, adding
the four operations this feature serves (`/reports/revenue`, `/reports/revenue/ranked`, `/reports/funnel`,
`/reports/funnel/ranked`) and the route rules that go with them. `routes_measure.go` and `openapi.yaml`
are never-edit files for a lane *except* at this step, which the build plan assigns to B — the same
exception 046 used for the wave B contract's first half.

**Why**: the build order puts the contract before the reports so 048 and 050 can code against it. Nothing
else in this feature touches a shared file.
