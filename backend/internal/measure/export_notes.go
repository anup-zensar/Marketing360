package measure

import (
	"slices"
	"strconv"
	"strings"

	"mapp/internal/allwork"
	"mapp/internal/platform/actor"
	"mapp/internal/platform/date"
)

// notes are the lines of About this export, and of a CSV before its rows (FR-022): the report, the year, the filters
// in words, the period, the currency, and who exported it and when; the rates used follow.
func (s *Service) notes(a actor.Actor, report string, f Filter, sc Scope) [][2]string {
	now := s.now().UTC()
	return [][2]string{{"Report", report}, {"Year", strconv.Itoa(sc.Year)}, {labelFilters, filterWords(f, sc)},
		{"Period", map[string]string{PeriodMonth: "Months", PeriodQuarter: "Quarters"}[f.Period]},
		{"Currency", sc.Currency.Code}, {"Exported by", a.Name},
		{"Exported at", date.Format(date.Today(now)) + " " + now.Format("15:04") + " UTC"}}
}

// afterFilters puts a report's own parameters in words right after the frame's filters.
func afterFilters(notes, own [][2]string) [][2]string {
	at := slices.IndexFunc(notes, func(n [2]string) bool { return n[0] == labelFilters }) + 1
	return slices.Insert(notes, at, own...)
}

func rateNotes(rates []string) [][2]string {
	notes := make([][2]string, 0, len(rates))
	for _, rate := range rates {
		notes = append(notes, [2]string{"Rate used", rate})
	}
	return notes
}

// filterWords names every restricted filter with its values' names, never an ID left out (FR-015): "None" for none.
func filterWords(f Filter, sc Scope) string {
	q := f.Query
	plans := map[string]string{}
	for _, n := range sc.Nodes {
		if n.Level == allwork.LevelPlan {
			plans[n.ID] = n.Name
		}
	}
	var parts []string
	add := func(title string, ids []string, names func(string) (string, bool)) {
		var words []string
		for _, id := range ids {
			if name, known := names(id); known {
				words = append(words, name)
			}
		}
		if len(words) > 0 {
			parts = append(parts, title+": "+strings.Join(words, ", "))
		}
	}
	add("Plans", q.Plans, func(id string) (string, bool) { name, known := plans[id]; return name, known })
	add("Plan type", q.PlanTypes, func(id string) (string, bool) { return id, true })
	add("Team", q.Teams, optionNamed(sc.Choices.Teams))
	add("Owner", q.Owners, optionNamed(sc.Choices.People))
	add("Activity type", q.ActivityTypes, optionNamed(sc.Choices.ActivityTypes))
	add("Activity status", q.ActivityStatuses, func(id string) (string, bool) { return id, true })
	for _, filter := range q.Attributes {
		if i := slices.IndexFunc(sc.Attributes, func(a allwork.Attribute) bool { return a.ID == filter.AttributeID }); i >= 0 {
			add(sc.Attributes[i].Label, filter.Values, valueNamed(sc.Attributes[i]))
		}
	}
	if len(parts) == 0 {
		return "None"
	}
	return strings.Join(parts, "; ")
}

func optionNamed(options []allwork.Option) func(string) (string, bool) {
	return func(id string) (string, bool) {
		i := slices.IndexFunc(options, func(o allwork.Option) bool { return o.ID == id })
		if i < 0 {
			return "", false
		}
		return options[i].Label, true
	}
}

func valueNamed(a allwork.Attribute) func(string) (string, bool) {
	words := map[string]string{allwork.AttrNotFilledIn: labelNotFilledIn, allwork.AttrYes: labelYes, allwork.AttrNo: labelNo}
	return func(id string) (string, bool) {
		if word, found := words[id]; found {
			return word, true
		}
		return optionNamed(a.Options)(id)
	}
}
