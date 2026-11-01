package customdashboard

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"mapp/internal/history"
	"mapp/internal/platform/date"
	"mapp/internal/platform/errs"
)

const (
	kindDashboard       = "customDashboard"
	kindDashboardPlural = "dashboards"
)

// Kinds returns the history kind of custom dashboards; a snapshot holds the dashboard with its widgets and
// shares, so Undo and Restore bring all three back.
func (s *Service) Kinds() []history.Kind {
	return []history.Kind{{
		Name: kindDashboard, Singular: dashboardNoun, Plural: kindDashboardPlural,
		Snapshot: s.snapshot, Remove: s.repo.Delete, Put: s.put, Describe: describeField, PrepareRestore: s.prepareRestore,
	}}
}

func (s *Service) snapshot(ctx context.Context, id string) (history.Snapshot, bool, error) {
	d, found, err := s.repo.Get(ctx, id)
	if err != nil || !found {
		return history.Snapshot{}, false, err
	}
	data, err := json.Marshal(d)
	if err != nil {
		return history.Snapshot{}, false, err
	}
	return history.Snapshot{ID: d.ID, Name: d.Name, Data: data}, true, nil
}

func (s *Service) put(ctx context.Context, snapshot history.Snapshot) error {
	var d Dashboard
	if err := json.Unmarshal(snapshot.Data, &d); err != nil {
		return err
	}
	return s.repo.Insert(ctx, normalised(d))
}

// prepareRestore refuses an Undo whose name the owner has meanwhile used, and renames on a Restore.
func (s *Service) prepareRestore(ctx context.Context, rc history.RestoreContext, snapshot *history.Snapshot) ([]string, error) {
	var d Dashboard
	if err := json.Unmarshal(snapshot.Data, &d); err != nil {
		return nil, err
	}
	taken, err := s.repo.NameTaken(ctx, d.OwnerID, d.Name, d.ID)
	if err != nil || !taken {
		return nil, err
	}
	if rc.Undo {
		return nil, errs.NameTaken(fmt.Sprintf(msgUndoNameTaken, d.Name))
	}
	base := shortened(d.Name, maxNameLength-len(restoredSuffix)) + restoredSuffix
	if d.Name, err = s.freeName(ctx, d.OwnerID, base, d.ID); err != nil {
		return nil, err
	}
	d.Touch(rc.Actor, rc.Now)
	data, err := json.Marshal(d)
	if err != nil {
		return nil, err
	}
	snapshot.Name, snapshot.Data = d.Name, data
	return []string{fmt.Sprintf(msgRenamedOnRestore, d.Name)}, nil
}

// describeField writes default filters in words; the widgets change arrives already written by logChanges.
func describeField(_ context.Context, field string, value any) (label, text string, err error) {
	switch field {
	case filtersField:
		return labelFilters, filtersText(value), nil
	case widgetsField:
		return labelWidgets, "", nil
	}
	return "", "", nil
}

func filtersText(value any) string {
	filters, _ := value.(map[string]any)
	var parts []string
	if from := dateText(filters["dateFrom"]); from != "" {
		parts = append(parts, fmt.Sprintf(textFiltersFrom, from))
	}
	if to := dateText(filters["dateTo"]); to != "" {
		parts = append(parts, fmt.Sprintf(textFiltersTo, to))
	}
	plans, _ := filters["planIds"].([]any)
	teams, _ := filters["teamIds"].([]any)
	parts = appendCount(parts, len(plans), "plan")
	parts = appendCount(parts, len(teams), "team")
	if len(parts) == 0 {
		return textNoFilters
	}
	text := strings.Join(parts, ", ")
	return strings.ToUpper(text[:1]) + text[1:]
}

func appendCount(parts []string, n int, noun string) []string {
	switch n {
	case 0:
		return parts
	case 1:
		return append(parts, "1 "+noun)
	}
	return append(parts, fmt.Sprintf("%d %ss", n, noun))
}

func dateText(value any) string {
	text, _ := value.(string)
	d, err := date.Parse(text)
	if err != nil {
		return ""
	}
	return date.Format(d)
}
