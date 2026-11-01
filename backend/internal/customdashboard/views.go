package customdashboard

import (
	"cmp"
	"context"
	"fmt"
	"slices"
	"strings"

	"mapp/internal/people"
	"mapp/internal/platform/actor"
	"mapp/internal/platform/authz"
)

func (s *Service) view(ctx context.Context, a actor.Actor, d Dashboard) (View, error) {
	isOwner := d.OwnerID == a.ID
	shareable := []string{}
	if isOwner {
		var err error
		if shareable, err = s.shareable(ctx, d); err != nil {
			return View{}, err
		}
	}
	owners, err := s.directory.PersonRefs(ctx, []string{d.OwnerID})
	if err != nil {
		return View{}, err
	}
	teams, err := s.directory.TeamRefs(ctx, append(slices.Clone(d.SharedWith), shareable...))
	if err != nil {
		return View{}, err
	}
	owner := personRef(owners, d.OwnerID)
	return View{
		ID: d.ID, Name: d.Name, Description: d.Description, Owner: owner, IsOwner: isOwner,
		SharedWith: teamList(teams, d.SharedWith), ShareableTeams: teamList(teams, shareable), Filters: d.Filters,
		Widgets: d.Widgets, Access: accessOf(ctx, isOwner, owner.Name), Meta: d.Meta,
	}, nil
}

func (s *Service) summaries(ctx context.Context, a actor.Actor, list []Dashboard) (List, error) {
	var ownerIDs, teamIDs []string
	for _, d := range list {
		ownerIDs = append(ownerIDs, d.OwnerID)
		teamIDs = append(teamIDs, d.SharedWith...)
	}
	owners, err := s.directory.PersonRefs(ctx, sortedIDs(ownerIDs))
	if err != nil {
		return List{}, err
	}
	teams, err := s.directory.TeamRefs(ctx, sortedIDs(teamIDs))
	if err != nil {
		return List{}, err
	}
	answer := List{Mine: []Summary{}, SharedWithMe: []Summary{}}
	for _, d := range slices.SortedFunc(slices.Values(list), byName) {
		owner := personRef(owners, d.OwnerID)
		summary := Summary{ID: d.ID, Name: d.Name, Description: d.Description, Owner: owner,
			SharedWith: teamList(teams, d.SharedWith), WidgetCount: len(d.Widgets), UpdatedAt: d.UpdatedAt,
			Access: accessOf(ctx, d.OwnerID == a.ID, owner.Name)}
		if d.OwnerID == a.ID {
			answer.Mine = append(answer.Mine, summary)
		} else {
			answer.SharedWithMe = append(answer.SharedWithMe, summary)
		}
	}
	return answer, nil
}

func byName(a, b Dashboard) int {
	return cmp.Or(strings.Compare(strings.ToLower(a.Name), strings.ToLower(b.Name)), strings.Compare(a.ID, b.ID))
}

// accessOf leaves Copy to anyone who builds dashboards and everything else to the owner.
func accessOf(ctx context.Context, isOwner bool, ownerName string) Access {
	build := authz.AvailableAnywhere(ctx, authz.DashboardsBuild)
	owned := build
	if !isOwner {
		owned = authz.Unavailable(fmt.Sprintf(msgNotOwner, ownerName))
	}
	return Access{Edit: owned, Share: owned, Delete: owned, Copy: build}
}

func (s *Service) ownerName(ctx context.Context, ownerID string) (string, error) {
	owners, err := s.directory.PersonRefs(ctx, []string{ownerID})
	return personRef(owners, ownerID).Name, err
}

func personRef(refs map[string]people.PersonRef, id string) people.PersonRef {
	ref, found := refs[id]
	if !found {
		return people.PersonRef{ID: id, Name: id}
	}
	return ref
}

func teamList(refs map[string]people.TeamRef, ids []string) []people.TeamRef {
	list := make([]people.TeamRef, 0, len(ids))
	for _, id := range ids {
		ref, found := refs[id]
		if !found {
			ref = people.TeamRef{ID: id, Name: id}
		}
		list = append(list, ref)
	}
	slices.SortFunc(list, func(a, b people.TeamRef) int {
		return cmp.Or(strings.Compare(strings.ToLower(a.Name), strings.ToLower(b.Name)), strings.Compare(a.ID, b.ID))
	})
	return list
}
