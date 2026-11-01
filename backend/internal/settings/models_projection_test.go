package settings

import (
	"errors"
	"slices"
	"testing"

	"mapp/internal/platform/errs"
)

func intp(n int) *int { return &n }

func TestCreateModel_VelocityAndStagesDefaultToZeroAndEveryStage_FR001_FR002(t *testing.T) {
	t.Parallel()
	f := newTestService(t)
	event := f.addType(t, "Event", true)

	m := f.createModel(t, modelInput("Lubricants 1.0", event.ID))

	if m.InqToMqlVelocityMonths != 0 || m.MqlToDealVelocityMonths != 0 {
		t.Errorf("velocities = %d, %d, want 0 and 0: nothing existing changes meaning", m.InqToMqlVelocityMonths, m.MqlToDealVelocityMonths)
	}
	if !slices.Equal(m.AllowedGoalStages, allGoalStages()) {
		t.Errorf("allowed stages = %v, want every stage", m.AllowedGoalStages)
	}
}

func TestUpdateModel_VelocityOutOfRangeRefusedTogether_FR001(t *testing.T) {
	t.Parallel()
	f := newTestService(t)
	event := f.addType(t, "Event", true)
	m := f.createModel(t, modelInput("Lubricants 1.0", event.ID))

	_, err := f.service.UpdateModel(t.Context(), alex, m.ID, ModelUpdate{
		InqToMqlVelocityMonths: intp(-1), MqlToDealVelocityMonths: intp(25), Version: version(1),
	})

	refusal, ok := errors.AsType[*errs.Error](err)
	if !ok || len(refusal.Fields) != 2 {
		t.Fatalf("UpdateModel = %v, want both velocities refused together as 0 to 24", err)
	}
}

func TestUpdateModel_StagesKeepFunnelOrderAndEmptyOrUnknownRefused_FR002(t *testing.T) {
	t.Parallel()
	f := newTestService(t)
	event := f.addType(t, "Event", true)
	m := f.createModel(t, modelInput("Lubricants 1.0", event.ID))

	saved, err := f.service.UpdateModel(t.Context(), alex, m.ID, ModelUpdate{
		AllowedGoalStages: &[]string{GoalStageRevenue, GoalStageINQ}, Version: version(1),
	})
	if err != nil || !slices.Equal(saved.AllowedGoalStages, []string{GoalStageINQ, GoalStageRevenue}) {
		t.Fatalf("saved stages = %v (%v), want the funnel's own order whatever the form sent", saved.AllowedGoalStages, err)
	}

	for name, stages := range map[string][]string{"none": {}, "unknown": {"deal"}} {
		if _, err := f.service.UpdateModel(t.Context(), alex, m.ID, ModelUpdate{
			AllowedGoalStages: &[]string{}, Version: version(2),
		}); name == "none" && err == nil {
			t.Error("an empty stage set was saved, want it refused")
		} else if name == "unknown" {
			if _, err := f.service.UpdateModel(t.Context(), alex, m.ID, ModelUpdate{
				AllowedGoalStages: &stages, Version: version(2),
			}); err == nil {
				t.Error("an unknown stage was saved, want it refused")
			}
		}
	}
}

func TestUpdateModel_VelocityChangeIsAuditedAndVersionChecked_FR003(t *testing.T) {
	t.Parallel()
	f := newTestService(t)
	event := f.addType(t, "Event", true)
	m := f.createModel(t, modelInput("Lubricants 1.0", event.ID))

	saved, err := f.service.UpdateModel(t.Context(), alex, m.ID, ModelUpdate{
		InqToMqlVelocityMonths: intp(1), MqlToDealVelocityMonths: intp(2), Version: version(1),
	})
	if err != nil || saved.InqToMqlVelocityMonths != 1 || saved.MqlToDealVelocityMonths != 2 {
		t.Fatalf("UpdateModel = %+v (%v), want the velocities saved", saved, err)
	}
	entry := f.lastEntry(t)
	if len(entry.Changes) != 2 || entry.Changes[0].Label != "INQ→MQL velocity" || entry.Changes[0].NewText != "1 month" {
		t.Errorf("changes = %+v, want each velocity logged in words", entry.Changes)
	}

	_, err = f.service.UpdateModel(t.Context(), alex, m.ID, ModelUpdate{InqToMqlVelocityMonths: intp(3), Version: version(1)})
	refusal, ok := errors.AsType[*errs.Error](err)
	if !ok || refusal.Code != "version_conflict" {
		t.Errorf("a stale save = %v, want version_conflict naming who saved last", err)
	}
}

func TestModelImpact_NarrowedStagesNameTheStrandedPlans_FR003(t *testing.T) {
	t.Parallel()
	f := newTestService(t)
	event := f.addType(t, "Event", true)
	m := f.createModel(t, modelInput("Lubricants 1.0", event.ID))
	mql := planRef("P-1001", "Plan A", m.ID, nil)
	mql.GoalStage = GoalStageMQL
	inq := planRef("P-1002", "Plan B", m.ID, nil)
	inq.GoalStage = GoalStageINQ
	f.refs = []Reference{mql, inq}

	impact, err := f.service.ModelImpact(t.Context(), alex, m.ID, ModelUpdate{
		AllowedGoalStages: &[]string{GoalStageINQ, GoalStageRevenue},
	})
	if err != nil {
		t.Fatalf("ModelImpact: %v", err)
	}
	if len(impact.Warnings) != 1 || impact.Warnings[0].Kind != ImpactGoalStageRemoved || impact.Warnings[0].Count != 1 {
		t.Fatalf("warnings = %+v, want the one plan whose MQL goal would be stranded", impact.Warnings)
	}

	after, err := f.service.UpdateModel(t.Context(), alex, m.ID, ModelUpdate{
		AllowedGoalStages: &[]string{GoalStageINQ, GoalStageRevenue}, Version: version(1),
	})
	if err != nil || len(after.AllowedGoalStages) != 2 {
		t.Errorf("the narrowed save = %+v (%v), want it saved: existing goals stand, only new entries are guided", after, err)
	}
}
