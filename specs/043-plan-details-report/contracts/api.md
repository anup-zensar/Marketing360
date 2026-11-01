# API Contract: Plan Details Report (043)

For the **wave A contract**, written into `backend/api/openapi.yaml` (tag **Reports**) by PR #57; `openapi.yaml`
is the reference where the two differ. It follows
039's convention (`specs/039-measure-foundation/contracts/api.md`, "The convention every report follows"):
`GET /api/v1/reports/<path>` with the report filter returns the report's widgets, and the report's
`measure.Report.Sheets` feeds 039's shared export, so 043 adds **no export operation of its own**. Errors use
the one envelope; 401 when signed out, 403 `forbidden` when the route rule refuses.

## Operations

| Operation | Query | Answer | Rule (`access/routes_measure.go`) |
|---|---|---|---|
| `GET /api/v1/reports/plan-details` (`getPlanDetailsReport`) | 039's report filter (data-model §2): `year`, `plan`, `planType`, `team`, `owner`, `activityType`, `activityStatus`, `attr`, `period`, `show`, `currency`, `table` | 200 `PlanDetailsReport`; 400 problems for a malformed filter, as the frame | `need(authz.ReportsView)` |

`show` is the level shown (`plan`, `campaign`, `initiative`, `activity`); **its default for this report is
`activity`** (039 data-model §2, "per report"). `period` sets the start periods of **Activities by start
month**; `currency` the currency of every planned-cost figure; `table` is read only by the web app.

**Export** is 039's `GET /api/v1/reports/plan-details/export` (`need(authz.DataExport)`): without `widget`, one
Excel sheet per widget of the level `show` names, then 039's **In view** and **About this export** sheets; with
`widget` (a `PlanDetailsWidget.key` of that level), that one widget as CSV or Excel. An unknown `widget` is 039's
400. An empty widget gives its header and total row.

The report counts only the plans in the person's `reports.view` scope, as the frame does (039 research R2); an
out-of-scope plan named in `plan` is left out and counted by the frame, never named. Planned cost needs no
`finance.view` (D-237), so the report adds no rule beyond `reports.view` (039 data-model §3). The empty states
("No plans you may see for <year>", "Your filters exclude everything", a widget's "No data yet") are rendered by
039's `ReportFrame` and `ReportWidget` from the frame's **In view** counts and the widget's `total`; this
response carries no state of its own.

## Schemas

```yaml
PlanDetailsReport:
  required: [show, currency, cards, builtIn, attributes, attributesSetUp]
  properties:
    show:       { enum: [plan, campaign, initiative, activity] }   # as applied, default activity
    currency:   { type: string }                 # the report filter's currency, ISO code (D-303)
    cards:      { $ref: PlanDetailsCards }
    builtIn:    { type: array, items: { $ref: PlanDetailsWidget } }     # FR-006, FR-007; [] when cards.records = 0
    attributes: { type: array, items: { $ref: PlanDetailsAttributeRow } } # FR-008, Settings' display order
    attributesSetUp: { type: boolean }           # false: "No <level> attributes are set up in Settings"

PlanDetailsCards:
  required: [records, plannedCost, allFilledIn, allFilledInPercent]
  properties:
    records:            { type: integer }        # plans, campaigns, initiatives or activities matching (= frame In view)
    plannedCost:        { $ref: Money }          # their planned cost (domain model §4.2), report currency
    allFilledIn:        { type: integer }        # records with Missing = 0 (domain model §2.6)
    allFilledInPercent: { type: integer, nullable: true }  # whole per cent, half up; null when records = 0

PlanDetailsAttributeRow:
  required: [attributeId, label, type, widgets]
  properties:
    attributeId: { type: string }                # SA-…; identity, never the label (client 019 FR-030)
    label:       { type: string }
    type:        { enum: [singleChoice, multipleChoice, multipleChoiceSplit, yesNo, text, longText, number] }
    widgets:     { type: array, items: { $ref: PlanDetailsWidget } }  # 3 for choice and yes/no, 1 (filledIn) otherwise

PlanDetailsWidget:
  required: [key, title, measure, values, total]
  properties:
    key:      { type: string, pattern: "^[a-z][A-Za-z0-9]{0,39}$" }  # 039's ReportWidget pattern (address
                                  # `table`, export `widget`), stable within a level: activityTypeCount,
                                  # activityTypePlannedCost, statusCount, progressCount, timingCount,
                                  # startPeriodCount, plannedCostEntered, goalEntered, and per Shell attribute
                                  # attr<number>Count | attr<number>PlannedCost | attr<number>FilledIn (SA-12 → attr12Count)
    title:    { type: string }    # "Activities by Company code", "Planned cost by Company code", "Company code filled in"
    measure:  { enum: [count, plannedCost, filledIn] }
    values:   { type: array, items: { $ref: PlanDetailsValue } }  # every value, largest first; then none,
                                  # doesNotApply (filledIn: filled, none, doesNotApply); start periods in time order
    other:    { $ref: PlanDetailsValue, nullable: true }  # the values after the ten largest, summed, for the chart (FR-017)
    total:    { $ref: PlanDetailsValue }   # each record once (FR-011); count 0 → the widget's "No data yet"
    filledInPercent: { type: integer, nullable: true }    # filledIn only: filled ÷ (filled + none), half up
    countsInEach:    { type: boolean }     # true: "<Records> with several values count in each group; the total counts them once"
    filterParam:     { enum: [attr, activityType, activityStatus], nullable: true }
                                  # the All work / report filter parameter a value opens (FR-014): attr opens
                                  # attr=<attributeId>:<value key>; null for status of other levels, progress,
                                  # timing, start period, planned cost entered, goal entered

PlanDetailsValue:
  required: [key, label, count, linkable]
  properties:
    key:         { type: string }  # the filter value where there is one: option id, "yes", "no", "none"
                                   # (Not filled in, 039 data-model §2), activity type id, activity status;
                                   # otherwise "2026-Q2" / "2026-04", doesNotApply, noDates, outsideYear,
                                   # filled, other, total
    label:       { type: string }  # "(inactive)" appended by the server for an inactive option
    inactive:    { type: boolean }
    count:       { type: integer }
    plannedCost: { $ref: Money, nullable: true }  # set when measure = plannedCost
    linkable:    { type: boolean } # true when filterParam is set and key is a filter value; false for
                                   # doesNotApply (All work counts it under none, 016 FR-039), other, total
```

## Export sheets (`measure.Report.Sheets`)

`Sheets(ctx, a, f measure.Filter) ([]measure.WidgetSheet, error)` returns one `WidgetSheet{Key, Sheet}` per
widget of `f.Show`, in screen order, `Key` = `PlanDetailsWidget.key`. Columns as the table view: **Value**, then
**Activities** (or Plans, Campaigns, Initiatives) for `count` and `filledIn`, **Planned cost (<currency>)** for
`plannedCost`, **Share** where the table shows it; every value, Not filled in, Does not apply and the total row
(FR-017; client 019 FR-014). 039's export adds In view, About this export (filters, period, currency, rates
used, "as of") and the file name.

## Go interfaces (`internal/measure/plandetails/ports.go`)

```go
type Transactor interface {
	Read(ctx context.Context, fn func(ctx context.Context) error) error
}

// Scoped is the frame's step for a report (039 research R1, R2): the year's nodes in the reports.view scope,
// matched by allwork.Match with the filter, and the filter's currency with its current rate.
type Scoped interface {
	Matched(ctx context.Context, a actor.Actor, f measure.Filter, columns []string) (Matching, error)
	Attributes(ctx context.Context) ([]allwork.Attribute, error)
}

type Matching struct {
	Nodes    []allwork.Node    // the year's nodes in scope
	Matched  allwork.Matched   // IDs matching f.Query per level (039's allwork.Match)
	Currency allwork.Currency  // f.Currency and its current rate (D-303)
}
```

`app/measure_plandetails.go` adapts `Scoped` from the frame's step if lane A exports it (plan, hand-offs), or
else from 039's `measure.Planning` (`Year` with columns `completeness` and `inq`, `Attributes`), `measure.Plans`
(`InReportsScope`) and `allwork.Match` — the same three calls the frame makes. Amounts are summed in USD and
converted once, as All work does.

No exported function is added outside lane D's paths. Proposed to the integrator (research R2):
`allwork.AttributeParts`, which would replace 043's local copy of 016's attribution rule.
