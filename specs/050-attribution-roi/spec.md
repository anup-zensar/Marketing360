# Feature Specification: Attribution and ROI

**Feature Branch**: `feature/050-attribution-roi`

**Created**: 2026-11-01

**Status**: Draft

**Input**: User description: "Attribution and ROI (lane A, issue #50): Settings › Attribution chooses the model that
credits revenue facts to activities (first touch, last touch, equally distributed, linear time decay, U-shaped — the
five that need no weights); attributed revenue and ROI appear on 048 and 049 with the model and the cost basis named
next to every figure; Use planned cost is a labelled choice, never silent." (feature map row 050)

**References**: Constitution 4.8 — Principles IV (the server works out every figure), VII (audit, versioned records),
IX, X · [measure-module-build.md](../../docs/sdd/measure-module-build.md) row 050 ·
[measure-metrics.md](../../docs/product/measure-metrics.md) §2 (ROI) · feature 046 (performance facts and journey
references, D-370) · 048 Summary and 049 Revenue (where the figures show) · Shell client spec 020 US8, FR-054 – FR-061 ·
Hive9 FA §E7.5.

## Clarifications

### Session 2026-11-01

Answered by the product owner for lane A; the integrator moves them to the decision log at the end of wave B.

- **050-C1** Q: Which models? → A: the **five that need no weights**: first touch, last touch, equally distributed,
  linear time decay and U-shaped. The six weighted models of client 020 FR-056 (last non-direct, interaction
  weighted, half-life time decay, W-shaped, full-path, custom) are a follow-up.
- **050-C2** Q: Is the Attribution view comparing models side by side (client 020 FR-060) part of 050? → A: **not
  now**; it is a follow-up, kept in the list below.

### Follow-ups (not built by 050)

- The Attribution view of client 020 FR-060: pick any model, credited revenue and conversions by activity, the
  unattributed amount, a cross-model comparison and freshness, without changing the designated model (050-C2).
- The six weighted models of client 020 FR-056 with their settings (half-life, touch-type weights) (050-C1).
- Grouping credited revenue by a multi-select attribute split by each activity's performance split (client 020
  FR-061), which needs the Attribution view.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Choose how revenue is credited (Priority: P1)

Alex, an administrator, opens Settings › Attribution. Equally distributed is the designated model. Alex picks
U-shaped; the page says every attributed revenue and ROI figure will be restated and asks for confirmation. After
saving, the activity log shows who changed it and when.

**Why this priority**: every attributed figure depends on one designated model (client 020 FR-058, FR-059).

**Independent Test**: change the model as an administrator, see the confirmation and the log entry; a planner sees the
designated model but cannot change it.

**Acceptance Scenarios**:

1. **Given** a fresh database, **When** Alex opens Settings › Attribution, **Then** Equally distributed is designated
   and the five models are listed, each with one sentence saying how it credits revenue.
2. **Given** Alex picks U-shaped, **When** Alex saves, **Then** a confirmation says every attributed revenue and ROI
   figure will be restated, and only after confirming is the change saved and logged with who and when.
3. **Given** a planner, **When** they open Settings › Attribution, **Then** they see the designated model without a
   Save action, with the reason ("Changing the attribution model needs Manage settings").
4. **Given** two administrators edit at once, **When** the second saves, **Then** it is refused as changed meanwhile.

---

### User Story 2 - Credit revenue to the activities of a journey (Priority: P1)

A conversion worth 1,000 EUR has a journey of three touches: a webinar, an e-mail and an event. Under the designated
model each activity receives its share; a conversion without a journey is credited wholly to its own activity.

**Why this priority**: attributed revenue is the basis of ROI and of 048's and 049's revenue figures (client 020 US8).

**Independent Test**: credit fixed journeys under each model and check every share to the cent.

**Acceptance Scenarios**:

1. **Given** touches on 1, 10 and 20 March and a conversion on 25 March of 1,000.00, **When** the model is first
   touch, last touch, equally distributed, linear time decay or U-shaped, **Then** the credits are (1,000/0/0),
   (0/0/1,000), (333.33/333.33/333.34), (166.66/333.33/500.01) and (400.00/200.00/400.00) as FR-003 defines.
2. **Given** a touch dated after the conversion, **When** revenue is credited, **Then** that touch gets nothing.
3. **Given** a revenue fact without a journey reference, **When** revenue is credited, **Then** its own activity gets
   100% under every model.
4. **Given** a touch that matched no activity, **When** revenue is credited, **Then** its share reports as
   *Unattributed*, and credited plus unattributed equals the conversion's revenue to the cent.

---

### User Story 3 - ROI with its basis named (Priority: P1)

On Summary and Revenue, every ROI figure says which model credited the revenue and which cost it divides by: actual
cost when the plans in view have payments in the year; otherwise "—" unless the person chooses *Use planned cost*,
which is then named next to the figure.

**Why this priority**: an ROI without its basis is misleading (measure-metrics §2; issue #50).

**Independent Test**: work out ROI with and without payments, with and without *Use planned cost*, and check the
label.

**Acceptance Scenarios**:

1. **Given** attributed revenue 1,500 and actual cost 1,000, **When** ROI is shown, **Then** it is 50% with "U-shaped ·
   on actual cost".
2. **Given** no payments in the year, **When** ROI is shown, **Then** it is "—" with "No actual cost yet".
3. **Given** *Use planned cost* is chosen, **When** ROI is shown, **Then** it divides by planned cost and says "on
   planned cost"; the choice lives in the address.

---

### User Story 4 - Change the model and see every figure follow (Priority: P2)

After the model changes, Summary, Revenue and every export show the new model's figures and name, with no stale
figure anywhere.

**Independent Test**: change the model, reopen the reports and an export, and compare.

### Edge Cases

- A conversion whose touches all fall after it: everything is Unattributed.
- A journey of one touch: that touch gets 100% under every model; U-shaped with two touches gives 50% each.
- Revenue in a currency without a rate is not credited and is named in a note (046-C7).
- Touches outside the plan year still take their share; the credit belongs to the conversion's date.
- A cost of zero gives ROI "—", never a division error.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Settings › Attribution MUST show the designated model (seed: Equally distributed) and the five models of
  050-C1, each with a one-sentence explanation in Shell words.
- **FR-002**: Only people with Manage settings MUST be able to change the designated model; the change MUST be
  confirmed after a message that every attributed figure will be restated, versioned (a stale save refused) and
  written to the activity log with who and when (client 020 FR-058, FR-059).
- **FR-003**: Each conversion's revenue MUST be credited to the activities of its journey's touches dated on or before
  the conversion, with weights summing to 100% (client 020 FR-054): first touch — all to the earliest; last touch —
  all to the latest; equally distributed — 1/n each; linear time decay — the k-th touch from the earliest gets k ÷
  (1 + … + n); U-shaped — 50% each to the first and last when n ≤ 2, otherwise 40% to the first, 40% to the last and
  20% split evenly over the others.
- **FR-004**: A revenue fact without a journey reference MUST be credited 100% to its own activity under every model
  (client 020 FR-055).
- **FR-005**: Credits MUST be in minor units; the residual of rounding MUST go to the largest weight, ties to the
  newest touch, so credited plus unattributed equals the conversion's revenue exactly (client 020 FR-057).
- **FR-006**: Touches that matched no activity MUST report as *Unattributed*.
- **FR-007**: ROI MUST be (attributed revenue − cost) ÷ cost as a whole per cent; cost is actual when the plans in view
  have payments in the year, otherwise "—"; *Use planned cost* MUST be a labelled choice in the address (measure-metrics
  §2).
- **FR-008**: Every attributed revenue and ROI figure MUST name the model, and every ROI its cost basis.
- **FR-009**: Attributed revenue MUST be worked out from the facts at the time of reading, so a model change restates
  every figure and export at once.
- **FR-010**: 048 and 049 MUST read attributed revenue and ROI through this feature's functions; 050 defines them once.

### Key Entities

- **Attribution settings**: the designated model; one record, versioned and logged.
- **Journey**: the touches (activity, date) sharing a pseudonymous reference, and the conversions with that reference.
- **Credit**: an activity's share of a conversion's revenue under a model; *Unattributed* when the touch matched none.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: For every model and journey of the fixtures, credited plus unattributed equals the converted revenue to
  the cent.
- **SC-002**: After a model change, no report or export shows a figure of the old model.
- **SC-003**: Every ROI on screen and in exports names its model and cost basis.

## Assumptions

- 046 provides revenue facts with an optional journey reference, and touch rows carrying the same reference with an
  activity and a date (046 column meaning `journey`, 046-C6). If 046 stores journeys differently, 050's reader adapts;
  the crediting rules above stay. To confirm with lane B (open point 050-Q1).
- 048 and 049 (lanes C and B) place the figures on their pages; 050 supplies the figures, their labels and the
  shared ROI parts.
