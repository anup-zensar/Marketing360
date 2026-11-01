# Implementation Plan: GSAP Link Reports

**Branch**: `feature/044-gsap-link-reports` (from `measure-module`) | **Date**: 2026-11-01 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `specs/044-gsap-link-reports/spec.md` (clarifications 044-C1–044-C5)

**Design artifacts**: [research.md](research.md) (R1–R6) · [contracts/api.md](contracts/api.md) ·
[quickstart.md](quickstart.md). No `data-model.md`: no new table, no migration (build guide §8).

## Summary

1. **A read-only `measure/gsaplink` package** builds one tab per request — SWBS not received, Waiting for GSAP or
   Purchase orders — with its cards, its chart and one page of rows, and the same rows unpaged for 039's export.
2. **The records in view are the frame's**: the year's nodes in the `reports.view` scope matched by
   `allwork.Match` (research R1). The SWBS tab is the matched line-item nodes whose All work SWBS status is
   `readyToSend`, so it equals All work by construction (044-C1, SC-002).
3. **Three reads keyed by the matched IDs** in `queries/measure_gsaplink.sql`: activity approval times, purchase
   requests waiting for GSAP by 032 FR-032 with their unanswered routing, and the purchase-order, invoice and
   payment transactions of the line items in the person's `finance.view` scope (R2, R3).
4. **039's frame** gives the filter, currency, saved views, empty states and export; the report adds `tab`, `sort`,
   `page` and `pageSize` (R4).
5. **Web** (`features/measure/gsap-link/`): the report page inside `ReportFrame`, three tabs, cards, the waiting
   band chart with its table, paged and sortable tables, rows linking to the line item or purchase request page.

Backend and web code wait for 039's hub on `measure-module` (build guide §3, about hour 16).

## Technical Context

**Language/Version**: Go 1.27 (`go.mod`); TypeScript 5.9 strict + React 19

**Primary Dependencies**: none new — 039's `measure` package, `allwork` types, `platform/money`, `platform/sheet`,
sqlc; web: TanStack Query, Zod, MSW, vitest-axe, Playwright, 039's `features/measure/shared`, 036's SVG bars

**Storage**: PostgreSQL. Reads only: All work's scoped year read (through the frame's step) and three new sqlc
queries in `queries/measure_gsaplink.sql` over `approval_rounds`, `purchase_requests`,
`purchase_request_lines`, `purchase_request_routings` and `financial_transactions`. No table, no migration.

**Testing**: Go unit tests of every rule on fakes; store tests on PostgreSQL (`testdb`) with a
scope-isolation test per read (IDs of another plan never come back) and the equality test with 032's waiting
view; handler and contract tests on `apitest` with personas (scope, no `reports.view`, no `finance.view` in one
plan, no `data.export`); Vitest + Testing Library + MSW `reply()` + vitest-axe, one test per tab state; Playwright
+ axe for US1 (`frontend/e2e/measure-gsap-link.spec.ts`)

**Target Platform**: developers' machines (127.0.0.1); GitHub Actions on `measure-module`

**Project Type**: web application (Go API + React web app)

**Performance Goals**: SC-001 — first page and every page or sort change within 2 s at 50 plans, 5,000
activities, 20,000 line items, 20,000 PR lines and 100,000 transactions; the tab adds under 300 ms to the year
read (`BenchmarkGsapLink_SC001`, R6). One year read and at most two of 044's statements per request.

**Constraints**: constitution VII (scope from the actor only; reads keyed by scoped IDs), IV (server works out
waiting times, bands, totals), F6 (the band chart has a table), X (chart tokens), V (≤ 300 non-blank lines per
file), VIII (comments ≤ 5 %); build guide §4 — only issue #37's "Owns" paths

**Scale/Scope**: 1 operation (the export is 039's); 1 Go package (~12 files); 1 query file; 1 wiring file; 1 web
folder (~14 files); 1 e2e spec

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.* — **Pass, before and after design.**

| Principle | Check | Result |
|---|---|---|
| I Spec first | Spec with 044-C1–C5; FR IDs in test names; every FR has a test below | Pass |
| II Shell language | SWBS code, GSAP PR number, purchase order, Ready to send to GSAP, Waiting for GSAP (032's words) | Pass |
| III Documented design | plan, research R1–R6, contracts/api.md, quickstart | Pass |
| IV Server is the source of truth | Rows, waiting times, bands, flags, totals and coverage from the API; the web app formats | Pass |
| V Simplicity | No table, no migration, no dependency; three reads by ID; one row builder for screen and export | Pass (one copied SQL predicate: Complexity Tracking) |
| VI Meaningful tests | A test per rule and tab state; scope-isolation test per read; equality with All work and 032's view; contract test; e2e for US1 | Pass |
| VII Security | `reports.view` route rule; IDs only from the scoped match; amounts only for `finance.view` plans (D-237); `tab`, `sort`, `pageSize` allowlisted; export needs `data.export`; files never stored | Pass |
| VIII Clean construction | Comment budget per file; one concept per file | Pass |
| IX Experience | Tab, sort and page in the address; states say why (039); rows open the record | Pass |
| X Palette | 036's bars with the chart tokens; flags labelled, never colour alone | Pass |
| B1–B9, F1–F10 | Gin only in `handler.go`; ctx first; contract first; sqlc for SQL; tests on fakes; TanStack Query hooks in `api.ts`; Zod for the address parameters; no chart library | Pass |

## Project Structure

### Documentation (this feature)

```text
specs/044-gsap-link-reports/
├── spec.md
├── plan.md              # this file
├── research.md          # R1–R6
├── quickstart.md
├── contracts/api.md     # for the wave A contract
├── checklists/requirements.md
└── tasks.md             # /speckit-tasks
```

### Source Code (repository root) — only issue #37's "Owns" paths

```text
backend/internal/platform/database/queries/measure_gsaplink.sql   # 3 reads (R2); sqlc output generated
backend/internal/measure/gsaplink/        # NEW, read-only
├── model.go          # GsapLinkReport, tabs, rows (API types)
├── ports.go          # Transactor, Scoped, Store
├── store_postgres.go # the three reads through sqlc
├── service.go        # Report(ctx, actor, measure.Filter, Table): one Read, one tab
├── swbs.go           # Ready-to-send line items, approval time, bands (FR-005–FR-007)
├── waiting.go        # waiting hours and working days, flags (R5; FR-006, FR-009)
├── requests.go       # Waiting for GSAP rows and cards (FR-008–FR-010)
├── orders.go         # rows per line item and PO, sums, totals, finance scope (FR-011–FR-013)
├── table.go          # sort keys, paging (FR-015)
├── sheets.go         # measure.Report: Key, Available, Sheets for 039's export (FR-019)
├── handler.go        # tab, sort, page, pageSize; 039's filter parser; GET /api/v1/reports/gsap-link
└── *_test.go         # rules, store on PostgreSQL, handler, benchmark
backend/internal/app/measure_gsaplink.go     # registerGsapLink replaces 039's stub: adapters, Register, mount
backend/internal/app/contract_routes_test.go # remove 044's line from notYetServed only (build guide §4.3)
frontend/src/features/measure/gsap-link/     # replaces 039's stub page
├── GsapLinkReport.tsx (039's stub), GsapTabs.tsx, SwbsTab.tsx, WaitingTab.tsx, OrdersTab.tsx,
│   WaitingBandsChart.tsx, PagedTable.tsx, waitingText.ts
├── api.ts, schemas.ts (tab, sort, page in the address), index.ts
└── *.test.tsx
frontend/e2e/measure-gsap-link.spec.ts
```

**Structure Decision**: web application; one business package and one feature folder as constitution B2 and F2
lay them out, wired through 039's stubs.

## Business rules and the tests that prove them

| Rule | Where | Test |
|---|---|---|
| Records from the frame's match in the `reports.view` scope (FR-003) | service, `Scoped` | `TestReport_OnlyPlansInScope_FR003` (contract, persona with one of two plans) |
| SWBS rows = matched line items Ready to send; received, off, not approved left out (FR-005) | `swbs.go` | `TestSwbs_ReadyToSendOnly_FR005`, `TestSwbs_EqualsAllWork_SC002` |
| Waiting from the latest approval, hours, never negative, flag > 48 h, bands add up (FR-006, FR-007) | `waiting.go`, `swbs.go` | `TestWaitingHours_FR006`, `TestSwbsBands_FR007` |
| Waiting for GSAP = 032 FR-032, the unanswered routing, resend ignored (FR-008, FR-009) | `requests.go`, store | `TestWaiting_EqualsPurchaseRequestsView_SC002` (PostgreSQL), `TestWaiting_ResendDoesNotRestart_FR009`, `TestWaiting_Modification_FR008` |
| Working days Monday to Friday, flag > 2 (FR-009) | `waiting.go` | `TestWorkingDays_FR009` (weekend, same day, Friday to Monday) |
| Cards and amount only for `finance.view` plans, "Covers N of M" (FR-004, FR-010) | `requests.go` | `TestWaiting_AmountOnlyWithFinanceView_FR004` |
| Orders: rows per line item and PO; PO amount, invoiced (credit notes negative), paid; unlinked ignored; totals over all rows (FR-011–FR-013) | `orders.go` | `TestOrders_SumsPerLineItemAndPO_FR012`, `TestOrders_PoOnTwoLineItems_FR011`, `TestOrders_TotalsAllPages_FR013`, `TestOrders_EqualsFinancials_SC003` |
| Orders only for `finance.view` plans (FR-004, 044-C4) | `orders.go`, store | `TestOrders_OnlyFinanceViewPlans_FR004` |
| Report currency at the current rate, summed in USD then converted once (FR-014) | `orders.go`, `requests.go` | `TestAmounts_InReportCurrency_FR014` |
| Sort keys, default sorts, ties by ID, page sizes 5–500 (FR-015) | `table.go`, `handler.go` | `TestTable_SortAndPage_FR015`, `TestHandler_RejectsUnknownTabSortPageSize_FR015` |
| Each read returns only the IDs asked for (constitution VII) | `store_postgres.go` | `TestStore_ScopeIsolation` (two plans, IDs of one) |
| Sheets: every row, default sort, hidden amounts empty, totals (FR-019, SC-005) | `sheets.go` | `TestSheets_EveryRow_FR019` |
| Access rules (FR-003) | 039's routes | contract tests `TestGsapLink_Contract`, `TestGsapLink_NoReportsView_FR003`, `TestGsapLinkExport_NeedsDataExport` |
| Each tab within 2 s at the 028 volume (SC-001) | — | `BenchmarkGsapLink_SC001` (R6) |
| Web: one test per tab state; tab, sort and page in the address; row opens the record (FR-001, FR-016, FR-018) | `features/measure/gsap-link` | `GsapLinkReport.test.tsx`, `SwbsTab.test.tsx`, `WaitingTab.test.tsx`, `OrdersTab.test.tsx` (axe in each) |
| US1 end to end on the demo data, axe (spec US1) | e2e | `measure-gsap-link.spec.ts` |

Benchmark result (2026-11-01, `BenchmarkGsapLink_SC001`, a 14-thread Windows developer machine, the 028 volume through
fakes): SWBS not received 32 ms, Waiting for GSAP 44 ms, Purchase orders 89 ms per request, each with its sort and
first page; all under the 300 ms the tab may add to the year read (R6).

## Dependencies and hand-offs

- **039 (lane A)** — used as merged in `measure-module`: `measure.Service.Scope` (the shared step), `measure.Report` with
  `measure.Parameterised` (`tab` and `sort` kept by saved views, given to `Sheets` in `Filter.Own`), the export,
  `ReportFrame`, `ReportWidget`. Still asked of lane A (tasks T029, T031): `plan` and `gsapPrNumber` in
  `GsapLinkSort`; `tab` and 044's `sort` values on `exportReport`.
- **Integrator** — record 044-C1–C6 in the decision log; 044-C1 goes to the product owner. Issue #37's client citation
  is corrected (FR-044, FR-045, FR-039).
- **#DEMO (lane D)** — demo data with line items Ready to send approved 1, 4 and 12 days before the demo date, a
  purchase request waiting for GSAP over a weekend, a top-up waiting (Modification), a PO with an invoice, a credit
  note and a payment, and a PO loaded from an extract with no purchase request, so every tab state shows.

## Complexity Tracking

| Violation | Why needed | Simpler alternative rejected because |
|---|---|---|
| 032's "waiting for GSAP" predicate copied into `measure_gsaplink.sql` (four lines) | `purchaserequest`'s list is scoped by `purchaseRequests.list` and is outside lane D's paths | Changing it needs lane C's agreement; the copy is pinned to it by `TestWaiting_EqualsPurchaseRequestsView_SC002` on PostgreSQL |
| The three reads answered twice: `measure_gsaplink.sql` on PostgreSQL, `gsapRecords` (`app/measure_gsaplink.go`) over the other areas' stores in memory | The API tests and a database-free run use the in-memory store set, which has no 044 store | Adding a 044 store to `stores.Set` edits shared files outside issue #37's paths that other lanes change too; the two are pinned equal by `TestGsapRecords_EqualPostgresStore` on PostgreSQL |
