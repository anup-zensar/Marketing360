# UI contract: Measure Foundation (039)

## Routes

| Route | Page | Needs |
|---|---|---|
| `/reports` | `ReportsPage` — the catalogue | `reports.view` |
| `/reports/plan-budget`, `/reports/approvals`, `/reports/purchase-requests`, `/reports/plan-details`, `/reports/gsap-link`, `/reports/user-access`, `/reports/summary`, `/reports/revenue`, `/reports/funnel` | each report's page, a stub until its lane builds it | `reports.view` (+ the report's rule) |

Sidebar: **Reports** (lucide icon `ChartColumnIncreasing`), permission `reports.view`, directly after
**Dashboards** (039-C3).
`app/navigation.ts`, `app/pages.tsx` and `app/router.tsx` carry every route from 039 on; lanes never edit them.

## Catalogue words (data-model §3)

| Report | Purpose (one line) |
|---|---|
| Plan budget | Budget, planned cost, committed and actual by month, and which plans, campaigns or activities lead or lag. |
| Approvals | How many approvals wait, how long decisions take, and every approval round. |
| Purchase requests | What is committed to each vendor and owner by delivery month, and what has been invoiced. |
| Plan details | How activities and planned cost spread across each Shell attribute, and what is not filled in. |
| GSAP link | SWBS codes not received, purchase requests waiting for GSAP, and the purchase orders of each line item. |
| User access | Who has which role and teams, and when they last signed in. |
| Summary | Whether the year is on track: projected against goal, the leading segments and the funnel. |
| Revenue | Revenue against goal by month, and the plans, campaigns or activities behind it. |
| Funnel stages | One funnel stage against goal by month, and where the funnel leaks. |

## Shared parts (`frontend/src/features/measure/shared/`, owned by 039)

| Part | What it does |
|---|---|
| `ReportFrame` | Page header (breadcrumb Reports › report, title, meta row: year, "Figures as of …"), Saved views, Export; left `FilterPanel`; `InViewLine`, `FilterChips`; the controls row; the report's widgets as children; the empty states of FR-031 |
| `FilterPanel` | Year, Plans, Plan type, Team, Owner, Activity type, Activity status, level tabs with attribute groups; counters, Select all, Clear all, search above ten values; **Apply** with "Changes not applied yet"; collapses to **Filters (n)** below 1024 px |
| `FilterChips` | One removable chip per restricted filter; **Reset filters** |
| `InViewLine` | "In view: 1 plan · 6 campaigns · …" |
| `PeriodControl`, `ShowControl`, `CurrencyControl` | Apply at once; in the address |
| `ReportWidget` | Title, ⓘ tip, **Chart · Table** switch (in the address), export menu (CSV, Excel), skeleton, error with **Try again**, "—" with reason; children draw the chart (036's SVG bars and lines, chart tokens) |
| `ReportTable` | The table view of a widget: same labels, totals row |
| `SavedViewsMenu` | List with star, active mark, "Modified" with **Update view** / **Save as new view**, **Save current filters**, rename, delete with Undo toast |
| `ExportMenu` | Whole-report **Export** (Excel) |
| `report-address.ts` | Address ↔ `ReportFilter` with Zod; malformed values dropped; canonical query for views |
| `api.ts` | Query hooks: `useReportCatalogue`, `useReportFrame(report, filter)`, `useReportViews(report)` + mutations; keys `['reports','catalogue']`, `['reports','frame',report,query]`, `['reports','views',report]` |

Reports import these from `@/features/measure/shared`; a part a report needs that is missing is asked of lane A,
never added by the report's lane (build guide §4).

## States every report page renders

Loading (skeletons), data, "No plans you may see for <year>", "Your filters exclude everything" with **Reset
filters**, a widget's "No data yet", "You can't see financial data" / "Covers N of M plans", a widget error with
**Try again**, "N filter values are no longer available and were left out", no access (the standard page).
