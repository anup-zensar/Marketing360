package customdashboard

// normalised gives every list of the dashboard a value, so answers hold [] and never null.
func normalised(d Dashboard) Dashboard {
	d.SharedWith = orEmpty(d.SharedWith)
	d.Filters.PlanIDs = orEmpty(d.Filters.PlanIDs)
	d.Filters.TeamIDs = orEmpty(d.Filters.TeamIDs)
	widgets := make([]Widget, 0, len(d.Widgets))
	for _, w := range d.Widgets {
		widgets = append(widgets, normalisedWidget(w))
	}
	d.Widgets = widgets
	return d
}

func normalisedWidget(w Widget) Widget {
	if w.Query != nil {
		q := *w.Query
		q.Measures = orEmpty(q.Measures)
		filters := make([]Filter, 0, len(q.Filters))
		for _, f := range q.Filters {
			f.Values = orEmpty(f.Values)
			filters = append(filters, f)
		}
		q.Filters = filters
		w.Query = &q
	}
	if w.Display.Table != nil {
		table := *w.Display.Table
		table.Columns = orEmpty(table.Columns)
		w.Display.Table = &table
	}
	return w
}

func orEmpty[T any](list []T) []T {
	if list == nil {
		return []T{}
	}
	return list
}
