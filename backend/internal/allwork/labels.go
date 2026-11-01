package allwork

type Labels struct {
	Status, Timing, Swbs, PlanType string
}

func LabelsOf(n *Node) Labels {
	return Labels{Status: labelOf(statusOf(n)), Timing: labelOf(timingOf(n)), Swbs: labelOf(swbsOf(n)),
		PlanType: planTypeWords[n.PlanType]}
}

func labelOf(s *Status) string {
	if s == nil {
		return ""
	}
	return s.Label
}

// PlannedByActivityInYears is PlannedByActivity over the years from–to, so months of other years keep their month.
func PlannedByActivityInYears(nodes []Node, from, to int) map[string]Months {
	return plannedByActivity(nodes, from, to)
}
