# Quickstart: Plan Budget Report

1. `go run ./cmd/seed -reset && go run ./cmd/api` in `backend/`; `pnpm run dev` in `frontend/`.
2. Sign in as Priya; open Reports › Plan budget for 2026. The cards equal the FY2026 Lubricants Europe plan page
   (budget, planned cost) and the Executive dashboard for her market (committed, actual).
3. The chart shows actual up to the current month only and marks it; Quarters regroups; Table lists running totals.
4. The ranked list opens on campaigns by planned cost. Open an initiative, then an activity: its line items show;
   Back and the browser's Back return.
5. Group by Owner and by Brand; totals per group; page size 25.
6. Export: the file has Cards, Periods and Ranked list sheets equal to the screen.
7. A person without Finance view for some plans sees committed and actual as "—" with the coverage note. No demo
   persona lacks it (Sara has it for both plans), so give someone a role without Finance view in Settings › Roles
   and access; `BudgetCards` and `planbudget` tests prove it.

Automated: `go test ./internal/measure/planbudget/...`, `pnpm exec vitest run src/features/measure`,
`pnpm run test:e2e -- measure-plan-budget` and `BenchmarkReportsFrame`'s plan budget cases (run in the end batch).
