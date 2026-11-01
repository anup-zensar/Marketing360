package measure

import (
	"slices"
	"strings"

	"mapp/internal/allwork"
)

const (
	labelNotFilledIn = "Not filled in"
	labelYes         = "Yes"
	labelNo          = "No"
)

func activityStatuses() []string {
	return []string{"created", "submitted", "declined", "approved", "inProgress", "complete"}
}

func planTypes() []string { return []string{"operational", "strategic", "crossMarket"} }

// choicesOf are the values the filter panel offers for the scope (contract ReportChoices, FR-008 – FR-011).
func choicesOf(sc Scope, f Filter, years []int, currencies []Currency) Choices {
	selected := selectedPlans(sc, f)
	teams := teamsOf(sc)
	return Choices{Years: years, Plans: planChoices(sc.Nodes), PlanTypes: planTypes(), Teams: teams,
		Owners: ownersOf(sc.Nodes, selected), ActivityTypes: typesOf(sc.Nodes, selected, sc.Choices.ActivityTypes),
		ActivityStatuses: activityStatuses(), AttributeGroups: attributeGroups(sc.Attributes, sc.Nodes),
		Currencies: currencies}
}

// selectedPlans are the plans of the filter, or every plan of the scope when it names none.
func selectedPlans(sc Scope, f Filter) map[string]bool {
	selected := map[string]bool{}
	for _, n := range sc.Nodes {
		if n.Level == allwork.LevelPlan && (len(f.Query.Plans) == 0 || slices.Contains(f.Query.Plans, n.ID)) {
			selected[n.ID] = true
		}
	}
	return selected
}

func planChoices(nodes []allwork.Node) []PlanChoice {
	plans := []PlanChoice{}
	for _, n := range nodes {
		if n.Level == allwork.LevelPlan {
			plans = append(plans, PlanChoice{ID: n.ID, Name: n.Name, PlanType: n.PlanType,
				TeamIDs: append([]string{}, n.TeamIDs...)})
		}
	}
	return plans
}

func ownersOf(nodes []allwork.Node, selected map[string]bool) []NamedChoice {
	byID := map[string]NamedChoice{}
	for _, n := range nodes {
		if n.Owner != nil && selected[n.PlanID] {
			byID[n.Owner.ID] = NamedChoice{ID: n.Owner.ID, Name: n.Owner.Name}
		}
	}
	return byName(byID)
}

func typesOf(nodes []allwork.Node, selected map[string]bool, types []allwork.Option) []NamedChoice {
	byID := map[string]NamedChoice{}
	for _, n := range nodes {
		if n.Level != allwork.LevelActivity || n.ActivityTypeID == "" || !selected[n.PlanID] {
			continue
		}
		name := n.ActivityTypeID
		if i := slices.IndexFunc(types, func(o allwork.Option) bool { return o.ID == n.ActivityTypeID }); i >= 0 {
			name = types[i].Label
		}
		byID[n.ActivityTypeID] = NamedChoice{ID: n.ActivityTypeID, Name: name}
	}
	return byName(byID)
}

func byName(byID map[string]NamedChoice) []NamedChoice {
	list := make([]NamedChoice, 0, len(byID))
	for _, c := range byID {
		list = append(list, c)
	}
	slices.SortFunc(list, func(x, y NamedChoice) int {
		return cmpOr(strings.Compare(strings.ToLower(x.Name), strings.ToLower(y.Name)), strings.Compare(x.ID, y.ID))
	})
	return list
}

func cmpOr(first, second int) int {
	if first != 0 {
		return first
	}
	return second
}

// attributeGroups are one group per active choice or yes/no attribute of each level, never merged by label; an
// inactive option is offered only while a record in scope has it; Not filled in comes last as none (FR-008).
func attributeGroups(attributes []allwork.Attribute, nodes []allwork.Node) []AttributeGroup {
	groups := []AttributeGroup{}
	for _, a := range attributes {
		if !a.Active || !filterable(a.Type) {
			continue
		}
		groups = append(groups, AttributeGroup{Level: string(a.Level), AttributeID: a.ID, Label: a.Label,
			Values: append(valuesOf(a, nodes), AttributeValue{ID: allwork.AttrNotFilledIn, Label: labelNotFilledIn, Active: true})})
	}
	return groups
}

func filterable(kind string) bool {
	switch kind {
	case allwork.TypeSingleChoice, allwork.TypeMultipleChoice, allwork.TypeMultipleChoiceSplit, allwork.TypeYesNo:
		return true
	}
	return false
}

func valuesOf(a allwork.Attribute, nodes []allwork.Node) []AttributeValue {
	if a.Type == allwork.TypeYesNo {
		return []AttributeValue{{ID: allwork.AttrYes, Label: labelYes, Active: true}, {ID: allwork.AttrNo, Label: labelNo, Active: true}}
	}
	values := []AttributeValue{}
	for _, o := range a.Options {
		if !o.Inactive || usedBy(nodes, a.ID, o.ID) {
			values = append(values, AttributeValue{ID: o.ID, Label: o.Label, Active: !o.Inactive})
		}
	}
	return values
}

func usedBy(nodes []allwork.Node, attributeID, optionID string) bool {
	return slices.ContainsFunc(nodes, func(n allwork.Node) bool {
		return slices.Contains(n.Attrs[attributeID].OptionIDs, optionID)
	})
}

// teamsOf are the teams of the plans in scope, so a market never sees another market's team names (SC-002).
func teamsOf(sc Scope) []TeamChoice {
	inScope := map[string]bool{}
	for _, n := range sc.Nodes {
		for _, id := range n.TeamIDs {
			inScope[id] = true
		}
	}
	teams := []TeamChoice{}
	for _, t := range sc.Choices.Teams {
		if inScope[t.ID] {
			teams = append(teams, TeamChoice{ID: t.ID, Name: t.Label, Mine: t.Mine})
		}
	}
	return teams
}
