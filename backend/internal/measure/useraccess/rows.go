package useraccess

import (
	"cmp"
	"slices"
	"strings"
	"time"

	"mapp/internal/platform/authz"
)

const (
	sortEmail    = "email"
	sortJobTitle = "jobTitle"
	sortRoles    = "roles"
	sortTeams    = "teams"
	sortStatus   = "status"
	sortSignedIn = "lastSignedIn"
)

// rows joins each person with their latest sign-in, if any since recording started (FR-005, FR-011).
func rows(people []Person, signIns map[string]time.Time) []Row {
	list := make([]Row, 0, len(people))
	for _, p := range people {
		row := Row{Person: p.Ref, Email: p.Email, Roles: p.Roles, Teams: slices.Clone(p.Teams)}
		if at, found := signIns[p.Ref.ID]; found {
			row.LastSignedInAt = &at
		}
		slices.SortFunc(row.Teams, func(a, b authz.TeamRef) int { return compareText(a.Name, b.Name) })
		list = append(list, row)
	}
	return list
}

// matching keeps the rows of the search (name, e-mail, job title, role, team; any case), the teams and the status.
func matching(list []Row, q Query) []Row {
	search := strings.ToLower(q.Search)
	return slices.DeleteFunc(slices.Clone(list), func(r Row) bool {
		return !statusMatches(r, q.Status) || !inTeams(r, q.Teams) || (search != "" && !strings.Contains(searchText(r), search))
	})
}

func statusMatches(r Row, status string) bool {
	return status == StatusAll || (status == StatusActive) == r.Person.Active
}

func inTeams(r Row, teams []string) bool {
	return len(teams) == 0 || slices.ContainsFunc(r.Teams, func(t authz.TeamRef) bool { return slices.Contains(teams, t.ID) })
}

func searchText(r Row) string {
	parts := []string{r.Person.Name, r.Email, r.Person.JobTitle, rolesText(r), teamsText(r)}
	return strings.ToLower(strings.Join(parts, "\n"))
}

func rolesText(r Row) string {
	names := make([]string, 0, len(r.Roles))
	for _, a := range r.Roles {
		names = append(names, a.Role.Name)
	}
	return strings.Join(names, ", ")
}

func teamsText(r Row) string {
	names := make([]string, 0, len(r.Teams))
	for _, t := range r.Teams {
		names = append(names, t.Name)
	}
	return strings.Join(names, ", ")
}

func columns() map[string]func(a, b Row) int {
	return map[string]func(a, b Row) int{
		sortName:     func(a, b Row) int { return compareText(a.Person.Name, b.Person.Name) },
		sortEmail:    func(a, b Row) int { return compareText(a.Email, b.Email) },
		sortJobTitle: func(a, b Row) int { return compareText(a.Person.JobTitle, b.Person.JobTitle) },
		sortRoles:    func(a, b Row) int { return compareText(rolesText(a), rolesText(b)) },
		sortTeams:    func(a, b Row) int { return compareText(teamsText(a), teamsText(b)) },
		sortStatus:   func(a, b Row) int { return cmp.Compare(inactive(a), inactive(b)) },
		sortSignedIn: func(a, b Row) int { return a.LastSignedInAt.Compare(*b.LastSignedInAt) },
	}
}

// sortRows sorts by one column, "-" first for descending; "Never" is always last, ties by person ID (FR-006).
func sortRows(list []Row, sort string) {
	key, sign := strings.TrimPrefix(sort, descending), 1
	if strings.HasPrefix(sort, descending) {
		sign = -1
	}
	by := columns()[key]
	slices.SortFunc(list, func(a, b Row) int {
		if key == sortSignedIn && (a.LastSignedInAt == nil || b.LastSignedInAt == nil) {
			return cmp.Or(cmp.Compare(never(a), never(b)), compareIDs(a.Person.ID, b.Person.ID))
		}
		return cmp.Or(sign*by(a, b), compareIDs(a.Person.ID, b.Person.ID))
	})
}

func never(r Row) int    { return boolInt(r.LastSignedInAt == nil) }
func inactive(r Row) int { return boolInt(!r.Person.Active) }

func boolInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func pageOf(list []Row, page, size int) []Row {
	from := min((page-1)*size, len(list))
	return list[from:min(from+size, len(list))]
}

func compareText(a, b string) int {
	return strings.Compare(strings.ToLower(a), strings.ToLower(b))
}

// compareIDs puts U-999 before U-1000.
func compareIDs(a, b string) int {
	return cmp.Or(cmp.Compare(len(a), len(b)), strings.Compare(a, b))
}
