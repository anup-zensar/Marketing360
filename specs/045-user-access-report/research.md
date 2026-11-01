# Research: User Access Report (045)

Only the real choices are recorded here (build guide §8). Each states the decision, why, and what was rejected.
Code references are as of `measure-module` at `03c8f6ab` and `feature/039-measure-hub` at `f5c15d74`.

## R1 — Roles, scopes and teams from the People tab's own read

**Decision**: the rows start from `access.Service.ListPeopleAccess` — the read behind Settings › Roles and access
› People (014 FR-017): every person with `authz.PersonRef` (name, job title, active), teams (`authz.TeamRef` with
`Active`) and role assignments with their scope text ("All markets", "My teams", named teams). E-mail comes from
`people.Store.List`; last sign-in from auth (R2). `app/measure_useraccess.go` adapts the three into
`useraccess`'s ports.

**Rationale**: SC-003 asks for exactly the People tab's roles, scopes and teams; reading through the same function
makes it true by construction and keeps 014's scope wording in one place. `access.manage` applies everywhere (014
Table A, not scoped), so "everyone" is the right set.

**Rejected**: a `measure_useraccess.sql` join of people, memberships, assignments and roles (issue #38 lists no
query file for 045; it would copy 014's scope text rule).

## R2 — Last sign-in: one column, written by auth in the session's step

**Decision**: migration **00029** adds `people.last_signed_in_at timestamptz NULL` (no default, no backfill). A new
query file `queries/sign_ins.sql` (auth's) holds `RecordSignIn` — `UPDATE people SET last_signed_in_at = $2 WHERE
id = $1` — and `ListLastSignIns` — `SELECT id, last_signed_in_at FROM people WHERE last_signed_in_at IS NOT NULL`.
`auth.Store` gains `RecordSignIn(ctx, personID, at)` and `LastSignIns(ctx) (map[string]time.Time, error)`;
`Service.startSession` calls `RecordSignIn(ctx, person.ID, now)` inside its existing `Write`, after the
credentials pass and before the session is put, with `now` in UTC.

**Rationale**: issue #38 names the column and the migration. Writing in the same `Write` makes the time and the
session one step: a sign-in that fails to start its session records nothing (FR-009). The update touches neither
`version` nor `updated_*`, and `PutPerson` names its columns explicitly, so a Settings edit never overwrites the
time and the time never makes an edit stale (045-C4). A separate query file leaves `sessions.sql` and
`people.sql` untouched.

**Rejected**: a `sign_ins` history table (out of scope: only the latest is asked); deriving it from `sessions`
(sessions are removed on sign-out and expiry); writing through `people.Store.Put` (bumps the version, logs a
change).

## R3 — Controls: the report page without the plan filter panel (045-C1)

**Decision**: the report uses 039's page and export but not `allwork`'s filter or the frame's matching: the
handler reads `q` (search), `team` (several), `status` (`active` default, `inactive`, `all`), `sort`, `page` and
`pageSize`. `Sheets` gets 039's `measure.Filter` and reads the same parameters from it only if 039 carries report
parameters; otherwise the export holds every active person in name order (hand-off below).

**Rationale**: plan filters would change nothing on people (spec 045-C1); client 019 FR-034 asks for team and
status.

**Hand-off to lane A**: `ReportFrame` must allow a report without the filter panel, In view line and Currency, and
039's export must pass report parameters (`q`, `team`, `status`, `sort`) to `Sheets` — the same question 044
raised for `tab` and `sort`. Neither blocks the spec or the backend; the web part waits for the answer.

## R4 — Filtering, bands, sort and paging in memory

**Decision**: the service builds every row (at most a few thousand people), filters by search, team and status,
counts the cards and bands over the filtered rows, sorts and pages. Bands use whole days between the sign-in and
`now` (injected clock): Within 30 days (≤ 30), 31 – 90 days, Over 90 days, Never.

**Rationale**: 2,000 people at client scale; one row builder for screen and export (SC-005).

## R5 — Benchmark inside the package

**Decision**: `BenchmarkUserAccess_SC001` in `useraccess` runs the service over 2,000 people, 200 teams, 5
roles and 4,000 assignments through fakes. Target: under 100 ms, keeping SC-001's 2 s with the three reads.

**Rationale**: `internal/platform/apibench` is not a lane D path (043 R5).

## R2 addendum (implementation) — a table of its own, and when recording started

**Decision**: migration 00029 creates `sign_ins (person_id, last_signed_in_at)`, one row per person upserted by sign-in,
instead of a `people` column: a new `people` column changes the rows sqlc generates for 003's people queries, and so
the people package, which is outside issue #38's paths. Apart from the person, the time cannot move their version or
"last changed" (FR-010) by construction. It also creates `sign_in_recording`, one row stamped with the moment the migration runs;
`GetSignInsRecordedSince` reads it for the contract's `recordedSince` ("Never" means "not since this date").

**Rationale**: the true start is when 00029 ran in that database. goose's own version table is outside the schema
sqlc reads, and the earliest stored sign-in moves forward as people sign in again.

**Rejected**: a date in code (wrong for every later environment); the earliest `last_signed_in_at` (moves).

## R3 answered by lane A (2026-11-01)

`measure.Parameterised` carries `q`, `team`, `status` and `sort` to saved views and, through `Filter.Own`, to
`Sheets`, which checks them. `access.manage` is checked once, in 039's `openable`, for the frame and the export, so
045 adds no check of its own. `ReportFrame` gains an optional prop that hides the filter panel, In view, chips and
Currency (045 adds it, with a test, after pinging lane A).
