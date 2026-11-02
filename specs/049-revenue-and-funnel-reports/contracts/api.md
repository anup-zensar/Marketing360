# API contract: Revenue and Funnel-Stage Reports

These four operations are **the wave B contract's report half**, which the build plan has lane B write
before 048 and 050 build against it (research R6). All paths under `/api/v1`. Both reports are read
through the frame's rule — signed in, scope applied in the service — and need **no new permission**.

## Operations

| Operation | Path | Permission |
|---|---|---|
| getRevenueReport | GET `/reports/revenue` | signed in (scoped in the service) |
| getRevenueRanked | GET `/reports/revenue/ranked` | signed in |
| getFunnelReport | GET `/reports/funnel` | signed in |
| getFunnelRanked | GET `/reports/funnel/ranked` | signed in |

Each takes the frame's shared parameters (year, plan, planType, team, owner, activityType,
activityStatus, attr, period, show, currency, table) plus this feature's own:

- `category` — an activity category id, `uncategorised`, or absent for all categories (049-C10)
- `stage` — the funnel report only: `inq` or `mql`, defaulting to the first the models in view allow
  (049-C11, 049-C12)
- the ranked parameters, as 040 defines them: `sort`, `order`, `group`, `size`, `page`, `q`, `drill`

`sort` accepts this feature's figures — `actualRevenue`, `projectedRevenue` on Revenue; `actual`,
`projected` on Funnel stages — not 040's money keys.

## Answers

**The report** carries the cards, the period series and the frame's own fields. Each series point is
`{ period, actual, projected, goal, inProgress }`, with `actual` and `projected` both present only in the
period in progress (047 FR-010). Every figure that cannot be worked out is `null` with a `reason`, never
a zero (FR-009); the report-level `reason` states why an empty report is empty (049-C20).

**The ranked answer** mirrors 040's shape — `total`, `trail`, `groups[{key,label,totals,entries}]`,
`drillLost` — with this feature's figures on each entry: `goal`, `actual`, `projected`, `performance`,
`actualCost`, `roi`, and on the funnel also `conversion`.

`actualCost` and `roi` are **always null with the reason "Attribution is not configured yet"** until
feature 050 lands (049-C6). The fields are in the contract from the start so 050 fills them without a
contract change.

Every projected figure is served with the model's name and `modelChangedAt` (047 FR-012).

## Not in this contract

No new permission keys. No stored record, so no create, update or delete. **Compare** (client 016 FR-067)
is not served (049-C7).

## Cross-feature notes for the integrator

1. **Two asks for feature 039's lane** (049-C19), neither of which this lane may make itself:
   `frontend/src/features/measure/shared/api.ts` hardcodes 040's sort keys when it builds an export URL,
   so a `revenue` or stage sort is dropped silently and the export falls back to the default order; and
   `shared/RankedList.tsx` has no **Expand all / Collapse all**, which client 016 FR-068 asks of every
   ranked list. Both belong in the shared component so every report gains them at once.
2. **Compare** (FR-067) is unbuilt anywhere and is recorded as out of scope here (049-C7); it belongs to
   the same shared list.
3. **measure-metrics.md** should gain 049 in the "Shown by" column of Actual, Projected, Actual or
   projected, Goal per stage, Performance and Conversion rate — a never-edit file for a lane.
4. **Feature 048** defines Performance's wording and Goal coverage, and **050** defines ROI; this feature
   shows the first and leaves the second stated as unavailable.
5. **One more ask for 040's lane**, found while writing `frontend/e2e/measure-revenue.spec.ts` (T052):
   `shared/RankedList.tsx`'s "Open X in Plan" row links fail WCAG 2.2's target-size rule (axe
   `target-size`) whenever a row's name is short, in both Revenue and Funnel's ranked lists — the first
   real-browser axe run against that shared component, since no e2e test had exercised it before (not
   even 040's own Plan budget report). The link needs a larger hit area, for example padding on `.open`
   in `RankedList.module.css`. T052's own axe check excludes `a[aria-label^="Open "]` so it still proves
   this feature's own markup without masking a defect it cannot fix in its own files.
