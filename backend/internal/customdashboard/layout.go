package customdashboard

import (
	"fmt"
	"strings"

	"mapp/internal/platform/errs"
)

// checkLayouts keeps every widget inside the 12-column grid and away from the widgets before it.
func checkLayouts(widgets []Widget, p *errs.Problems) {
	for i, w := range widgets {
		at := fmt.Sprintf(widgetFieldFormat, i) + ".layout"
		if message := boundsProblem(w.Layout); message != "" {
			p.Add(at, message)
			continue
		}
		for _, before := range widgets[:i] {
			if overlaps(before.Layout, w.Layout) {
				p.Add(at, fmt.Sprintf(msgOverlaps, widgetName(before)))
				break
			}
		}
	}
}

func boundsProblem(l Layout) string {
	switch {
	case outside(l.X, 0, gridColumns-1):
		return msgColumnRange
	case outside(l.Y, 0, gridRows-1):
		return msgRowRange
	case outside(l.W, minSize, maxSize):
		return msgWidthRange
	case outside(l.H, minSize, maxSize):
		return msgHeightRange
	case l.X+l.W > gridColumns:
		return msgPastRightEdge
	case l.Y+l.H > gridRows:
		return msgPastBottomEdge
	}
	return ""
}

func outside(value, least, most int) bool {
	return value < least || value > most
}

func overlaps(a, b Layout) bool {
	return a.X < b.X+b.W && b.X < a.X+a.W && a.Y < b.Y+b.H && b.Y < a.Y+a.H
}

// widgetName names a widget in messages and the log: its title, or its type and id when untitled.
func widgetName(w Widget) string {
	if title := strings.TrimSpace(w.Title); title != "" {
		return "'" + title + "'"
	}
	label := w.Type
	if rule, known := typeRules()[w.Type]; known {
		label = rule.label
	}
	return fmt.Sprintf(textUntitledWidget, label, w.ID)
}
