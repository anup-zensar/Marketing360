# Quickstart: GSAP Link Reports (044)

How to check the report end to end once 039's hub and 044 are on the branch. Rules and tests are in
[plan.md](plan.md); the operation in [contracts/api.md](contracts/api.md).

## Prerequisites

- `measure-module` with 039 merged; this branch rebased on it; `sqlc generate` run (build guide §4.6)
- Demo data with the Measure story (#DEMO): `cd backend && go run ./cmd/seed`

## Run

```powershell
cd backend;  go run ./cmd/api
cd frontend; pnpm run dev
```

## Check

| Step | Expected |
|---|---|
| Sign in as a planner with one team; open **Reports › GSAP link** | SWBS not received: cards Waiting and Over 48 hours, the band chart, rows longest waiting first |
| Compare with All work's SWBS request column, same filters | The same line items as "Ready to send to GSAP" |
| Open **Waiting for GSAP** | The same requests as Purchase requests › Waiting for GSAP; a request routed on Friday shows 1 working day on Monday |
| Open **Purchase orders**; compare one line item with its Financials section | PO amount, invoiced and paid equal to the cent; totals row over all pages |
| Sign in as a person without "View financial data" in one plan | Waiting amounts "—" for that plan; Purchase orders without it; "Covers 1 of 2 plans" |
| Choose a team with no plans this year | "Your filters exclude everything" with **Reset filters** on every tab |
| Sort, change page size, press Back | The address holds `tab`, `sort`, `page`, `pageSize`; Back restores them |
| Export each tab as CSV and Excel; then the report's **Export** | Every row, same columns and totals; four sheets plus In view and About this export |

## Automated

```powershell
cd backend;  go test -race -short ./internal/measure/gsaplink/... ./internal/app/...
             go test -run '^$' -bench BenchmarkGsapLink_SC001 -benchtime 5x ./internal/measure/gsaplink/
cd frontend; pnpm run check; pnpm run test:e2e -- measure-gsap-link
scripts/check-changed.ps1
```
