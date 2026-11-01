# Implementation Plan: Approvals Report

**Branch**: `feature/041-approvals-report` (from `measure-module`) | **Date**: 2026-11-01 | **Spec**: [spec.md](spec.md)

**Design artifacts**: [research.md](research.md) (R1–R5) · [contracts/api.md](contracts/api.md) ·
[measure-module-build.md](../../docs/sdd/measure-module-build.md). No `data-model.md`: no new table.

## Summary

A read-only report on 039's frame. `GET /api/v1/reports/approvals` returns three widgets — status mix, average days
to a decision, approvals grid — and `Report.Sheets` feeds 039's shared export. Rounds and their actions are read
by the existing approval-store reads for the plans in view (R2), figures reuse 017's rules (R1), nothing is
written, no migration, no new SQL.

## Technical Context

Go 1.27 + TypeScript 5.9/React 19; no new dependency. Reads: the year's nodes (`Planning.Year`, shared with 039)
for the plans in the person's `reports.view` scope, then **one** `RoundsOfSubjects` and **one**
`ActionsOfSubjects` for those subjects, then one purchase-request read (R4). No statement per row.
Tests: Go on fakes; Vitest + MSW + vitest-axe; Playwright + axe for US1 (`e2e/measure-approvals.spec.ts`).

## Files (this lane's Owns only)

```text
backend/internal/measure/approvalreport/
  model.go ports.go service.go handler.go outcomes.go average.go rows.go grid.go   # + tests, + approvalreporttest/
backend/internal/app/measure_approvals.go        # body of registerApprovals; adapter on stores.ApprovalStore
frontend/src/features/measure/approvals/         # ApprovalsReport, StatusWidget, AverageWidget, ApprovalsGrid, api.ts, schemas.ts
frontend/e2e/measure-approvals.spec.ts
```

`queries/measure_approvals.sql` is **not created** (R2). Not touched: `openapi.yaml`, `routes_measure.go`,
`app/measure.go`, `internal/approval`, `purchaserequest`, `features/measure/index.ts`, the decision log.

## Business rules and the tests that prove them

| Rule | Where | Test |
|---|---|---|
| Engine outcome → one of six words; every round counted once; mix adds to total (041-C1, C6) | `outcomes.go` | `TestOutcomeWord_041_FR006` (table, every engine value) |
| Status mix and share per record type; only types with rounds offered (FR-006, FR-007) | `service.go` | `TestStatusMix_041_FR006_FR007` |
| Waiting count and oldest waiting of the selected type; the three types add up to 036's (FR-008, 041-C10) | `service.go` | `TestWaiting_ByType_041_FR008` (service) |
| Overall average is `approval.AverageDecisionDays` over person-decided rounds, auto-approved left out (FR-009, R1) | `average.go` | `TestAverage_041_FR009` |
| Per-approver average from submission to own decision, slowest first, decider credited incl. backup, delegate, on behalf (FR-010, C2, C7) | `average.go` | `TestApproverAverages_041_FR010` |
| No decision → `null`, never zero (FR-011, FR-019) | `average.go` | `TestAverage_NoDecisions_041_FR011` |
| One row per named approver per level; Not needed once another decided; automatic level one row with no approver; skipped level no row (FR-012, C2, C3) | `rows.go` | `TestRows_041_FR012_C2_C3` |
| Filters (record type, approver, outcome), unknown value → empty not error; search; sort across all rows with stable ties; paging and total (FR-013–FR-015) | `rows.go` | `TestGridQuery_041_FR013_FR015` |
| Amount: activity planned cost now, request amount 032, display currency; `null` without `finance.view` with plans covered (FR-018, C4) | `rows.go` | `TestAmounts_041_FR018` |
| Name is a link only when the person may open the record (FR-017, C5) | `rows.go` | `TestCanOpen_041_FR017` |
| A purchase request is in view when one of its line items is, so the frame's filters narrow it (FR-005, 041-C8) | `service.go` | `TestPurchaseRequestsInView_041_C8` |
| Unknown record type, outcome or approver answers empty, not an error; the approver filter narrows the grid only (041-C9, C10) | `grid.go`, `handler.go` | `TestUnknownValues_041_C10`, `TestApproverFilterGridOnly_041_C9` |
| Same figures as 036's Approvals panel for the same plans: waiting summed over types, oldest waiting, average (SC-002) | — | contract `TestApprovalsReport_EqualsDashboard_041_SC002` on `apitest` (DemoDB; calls both endpoints) |
| Only plans in `reports.view` scope; a record outside never read or counted (FR-003) | `service.go` | `TestScope_PerPersona_041_FR003` (service, fakes) · contract `TestApprovalsReport_Scope_041_FR003` |
| Sheets: three, every matching row, no paging, currency in heading (FR-023) | `service.go` | `TestSheets_041_FR023` |
| Contract: 200 shape, 400 only for a malformed `sort`, `order`, `page`, `pageSize` or frame filter, 401/403 via the route table (FR-001) | `handler.go` | `TestApprovalsReport_Contract_041` (`apitest`) |
| Each state a widget renders: data, no plans, no approvals, no financial access, filters exclude everything, loading, failure (FR-020) | `approvals/*.tsx` | one MSW test per state per widget |
| Status widget type switch, grid search/sort/filter/page, comment expand, plain-text name, export menu | `approvals/*.tsx` | component tests with `vitest-axe` |
| US1 end to end: Alex opens Reports › Approvals; figures equal Monitoring; axe clean | — | `e2e/measure-approvals.spec.ts` |

No scope-isolation test on PostgreSQL: the feature adds **no store read**; the reads it uses are scope-tested
in `approval` and `purchaserequest`, and the scope here is the plans in view (service and contract tests above).

## Permissions

`need(authz.ReportsView)` on `GET /api/v1/reports/approvals`; the shared export keeps `need(authz.DataExport)`.
Rows come only from plans in the actor's `reports.view` scope; `finance.view` per plan decides each amount.
Lane A adds the rule with the wave A contract.

## Benchmark

`BenchmarkApprovalsReport` at the 028 volumes, one request for the demo year with no filter: statements fixed
(year nodes, rounds, actions, purchase requests) and recorded here after US1. File:
`internal/platform/apibench/measure_approvals_test.go` (a new file beside 039's `reports_test.go`; lane A to
confirm it may live there, otherwise in the package).

## Constitution check

Spec first ✔ (C1–C7) · server works out every figure ✔ · reuses 017/031/032/036 definitions ✔ · money exact
decimal ✔ · Shell vocabulary (Reports, Approvals, Not filled in not used) ✔ · palette tokens, 036 SVG bars, no
chart library ✔ · files ≤ 300 lines, comments ≤ 5% ✔ (checked per story) · no new dependency ✔ · read-only, no
yellow primary button ✔.

## Open questions for lane A (and a mismatch to settle)

1. **Report-specific parameters.** Can the approver, outcome, record type, search, sort, page and page size (and
   the status widget's type) ride in the page address, and can saved views store them? 039's `ReportFilter` has
   none of them. Provisional: address only, parsed in 041's handler (R5).
2. **Matching helper and export.** Where do `allwork.Match` / the frame's "plans in view" and the shared export
   settle (039 T022, T025, T044)? 041 calls them as they appear; until then the service is written against its
   own ports.
3. **File name.** The hub's stub is `app/measure_approvals.go` (`registerApprovals`); issue #34's Owns list says
   `app/measure_approvalreport.go`. 041 uses the hub's name; the issue should be corrected.
4. **Purchase request name, owner, PO number** (R4): `purchaserequest.Summaries` carries none; if no existing
   read gives them in one statement, may the integrator widen `Summary`?
5. **Benchmark file** location (above).
6. **`restarted` and `cancelled`** shown as Replaced and Recalled (041-C6): confirm with the 017/031 owners.
7. **Controls row.** This report has no time widget or ranked list, so it needs no Period or Show: can
   `ReportFrame` omit them per report (039 FR-016)? Provisional: the contract accepts and ignores both.
