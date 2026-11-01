# Feature Specification: User Access Report

**Feature Branch**: `feature/045-user-access-report`

**Created**: 2026-11-01

**Status**: Draft

**Input**: User description: "People with role, teams, job title and last sign-in; search, sort, export; opens
with `access.manage`. Demo sign-in records the time of a successful sign-in (migration 00029,
`people.last_signed_in_at`, UTC). Stories: US1 the report with last sign-in (P1) · US2 export (P2)." (GitHub
issue #38, `docs/sdd/measure-issues.md` §#045)

**References**: Constitution 4.8.0 — Current Phase & Scope (Measure wave A, D-366), Principles IV (the server
works out every figure), VII (access and audit), F3 (state in the address), F6 (every chart has a table) · Hive9
functional analysis §E3.9.5 (User Details and User Detail Report columns) · Shell's client spec 019 US7,
**FR-034** (User Detail Report), **FR-035** (User Details: by last sign-in band), FR-013, FR-014 (table paging and
export), FR-020 (empty causes) and the Assumption that this report alone shows e-mail addresses · features 003
(people and teams), 014 (roles, scopes, Settings › Roles and access › People; D-367 `access.manage`), 039 (the
report frame: page header, saved views, widget table and export; D-367, D-368).

The issue cites client FR-040 – FR-044; those are 019's further catalogue dashboards. The user access pages are
019 FR-034 and FR-035, cited above.

## Clarifications

Answers given while specifying are recorded here as `045-Cn` with their reasoning; the integrator moves them
into the decision log at the end of wave A (D-369).

### Session 2026-11-01

Settled with the recommended answer, as the developer asked for planned work to go ahead without stopping;
each stands until the developer or the integrator overrides it.

- **045-C1** Q: 039's filter panel selects plans (year, plan, team, owner, activity type, status, attributes),
  and people do not belong to plans. Which controls does User access have? → A: 039's page header, breadcrumb,
  saved views, widget table view and export, without the plan filter panel, Period, Show or Currency; its own
  controls are **Search**, **Team**, **Status** (Active, Inactive, All; default Active) and the table's sort and
  paging, all in the page address. *Why*: client 019 FR-034 asks for team and status filters; plan filters would
  change nothing on a list of people and would mislead. Nothing of the frame is redefined, only left out.
- **045-C2** Q: Does 045 include client FR-035's User Details summary? → A: Only its **last sign-in band**
  widget (Within 30 days, 31 – 90 days, Over 90 days, Never) with the cards People, Signed in within 30 days and
  Never signed in. *Why*: last sign-in is the one figure 014's People tab does not show and is what an access
  review asks first; people by team and by role are the table's own filters.
- **045-C3** Q: A person may hold several roles, each with a scope (014 FR-002, FR-004); Hive9 had one role. What
  does the Role column show? → A: Every role with its scope, as 014's People tab shows it ("Planner · My teams;
  Global planner · All markets"), or "No role". *Why*: MAPP's access model; one role would hide access.
- **045-C4** Q: Does recording the sign-in time change the person record? → A: No. It writes only
  `last_signed_in_at`, in the same step as the session, with no new version, no "changed by" and no activity-log
  entry. *Why*: a sign-in is not an edit; bumping the version would make an administrator's open edit of that
  person fail as changed in another tab, and the log would fill with sign-ins.

## User Scenarios & Testing *(mandatory)*

The report is **read-only**: nothing on it changes a record and opening it writes nothing except what 039's
saved views write. Its scenarios are the states the page can be in — data, empty, no access, filters exclude
everything — not create, edit or delete; people, teams and roles are changed in Settings (003, 014). The only
write of the feature is the sign-in time (US1 scenario 7), made by signing in, not by the report.

### User Story 1 - Review who has access and when they last signed in (Priority: P1)

An administrator opens **Reports › User access** to review access: every person with their e-mail, job title,
roles with scopes, teams, status and last sign-in, and how many have not signed in for a long time, so that they
can remove access nobody uses.

**Why this priority**: an access audit is a control Shell must evidence (client 019 US7); last sign-in is the
figure no other MAPP screen shows.

**Independent Test**: On the demo data, sign in as two demo people, then open the report as the administrator;
check both last sign-in times, the "Never" of the others, the band widget and the cards by hand.

**Acceptance Scenarios**:

*Data*

1. **Given** the demo people, **When** Alex Morgan (Administrator) opens the report, **Then** the table lists
   every active person with Name, E-mail, Job title, Roles ("Administrator · All markets"), Teams
   (comma-separated, alphabetical), Status and Last sign-in, sorted by name.
2. **Given** Priya signed in today at 09:12 UTC, **Then** her Last sign-in shows today's date and time in Alex's
   date format and time zone; **given** a person who never signed in, it shows "Never".
3. **Given** 6 active people, 2 signed in within 30 days and 4 never, **Then** the cards show People 6, Signed
   in within 30 days 2, Never signed in 4, and the widget **Last sign-in** shows Within 30 days 2 · 31 – 90 days
   0 · Over 90 days 0 · Never 4.
4. **Given** Team set to DE_LUBES, **Then** only members of that team are listed and counted; **given** Status
   Inactive, only inactive people; **given** All, both, inactive ones marked "Inactive".
5. **Given** a search for "planner", **Then** people whose name, e-mail, job title, role or team contains it
   are listed; the search ignores letter case.
6. **Given** sorting by Last sign-in, **Then** the most recent comes first and "Never" last; any column sorts;
   20 rows per page with page sizes 5 to 500 and "1–20 of 64".
7. **Given** a person signs in successfully, **Then** their last sign-in becomes that moment (UTC), with no
   change to their version, "last changed" or the activity log; a refused sign-in records nothing (045-C4).
8. **Given** a person with two roles, **Then** the Roles cell lists both with their scopes; **given** none, "No
   role".

*Empty*

9. **Given** no person matches the team or status chosen, **Then** the page shows "No people match these
   filters" with **Reset filters**, and the cards show 0.
10. **Given** nobody has signed in since sign-in times were first recorded, **Then** every Last sign-in shows
    "Never" and the widget says "Sign-in times are recorded from <date of migration 00029> on".

*No access*

11. **Given** a person without "Give roles and change roles and access rules", **Then** User access is not
    offered on the Reports page and its address answers with the standard "no access" page naming that
    permission (039 FR-003).
12. **Given** a person with that permission, **Then** every person is listed, whatever teams the
    administrator belongs to, because the permission applies everywhere (014 Table A).

*No financial access*

13. **Given** an administrator without "View financial data", **Then** the report is shown in full: it holds
    no financial data.

*Filters exclude everything*

14. **Given** a search that matches nobody, **Then** the table says "No people match "xyz"" with **Clear
    search**, and the cards and widget show 0.

---

### User Story 2 - Export the access list (Priority: P2)

The administrator exports the list as CSV or Excel to attach to an access review.

**Why this priority**: access reviews are filed as spreadsheets (client 019 US7, FR-014); the frame gives the
export, the report supplies the rows.

**Independent Test**: Export with a team filter and a search as CSV and as Excel and compare with the table.

**Acceptance Scenarios**:

1. **Given** a person holding "Export data", **When** they export the table, **Then** the file holds every row
   matching the search, team and status, in the current sort, with the screen's columns, Last sign-in as a date
   and time in UTC named in the heading, "Never" as an empty cell.
2. **Given** the report's **Export**, **Then** the Excel file holds the people sheet and the Last sign-in band
   sheet, then 039's About this export sheet with the search, team and status in words.
3. **Given** a person without "Export data", **Then** no export is offered (039).

---

### Edge Cases

- **Person made inactive**: listed under Inactive with their last sign-in kept; they cannot sign in (auth
  refuses inactive people), so it no longer changes.
- **Person reactivated**: listed under Active again with the last sign-in from before.
- **Person with the same name as another**: both listed; the e-mail tells them apart.
- **Team made inactive**: still named in its members' Teams cell, marked "(inactive)".
- **Role renamed or deleted in Settings**: the new name on the next load; a deleted role no longer appears.
- **Sign-in in a second tab while signed in elsewhere**: each successful sign-in records its own time; the latest
  wins.
- **Clock**: times are stored in UTC and shown in the viewer's time zone; the band uses whole days from now.
- **Demo data**: no demo person has a sign-in time until someone signs in as them; the demo data's people are
  not changed (build guide: demo data is additive).

## Requirements *(mandatory)*

### Functional Requirements

**Report and access**

- **FR-001**: The Reports area MUST offer **User access** inside 039's report page (header, breadcrumb, saved
  views, widget table view and export) without the plan filter panel, Period, Show or Currency (045-C1).
- **FR-002**: User access MUST need "Give roles and change roles and access rules" (`access.manage`); without it
  the report is not offered and its address answers with the standard "no access" page. *(D-367; 039 FR-003)*
- **FR-003**: The report MUST list every person in MAPP, active or not as the Status filter says, because
  `access.manage` applies everywhere. *(014 Table A; client 019 FR-034)*
- **FR-004**: The report MUST be read-only. *(client 019 FR-022; D-366)*

**The list**

- **FR-005**: Each row MUST show Name, E-mail, Job title, Roles (each role with its scope, or "No role"), Teams
  (every team, comma-separated, alphabetical, inactive ones marked), Status (Active or Inactive) and Last sign-in
  (date and time, or "Never"). *(client 019 FR-034; FA §E3.9.5; 045-C3)*
- **FR-006**: The report MUST offer Search (name, e-mail, job title, role, team; ignoring letter case), Team (one
  or more teams), Status (Active default, Inactive, All), sort by any column (name by default; "Never" last when
  sorting by Last sign-in) and paging (5 to 500 rows, default 20), all in the page address. *(client 019 FR-013,
  FR-034; constitution F3; 045-C1)*
- **FR-007**: The report MUST show the cards **People**, **Signed in within 30 days** and **Never signed in** and
  the widget **Last sign-in** with the bands Within 30 days, 31 – 90 days, Over 90 days and Never, over the
  people matching the search, team and status, adding up to the People card. *(client 019 FR-035; 045-C2)*
- **FR-008**: Selecting a person MUST open their page in Settings › People (014), where access is changed.

**Recording the sign-in time**

- **FR-009**: Every successful sign-in MUST record its moment, in UTC, as the person's last sign-in, in the same
  step that starts the session; a refused sign-in (wrong password, inactive person, locked attempt) MUST record
  nothing. *(issue #38; client 019 FR-034)*
- **FR-010**: Recording the sign-in time MUST NOT change the person's version, "last changed by" or "last changed
  at", and MUST NOT write an activity-log entry. *(045-C4)*
- **FR-011**: Times recorded before this feature MUST NOT be invented: people who have not signed in since show
  "Never". *(constitution IV)*

**Frame, states and export (039)**

- **FR-012**: The page MUST render one of these states: data; "No people match these filters" with **Reset
  filters**; "No people match "<search>"" with **Clear search**; no access (the standard page). *(client 019
  FR-020; 039)*
- **FR-013**: The table and band widget exports, and the report's **Export**, MUST hold every matching row in the
  current sort with the screen's columns, as 039 defines the export. *(client 019 FR-014; 039 FR-022)*

### Key Entities

- **Person's last sign-in** (new stored value): the moment of the person's latest successful sign-in, in UTC;
  empty until the first one after this feature.
- **Access row** (calculated): a person with e-mail, job title, roles with scopes, teams, status and last sign-in.
- Read from existing records: people and teams (003), roles and role assignments with scopes (014).

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: With 2,000 people (client production scale), the report shows its first page within 2 seconds, and
  each search, filter, sort or page change within 2 seconds, for 95 % of requests (client 019 SC-002).
- **SC-002**: After a successful sign-in, the report shows that person's last sign-in to the minute on the next
  load, for 100 % of sign-ins; after a refused one it is unchanged.
- **SC-003**: Every row's roles, scopes and teams equal Settings › Roles and access › People for that person, for
  100 % of people.
- **SC-004**: An administrator lists the active people who have not signed in for over 90 days and exports them,
  within 2 minutes on the first attempt.
- **SC-005**: Every export equals the table, row for row.
- **SC-006**: The page passes WCAG 2.1 AA with no critical or serious finding, and the band widget has a table
  view (client 019 SC-010).

## Assumptions

- **One Name column**: MAPP keeps one name per person (003); Hive9's first and last name columns become Name.
- **E-mail addresses are shown**: this report is an access audit, the one place client 019 allows them.
- **No sign-in history**: only the latest sign-in is kept; a history of sign-ins is out of scope.
- **Sign-in is the demo sign-in** (auth, D-361); a single sign-on added later records the same value.
- 039's report page and export are in place before 045's web part is built (build guide §3); names follow 039's
  contract as pushed on `feature/039-measure-hub` (checked 2026-11-01).
