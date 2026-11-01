package customdashboard

import (
	"cmp"
	"context"
	"fmt"
	"slices"
	"strings"
	"unicode/utf8"

	"mapp/internal/history"
	"mapp/internal/platform/actor"
	"mapp/internal/platform/authz"
	"mapp/internal/platform/errs"
)

// Share sets the teams a dashboard is shared with: each must be one the owner belongs to now or one already
// shared, so leaving a team does not force the owner to unshare it.
func (s *Service) Share(ctx context.Context, a actor.Actor, id string, in SharingInput) (View, error) {
	var view View
	err := s.tx.Write(ctx, func(ctx context.Context) error {
		stored, err := s.owned(ctx, a, id)
		if err != nil {
			return err
		}
		teams := sortedIDs(in.TeamIDs)
		if err = s.checkSharing(ctx, stored, teams, in.Version); err != nil {
			return err
		}
		if err = stored.CheckVersion(*in.Version, dashboardNoun); err != nil {
			return err
		}
		changed := stored
		if !slices.Equal(stored.SharedWith, teams) {
			if changed, err = s.reshare(ctx, a, stored, teams); err != nil {
				return err
			}
		}
		view, err = s.view(ctx, a, changed)
		return err
	})
	return view, err
}

func (s *Service) checkSharing(ctx context.Context, d Dashboard, teams []string, version *int) error {
	allowed, err := s.shareable(ctx, d)
	if err != nil {
		return err
	}
	var problems errs.Problems
	switch {
	case len(teams) > maxSharedTeams:
		problems.Add(teamIDsField, msgTooManyTeams)
	case slices.ContainsFunc(teams, func(t string) bool { return !slices.Contains(allowed, t) }):
		problems.Add(teamIDsField, msgShareOwnTeams)
	}
	if version == nil {
		problems.Add(versionField, msgEnterVersion)
	}
	return problems.Err()
}

// shareable is the owner's teams now and the teams already shared, sorted by id.
func (s *Service) shareable(ctx context.Context, d Dashboard) ([]string, error) {
	mine, err := s.directory.TeamsOf(ctx, d.OwnerID)
	if err != nil {
		return nil, err
	}
	return sortedIDs(append(slices.Clone(mine), d.SharedWith...)), nil
}

func (s *Service) reshare(ctx context.Context, a actor.Actor, stored Dashboard, teams []string) (Dashboard, error) {
	changed := stored
	changed.SharedWith = teams
	changed.Touch(a, s.now())
	if err := s.repo.Update(ctx, changed, stored.Version); err != nil {
		return Dashboard{}, err
	}
	if err := s.repo.ReplaceShares(ctx, changed.ID, teams); err != nil {
		return Dashboard{}, err
	}
	details, err := s.sharingText(ctx, stored.SharedWith, teams)
	if err != nil {
		return Dashboard{}, err
	}
	_, err = s.history.Log(ctx, a, history.LogInput{
		Kind: kindDashboard, RecordID: changed.ID, RecordName: changed.Name, Action: history.ActionChanged, Details: details,
	})
	return changed, err
}

// sharingText says "Shared with Lubricants Europe. No longer shared with Mobility China" for the log.
func (s *Service) sharingText(ctx context.Context, before, after []string) (string, error) {
	refs, err := s.directory.TeamRefs(ctx, append(slices.Clone(before), after...))
	if err != nil {
		return "", err
	}
	namesOf := func(ids []string, other []string) string {
		var names []string
		for _, id := range ids {
			if !slices.Contains(other, id) {
				names = append(names, cmp.Or(refs[id].Name, id))
			}
		}
		return strings.Join(names, ", ")
	}
	var parts []string
	if added := namesOf(after, before); added != "" {
		parts = append(parts, fmt.Sprintf(msgSharedWith, added))
	}
	if removed := namesOf(before, after); removed != "" {
		parts = append(parts, fmt.Sprintf(msgNoLongerSharedWith, removed))
	}
	return strings.Join(parts, ". "), nil
}

// Copy makes a dashboard owned by the person, shared with nobody, from one they may see.
func (s *Service) Copy(ctx context.Context, a actor.Actor, id string, in CopyInput) (View, error) {
	if err := authz.Require(ctx, authz.DashboardsBuild); err != nil {
		return View{}, err
	}
	var view View
	err := s.tx.Write(ctx, func(ctx context.Context) error {
		original, err := s.visible(ctx, a, id)
		if err != nil {
			return err
		}
		name, err := s.copyName(ctx, a, original.Name, in.Name)
		if err != nil {
			return err
		}
		d := normalised(Dashboard{Name: name, Description: original.Description, OwnerID: a.ID,
			Filters: original.Filters, Widgets: original.Widgets})
		if d, err = s.insert(ctx, a, d); err != nil {
			return err
		}
		if _, err = s.history.Log(ctx, a, history.LogInput{Kind: kindDashboard, RecordID: d.ID, RecordName: d.Name,
			Action: history.ActionCopied, Details: fmt.Sprintf(msgCopiedFrom, original.Name, original.ID)}); err != nil {
			return err
		}
		view, err = s.view(ctx, a, d)
		return err
	})
	return view, err
}

// copyName checks a chosen name, or numbers "Copy of <name>" until it is free, within 80 characters.
func (s *Service) copyName(ctx context.Context, a actor.Actor, original string, chosen *string) (string, error) {
	if chosen != nil {
		name := strings.TrimSpace(*chosen)
		var problems errs.Problems
		checkName(name, &problems)
		return name, problems.Err()
	}
	return s.freeName(ctx, a.ID, copyPrefix+original, "")
}

// freeName returns base, or base numbered " (2)", " (3)", …, shortened to fit 80 characters and not used by
// another dashboard of ownerID than exceptID.
func (s *Service) freeName(ctx context.Context, ownerID, base, exceptID string) (string, error) {
	candidate := shortened(base, maxNameLength)
	for n := 2; n < maxRestoredNameAttempt; n++ {
		taken, err := s.repo.NameTaken(ctx, ownerID, candidate, exceptID)
		if err != nil || !taken {
			return candidate, err
		}
		suffix := fmt.Sprintf(" (%d)", n)
		candidate = shortened(base, maxNameLength-len(suffix)) + suffix
	}
	return "", errs.NameTaken(fmt.Sprintf(msgNameTaken, base))
}

func shortened(text string, limit int) string {
	if utf8.RuneCountInString(text) <= limit {
		return text
	}
	return strings.TrimSpace(string([]rune(text)[:limit]))
}

// UsingDataset lists the dashboards the person may see whose widgets read the dataset.
func (s *Service) UsingDataset(ctx context.Context, a actor.Actor, datasetID string) ([]Ref, error) {
	var refs []Ref
	err := s.tx.Read(ctx, func(ctx context.Context) error {
		var err error
		refs, err = s.repo.UsingDataset(ctx, a.ID, datasetID)
		return err
	})
	return refs, err
}

func sortedIDs(list []string) []string {
	sorted := slices.Clone(orEmpty(list))
	slices.SortFunc(sorted, func(a, b string) int {
		if len(a) != len(b) {
			return len(a) - len(b)
		}
		return strings.Compare(a, b)
	})
	return slices.Compact(sorted)
}
