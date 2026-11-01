package approvalreport_test

import (
	"testing"

	"mapp/internal/approval"
	"mapp/internal/measure/approvalreport"
)

func TestOutcomeWord_041_FR006(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		round approval.Round
		want  approvalreport.Outcome
	}{
		{name: "open is waiting", round: approval.Round{Outcome: approval.OutcomeOpen}, want: approvalreport.OutcomeWaiting},
		{name: "approved by a person", round: approval.Round{Outcome: approval.OutcomeApproved}, want: approvalreport.OutcomeApproved},
		{name: "approved with no person is automatic", round: approval.Round{Outcome: approval.OutcomeApproved, AutoApproved: true},
			want: approvalreport.OutcomeAutoApproved},
		{name: "declined", round: approval.Round{Outcome: approval.OutcomeDeclined}, want: approvalreport.OutcomeDeclined},
		{name: "recalled", round: approval.Round{Outcome: approval.OutcomeRecalled}, want: approvalreport.OutcomeRecalled},
		{name: "041-C6 restarted is replaced", round: approval.Round{Outcome: approval.OutcomeRestarted}, want: approvalreport.OutcomeReplaced},
		{name: "041-C6 superseded is replaced", round: approval.Round{Outcome: approval.OutcomeSuperseded}, want: approvalreport.OutcomeReplaced},
		{name: "041-C6 cancelled is recalled", round: approval.Round{Outcome: approval.OutcomeCancelled}, want: approvalreport.OutcomeRecalled},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := approvalreport.OutcomeWord(tt.round); got != tt.want {
				t.Errorf("OutcomeWord() = %q, want %q", got, tt.want)
			}
		})
	}
}
