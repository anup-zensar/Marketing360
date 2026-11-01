# Research: Projections

Decisions behind plan.md. Sources: client 020 FR-048–FR-053, client 016 FR-033–FR-039, FA §E3.2.6 and
§E7.5.1, MA §5–§6, domain model §4.5, the 002 settings package, the 046 performance package and 039's
measure frame.

## R1 — The projection recurrence is client 020's documented default, computed monthly

**Decision**: one pure function: phase the goal-stage series day-exactly across the activity's months of
the plan year, collapse to months, then for each later stage multiply the previous series by the
transition's rate (half-up per month) and shift it by the velocity in whole months, dropping what lands
past December; revenue = deals × average deal size.

**Why**: Hive9's own algorithm is undocumented (016 FR-036) and 020 FR-048–FR-053 is the agreed default.
Monthly is the grain 046's facts aggregate to (`StageTotals.Month`, "2006-01") and 039's frame sums months
into quarters, so nothing else is needed. Day-exact phasing inside the goal stage keeps partial months
honest (an activity starting the 20th gets a small first month) while the month shift stays exact.

**Alternatives rejected**: a run-rate forecast (016 explicitly does not adopt it); storing projections
(020 FR-052's recalculation rule makes stored values stale the moment anything changes — derived-on-read
is 046's freshness rule applied again); day-grain serving (no consumer asks for days; the in-progress
month carries both series, which is all 016 FR-028 needs).

## R2 — Velocity and allowed stages are model columns; categories are a Settings record

**Decision**: migration 00033 adds `inq_to_mql_velocity_months`, `mql_to_deal_velocity_months`
(integer NOT NULL CHECK 0–24, backfilled 0) and `allowed_goal_stages` (text[] NOT NULL, each element in
inq·mql·revenue, non-empty, backfilled all three) to `models`; `activity_categories` is a new table with
the full record rules plus `activity_category_types` assigning each activity type to at most one category
(PK on the type), and a partial unique index enforcing at most one default category.

**Why**: the two velocities sit beside the two rate columns they qualify (the schema already hard-codes
the two transitions); backfills keep every existing model meaning exactly what it meant. Categories mirror
`activity_types` (feature 002's shape: id, name unique ignoring case, active, audit columns).

**Alternatives rejected**: a per-transition table (two fixed transitions do not need a table — the rates
prove it); velocity as days (no source states a unit; the feature map and glossary say months; 047-C1);
category as an enum CHECK (administrators manage the list at run time, so it is data, not schema).

## R3 — The engine is a pure package with a port, fed by 046

**Decision**: `backend/internal/projection` holds types and the calculation only — no Gin, no SQL, no
clock. 046's Figures port (`performance.GoalFacts`) gains the activity's dates (StartDate, EndDate,
Ongoing), and `performance` calls `projection.Project(in)` to fill the panel's reserved `projected`
fields. 048–050 call the same package through their own adapters.

**Why**: the build plan names `measure/projection (pure)`; a pure package makes SC-001's hand-calculated
table the whole test harness and keeps 048/049 free of HTTP round-trips (constitution: the server is the
source of truth, reports compute server-side). GoalFacts already carries ModelID and PlanYear; it gains
the activity's dates **and the model's own parameters**, so the panel projects without `performance`
reaching into Settings itself. The port is lane B's own file.

**Alternatives rejected**: an HTTP projection endpoint for the reports (server-internal consumers need a
function, not a route); putting the engine inside `performance` (046 is matching and facts; 048–050 would
then import the fact store to compute arithmetic).

## R4 — The panel serves projections; Settings serves configuration; no report page

**Decision**: the only contract changes are the model schema's three new optional fields, the
activity-categories CRUD under Settings, and the panel's projected fields (046 reserved them). Reading
stays under the panel's rules (signed in, scoped in the service); writing stays under `settings.manage`.

**Why**: 047-C11 — the registry entries for summary/revenue/funnel belong to 048/049. No new Table A key:
D-367 reserved nothing for 047 because nothing new is needed.
