package plandetails

import (
	"mapp/internal/allwork"
	"mapp/internal/platform/money"
)

func filledPart() part {
	return part{key: KeyFilled, label: "Filled in", order: 0, counted: true}
}

// rows returns one row per active Shell attribute of the level, in Settings' order (FR-008).
func (v view) rows(attributes []allwork.Attribute) []AttributeRow {
	rows := []AttributeRow{}
	for _, a := range attributes {
		if !a.Active || a.Level != allwork.Level(v.level) {
			continue
		}
		row := AttributeRow{AttributeID: a.ID, Label: a.Label, Type: a.Type}
		if choiceKind(a.Type) {
			row.Widgets = v.choiceWidgets(a)
		} else {
			row.Widgets = []Widget{v.fieldFilledIn(a)}
		}
		rows = append(rows, row)
	}
	return rows
}

func choiceKind(kind string) bool {
	switch kind {
	case allwork.TypeSingleChoice, allwork.TypeMultipleChoice, allwork.TypeMultipleChoiceSplit, allwork.TypeYesNo:
		return true
	}
	return false
}

// choiceWidgets are the records and the planned cost per value and the filled-in widget (FR-009). The planned cost
// is divided activity by activity, as All work's Breakdown divides it (FR-011, SC-002).
func (v view) choiceWidgets(a allwork.Attribute) []Widget {
	values, filled := newTally(), newTally()
	filled.seed(filledPart(), notFilledInPart())
	for _, r := range v.records {
		units := v.units[r.PlanID]
		parts, each := partsOf(a, r, units, money.Amount{})
		values.add(r.ID, parts, money.Amount{})
		for _, activity := range v.pieces[r.ID] {
			amount := plannedCostOf(activity)
			divided, _ := partsOf(a, r, units, amount)
			values.add(r.ID, divided, amount)
		}
		values.countsInEach = values.countsInEach || each
		filled.add(r.ID, []part{filledIn(parts[0])}, money.Amount{})
	}
	return []Widget{
		values.widget(spec{key: attrKey(a.ID, suffixCount), title: v.recordsWord() + " by " + a.Label,
			measure: MeasureCount, filter: FilterAttr}, v.currency),
		values.widget(spec{key: attrKey(a.ID, suffixPlannedCost), title: "Planned cost by " + a.Label,
			measure: MeasurePlannedCost, filter: FilterAttr}, v.currency),
		v.filledInWidget(a, filled),
	}
}

// filledIn turns a record's first value into Filled in, Not filled in or Does not apply.
func filledIn(p part) part {
	if p.key == KeyNotFilledIn || p.key == KeyDoesNotApply {
		return p
	}
	return filledPart()
}

// fieldFilledIn is the only widget of a text, long text or number attribute (FR-010, 043-C3).
func (v view) fieldFilledIn(a allwork.Attribute) Widget {
	filled := newTally()
	filled.seed(filledPart(), notFilledInPart())
	for _, r := range v.records {
		p := filledPart()
		switch {
		case !shownFor(a, v.units[r.PlanID]):
			p = doesNotApplyPart()
		case empty(r.Fields["attr."+a.ID]):
			p = notFilledInPart()
		}
		filled.add(r.ID, []part{p}, money.Amount{})
	}
	return v.filledInWidget(a, filled)
}

func empty(value allwork.Value) bool {
	return value.Text == "" && value.Number == nil && value.USD == nil && value.Money == nil && value.Yes == nil &&
		len(value.Choices) == 0
}

func (v view) filledInWidget(a allwork.Attribute, filled *tally) Widget {
	w := filled.widget(spec{key: attrKey(a.ID, suffixFilledIn), title: a.Label + " filled in", measure: MeasureFilledIn,
		filter: FilterAttr, timeOrder: true}, v.currency)
	yes, no := len(filled.buckets[KeyFilled].records), len(filled.buckets[KeyNotFilledIn].records)
	w.FilledInPercent = wholePercent(yes, yes+no)
	return w
}
