# Quickstart: User Access Report (045)

How to check the report end to end once 039's hub and 045 are on the branch. Rules and tests are in
[plan.md](plan.md); the operation in [contracts/api.md](contracts/api.md).

## Prerequisites

- `measure-module` with 039 merged; this branch rebased on it; `sqlc generate` run (build guide §4.6)
- Demo data: `cd backend && go run ./cmd/seed` (it applies migration 00029 through `database.Migrate`)
- Once, by hand: `goose -dir internal/platform/database/migrations postgres "$env:DATABASE_URL" down-to 28`, then
  `up` again — the column goes and comes back

## Run

```powershell
cd backend;  go run ./cmd/api
cd frontend; pnpm run dev
```

## Check

| Step | Expected |
|---|---|
| Sign in as Priya, sign out; sign in with a wrong password as Carlos | — |
| Sign in as Alex (Administrator); open **Reports › User access** | Every active person; Priya's last sign-in now; Carlos "Never"; cards and the band widget add up |
| Open Settings › People, change Priya's job title and save | Saves without "changed in another tab"; the report keeps her last sign-in |
| Choose a team, then Status Inactive, then search "planner" | The list, cards and bands follow; the address holds `team`, `status`, `q` |
| Sort by Last sign-in | Most recent first, "Never" last |
| Sign in as a person with "See reports" but without "Give roles…" | User access is not on the Reports page; its address shows the no-access page |
| Export the table as CSV and Excel; then the report's **Export** | Every matching row, same columns; Last sign-in in UTC, empty for Never |

## Automated

```powershell
cd backend;  go test -race -short ./internal/auth/... ./internal/measure/useraccess/... ./internal/app/...
             go test -run '^$' -bench BenchmarkUserAccess_SC001 -benchtime 5x ./internal/measure/useraccess/
cd frontend; pnpm run check; pnpm run test:e2e -- measure-user-access
scripts/check-changed.ps1
```
