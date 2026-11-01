package dataset

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"mapp/internal/history"
	"mapp/internal/people"
	"mapp/internal/platform/errs"
)

const (
	kindDataset      = "dataset"
	kindSingular     = "dataset"
	kindPlural       = "datasets"
	labelName        = "Name"
	labelDesc        = "Description"
	labelColumns     = "Column names"
	labelShared      = "Shared with"
	textNotShared    = "Nobody"
	listSeparator    = ", "
	restoredSuffix   = " (restored)"
	restoredNumbered = "%s (restored %d)"
	restoredLongest  = " (restored 99)"
)

// snapshotData keeps the whole dataset in the recycle bin: its row, columns, shares and rows.
type snapshotData struct {
	Dataset Dataset `json:"dataset"`
	Rows    []Row   `json:"rows"`
}

func (s *Service) Kinds() []history.Kind {
	return []history.Kind{{
		Name: kindDataset, Singular: kindSingular, Plural: kindPlural, Snapshot: s.snapshot, Remove: s.repo.Delete,
		Put: s.put, Describe: s.describe, PrepareRestore: s.prepareRestore,
	}}
}

func (s *Service) snapshot(ctx context.Context, id string) (history.Snapshot, bool, error) {
	d, found, err := s.repo.Get(ctx, id)
	if err != nil || !found {
		return history.Snapshot{}, false, err
	}
	rows, err := s.repo.Rows(ctx, id)
	if err != nil {
		return history.Snapshot{}, false, err
	}
	data, err := json.Marshal(snapshotData{Dataset: d, Rows: rows})
	if err != nil {
		return history.Snapshot{}, false, err
	}
	return history.Snapshot{ID: d.ID, Name: d.Name, Data: data}, true, nil
}

func (s *Service) put(ctx context.Context, snapshot history.Snapshot) error {
	var data snapshotData
	if err := json.Unmarshal(snapshot.Data, &data); err != nil {
		return err
	}
	return s.repo.Insert(ctx, data.Dataset, data.Rows)
}

func (s *Service) describe(ctx context.Context, field string, value any) (label, text string, err error) {
	switch field {
	case nameField:
		return labelName, "", nil
	case descriptionField:
		return labelDesc, "", nil
	case columnsField:
		return labelColumns, strings.Join(textsOf(value), listSeparator), nil
	case "sharedWith":
		ids := textsOf(value)
		refs, err := s.directory.TeamRefs(ctx, ids)
		if err != nil || len(ids) == 0 {
			return labelShared, textNotShared, err
		}
		return labelShared, teamNames(refs, ids), nil
	}
	return "", "", nil
}

func textsOf(value any) []string {
	list, _ := value.([]any)
	texts := make([]string, 0, len(list))
	for _, v := range list {
		if text, ok := v.(string); ok {
			texts = append(texts, text)
		}
	}
	return texts
}

func teamNames(refs map[string]people.TeamRef, ids []string) string {
	names := make([]string, 0, len(ids))
	for _, id := range ids {
		if t, found := refs[id]; found {
			names = append(names, t.Name)
		} else {
			names = append(names, id)
		}
	}
	return strings.Join(names, listSeparator)
}

// prepareRestore refuses an Undo when the owner now has another dataset of the name, and a Restore renames it.
func (s *Service) prepareRestore(ctx context.Context, rc history.RestoreContext, snapshot *history.Snapshot) ([]string, error) {
	var data snapshotData
	if err := json.Unmarshal(snapshot.Data, &data); err != nil {
		return nil, err
	}
	d := data.Dataset
	names, err := s.namesOfOwner(ctx, d.OwnerID, d.ID)
	if err != nil || !slices.Contains(names, strings.ToLower(d.Name)) {
		return nil, err
	}
	if rc.Undo {
		return nil, errs.NameTaken(fmt.Sprintf(msgUndoTaken, d.Name))
	}
	d.Name = restoredName(names, d.Name)
	d.Touch(rc.Actor, rc.Now)
	data.Dataset = d
	encoded, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	snapshot.Name, snapshot.Data = d.Name, encoded
	return []string{fmt.Sprintf(msgRenamed, d.Name)}, nil
}

func (s *Service) namesOfOwner(ctx context.Context, ownerID, exceptID string) ([]string, error) {
	datasets, err := s.repo.ListVisible(ctx, ownerID)
	names := []string{}
	for _, d := range datasets {
		if d.OwnerID == ownerID && d.ID != exceptID {
			names = append(names, strings.ToLower(d.Name))
		}
	}
	return names, err
}

func restoredName(taken []string, name string) string {
	runes := []rune(name)
	base := string(runes[:min(len(runes), maxNameRunes-len(restoredLongest))])
	candidate := base + restoredSuffix
	for n := 2; slices.Contains(taken, strings.ToLower(candidate)); n++ {
		candidate = fmt.Sprintf(restoredNumbered, base, n)
	}
	return candidate
}
