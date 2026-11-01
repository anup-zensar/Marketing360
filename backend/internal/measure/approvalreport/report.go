package approvalreport

import (
	"context"

	"mapp/internal/measure"
	"mapp/internal/platform/actor"
	"mapp/internal/platform/sheet"
)

// Report is the Approvals report as the hub's registry holds it.
type Report struct {
	service *Service
}

func NewReport(s *Service) *Report {
	return &Report{service: s}
}

func (r *Report) Key() string      { return measure.KeyApprovals }
func (r *Report) Available() bool  { return true }
func (r *Report) Params() []string { return Params() }

func (r *Report) Sheets(ctx context.Context, a actor.Actor, f measure.Filter) ([]measure.WidgetSheet, error) {
	sn, err := r.service.snapshot(ctx, a, f)
	if err != nil {
		return nil, err
	}
	return []measure.WidgetSheet{
		{Key: "status", Sheet: statusSheet(sn)},
		{Key: "average", Sheet: averageSheet(sn)},
	}, nil
}

func statusSheet(sn snapshot) sheet.Sheet {
	rows := [][]sheet.Cell{}
	for _, t := range allRecordTypes() {
		mix := statusMix(sn.entries, t)
		rows = append(rows, []sheet.Cell{sheet.Bold(recordTypeLabel(t))},
			[]sheet.Cell{sheet.Bold("Status"), sheet.Bold("Count"), sheet.Bold("Share (%)")})
		for _, item := range mix.Items {
			rows = append(rows, []sheet.Cell{sheet.Text(outcomeLabel(item.Outcome)), sheet.Number(int64(item.Count)),
				sheet.Percent(int64(item.SharePercent))})
		}
	}
	return sheet.Sheet{Name: "Current approval status", Rows: rows}
}

func averageSheet(sn snapshot) sheet.Sheet {
	avg := averages(sn.entries, sn.actions)
	rows := [][]sheet.Cell{
		{sheet.Bold("Approver"), sheet.Bold("Days"), sheet.Bold("Decisions")},
		{sheet.Text("All approvers"), sheet.Text(daysText(avg.Days)), sheet.Number(int64(avg.Decisions))},
	}
	for _, p := range avg.Approvers {
		rows = append(rows, []sheet.Cell{sheet.Text(p.Person.Name), sheet.Text(p.Days), sheet.Number(int64(p.Decisions))})
	}
	return sheet.Sheet{Name: "Average days to a decision", Rows: rows}
}

func daysText(days *string) string {
	if days == nil {
		return "—"
	}
	return *days
}
