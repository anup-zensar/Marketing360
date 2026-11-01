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

const versionField = "version"

// logged is what the activity log compares on a change: column types never change, so only their names.
type logged struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Columns     []string `json:"columns"`
}

func loggedOf(d Dataset) logged {
	labels := make([]string, 0, len(d.Columns))
	for _, c := range d.Columns {
		labels = append(labels, c.Label)
	}
	return logged{Name: d.Name, Description: d.Description, Columns: labels}
}

// Update changes the name, description and column names the owner sends; a save that changes nothing creates no
// version and no log entry.
func (s *Service) Update(ctx context.Context, a actor.Actor, id string, in UpdateInput) (View, error) {
	var view View
	err := s.tx.Write(ctx, func(ctx context.Context) error {
		stored, err := s.owned(ctx, a, id)
		if err != nil {
			return err
		}
		changed, err := s.changed(ctx, stored, in)
		if err != nil {
			return err
		}
		changes, err := history.Diff(loggedOf(stored), loggedOf(changed))
		if err != nil {
			return err
		}
		if len(changes) == 0 {
			view, err = s.view(ctx, a, stored)
			return err
		}
		if changed, err = s.save(ctx, a, stored, changed); err != nil {
			return err
		}
		if _, err := s.history.Log(ctx, a, history.LogInput{Kind: kindDataset, RecordID: id, RecordName: changed.Name,
			Action: history.ActionChanged, Changes: changes}); err != nil {
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

func (s *Service) changed(ctx context.Context, stored Dataset, in UpdateInput) (Dataset, error) {
	var problems errs.Problems
	changed := stored
	if in.Name != nil {
		changed.Name = trimmed(*in.Name)
		checkName(changed.Name, &problems)
	}
	if in.Description != nil {
		changed.Description = trimmed(*in.Description)
		checkDescription(changed.Description, &problems)
	}
	changed.Columns = relabelled(stored.Columns, in.Columns, &problems)
	if in.Version == nil {
		problems.Add(versionField, msgVersion)
	}
	if err := problems.Err(); err != nil {
		return Dataset{}, err
	}
	if err := stored.CheckVersion(*in.Version, kindSingular); err != nil {
		return Dataset{}, err
	}
	if !strings.EqualFold(changed.Name, stored.Name) {
		return changed, s.checkNameFree(ctx, stored.OwnerID, changed.Name, stored.ID)
	}
	return changed, nil
}

// relabelled gives the columns their new names; names stay unique ignoring case, kinds and currencies stay.
func relabelled(columns []Column, inputs []LabelInput, problems *errs.Problems) []Column {
	result := slices.Clone(columns)
	for _, in := range inputs {
		if in.Position < 0 || in.Position >= len(result) {
			problems.Add(columnsField, fmt.Sprintf(msgNoSuchColumn, in.Position+1))
			continue
		}
		result[in.Position].Label = trimmed(in.Label)
	}
	taken := map[string]bool{}
	for i, c := range result {
		checkLabel(c.Label, i, taken, problems)
	}
	return result
}

func (s *Service) save(ctx context.Context, a actor.Actor, stored, changed Dataset) (Dataset, error) {
	changed.Touch(a, s.now())
	if err := s.repo.Update(ctx, changed); err != nil {
		return Dataset{}, err
	}
	if slices.Equal(loggedOf(stored).Columns, loggedOf(changed).Columns) {
		return changed, nil
	}
	return changed, s.repo.UpdateLabels(ctx, changed.ID, changed.Columns)
}

// Delete moves the dataset to the recycle bin; widgets that read it then answer that it is no longer available.
func (s *Service) Delete(ctx context.Context, a actor.Actor, id string) (history.BinEntry, error) {
	var entry history.BinEntry
	err := s.tx.Write(ctx, func(ctx context.Context) error {
		if _, err := s.owned(ctx, a, id); err != nil {
			return err
		}
		var err error
		entry, err = s.history.MoveToBin(ctx, a, kindDataset, id)
		return err
	})
	if err != nil {
		return history.BinEntry{}, err
	}
	return entry, nil
}
