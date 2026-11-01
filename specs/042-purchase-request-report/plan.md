# Implementation Plan: Purchase Request Report

**Branch**: `feature/042-purchase-request-report` (from `measure-module`) | **Date**: 2026-11-02 | **Spec**: [spec.md](spec.md)

**Design artifacts**: [contracts/api.md](contracts/api.md) · [measure-module-build.md](../../docs/sdd/measure-module-build.md).
No `data-model.md` and no `research.md`: no new table, and the reads are those of 041 (the decisions are in the
spec's Clarifications 042-C1 to C6).

## Summary

A read-only report on 039's frame. `GET /api/v1/reports/purchase-requests` returns the cards, the status mix, two
commitment charts (by vendor, by owner, per delivery month) and the vendor › request › line › line item grid;
`Report.Sheets` feeds 039's shared export. One `Read`: the hub's scope, then **one** read of the purchase requests
on the line items in view (with their lines and PO lines), **one** read of the invoices of their PO lines, and the
approval workflow names. Nothing is written, no migration, no SQL of its own.

## Technical Context

Go 1.27 + TypeScript 5.9/React 19; no new dependency. The adapter in `app/measure_purchasereport.go` reuses 032's
store read `OnLineItems` (scoped as 036 reads it), 032's `prInvoices` (015 invoices per PO line), `settings.Rates`
for today's rates and the approval store's workflow list. Amounts convert line by line to USD as 032's reader does,
then to the display currency of the hub's scope. Tests: Go on fakes; Vitest + MSW + vitest-axe; Playwright for US1.

## Files (this lane's Owns only)

```text
backend/internal/measure/purchasereport/
  model.go ports.go service.go handler.go report.go query.go months.go charts.go grid.go rows.go sheets.go  # + tests, + purchasereporttest/
backend/internal/app/measure_purchasereport.go        # body of registerPurchaseRequests; the adapter
frontend/src/features/measure/purchase-requests/      # PurchaseRequestsReport, CardsRow, StatusWidget, CommitmentChart, RequestsGrid, api.ts, schemas.ts
frontend/e2e/measure-purchase-requests.spec.ts
```

Edited with the feature as lane A delegated (T013): `openapi.yaml` (the operation and schemas), `routes_measure.go`
(the rule), `contract_routes_test.go` (the `notYetServed` line, removed again once served). Not touched:
`internal/purchaserequest`, `app/measure.go`, `features/measure/index.ts`, the decision log.

## Business rules and the tests that prove them

| Rule | Where | Test |
|---|---|---|
| Six statuses with count, amount and whole-number shares adding to 100; waiting for GSAP (FR-007) | `charts.go` | `TestStatusMix_042_FR007` |
| Commitment = Approved and Closed lines by delivery month; every month between first and last; blank date last (042-C1, C2) | `months.go`, `charts.go` | `TestCommitment_042_FR009_C1`, `TestMonths_042_C2` |
| Top ten and Other in the chart, everything in the table (FR-011); no vendor = Not filled in (FR-012) | `charts.go` | `TestTopTenAndOther_042_FR011`, `TestNoVendor_042_FR012` |
| Monthly values add up to the card (FR-013) | `service.go` | `TestChartsAddUpToTheCard_042_FR013` |
| Amounts "—" for a plan without finance.view, with the plans covered (042-C4) | `service.go` | `TestNoFinanceView_042_FR020` |
| Grid hierarchy, search keeping the path, sort per level, paging by vendors (042-C3) | `grid.go` | `TestGridHierarchy_042_FR014`, `TestGridSearch_042_FR017`, `TestGridSortAndPaging_042_FR017` |
| Invoiced per line from PO lines; 0.00 with a PO and no invoice, "—" without a PO (042-C5) | `rows.go` | `TestInvoiced_042_FR019_C5` |
| Filters vendor, request owner, status; unknown values match nothing (042-C6) | `service.go` | `TestFilters_042_FR004_C6` |
| Columns choice is validated and kept in order (FR-016) | `query.go` | `TestParseQuery_042_FR016` |
| Export sheets hold every vendor and every matching line, chosen columns in order (FR-026, FR-027) | `sheets.go` | `TestSheets_042_FR026_FR027` |
| Contract: 200 shape, 400 for malformed sort/order/page/pageSize/columns, 401, 403 | handler | `TestPurchaseRequestsReport_Contract_042` |
| Status counts equal 036's Purchase requests panel (FR-008, SC-002) | handler | `TestPurchaseRequestsReport_EqualsDashboard_042_SC002` |

## Benchmark

`BenchmarkPurchaseRequestsReport` in `internal/platform/apibench/measure_purchasereport_test.go`: statements fixed
(year nodes, purchase requests on the line items, invoices, workflows), recorded here once PostgreSQL is available.

## Constitution check

Spec first ✔ · server works out every figure ✔ · reuses 032, 033, 015, 036 definitions ✔ · money exact decimal ✔ ·
Shell vocabulary (Reports, Purchase requests, Not filled in) ✔ · palette tokens, SVG bars, no chart library ✔ ·
files ≤ 300 lines, comments ≤ 5% (checked per story) ✔ · no new dependency ✔ · read-only, no yellow primary button ✔.
