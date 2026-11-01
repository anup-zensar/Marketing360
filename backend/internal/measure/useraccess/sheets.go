package useraccess

import (
	"cmp"
	"context"
	"net/url"
	"strings"

	"mapp/internal/measure"
	"mapp/internal/platform/actor"
	"mapp/internal/platform/sheet"
)

// The widgets of the export, in screen order (contract: people, signInBands).
const (
	widgetPeople = "people"
	widgetBands  = "signInBands"
)

func (s *Service) Key() string     { return measure.KeyUserAccess }
func (s *Service) Available() bool { return true }

// Params are the report's own address parameters that saved views keep and the export passes on; never the page.
func (s *Service) Params() []string { return []string{paramSearch, paramTeam, paramStatus, paramSort} }

// Sheets gives both widgets with every matching row in the address's sort (FR-013); 039 checks access.manage.
func (s *Service) Sheets(ctx context.Context, a actor.Actor, f measure.Filter) ([]measure.WidgetSheet, error) {
	q, err := ParseQuery(f.Own)
	if err != nil {
		return nil, err
	}
	var sheets []measure.WidgetSheet
	err = s.tx.Read(ctx, func(ctx context.Context) error {
		list, _, readErr := s.matching(ctx, a, q)
		bands, _ := bandsAndCards(list, s.now())
		sheets = []measure.WidgetSheet{
			{Key: widgetPeople, Sheet: sheet.Sheet{Name: "People", Rows: peopleTable(list)}},
			{Key: widgetBands, Sheet: sheet.Sheet{Name: "Last sign-in", Rows: bandsTable(bands)}},
		}
		return readErr
	})
	if err != nil {
		return nil, err
	}
	return sheets, nil
}

// peopleTable has the screen's columns; Never is an empty cell.
func peopleTable(list []Row) [][]sheet.Cell {
	table := [][]sheet.Cell{{sheet.Bold("Name"), sheet.Bold("E-mail"), sheet.Bold("Job title"), sheet.Bold("Roles"),
		sheet.Bold("Teams"), sheet.Bold("Status"), sheet.Bold("Last sign-in (UTC)")}}
	for _, r := range list {
		table = append(table, []sheet.Cell{sheet.Text(r.Person.Name), sheet.Text(r.Email), sheet.Text(r.Person.JobTitle),
			sheet.Text(rolesWithScopes(r)), sheet.Text(teamsMarked(r)), sheet.Text(statusWord(r)), signedIn(r)})
	}
	return table
}

func bandsTable(bands []Band) [][]sheet.Cell {
	table := [][]sheet.Cell{{sheet.Bold("Last sign-in"), sheet.Bold("People")}}
	total := 0
	for _, b := range bands {
		table = append(table, []sheet.Cell{sheet.Text(b.Label), sheet.Number(int64(b.Count))})
		total += b.Count
	}
	return append(table, []sheet.Cell{sheet.Bold("Total"), sheet.Number(int64(total))})
}

func rolesWithScopes(r Row) string {
	if len(r.Roles) == 0 {
		return "No role"
	}
	parts := make([]string, 0, len(r.Roles))
	for _, a := range r.Roles {
		parts = append(parts, a.Role.Name+" · "+a.Scope.Text)
	}
	return strings.Join(parts, "; ")
}

func teamsMarked(r Row) string {
	parts := make([]string, 0, len(r.Teams))
	for _, t := range r.Teams {
		name := t.Name
		if !t.Active {
			name += " (inactive)"
		}
		parts = append(parts, name)
	}
	return strings.Join(parts, ", ")
}

func statusWord(r Row) string {
	if r.Person.Active {
		return "Active"
	}
	return "Inactive"
}

func signedIn(r Row) sheet.Cell {
	if r.LastSignedInAt == nil {
		return sheet.Cell{}
	}
	return sheet.DateTime(*r.LastSignedInAt)
}

// Describe says the report's own filters in About this export, after Filters (US2-2): search, teams and status.
func (s *Service) Describe(ctx context.Context, a actor.Actor, own url.Values) ([][2]string, error) {
	q, err := ParseQuery(own)
	if err != nil {
		return nil, err
	}
	var lines [][2]string
	if q.Search != "" {
		lines = append(lines, [2]string{"Search", q.Search})
	}
	if len(q.Teams) > 0 {
		names, readErr := s.teamNames(ctx, a, q.Teams)
		if readErr != nil {
			return nil, readErr
		}
		lines = append(lines, [2]string{"Team", strings.Join(names, ", ")})
	}
	status := map[string]string{StatusActive: "Active", StatusInactive: "Inactive", StatusAll: "All"}[q.Status]
	return append(lines, [2]string{"Status", status}), nil
}

// teamNames names the chosen teams as people's teams show them; a team no one is in keeps its ID.
func (s *Service) teamNames(ctx context.Context, a actor.Actor, ids []string) ([]string, error) {
	known := map[string]string{}
	err := s.tx.Read(ctx, func(ctx context.Context) error {
		people, readErr := s.people.PeopleAccess(ctx, a)
		for _, p := range people {
			for _, t := range p.Teams {
				known[t.ID] = t.Name
			}
		}
		return readErr
	})
	names := make([]string, 0, len(ids))
	for _, id := range ids {
		names = append(names, cmp.Or(known[id], id))
	}
	return names, err
}
