# Quickstart: Measure Foundation (039)

## Run

```bash
cd backend && go run ./cmd/seed -reset && go run ./cmd/api
```

```bash
cd frontend && pnpm run dev
```

Sign in as a demo person (docs/product/demo-users-and-data.md): Alex (every market) and Priya (Lubricants
Europe only).

## Check

1. **Reports in the sidebar** after Dashboards; `/reports` lists the nine reports in order, those not built yet
   as "Coming soon" (FR-001, 039-C3).
2. Open `/reports/plan-budget`: the frame shows the current year, every plan in scope, Months, your currency, and
   "In view: …" equal to All work's counts for the same year with no filters (FR-007, SC-001).
3. Tick a brand in the **Activity** tab, see "Changes not applied yet", press **Apply**: chips, "In view" and the
   address change together; copy the address into another window — same report (FR-013–FR-015, SC-005).
4. Edit the address to name a plan of another market: "1 filter value is no longer available and was left out",
   nothing named (FR-015, SC-002).
5. **Saved views**: save, star, reload `/reports/plan-budget` without a query (the view applies), change a
   filter ("Modified"), update, rename, delete and **Undo** (FR-024–FR-029).
6. **Export**: the Excel file holds **In view** and **About this export** with the filters in words (FR-022).
7. Without financial access, "You can't see financial data" shows where money would (FR-004). Every demo role
   with `reports.view` also has `finance.view`, so to see it give someone a role without Finance view in
   Settings › Roles and access; `ReportFrame.test.tsx` and the frame's contract tests prove it.

## Tests

```bash
cd backend && go test -race -short ./internal/measure/... ./internal/allwork/... ./internal/access/... ./internal/app/...
```

```bash
cd backend && go test -race -run ReportViews ./internal/measure/...
```

(the second needs `TEST_DATABASE_URL`: the person-isolation test of `report_views`)

```bash
cd frontend && pnpm run test -- measure && pnpm run test:e2e -- measure.spec.ts
```

```bash
cd backend && go test -run '^$' -bench 'BenchmarkReportsFrame' -benchtime 5x ./internal/platform/apibench/
```
