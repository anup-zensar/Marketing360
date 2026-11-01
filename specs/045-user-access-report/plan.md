# Implementation Plan: User Access Report

**Branch**: `feature/045-user-access-report` (from `measure-module`) | **Date**: 2026-11-01 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `specs/045-user-access-report/spec.md` (clarifications 045-C1–045-C4)

**Design artifacts**: [research.md](research.md) (R1–R5) · [contracts/api.md](contracts/api.md) ·
[quickstart.md](quickstart.md). No `data-model.md`: one nullable column on an existing table, described in R2
(build guide §8 asks for it only for a new table).

## Summary

1. **Last sign-in** (R2): migration **00029** adds `people.last_signed_in_at` (timestamptz, null); auth records
   it in `startSession`'s existing `Write`, after the credentials pass, through a new auth query file
   `sign_ins.sql` — no version bump, no "changed by", no activity-log entry (045-C4).
2. **A read-only `measure/useraccess` package** builds the list from 014's People tab read
   (`access.Service.ListPeopleAccess`), e-mail from `people.Store.List` and the times from auth, then filters,
   counts, sorts and pages in memory (R1, R4).
3. **039's report page and export**, without the plan filter panel, Period, Show or Currency (045-C1, R3); its
   own Search, Team and Status in the address. `Sheets` checks `access.manage` itself because 039's export route
   needs only `data.export`.
4. **Web** (`features/measure/user-access/`): the page, cards, the last sign-in band widget with its table, the
   paged and sortable table, rows linking to Settings › People.

All code, the migration and the auth change included, waits for 039's hub on `measure-module` (build guide §3,
about hour 16), as lane D's instructions say, although the migration and the auth change do not depend on it.

## Technical Context

**Language/Version**: Go 1.27 (`go.mod`); TypeScript 5.9 strict + React 19

**Primary Dependencies**: none new — 039's `measure` package, `access`, `people`, `auth`, `platform/sheet`, sqlc,
goose; web: TanStack Query, Zod, MSW, vitest-axe, Playwright, 039's `features/measure/shared`, 036's SVG bars

**Storage**: PostgreSQL. Migration 00029 (one column, no index: the report reads every person). Two new sqlc
queries in `queries/sign_ins.sql`; existing reads of people, teams, memberships, roles and assignments.

**Testing**: Go unit tests of every rule on fakes; auth store tests on PostgreSQL (`testdb`) for `RecordSignIn`
(only that column changes: version and `updated_*` equal before and after) and `LastSignIns`; auth service tests
for success, wrong password, inactive and locked; handler and contract tests on
`apitest` with personas (administrator; `reports.view` without `access.manage`; no `data.export`); Vitest + MSW +
vitest-axe per page state; Playwright + axe for US1 (`frontend/e2e/measure-user-access.spec.ts`: sign in as
two people, then the administrator sees both times)

**Target Platform**: developers' machines (127.0.0.1); GitHub Actions on `measure-module`

**Project Type**: web application (Go API + React web app)

**Performance Goals**: SC-001 — first page and every change within 2 s at 2,000 people; the service adds under
100 ms to its three reads (`BenchmarkUserAccess_SC001`, R5). Sign-in adds one `UPDATE` by primary key.

**Constraints**: constitution VII (route rule `access.manage`; `Sheets` checks it again; e-mail shown only here,
client 019 Assumptions), IV (bands and counts from the API), F6 (the band widget has a table), V (≤ 300 non-blank
lines per file), VIII (comments ≤ 5 %); build guide §4 — only issue #38's "Owns" paths

**Scale/Scope**: 1 operation (the export is 039's); 1 migration; 1 query file and 2 store methods in auth; 1 Go
package (~8 files); 1 wiring file; 1 web folder (~10 files); 1 e2e spec

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.* — **Pass, before and after design.**

| Principle | Check | Result |
|---|---|---|
| I Spec first | Spec with 045-C1–C4; FR IDs in test names; every FR has a test below | Pass |
| II Shell language | Roles and scopes in 014's words; "Never"; Active and Inactive | Pass |
| III Documented design | plan, research R1–R5, contracts/api.md, quickstart | Pass |
| IV Server is the source of truth | Rows, counts, bands from the API; the web app formats dates in the viewer's format and zone | Pass |
| V Simplicity | One nullable column; reuses 014's People read; no new table, no dependency | Pass |
| VI Meaningful tests | A test per rule and page state; PostgreSQL test that sign-in changes only the time; contract tests; e2e for US1 | Pass |
| VII Security | `access.manage` on the route and again in `Sheets`; refused sign-ins record nothing; the time never reveals whether an e-mail exists (written only after the constant-time check passes) | Pass |
| VIII Clean construction | Comment budget per file; one concept per file | Pass |
| IX Experience | Search, team, status, sort and page in the address; states say why; rows open Settings › People | Pass |
| X Palette | 036's bars with the chart tokens; Never in a neutral token, labelled | Pass |
| B1–B9, F1–F10 | Gin only in `handler.go`; ctx first; contract first; sqlc for SQL; goose migration with Down; TanStack Query hooks in `api.ts`; Zod for the address; no chart library | Pass |

## Project Structure

### Documentation (this feature)

```text
specs/045-user-access-report/
├── spec.md
├── plan.md              # this file
├── research.md          # R1–R5
├── quickstart.md
├── contracts/api.md     # for the wave A contract
├── checklists/requirements.md
└── tasks.md             # /speckit-tasks
```

### Source Code (repository root) — only issue #38's "Owns" paths

```text
backend/internal/platform/database/migrations/00029_last_sign_in.sql   # Up: ADD COLUMN; Down: DROP COLUMN
backend/internal/platform/database/queries/sign_ins.sql                # RecordSignIn, ListLastSignIns (sqlc output generated)
backend/internal/auth/                     # the sign-in change only
├── store.go          # Store += RecordSignIn, LastSignIns
├── store_postgres.go # the two methods through sqlc
├── store_fake_test.go# the fake's two methods
├── service.go        # startSession: RecordSignIn(person.ID, now) inside the Write
└── *_test.go         # sign-in records, refusals record nothing, store changes only the column
backend/internal/measure/useraccess/       # NEW, read-only
├── model.go          # UserAccessReport, rows, cards, bands (API types)
├── ports.go          # Access, People, SignIns
├── service.go        # Report(ctx, actor, Query): rows, filter, cards, bands, sort, page
├── query.go          # q, team, status, sort, page, pageSize: parse and allowlist
├── bands.go          # sign-in bands from an injected clock (R4)
├── sheets.go         # measure.Report: Key, Available, Sheets (checks access.manage)
├── handler.go        # GET /api/v1/reports/user-access
└── *_test.go         # rules, handler, benchmark
backend/internal/app/measure_useraccess.go   # registerUserAccess replaces 039's stub: adapters, Register, mount
backend/internal/app/contract_routes_test.go # remove 045's line from notYetServed only (build guide §4.3)
frontend/src/features/measure/user-access/   # replaces 039's stub page
├── UserAccessReport.tsx (039's stub), AccessControls.tsx, AccessCards.tsx, SignInBands.tsx, AccessTable.tsx,
│   lastSignIn.ts
├── api.ts, schemas.ts (address), index.ts
└── *.test.tsx
frontend/e2e/measure-user-access.spec.ts
```

**Structure Decision**: web application; one business package and one feature folder as constitution B2 and F2
lay them out, wired through 039's stubs; auth changed only where it starts a session.

## Business rules and the tests that prove them

| Rule | Where | Test |
|---|---|---|
| A successful sign-in records its moment in UTC in the session's step (FR-009) | `auth/service.go` | `TestSignIn_RecordsLastSignIn_FR009` |
| Wrong password, inactive person, locked attempt record nothing (FR-009) | `auth/service.go` | `TestSignIn_RefusedRecordsNothing_FR009` |
| Recording changes only `last_signed_in_at`: version and `updated_*` unchanged, no log entry (FR-010) | `auth/store_postgres.go` | `TestRecordSignIn_OnlyTheTime_FR010` (PostgreSQL) |
| Migration 00029 applies through `database.Migrate` (testdb); existing people read "Never" (FR-011) | migration, `auth/store_postgres.go` | `TestLastSignIns_NoneBeforeFirstSignIn_FR011` (PostgreSQL); the Down step checked by hand once (quickstart) |
| Rows = People tab roles, scopes and teams, plus e-mail, status, last sign-in (FR-003, FR-005, SC-003) | `service.go` | `TestRows_EqualPeopleTab_SC003`, `TestRows_NoRole_FR005`, `TestRows_InactiveTeamMarked_FR005` |
| Search, Team, Status (default Active), sort ("Never" last), paging 5–500 (FR-006) | `query.go`, `service.go` | `TestQuery_Allowlist_FR006`, `TestFilter_SearchTeamStatus_FR006`, `TestSort_NeverLast_FR006`, `TestPage_FR006` |
| Cards and bands over the filtered rows, adding up (FR-007) | `bands.go` | `TestBands_Boundaries_FR007` (30, 31, 90, 91 days, never), `TestCards_FR007` |
| `access.manage` on the route and in `Sheets` (FR-002) | 039's routes, `sheets.go` | `TestUserAccess_NeedsAccessManage_FR002` (contract), `TestSheets_RefusesWithoutAccessManage_FR002` |
| Sheets: every matching row, screen columns, Never empty (FR-013, SC-005) | `sheets.go` | `TestSheets_EveryRow_FR013` |
| States: data, no match for filters, no match for search (FR-012) | web | `UserAccessReport.test.tsx` (one test per state, axe) |
| 2,000 people within SC-001 | — | `BenchmarkUserAccess_SC001` (R5) |
| US1 end to end on the demo data, axe | e2e | `measure-user-access.spec.ts` |

Benchmark result: _recorded here after implementation_.

## Dependencies and hand-offs

- **039 (lane A)** — `measure.Report`/`WidgetSheet`, the `registerUserAccess` stub, the shared export, the report
  page parts. Asked of lane A (R3): `ReportFrame` without the filter panel, In view and Currency for this report;
  report parameters (`q`, `team`, `status`, `sort`) passed to `Sheets`; the rule
  `GET /api/v1/reports/user-access: need(authz.AccessManage)` and the operation in `notYetServed` with the wave A
  contract.
- **Integrator** — fold `contracts/api.md` into the wave A contract; reserve migration number 00029 for 045 (039
  has 00028); record 045-C1–C4; correct issue #38's client citation (FR-034, FR-035). `sqlcdb/models.go` gains the
  column when regenerated (build guide §4.6).
- **#DEMO (lane D)** — none: the demo people are not changed (demo data is additive); the e2e test signs in to
  create the times.

## Complexity Tracking

| Violation | Why needed | Simpler alternative rejected because |
|---|---|---|
| Two tables, `sign_ins` and the one-row `sign_in_recording`, instead of the `people` column (research R2 addendum) | A `people` column changes the people package's generated rows; `recordedSince` is when recording started in this database | The column needs 003's people store changed (outside issue #38's paths); goose's version table is outside sqlc's schema |
