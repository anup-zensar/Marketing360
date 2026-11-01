package customdashboard

import (
	"fmt"
	"regexp"
	"slices"
	"strings"
	"unicode/utf8"

	"mapp/internal/platform/errs"
	"mapp/internal/platform/record"
)

// validate returns every problem of a dashboard at once (FR-033); fields are checked against the data
// catalogue only when a widget's data is asked for.
func validate(d Dashboard) errs.Problems {
	var p errs.Problems
	checkName(d.Name, &p)
	if tooLong(d.Description, maxDescriptionLength) {
		p.Add(descriptionField, msgDescriptionTooLong)
	}
	checkFilters(d.Filters, &p)
	checkWidgets(d.Widgets, &p)
	return p
}

func checkName(name string, p *errs.Problems) {
	switch {
	case strings.TrimSpace(name) == "":
		p.Add(nameField, msgEnterName)
	case tooLong(name, maxNameLength):
		p.Add(nameField, msgNameTooLong)
	case record.CheckSingleLine(name) != "":
		p.Add(nameField, record.CheckSingleLine(name))
	}
}

func checkFilters(f Filters, p *errs.Problems) {
	if f.DateFrom != nil && f.DateTo != nil && f.DateFrom.After(*f.DateTo) {
		p.Add("filters.dateTo", msgDateRange)
	}
	checkIDs("filters.planIds", f.PlanIDs, planIDPattern, [2]string{msgTooManyPlans, msgUnknownPlan}, p)
	checkIDs("filters.teamIds", f.TeamIDs, teamIDPattern, [2]string{msgTooManyTeams, msgUnknownTeam}, p)
}

// checkIDs refuses more than 20 ids (messages[0]) or an id that is not one (messages[1]).
func checkIDs(field string, ids []string, pattern *regexp.Regexp, messages [2]string, p *errs.Problems) {
	if len(ids) > maxFilterIDs {
		p.Add(field, messages[0])
		return
	}
	if slices.ContainsFunc(ids, func(id string) bool { return !pattern.MatchString(id) }) {
		p.Add(field, messages[1])
	}
}

func checkWidgets(widgets []Widget, p *errs.Problems) {
	if len(widgets) > maxWidgets {
		p.Add(widgetsField, msgTooManyWidgets)
	}
	seen := map[string]bool{}
	for i, w := range widgets {
		at := fmt.Sprintf(widgetFieldFormat, i)
		if seen[w.ID] || !widgetIDPattern.MatchString(w.ID) {
			p.Add(at+".id", msgWidgetID)
		}
		seen[w.ID] = true
		checkWidget(at, w, p)
	}
	checkLayouts(widgets, p)
}

func checkWidget(at string, w Widget, p *errs.Problems) {
	rule, known := typeRules()[w.Type]
	if !known {
		p.Add(at+".type", msgWidgetType)
		return
	}
	checkTexts(at, w, p)
	switch {
	case w.Type == typeText && w.Query != nil:
		p.Add(at+".query", msgTextHasNoQuery)
	case w.Type != typeText && w.Query == nil:
		p.Add(at+".query", msgChooseMeasure)
	case w.Query != nil:
		checkQuery(at+".query", rule, *w.Query, p)
	}
	checkDisplay(at+".display", w, rule, p)
}

func checkTexts(at string, w Widget, p *errs.Problems) {
	if tooLong(w.Title, maxTitleLength) {
		p.Add(at+".title", msgTitleTooLong)
	}
	if tooLong(w.Subtitle, maxSubtitleLength) {
		p.Add(at+".subtitle", msgSubtitleTooLong)
	}
	switch {
	case tooLong(w.Note, maxNoteLength):
		p.Add(at+".note", msgNoteTooLong)
	case w.Type != typeText && w.Note != "":
		p.Add(at+".note", msgNoteOnlyText)
	}
}

func tooLong(text string, limit int) bool {
	return utf8.RuneCountInString(text) > limit
}

// withArticle starts a type's noun with "A" or "An", as messages begin.
func withArticle(noun string) string {
	if strings.ContainsRune("aeiou", rune(noun[0])) {
		return "An " + noun
	}
	return "A " + noun
}

func indexed(at, list string, i int) string {
	return fmt.Sprintf(indexedFieldFormat, at, list, i)
}
