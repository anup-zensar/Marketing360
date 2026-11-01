package dataset

import (
	"context"
	"slices"

	"mapp/internal/people"
	"mapp/internal/platform/actor"
	"mapp/internal/platform/money"
)

func (s *Service) view(ctx context.Context, a actor.Actor, d Dataset) (View, error) {
	isOwner := d.OwnerID == a.ID
	shareable, err := s.shareableIDs(ctx, a, d)
	if err != nil {
		return View{}, err
	}
	persons, teams, err := s.refs(ctx, []Dataset{d}, shareable)
	if err != nil {
		return View{}, err
	}
	usedBy, err := s.usage.UsingDataset(ctx, a, d.ID)
	if err != nil {
		return View{}, err
	}
	return View{
		ID: d.ID, Name: d.Name, Description: d.Description, Owner: persons[d.OwnerID], IsOwner: isOwner,
		SharedWith: teamsOf(teams, d.SharedWith), ShareableTeams: teamsOf(teams, shareable), FileName: d.FileName,
		SheetName: d.SheetName, RowCount: d.RowCount, Columns: d.Columns, UsedBy: usedBy, Access: accessOf(isOwner),
		Meta: d.Meta,
	}, nil
}

// shareableIDs are the owner's teams now and the teams already shared; nobody else may share.
func (s *Service) shareableIDs(ctx context.Context, a actor.Actor, d Dataset) ([]string, error) {
	if d.OwnerID != a.ID {
		return []string{}, nil
	}
	own, err := s.directory.TeamIDsOf(ctx, a.ID)
	if err != nil {
		return nil, err
	}
	ids := slices.Concat(own, d.SharedWith)
	slices.Sort(ids)
	return slices.Compact(ids), nil
}

func (s *Service) summaries(ctx context.Context, a actor.Actor, datasets []Dataset) ([]Summary, error) {
	persons, teams, err := s.refs(ctx, datasets, nil)
	if err != nil {
		return nil, err
	}
	summaries := make([]Summary, 0, len(datasets))
	for _, d := range datasets {
		summaries = append(summaries, Summary{
			ID: d.ID, Name: d.Name, Description: d.Description, Owner: persons[d.OwnerID],
			SharedWith: teamsOf(teams, d.SharedWith), RowCount: d.RowCount, ColumnCount: len(d.Columns),
			UpdatedAt: d.UpdatedAt, Access: accessOf(d.OwnerID == a.ID),
		})
	}
	return summaries, nil
}

func (s *Service) refs(ctx context.Context, datasets []Dataset, moreTeams []string) (map[string]people.PersonRef,
	map[string]people.TeamRef, error) {
	ownerIDs, teamIDs := []string{}, slices.Clone(moreTeams)
	for _, d := range datasets {
		ownerIDs = append(ownerIDs, d.OwnerID)
		teamIDs = append(teamIDs, d.SharedWith...)
	}
	slices.Sort(ownerIDs)
	slices.Sort(teamIDs)
	persons, err := s.directory.PersonRefs(ctx, slices.Compact(ownerIDs))
	if err != nil {
		return nil, nil, err
	}
	teams, err := s.directory.TeamRefs(ctx, slices.Compact(teamIDs))
	return persons, teams, err
}

func teamsOf(refs map[string]people.TeamRef, ids []string) []people.TeamRef {
	teams := make([]people.TeamRef, 0, len(ids))
	for _, id := range ids {
		if t, found := refs[id]; found {
			teams = append(teams, t)
		}
	}
	return teams
}

func rowTexts(columns []Column, r Row) []*string {
	texts := make([]*string, len(columns))
	for i := range columns {
		if i >= len(r) || r[i].Empty {
			continue
		}
		text := r[i].Text
		if r[i].Minor != nil {
			text = money.FromCents(*r[i].Minor).String()
		}
		texts[i] = &text
	}
	return texts
}
