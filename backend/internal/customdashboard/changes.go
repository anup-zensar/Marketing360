package customdashboard

import (
	"bytes"
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"mapp/internal/history"
)

type loggedFields struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Filters     Filters `json:"filters"`
}

// logChanges lists the changed fields for the activity log; widgets come as one change written in words,
// "4 widgets → 5 widgets; added 'Leads'", because the list as stored is unreadable.
func logChanges(before, after Dashboard) ([]history.Change, error) {
	changes, err := history.Diff(loggedOf(before), loggedOf(after))
	if err != nil {
		return nil, err
	}
	text := widgetChanges(before.Widgets, after.Widgets)
	if text == "" {
		return changes, nil
	}
	oldText, err := json.Marshal(widgetCount(len(before.Widgets)))
	if err != nil {
		return nil, err
	}
	newText, err := json.Marshal(widgetCount(len(after.Widgets)) + "; " + text)
	if err != nil {
		return nil, err
	}
	return append(changes, history.Change{Field: widgetsField, Old: oldText, New: newText}), nil
}

func loggedOf(d Dashboard) loggedFields {
	return loggedFields{Name: d.Name, Description: d.Description, Filters: d.Filters}
}

func widgetCount(n int) string {
	if n == 1 {
		return textOneWidget
	}
	return fmt.Sprintf(textWidgetCount, n)
}

func widgetChanges(before, after []Widget) string {
	var added, removed, changed []string
	for _, w := range after {
		i := slices.IndexFunc(before, func(b Widget) bool { return b.ID == w.ID })
		switch {
		case i < 0:
			added = append(added, widgetName(w))
		case !isSameWidget(before[i], w):
			changed = append(changed, widgetName(w))
		}
	}
	for _, w := range before {
		if !slices.ContainsFunc(after, func(a Widget) bool { return a.ID == w.ID }) {
			removed = append(removed, widgetName(w))
		}
	}
	parts := slices.Concat(named(textAdded, added), named(textRemoved, removed), named(textChanged, changed))
	return strings.Join(parts, "; ")
}

func named(format string, names []string) []string {
	if len(names) == 0 {
		return nil
	}
	return []string{fmt.Sprintf(format, strings.Join(names, ", "))}
}

func isSameWidget(a, b Widget) bool {
	encodedA, errA := json.Marshal(a)
	encodedB, errB := json.Marshal(b)
	return errA == nil && errB == nil && bytes.Equal(encodedA, encodedB)
}
