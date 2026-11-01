package approvalreport

import "mapp/internal/approval"

// OutcomeWord is the one word shown for a round, whatever the engine called it (041-C1, C6).
func OutcomeWord(r approval.Round) Outcome {
	switch r.Outcome {
	case approval.OutcomeOpen:
		return OutcomeWaiting
	case approval.OutcomeApproved:
		if r.AutoApproved {
			return OutcomeAutoApproved
		}
		return OutcomeApproved
	case approval.OutcomeDeclined:
		return OutcomeDeclined
	case approval.OutcomeRestarted, approval.OutcomeSuperseded:
		return OutcomeReplaced
	case approval.OutcomeRecalled, approval.OutcomeCancelled:
		return OutcomeRecalled
	}
	return OutcomeRecalled
}

func recordTypeOf(kind approval.SubjectKind) RecordType {
	switch kind {
	case approval.KindPurchaseRequest:
		return RecordPurchaseRequest
	case approval.KindModifiedPurchaseRequest:
		return RecordModifiedPurchaseRequest
	case approval.KindActivity:
		return RecordActivity
	}
	return RecordActivity
}

func allRecordTypes() []RecordType {
	return []RecordType{RecordActivity, RecordPurchaseRequest, RecordModifiedPurchaseRequest}
}

func statusOutcomes() []Outcome {
	return []Outcome{OutcomeWaiting, OutcomeApproved, OutcomeDeclined, OutcomeRecalled, OutcomeReplaced, OutcomeAutoApproved}
}
