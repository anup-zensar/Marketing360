# Feature Specification: Projections

**Feature Branch**: `047-projections`

**Created**: 2026-11-01

**Status**: Draft

**Input**: Feature 047 of the Measure module (docs/sdd/measure-module-build.md): velocity and allowed goal
stages on models, activity categories in Settings, and projected INQ, MQL and revenue per period from the
model's rates — the projection engine features 048, 049 and 050 read. Sources: client 020 FR-048–FR-053 and
US7, client 016 FR-033–FR-039, FA §E3.2.6 and §E7.5.1, MA §5–§6, domain model §4.5, measure-metrics.md §2.

## Clarifications

- **047-C1** — Q: No source states velocity's unit (FA says only "the expected time lag between stages").
  → A: **Whole months per transition**, as the feature map and glossary already say: one integer for
  INQ→MQL and one for MQL→deal, 0–24, zero meaning the conversion lands in the same month. Client 020
  FR-053's day grain concerns how a phased series is *read*, not the lag: the goal stage phases evenly
  across its days (FR-049) and a month lag shifts that series by whole calendar months, so day totals stay
  exact under the shift.

- **047-C2** — Q: Where do velocity and the allowed stages live? → A: **On the model record**, like the two
  conversion rates they belong with: migration 00033 adds `inq_to_mql_velocity_months` and
  `mql_to_deal_velocity_months` (integer, NOT NULL, CHECK 0–24, backfilled 0 = Hive9's observable
  same-month behaviour) and `allowed_goal_stages` (text[], CHECK ⊆ {inq, mql, revenue}, NOT NULL,
  backfilled to all three so nothing existing changes meaning). The contract adds them as optional inputs
  with those defaults, so existing clients keep working.

- **047-C3** — Q: FA says a goal "can only be entered at an allowed stage", but goal entry lives on Plan
  module screens other lanes own. → A: **047 stores, validates and serves the configuration; enforcement at
  goal entry is a recorded cross-folder ask** (like 046's FR-036), with one exception 047 does own: the
  model-edit impact check warns when narrowing the allowed stages strands existing goals. An existing goal
  at a now-disallowed stage stays valid and keeps projecting — configuration guides new entries, it never
  silently rewrites old ones.

- **047-C4** — Q: Hive9's projection algorithm is undocumented (016 FR-036 "still open with the vendor").
  → A: **MAPP adopts the documented default as written in client 020 FR-048–FR-053** and records it here as
  the single definition: the goal stage follows its goal phased evenly across the activity's months (an
  evergreen goal runs to plan-year end); each later stage is the previous stage's series × that
  transition's conversion rate, shifted forward by its velocity, rounded half-up per month; revenue is the
  deal series × average deal size. Nothing projects past the plan year: a conversion shifted beyond it is
  dropped, not bunched into December. SC-001's hand-calculated table is the sign-off artefact.

- **047-C5** — Q: What grain does the engine serve? → A: **Months** ("2026-01" … "2026-12"), the grain 046's
  facts already aggregate to; quarters are sums of months, which the 039 frame already does. The month
  containing today is *in progress*: it carries both its actual-to-date and its projected value (016
  FR-028); whole months before it are actual only, months after it projected only.

- **047-C6** — Q: How do "Set later stages myself" activities project? → A: **Each typed stage phases its
  own goal independently** — the typed values replace the rate arithmetic (that is what the override
  means), velocity shifts nothing, and revenue uses the typed revenue goal. A calculated activity projects
  from INQ through the model exactly as 047-C4 says.

- **047-C7** — Q: What is an activity category, exactly? → A: **A Settings record grouping activity
  types** (Hive9 "Tactic Category", FA L89): own table `activity_categories` (00033), name unique ignoring
  case, active flag, full record rules. **An activity type belongs to at most one category**; types in no
  category count under the implicit *Uncategorised* bucket, so report totals never lose activities
  (016 FR-022 counts "only activities whose type belongs to it" — Uncategorised is a selectable bucket, not
  a gap). **At most one category is the default** (the organisation setting 016 FR-022's selector opens
  on); with none marked, reports open on the alphabetically first active category.

- **047-C8** — Q: 020 FR-052 wants projections recalculated whenever goals, dates, model parameters or
  facts change, stamped with the model and its rates' last change. → A: **Projections are derived at read
  time and never stored** — the same rule as 046's freshness (a stored value would go stale by itself) —
  so recalculation is automatic; the stamp is the model's name and its record's `updatedAt`, served with
  every projection. A separate "parameters changed" timestamp was rejected: it needs a second column that
  every writer must remember to set, and an `updatedAt` that moves on a rename only ever over-reports.

- **047-C9** — Q: 016 FR-026 wants period boundaries in the acting user's time zone. → A: **The engine uses
  the server's injected clock, UTC calendar dates** (constitution: timestamps UTC, deterministic clock).
  Where the marker is *drawn* is the report frame's concern (D-368, feature 039); a per-user time zone is
  that frame's future refinement, not a projection input.

- **047-C10** — Q: 016 FR-035 credits actuals through the designated attribution model — feature 050, which
  lands after 047. → A: **Until 050, actuals are 046's counted facts as they stand** (held, unmatched and
  replaced facts excluded; revenue rows without a rate excluded and counted, 046-C7). 050 swaps the actual
  series' source, not the recurrence.

- **047-C11** — Q: Does 047 ship a report page? → A: **No.** The measure registry's summary, revenue and
  funnel entries belong to 048 and 049. 047 delivers the Settings records, the pure projection engine, its
  figures on 046's activity panel (whose contract reserved `projected` for 047), and the Go port the
  reports call — no new report route.

- **047-C12** — Q: Can a category be deleted while it groups activity types? → A: **Yes, and it is never
  refused.** FR-021 says deleting releases its types, so unlike an activity type there is no `in_use`
  refusal: the category's usage line states what it groups before the delete, and Undo brings both the
  category and its assignments back. A type released this way is *Uncategorised*, which is a bucket the
  reports can select, not a gap (047-C7).

- **047-C13** — Q: Moving a type between categories changes two categories — does the one losing it get a
  version bump and a log entry? → A: **No.** The assignment lives in its own table keyed by the type, so
  the losing category's list is derived on read; only the category being edited is versioned and logged.
  One change, one entry (constitution, every new record type).

- **047-C14** — Q: What happens when a category that was the default is restored while another is the
  default now? → A: **It returns not the default, with a note saying so.** The schema allows one default
  and the record in hand is the newer decision; silently moving the mark would undo a choice nobody made
  here. The same rule covers Undo and Restore.

- **047-C15** — Q: Does a category need its own activate/deactivate operation, as activity types have? →
  A: **No.** `active` is a field of the category's PATCH, so the five CRUD operations are the whole
  surface; nothing else in Settings needs a sixth.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Velocity and allowed goal stages on the model (Priority: P1)

An administrator opens Settings › Models and gives a model its velocity — how many months a conversion
takes from INQ to MQL and from MQL to deal — and ticks which stages a goal may be set against. The live
waterfall example shows where the lag places each stage. Saving warns when narrowing the allowed stages
strands an existing goal.

**Why this priority**: every projected number is placed in time by these fields; without them the engine
has no lag and no starting rule, and 048/049 cannot land.

**Independent Test**: edit a model's velocity and allowed stages in Settings and see the saved values,
the audit entry, the version check and the impact warning — no projection needed yet.

**Acceptance Scenarios**:

1. **Given** a model with rates 2 % and 5 %, **When** the administrator sets INQ→MQL velocity 1 and
   MQL→deal velocity 2 and saves, **Then** the model shows both lags, the change is audited, and a stale
   version is refused naming who saved last (version_conflict).
2. **Given** velocity typed as -1, 25 or 1.5, **When** saving, **Then** every fault is reported together:
   velocity is a whole number of months from 0 to 24.
3. **Given** a model whose allowed stages are INQ · MQL · Revenue, **When** the administrator unticks MQL
   while a plan goal is set at MQL, **Then** the impact check names that goal before saving, and after the
   save the existing goal stands and keeps projecting.
4. **Given** every stage unticked, **When** saving, **Then** the save is refused: at least one stage must
   allow a goal.

---

### User Story 2 - Projected INQ, MQL and revenue per month (Priority: P1)

A marketer opens an approved activity with a goal and sees, beside each stage's goal and actual, its
projected value for the plan year, and the months behind it: actual up to today, projected from today, the
current month in progress with both. Every projected figure names the model and when its parameters last
changed, so a reviewer can reproduce it.

**Why this priority**: this is the feature — the engine 048 and 049 consume and the panel's reserved
`projected` field.

**Independent Test**: with a goal, a model and a fixed clock, the panel's projected figures match a
hand-calculated table (SC-001) without any report page.

**Acceptance Scenarios**:

1. **Given** an activity goal of 365,000 INQ across the whole of 2026 — 1,000 a day — a model with rates
   2 % and 5 %, velocities 1 and 2 and deal size 10,000 USD, **When** the panel is read with today fixed
   mid-year, **Then** each month carries its own day count × 1,000 (January 31,000, February 28,000 …),
   MQL is 2 % of each month placed one month later, deals are 5 % of each MQL month placed two further
   months on, revenue = deals × 10,000 USD, and the conversions shifted past December are dropped.
2. **Given** the same activity, **When** today's month is read, **Then** that month carries both its
   actual-to-date and its projected value, months before it only actual, months after it only projected.
3. **Given** an activity with "Set later stages myself" and typed INQ, MQL and revenue, **When** the panel
   is read, **Then** each stage projects its own typed goal phased over the activity's months and the
   model's rates change nothing.
4. **Given** an activity without a goal, **When** the panel is read, **Then** projected shows "—" with the
   reason *no goal* — never 0 (measure-metrics: a figure that cannot be computed states why).
5. **Given** a projected figure, **When** it is shown, **Then** the model's name and the date its
   parameters last changed are served with it (020 FR-052).
6. **Given** facts and goals of another team's activity, **When** a viewer outside its scope asks,
   **Then** the activity answers as missing (constitution 7; same rule as 046's panel).

---

### User Story 3 - Activity categories in Settings (Priority: P2)

An administrator opens Settings › Activity categories, creates the categories the reports group by
(Asset, Promotion, …), assigns each activity type to at most one and marks one category as the default the
reports open on. Deleting goes to the recycle bin; a category in use warns first.

**Why this priority**: 048's cards and 049's summaries group by category; the list can land after the
engine because reports (048/049) are the only consumers.

**Independent Test**: full CRUD on the Settings page with the record rules — audit, bin, undo, version
check — plus the type assignment, without any report.

**Acceptance Scenarios**:

1. **Given** no categories, **When** the administrator creates "Asset" and assigns two activity types,
   **Then** it lists with its type count, the change is audited, and creating "asset" again is refused as
   a duplicate ignoring case.
2. **Given** a type already in "Asset", **When** it is assigned to "Promotion", **Then** it moves — a type
   belongs to at most one category — and both categories' counts update.
3. **Given** "Asset" is marked default, **When** "Promotion" is marked default, **Then** Asset loses the
   mark: at most one default exists.
4. **Given** a category, **When** it is deleted, **Then** it goes to the recycle bin with Undo, its types
   become uncategorised, and restoring it restores the assignments it held.
5. **Given** a viewer without `settings.manage`, **When** they open the page, **Then** they can read it
   with the standard view-only notice and every change is refused.

---

### Edge Cases

- An activity has no start date or is evergreen: phasing starts at the plan year's start (or the
  activity's start when inside the year) and runs to plan-year end (016 FR-034, 020 FR-049).
- A velocity shift lands a conversion after December: it is dropped, never bunched into the last month —
  nothing projects past the plan year (020 FR-051).
- The goal stage is MQL or Revenue (plan goals work backwards): projection still runs **forward only**
  from the goal stage (020 FR-050); earlier stages show no projection.
- Rates or deal size change mid-year: every projection reflects the new parameters at once (047-C8) and
  the served stamp moves.
- Today is before the activity starts: everything is projected; today after the plan year: everything is
  actual and projected is zero-length, with the panel figure equal to the actual.
- A model is deactivated: its plans' projections still compute (the model's parameters remain), exactly as
  goals do today.
- The plan year has no facts at all: actual months are empty, projection follows the phased goal alone.

## Requirements *(mandatory)*

**Model velocity and allowed goal stages** *(FA §E7.5.1; 016 FR-034)*

- **FR-001**: A model MUST carry one whole-month velocity per transition — INQ→MQL and MQL→deal — each 0
  to 24, validated with every fault reported at once (047-C1, 047-C2).
- **FR-002**: A model MUST carry its allowed goal stages, a non-empty subset of INQ · MQL · Revenue,
  defaulting to all three (047-C2).
- **FR-003**: Editing velocity or allowed stages MUST follow the model's existing record rules: one audit
  entry per change, version_conflict on a stale save, and the pre-save impact check naming the goals a
  narrowed stage set strands (047-C3).
- **FR-004**: The served model MUST expose velocity and allowed stages wherever the model is read today,
  and the Settings form MUST show the lag's effect in the live waterfall example.
- **FR-005**: Enforcing allowed stages at goal entry is the Plan module's screens' work and MUST be
  recorded as a cross-folder ask; 047 MUST NOT rewrite or invalidate existing goals (047-C3).

**The projection** *(client 020 FR-048–FR-053, 016 FR-033–FR-037; domain model §4.5)*

- **FR-006**: The goal-stage series MUST phase the activity's goal **day-proportionally** across the
  activity's days inside the plan year, collapsed to months, so a part month carries only its own days; an
  evergreen or undated activity phases to plan-year end (047-C4, edge cases).
- **FR-007**: Each later stage MUST be the previous stage's series × that transition's conversion rate,
  shifted forward by its velocity in whole months, rounded half-up per month; projected revenue MUST be
  the deal series × average deal size in USD (047-C4).
- **FR-008**: Projection MUST run forward from the goal stage only, and MUST NOT produce anything past the
  plan year's end — shifted conversions landing beyond it are dropped (020 FR-050, FR-051).
- **FR-009**: A "Set later stages myself" activity MUST project each typed stage from its own phased goal,
  ignoring the rates and velocity (047-C6).
- **FR-010**: The plan-year figure MUST be actual to the end of the last whole month before today plus
  projected from today, and that single figure — not the from-today part alone — is what the panel serves
  as `projected`, so a closed plan year shows the actual and nothing is ever added twice. The month
  containing today MUST carry both its actual and its projected value (047-C5, 016 FR-028/FR-037).
- **FR-011**: Projections MUST be derived at read time, never stored, so every change to goals, dates,
  model parameters or facts is reflected at once (047-C8, 020 FR-052).
- **FR-012**: Every projected figure MUST be served with the model's name and the date **the model was
  last changed** (047-C8, 020 FR-052, SC-011 of client 020). The record's own `updatedAt` is that date: it
  moves on any edit to the model, which is the honest conservative answer — it never says "unchanged" when
  a rate did change.
- **FR-013**: An activity without a goal MUST show no projection, with the reason *no goal* — never a zero
  (016 FR-036, measure-metrics header). That reason MUST be carried separately from the reason a figure
  has no **actual**: a stage can have facts and still no goal, and an actual must never swallow the
  explanation of a missing projection.
- **FR-014**: Actual inputs MUST be 046's counted facts — held, unmatched and replaced facts excluded;
  revenue facts without a rate excluded and counted (047-C10, 046-C7) — until feature 050 swaps the source.
- **FR-015**: The engine MUST live in a pure package with no HTTP, SQL or clock of its own: inputs in,
  series out, so 048/049/050 call it as a port and SC-001's table drives it directly.
- **FR-016**: 046's activity panel MUST gain the projected figures its contract reserved: a projected
  value for each stage **the panel shows** (INQ and MQL), projected revenue, and the months split per
  FR-010, under 046's existing scope and approval rules (047-C11). Deals stay an internal stage of the
  waterfall, observable through revenue and through the Go port that 048–050 call.
- **FR-017**: Today MUST come from the injected clock as a UTC calendar date (047-C9).

**Activity categories** *(016 FR-003, FR-022; FA L89, L2622)*

- **FR-018**: Administrators MUST manage activity categories in Settings: name 1–80 unique ignoring case,
  active flag, created like every record type (audit, version, recycle bin with Undo, history kind)
  (047-C7).
- **FR-019**: An activity type MUST belong to at most one category; assigning it elsewhere moves it; types
  in none count as *Uncategorised*, a bucket the reports can select so no activity is ever lost (047-C7).
- **FR-020**: At most one category MUST be the default, and marking a new default MUST unmark the old in
  the same change (047-C7). What a report opens on when no category is marked is that report's rule
  (048/049), not this feature's.
- **FR-021**: Deleting a category MUST move it to the recycle bin, release its types to Uncategorised and
  restore its assignments on restore; a category's deletion MUST warn when the category is in use.
- **FR-022**: The categories, and each activity type's category, MUST be **served** so the reports and the
  activity-type Settings page can read them; how either chooses to show it is theirs.

**Access and records** *(constitution 7; 014 Table A)*

- **FR-023**: Editing a model's new fields and managing categories MUST require `settings.manage`; reading
  projections MUST follow the activity panel's existing rule — signed in, scope applied in the service, a
  record outside the scope answering as missing. No new permission keys (Table A unchanged).
- **FR-024**: Every operation MUST have a route rule and fail closed; every changed record type MUST keep
  one activity-log entry per change inside the same transaction.

### Key Entities

- **Model (extended)**: the feature-002 record gains two velocities (whole months, 0–24) and its allowed
  goal stages (non-empty subset of INQ · MQL · Revenue). Everything else — rates, deal size, activity
  types, record rules — stands.
- **Activity category**: a named grouping of activity types for the Measure reports — name, active,
  default flag (at most one), member types (each type in at most one category), full record rules.
- **Projection**: a derived, never-stored series per activity and stage — month, projected value — plus
  projected revenue, the split at today and the model stamp. Not a record: no identity, no audit.

## Success Criteria *(mandatory)*

- **SC-001**: A table-driven test holds hand-calculated projections (the domain model §4.5 example
  extended with velocities 1 and 2) and the engine matches every row exactly — the sign-off artefact for
  047-C4 (client 020 SC-006).
- **SC-002**: With a fixed clock, reading the same activity twice gives identical projections; changing a
  rate, a goal or today changes them with no other action (FR-011).
- **SC-003**: The goal stage's months sum to the goal exactly. A later stage's months sum to the earlier
  stage's kept months at its rate, **within the per-month rounding** (each month is rounded half-up on its
  own, so the total may differ from rounding the sum by less than one unit per month), with the
  conversions shifted past December dropped — those are exactly the shifted tail (FR-007, FR-008).
- **SC-004**: The activity panel with projections answers within the panel's existing budget (046 SC-008's
  one read, no per-month queries) — the engine adds arithmetic, not queries.
- **SC-005**: Every rule above has a test naming its FR, and the engine gets a property test driven by Go's
  fuzzer over any goal, any dates and any velocities: twelve months, no negative month, nothing above the
  goal, and the goal stage summing to the goal exactly (SC-003).
- **SC-006**: A reviewer can reproduce any projected figure from what is served with it: the goal, the
  model's parameters, their last-changed date and this spec's rule (020 SC-011).

## Out of scope

- Report pages, cards and league tables (048, 049); attribution and ROI (050); the today-marker rendering
  and per-user time zones (039's frame, D-368); enforcement of allowed stages at goal entry (Plan module —
  recorded ask); custom goals (none exist in MAPP); category-level visibility (Hive9's AdminOnly hint —
  not adopted); day-grain serving (months served; quarters summed by the frame).
