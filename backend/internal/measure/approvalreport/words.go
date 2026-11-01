package approvalreport

func recordTypeLabel(t RecordType) string {
	switch t {
	case RecordPurchaseRequest:
		return "Purchase request"
	case RecordModifiedPurchaseRequest:
		return "Modified purchase request"
	case RecordActivity:
		return "Activity"
	}
	return string(t)
}

func outcomeLabel(o Outcome) string {
	switch o {
	case OutcomeWaiting:
		return "Waiting"
	case OutcomeApproved:
		return "Approved"
	case OutcomeDeclined:
		return "Declined"
	case OutcomeRecalled:
		return "Recalled"
	case OutcomeReplaced:
		return "Replaced by a new submission"
	case OutcomeAutoApproved:
		return "Approved automatically"
	case OutcomeNotNeeded:
		return "Not needed"
	}
	return string(o)
}
