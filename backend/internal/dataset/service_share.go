package dataset

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"mapp/internal/history"
	"mapp/internal/platform/actor"
	"mapp/internal/platform/errs"
)

const (
	teamIDsField = "teamIds"
	maxShares    = 20
)

type sharing struct {
	SharedWith []string `json:"sharedWith"`
}

// Share sets the teams the dataset is shared with, view-only: the owner's teams now, or teams already shared.
func (s *Service) Share(ctx context.Context, a actor.Actor, id string, in SharingInput) (View, error) {
	var view View
	err := s.tx.Write(ctx, func(ctx context.Context) error {
		stored, err := s.owned(ctx, a, id)
		if err != nil {
			return err
		}
		teamIDs, err := s.checkSharing(ctx, a, stored, in)
		if err != nil {
			return err
		}
		if slices.Equal(teamIDs, stored.SharedWith) {
			view, err = s.view(ctx, a, stored)
			return err
		}
		changed, err := s.reshare(ctx, a, stored, teamIDs)
		if err != nil {
			return err
		}
		view, err = s.view(ctx, a, changed)
		return err
	})
	if err != nil {
		return View{}, err
	}
	return view, nil
}

func (s *Service) checkSharing(ctx context.Context, a actor.Actor, stored Dataset, in SharingInput) ([]string, error) {
	allowed, err := s.shareableIDs(ctx, a, stored)
	if err != nil {
		return nil, err
	}
	teamIDs := slices.Clone(in.TeamIDs)
	slices.Sort(teamIDs)
	teamIDs = slices.Compact(teamIDs)
	var problems errs.Problems
	if len(teamIDs) > maxShares || slices.ContainsFunc(teamIDs, func(t string) bool { return !slices.Contains(allowed, t) }) {
		problems.Add(teamIDsField, msgOnlyYourTeams)
	}
	if in.Version == nil {
		problems.Add(versionField, msgVersion)
	}
	if err := problems.Err(); err != nil {
		return nil, err
	}
	return teamIDs, stored.CheckVersion(*in.Version, kindSingular)
}

func (s *Service) reshare(ctx context.Context, a actor.Actor, stored Dataset, teamIDs []string) (Dataset, error) {
	changed := stored
	changed.SharedWith = teamIDs
	changed.Touch(a, s.now())
	if err := s.repo.Update(ctx, changed); err != nil {
		return Dataset{}, err
	}
	if err := s.repo.ReplaceShares(ctx, changed.ID, teamIDs); err != nil {
		return Dataset{}, err
	}
	changes, err := history.Diff(sharing{stored.SharedWith}, sharing{teamIDs})
	if err != nil {
		return Dataset{}, err
	}
	details, err := s.sharingDetails(ctx, stored.SharedWith, teamIDs)
	if err != nil {
		return Dataset{}, err
	}
	_, err = s.history.Log(ctx, a, history.LogInput{Kind: kindDataset, RecordID: changed.ID, RecordName: changed.Name,
		Action: history.ActionChanged, Changes: changes, Details: details})
	return changed, err
}

// sharingDetails says it as the log reads: "Shared with Lubricants Europe; No longer shared with Mobility China".
func (s *Service) sharingDetails(ctx context.Context, before, after []string) (string, error) {
	added := slices.DeleteFunc(slices.Clone(after), func(t string) bool { return slices.Contains(before, t) })
	removed := slices.DeleteFunc(slices.Clone(before), func(t string) bool { return slices.Contains(after, t) })
	refs, err := s.directory.TeamRefs(ctx, slices.Concat(added, removed))
	if err != nil {
		return "", err
	}
	parts := []string{}
	if len(added) > 0 {
		parts = append(parts, fmt.Sprintf(msgSharedWith, teamNames(refs, added)))
	}
	if len(removed) > 0 {
		parts = append(parts, fmt.Sprintf(msgNoLonger, teamNames(refs, removed)))
	}
	return strings.Join(parts, "; "), nil
}
