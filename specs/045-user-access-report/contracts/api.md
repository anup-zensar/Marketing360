# API Contract: User Access Report (045)

For the **wave A contract** (lane A writes it into `backend/api/openapi.yaml`, tag **Reports**). It follows
039's convention: `GET /api/v1/reports/<path>` returns the report's widgets, and the report's
`measure.Report.Sheets` feeds 039's shared export, so 045 adds **no export operation of its own**. Errors use the
one envelope; 401 when signed out, 403 `forbidden` when the route rule refuses.

## Operations

| Operation | Query | Answer | Rule (`access/routes_measure.go`) |
|---|---|---|---|
| `GET /api/v1/reports/user-access` (`getUserAccessReport`) | `q` (search, at most 100 characters), `team` (several team IDs), `status` (`active` default, `inactive`, `all`), `sort` (a column key, `-` for descending; default `name`), `page` (from 1), `pageSize` (5–500, default 20). 039's plan filter parameters are ignored | 200 `UserAccessReport`; 400 problems for an unknown `status` or `sort`, `pageSize` out of range or `q` too long | `need(authz.AccessManage)` (D-367; the catalogue lists the report only to holders of `access.manage`) |

**Export** is 039's `GET /api/v1/reports/user-access/export` (`need(authz.DataExport)`): widgets `people` and
`signInBands`; every matching row in the current sort (research R3). That route needs only `data.export`, so
`useraccess`'s `Sheets` MUST refuse an actor without `access.manage` with 403 `forbidden` and the permission's
refusal noun before reading anything (constitution VII; `TestSheets_RefusesWithoutAccessManage_FR002`).

Sign-in is unchanged on the wire: `POST /api/v1/auth/sign-in` answers as today; it now also records the time
(research R2). No contract change.

## Schemas

```yaml
UserAccessReport:
  required: [total, page, pageSize, sort, cards, bands, rows, recordedSince]
  properties:
    total:    { type: integer }           # matching people across all pages
    page:     { type: integer }
    pageSize: { type: integer }
    sort:     { type: string }            # as applied
    cards:
      required: [people, signedInWithin30Days, neverSignedIn]
      properties: { people: integer, signedInWithin30Days: integer, neverSignedIn: integer }
    bands:    { type: array, items: { key: { enum: [within30Days, days31to90, over90Days, never] }, label: string, count: integer } }
    recordedSince: { type: string, format: date }   # when sign-in times started to be recorded (migration 00029)
    rows:     { type: array, items: { $ref: UserAccessRow } }

UserAccessRow:
  required: [person, email, roles, teams, lastSignedInAt]
  properties:
    person:         { $ref: PersonRef }   # id, name, initials, jobTitle, active (014)
    email:          { type: string }
    roles:          { type: array, items: { $ref: RoleAssignment } }   # role with scope, as the People tab (014)
    teams:          { type: array, items: { $ref: TeamRef } }          # alphabetical; active false → "(inactive)"
    lastSignedInAt: { type: string, format: date-time, nullable: true } # UTC; null → "Never"
```

Sort keys: `name`, `email`, `jobTitle`, `roles` (first role's name), `teams` (first team's name), `status`,
`lastSignedIn` (most recent first when descending; null always last). Ties break by person ID.

## Export sheets (`measure.Report.Sheets`)

`people`: Name, E-mail, Job title, Roles ("Planner · My teams; Global planner · All markets"), Teams
(comma-separated), Status, Last sign-in (UTC) — empty for Never. `signInBands`: Band, People.

## Go interfaces (`internal/measure/useraccess/ports.go`)

```go
// Access is 014's People tab read (research R1).
type Access interface {
	ListPeopleAccess(ctx context.Context, a actor.Actor) ([]access.PersonRoles, error)
}

// People gives the e-mail addresses (003).
type People interface {
	List(ctx context.Context) ([]people.Person, error)
}

// SignIns is auth's read of the last sign-in times (research R2).
type SignIns interface {
	LastSignIns(ctx context.Context) (map[string]time.Time, error)
}
```

`auth.Store` gains `RecordSignIn(ctx context.Context, personID string, at time.Time) error` and
`LastSignIns(ctx context.Context) (map[string]time.Time, error)`.
