# Feature Specification: Revenue and Funnel-Stage Reports

**Feature Branch**: `049-revenue-and-funnel-reports`

**Created**: 2026-11-02

**Status**: Draft

**Input**: Feature 049 of the Measure module (docs/sdd/measure-module-build.md): two reports — **Revenue**
and **Funnel stages** — each with summary cards, a series of actual, projected and goal by period split at
today, and 040's ranked list sorted by revenue or by a stage. Sources: client 016 FR-054–FR-073, client
020 US7, measure-metrics.md §2, the 039 frame, the 040 ranked list, 046's facts and 047's engine.

## Clarifications

- **049-C1** — Q: 046 aggregates facts for **one** activity; these reports roll up across every activity
  in view. → A: **A new set-based query**, `queries/measure_revenue.sql`, returning one row per
  (activity, stage, month, currency) for the activities in scope, filtered by `activity_in_scope` exactly
  as 046's is. One statement whatever the number of activities: calling 046's per-activity totals in a
  loop would grow the statement count with the records, which the build guide forbids.

- **049-C2** — Q: Goals and models in bulk do not exist either — `allwork.Node` carries each activity's
  INQ, MQL, revenue, dates and type, but not its plan's model or its "set later stages myself" flag, and
  widening `allwork` is feature 010's, not ours. → A: **A port of 049's own**, `Goals`, answered by an
  adapter in `app/measure_revenue.go`: for the activities in view it returns each one's goal facts and its
  plan's model. The adapter reads what already exists (plan → model, activity → goals); the port keeps
  `measure/revenue` free of those packages, as 046's Figures port does.

- **049-C3** — Q: Does the projection run per activity or on the summed goals? → A: **Per activity, then
  summed.** Activities differ in start date, model and velocity, so projecting a roll-up from summed goals
  would place conversions in months no activity ever reaches. 047's engine is pure and cheap; the report
  calls it once per activity in view and adds the series month by month.

- **049-C4** — Q: 047 serves months; the frame offers months **and quarters**. → A: **The report sums
  months into quarters**, because 046's facts and 047's series are both month-grained (047-C5). A quarter
  is actual when all three of its months are behind today, in progress when today falls inside it, and
  planned otherwise — the same rule one month down.

- **049-C5** — Q: Which figures does 049 define, and which does it only show? → A: It **defines none**.
  Every figure comes from measure-metrics: actual from 046, projected and "actual or projected" from 047,
  goal per stage from the domain model. Two figures it shows are defined elsewhere and are not yet
  computable here — Performance (048) and ROI (050) — and are handled by 049-C6.

- **049-C6** — Q: The Revenue report is specified with ROI and actual-cost columns, but ROI belongs to
  feature 050, which depends on this one. → A: **The columns exist and every value is "—" with the reason
  "Attribution is not configured yet".** A figure that cannot be worked out states why and never shows a
  zero (measure-metrics). The contract carries the fields from the start, so 050 fills them without a
  contract change and without this report's shape moving under its readers.

- **049-C7** — Q: Client 016 FR-067's **Compare** control exists nowhere in MAPP — feature 040 shipped its
  ranked list without it. → A: **Out of scope for 049**, recorded here and raised with the integrator.
  Building it in one lane would make the Revenue and Funnel ranked lists behave differently from Plan
  budget's, which is the opposite of what a shared list is for. It belongs in 039's shared component when
  a lane is funded to add it to every report at once.

- **049-C8** — Q: FR-072 wants a summary-card strip each person can show, hide, reorder and have
  remembered; measure-metrics §3 records "Fixed cards per report until Shell asks (Q-029)". → A: **Fixed
  cards**, as the project's own decision says. The configurable strip is a product question still open with
  Shell, and it needs per-user layout storage this feature has no table for.

- **049-C9** — Q: Which cards, then? → A: **Revenue**: goal, actual, projected (actual or projected) and
  performance, for revenue. **Funnel stages**: the same four for the selected stage, plus the conversion
  from the previous stage. Both strips respect the filter panel. Cards per **activity category** (FR-070's
  "<stage> <category>" set) follow once a report can group by category — see 049-C10.

- **049-C10** — Q: FR-055 puts an **Activity category** selector in the control bar, but no such parameter
  exists in the frame's filter. → A: **An own parameter of each report**, `category`, parsed by the report
  the way 040 parses its ranked-list parameters, listed in `Params()` so saved views keep it, and defaulting
  to **All categories**. An activity whose type belongs to no category counts under **Uncategorised**, a
  selectable value, so no activity is ever lost from a total (047-C7).

- **049-C11** — Q: Which stages can the funnel report select, and what happens when plans in view use
  models that allow different ones? → A: **The selector lists every stage any model in view allows**, in
  funnel order, defaulting to the first. A plan whose model does not carry the selected stage shows "—",
  and the report states how many plans are uncovered — the frame already carries a "left out" count for
  exactly this kind of statement.

- **049-C12** — Q: Is **deals** a selectable funnel stage? → A: **No.** A model's allowed goal stages are
  INQ, MQL and Revenue (047-C2); deals is an internal step of the waterfall that prices revenue (047's
  FR-016). The selector offers what a goal can be set against.

- **049-C13** — Q: FR-062 puts **conversion from the previous stage** on the funnel's ranked rows, but
  measure-metrics assigns "Conversion rate" to feature 048. → A: **049 shows it, 048 still defines it.**
  The definition is one line — a stage's actual over the previous stage's actual — and 046's activity panel
  already computes it that way. 049 reuses that rule rather than inventing a second one, and the
  measure-metrics row gains 049 to its "Shown by" column when the integrator folds these in.

- **049-C14** — Q: Client 016 FR-056 lists an "actual at monthly FX" row. → A: **Dropped.** measure-metrics
  §3 records it among the Hive9 figures MAPP does not show: everything converts at the current rate
  (D-303, Q-023). The spec says so rather than leaving a reader to wonder.

- **049-C15** — Q: Performance is "(actual + projected) ÷ goal", defined by 048, which has not landed. Can
  049 show it? → A: **Yes.** Unlike ROI it needs nothing 049 does not already hold: the actual and projected
  series are this feature's own output and the goal comes with them. 048 still owns the definition and the
  variance wording; 049 shows the figure and states the same words.

- **049-C16** — Q: Which goal is Performance measured against on a plan row — the plan's own goal or the
  sum of its activities' goals? → A: **The sum of the activities' goals in view**, because the actual and
  projected figures beside it are also the activities' own. Comparing an activity roll-up against a plan
  goal is Goal coverage, which is 048's figure, not this one's.

- **049-C17** — Q: Two reports, one package or two? → A: **One package, `measure/revenue`**, as the build
  plan allots. The two reports share their whole data path — the same facts, the same projection, the same
  ranked list — and differ only in which measure they show. Two packages would duplicate all of it.

- **049-C18** — Q: 040's ranked list is a Go package-private implementation welded to money figures; only
  its React component is generic. → A: **Copy the Go, reuse the React.** `measure/revenue` gets its own
  ranked implementation with a figures type of its own, keeping 040's rules exactly (two-level drill, the
  drill trail, paging across groups, attribute groups inherited from the nearest ancestor, entries without
  the sort figure last). The web imports 039's `RankedList` as it stands.

- **049-C19** — Q: Two things these reports need live in files lane B may not edit: the shared export
  helper hardcodes 040's sort keys, so a `revenue` sort would be dropped silently from an export URL, and
  the shared ranked list has no Expand all / Collapse all (FR-068). → A: **Both are recorded as asks for
  039's lane**, and 049 ships without them: the reports work, the export of a *re-sorted* ranked list falls
  back to the default order until the ask lands, and the spec says so rather than letting a reader assume.

- **049-C20** — Q: What does an empty report say (FR-074)? → A: **The reason, never an empty grid**: no
  goals set · the activities are not approved yet, so their facts are held (046 FR-020) · no performance
  data has been loaded · no plan in view uses a model with this stage. The backend states which, as 046's
  panel does with its reasons, because only the server knows which one is true.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Revenue against goal (Priority: P1)

A marketing manager opens Reports › Revenue, picks a year and a period, and sees revenue for the plans in
view: cards for goal, actual, projected and performance; a series by month or quarter that is actual up to
today and projected after it; and beneath it the ranked list of plans, campaigns, initiatives or activities
behind those numbers, sorted by actual or projected revenue.

**Why this priority**: it is the report the module exists to answer — what revenue did the plan expect,
what has arrived, and where is it coming from.

**Independent Test**: with seeded facts and goals, open the report and check the cards, the series split at
today and the ranked list against hand-worked figures; no other report needed.

**Acceptance Scenarios**:

1. **Given** plans with revenue goals and loaded facts, **When** the report is opened for the plan year,
   **Then** each month before this one shows actual revenue, each month after shows projected, the month in
   progress shows both, and the card total is actual to the end of last month plus projected from today.
2. **Given** the same report, **When** the ranked list is sorted by projected revenue and grouped by plan,
   **Then** entries are ordered by that figure within each group, each group shows its own total, and each
   entry lists the level below it.
3. **Given** a plan whose activities have no goal, **When** the report is opened, **Then** its projected
   figures show "—" with the reason, never 0.
4. **Given** attribution is not configured, **When** the ROI and actual-cost columns are read, **Then**
   every value is "—" with the reason "Attribution is not configured yet" (049-C6).
5. **Given** a viewer whose scope covers one team, **When** they open the report, **Then** only their teams'
   plans are counted anywhere in it, and the frame states how many records were left out.

---

### User Story 2 - One funnel stage against goal (Priority: P1)

The same manager opens Reports › Funnel stages and picks a stage — INQ or MQL — and sees that stage's goal,
actual and projected volumes by period, its conversion from the stage before it, and the ranked list
sorted by that stage's volume, so they can see where the funnel leaks.

**Why this priority**: the funnel is the other half of the same question and shares the whole data path;
shipping one without the other would leave the module answering revenue with no way to see why.

**Independent Test**: pick each stage in turn and check the cards, series and ranked rows against the same
seeded data.

**Acceptance Scenarios**:

1. **Given** models that allow INQ and MQL goals, **When** the report is opened, **Then** the stage
   selector lists those stages in funnel order and opens on the first.
2. **Given** a stage is selected, **When** the ranked list is read, **Then** each entry shows that stage's
   goal, actual, projected and performance, and its conversion from the previous stage.
3. **Given** plans in view whose models do not allow the selected stage, **Then** those rows show "—" and
   the report states how many plans are uncovered (049-C11).
4. **Given** the previous stage has no actual, **When** conversion is worked out, **Then** it shows "—",
   never 0 % and never a division by zero.

---

### User Story 3 - Grouping by activity category (Priority: P2)

A planner narrows either report to one activity category — the groupings an administrator set up in 047 —
to compare, say, Asset against Promotion, and the cards, series and ranked list all follow the choice.

**Why this priority**: the reports are useful without it; the category selector makes them comparable, and
it is the control 048's cards will reuse.

**Independent Test**: set up two categories, assign types, and check that choosing one narrows every part
of the report and that *Uncategorised* is selectable and loses nothing.

**Acceptance Scenarios**:

1. **Given** categories exist, **When** one is chosen, **Then** only activities whose type belongs to it
   are counted in the cards, the series and the ranked list.
2. **Given** an activity whose type is in no category, **When** *Uncategorised* is chosen, **Then** it is
   counted there, and with *All categories* every activity is counted exactly once.
3. **Given** a category is chosen, **When** the view is saved, **Then** reopening the saved view restores
   the category (049-C10).

---

### Edge Cases

- No facts have been loaded at all: actual is empty and the series is projection alone, with the reason
  stated (049-C20).
- Every activity in view is unapproved: its facts are held and count nowhere (046 FR-020); the report says
  so rather than showing zeros.
- Plans in view use different models: each activity projects through its own plan's model (049-C3); the
  stage selector is the union of what those models allow (049-C11).
- An activity spans two plan years: its facts count in the year of each fact's date (046 FR-015), so a
  year's report never borrows another year's.
- The ranked list is drilled into a record that the filter then hides: the list falls back to the top level
  and says so, exactly as 040's does.
- Revenue facts in a currency with no rate: excluded from the converted total and counted, as 046-C7 has
  it, with the count shown.

## Requirements *(mandatory)*

**Both reports** *(client 016 FR-054, FR-063–FR-069, FR-074)*

- **FR-001**: Each report MUST plug into the 039 frame: its filter, year, period, show level, currency,
  saved views and whole-report export, with no filter parsing of its own beyond its own parameters.
- **FR-002**: Each report MUST serve its figures for the **plans in the actor's scope only**, filtered in
  SQL, with a test proving a record outside the scope is invisible.
- **FR-003**: Each report MUST offer its figures by **month or quarter**, a quarter being the sum of its
  months (049-C4).
- **FR-004**: Every series MUST be split at today: actual to the end of the last whole period, projected
  from today, the period in progress carrying both and counted once (047 FR-010).
- **FR-005**: Each report MUST carry 040's ranked list with the same rules — two levels at once, drill with
  a trail and Back, grouping with group totals, paging across groups, name search, and a link that opens
  the record in the Plan workspace (049-C18).
- **FR-006**: The ranked list MUST be sortable by the report's own figures: actual and projected revenue on
  Revenue, the selected stage's actual and projected volume on Funnel stages.
- **FR-007**: Every widget MUST offer a table view and an export, per the module's report pattern.
- **FR-008**: An empty report MUST state **why** it is empty, from the causes in 049-C20, and MUST NOT show
  a grid of zeros.
- **FR-009**: Figures the feature cannot work out MUST show "—" with a reason — never 0, never a blank
  (measure-metrics).

**Revenue** *(client 016 FR-054–FR-058)*

- **FR-010**: The Revenue report MUST show cards for revenue goal, actual, projected and performance
  (049-C9).
- **FR-011**: It MUST show revenue by period — actual, projected and goal — with a table view of the same.
- **FR-012**: Each ranked entry MUST show its name, level, owner, revenue goal, actual, projected,
  performance, actual cost and ROI.
- **FR-013**: Actual cost and ROI MUST show "—" with the reason "Attribution is not configured yet" until
  feature 050 lands (049-C6); the contract MUST carry the fields from the start.
- **FR-014**: An information tip MUST explain, in Shell's words, that the plan's model defines the stages,
  rates and velocity, how a goal projects the later stages within the plan year, and what the list compares.

**Funnel stages** *(client 016 FR-059–FR-062)*

- **FR-015**: The report MUST offer a stage selector listing every stage the models in view allow, in
  funnel order, defaulting to the first (049-C11, 049-C12).
- **FR-016**: It MUST show the selected stage's goal, actual and projected by period, named from the stage.
- **FR-017**: Each ranked entry MUST show the stage's goal, actual, projected, performance and its
  **conversion from the previous stage**; conversion MUST show "—" when the previous stage has no actual
  (049-C13).
- **FR-018**: A plan whose model does not allow the selected stage MUST show "—", and the report MUST state
  how many plans are uncovered (049-C11).

**Activity categories** *(client 016 FR-055, FR-073)*

- **FR-019**: Both reports MUST take an **activity category** parameter defaulting to all categories, kept
  by saved views, which narrows every part of the report (049-C10).
- **FR-020**: An activity whose type is in no category MUST be countable as **Uncategorised**, so that with
  all categories chosen every activity is counted exactly once.

**The figures** *(measure-metrics §2)*

- **FR-021**: Actual MUST be 046's counted facts for the activities in view, by the fact's date — held,
  unmatched and replaced facts count nowhere.
- **FR-022**: Projected MUST come from 047's engine, run **per activity** and summed (049-C3), with the
  model named beside the figure as 047 requires.
- **FR-023**: Goal per stage MUST be the sum of the activities' goals in view, and Performance MUST be
  (actual + projected) ÷ goal as a whole per cent with 048's variance wording (049-C15, 049-C16).
- **FR-024**: Money MUST be shown in the frame's display currency, converted once at the current rate, with
  facts in a currency that has no rate excluded and counted (046-C7, D-303).
- **FR-025**: Reading these reports MUST need no new permission: the frame's rule governs, and the scope is
  applied in the service.

### Key Entities

- **Report figures (derived)**: per entry and period — goal, actual, projected, performance, and for the
  funnel the conversion from the previous stage. Nothing is stored; every figure is worked out per request
  from 046's facts and 047's engine.
- **Activity category selection**: a report parameter, not a record; the categories themselves are 047's.

## Success Criteria *(mandatory)*

- **SC-001**: With seeded goals and facts, both reports match a hand-worked table for one plan, one
  campaign and one activity, including the split at today.
- **SC-002**: The totals of the ranked list's top-level entries, with no grouping, equal the series totals
  for the same filter (client 016 FR-069).
- **SC-003**: The number of SQL statements a report runs MUST NOT grow with the number of activities,
  plans or periods in view; a test pins it as 046's does.
- **SC-004**: Every scoped read has an isolation test; a viewer sees no plan outside their teams in any
  figure, card or row.
- **SC-005**: Every rule above has a test naming its FR, and both P1 stories have an end-to-end browser
  test with axe.
- **SC-006**: A reviewer can reproduce any projected figure from what is served beside it: the goal, the
  model and its last change (047 FR-012).

## Out of scope

- **Compare** (client 016 FR-067) — belongs to the shared ranked list, recorded as an ask (049-C7).
- **Expand all / Collapse all** and a re-sorted export URL — both live in 039's shared files (049-C19).
- A **configurable card strip** (FR-072) — fixed cards until Shell answers Q-029 (049-C8).
- **Actual at monthly FX** (FR-056) — MAPP converts at the current rate everywhere (049-C14).
- **ROI's definition** (050), **Performance's definition** and **Goal coverage** (048) — shown here where
  possible, defined there.
- The **Summary report** (048) and **attribution settings** (050).
