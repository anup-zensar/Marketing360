# Tasks: User Access Report

**Input**: Design documents from `specs/045-user-access-report/` — [plan.md](plan.md), [spec.md](spec.md),
[research.md](research.md), [contracts/api.md](contracts/api.md), [quickstart.md](quickstart.md); the operation and
schemas as merged in `backend/api/openapi.yaml` (PR #57); 039 as merged in `measure-module` (report parameters
through `measure.Parameterised` and `Filter.Own`; `access.manage` checked once, in `openable`, for the frame and the
export).

**Tests**: required by constitution VI — a test per rule and per page state, the PostgreSQL tests that a sign-in
changes only its time, contract tests, an end-to-end test for US1. Write each test first and see it fail.

**Lane rules** (build guide §4): only issue #38's "Owns" paths — `measure/useraccess`, `app/measure_useraccess.go`,
the sign-in change in `auth`, migration 00029 (and `queries/sign_ins.sql` with its generated `sqlcdb` file),
`features/measure/user-access/`, `e2e/measure-user-access.spec.ts` — plus removing 045's line from
`notYetServed`, and the optional `ReportFrame` prop lane A agreed to (pinged before the edit). Commit and push after
each story.

## Phase 1: Setup

- [X] T001 Branch `feature/045-user-access-report` has `measure-module` with 039's hub, frame, export, saved views and
  `measure.Parameterised`; the backend builds

## Phase 2: Foundational — the sign-in time (blocks US1)

- [X] T002 `backend/internal/platform/database/migrations/00029_last_sign_in.sql`: `sign_ins (person_id, last_signed_in_at)`
  (research R2 addendum); a one-row `sign_in_recording (since)` stamped when the migration runs (research R2,
  `recordedSince`); Down drops both
- [X] T003 `backend/internal/platform/database/queries/sign_ins.sql`: `RecordSignIn` (the column only),
  `ListLastSignIns`, `GetSignInsRecordedSince`; `sqlc generate`
- [X] T004 Tests first: `auth/service_test.go` `TestSignIn_RecordsLastSignIn_FR009`,
  `TestSignIn_RefusedRecordsNothing_FR009` (wrong password, inactive, locked); `auth/store_postgres_test.go`
  `TestRecordSignIn_OnlyTheTime_FR010` (version and `updated_*` unchanged, no log entry),
  `TestLastSignIns_NoneBeforeFirstSignIn_FR011`
- [X] T005 `auth.Store` += `RecordSignIn`, `LastSignIns`, `SignInsRecordedSince`: `store_postgres.go`, the fakes
  (`store_fake_test.go`, `authtest/store.go`); `service.go` `startSession` records the time inside its `Write`

## Phase 3: User Story 1 — The report with last sign-in (P1) 🎯 MVP

- [X] T006 [P] [US1] `backend/internal/measure/useraccess/query_test.go`: `TestQuery_Allowlist_FR006` (q length,
  team IDs, status, sort keys, page, page size 5 – 500)
- [X] T007 [P] [US1] `rows_test.go`: `TestRows_NoRole_FR005`, `TestRows_InactiveTeamMarked_FR005`,
  `TestFilter_SearchTeamStatus_FR006`, `TestSort_NeverLast_FR006`, `TestPage_FR006`
- [X] T008 [P] [US1] `bands_test.go`: `TestBands_Boundaries_FR007` (30, 31, 90, 91 days, never), `TestCards_FR007`
- [X] T009 [US1] `model.go`, `ports.go`, `query.go`, `rows.go`, `bands.go`, `service.go`: the rules of T006 – T008,
  one row builder for screen and export
- [X] T010 [US1] `handler.go`: `GET /api/v1/reports/user-access`; `routes_test.go` (`TestRoutes_ReadOnly_FR004`)
- [X] T011 [US1] `backend/internal/app/measure_useraccess.go`: adapters over `access.Service.ListPeopleAccess`, the
  people store and auth's sign-in reads; register and mount; remove 045's line from `notYetServed`; contract tests in
  `useraccess/handler_test.go`: `TestUserAccess_Contract`, `TestUserAccess_NeedsAccessManage_FR002`,
  `TestRows_EqualPeopleTab_SC003`, `TestSignIn_ShowsInReport_SC002`
- [X] T012 [US1] `frontend/src/features/measure/shared/ReportFrame.tsx`: the optional prop lane A agreed to — hides
  the filter panel, In view, chips and Currency, keeps the header, saved views and export — with its test in
  `ReportFrame.test.tsx` (git checked for lane A's latest first; lane A pinged)
- [X] T013 [US1] `frontend/src/features/measure/user-access/`: `api.ts`, `access-address.ts` (q, team, status, sort,
  page, page size in the address), `UserAccessReport.tsx` (replaces 039's stub), `AccessControls.tsx`,
  `SignInBands.tsx` (the Last sign-in widget with its table view), `AccessTable.tsx` (sortable, paged, a row opens
  Settings › People), `access-text.ts`
- [X] T014 [US1] Web tests with vitest-axe: `UserAccessReport.test.tsx` (the states of FR-012, the address),
  `AccessTable.test.tsx`
- [X] T015 [US1] `frontend/e2e/measure-user-access.spec.ts`: two people sign in, the administrator sees both times;
  axe

## Phase 4: User Story 2 — Export (P2)

- [X] T016 [P] [US2] `sheets_test.go`: `TestSheets_EveryRow_FR013` (both widgets, every matching row in the sort,
  Never empty), `TestSheets_EqualTable_SC005`, `TestSheets_RejectsUnknownParameters`
- [X] T017 [US2] `sheets.go`: `measure.Report` and `measure.Parameterised` (`q`, `team`, `status`, `sort`), reading
  and checking them from `Filter.Own`
- [X] T018 [US2] e2e US2 step: the table as CSV and Excel, the report's Export

## Phase 5: Polish

- [X] T019 `BenchmarkUserAccess_SC001` (2,000 people, 200 teams, 4,000 assignments; research R5); result in plan.md
- [X] T020 Comment budget and file length of every file of this feature; lint and format
- [X] T021 `/speckit-converge` (nothing left); spec status Built; pull request #66 into `measure-module` with the feature-map status

## Dependencies

T001 → T002 – T005 → US1 (T006 – T015) → US2 (T016 – T018) → Polish. Tests before implementation in each story.
