package settings

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"mapp/internal/platform/errs"
)

// maxVelocityMonths bounds a transition's lag: two years covers every sales cycle the sources describe.
const maxVelocityMonths = 24

const (
	velocityMessage   = "Enter a whole number of months from 0 to 24."
	goalStagesMessage = "Tick at least one of INQ, MQL and Revenue."
)

func applyProjectionFields(m Model, update ModelUpdate, problems *errs.Problems) Model {
	if update.InqToMqlVelocityMonths != nil {
		m.InqToMqlVelocityMonths = parseVelocity(problems, "inqToMqlVelocityMonths", *update.InqToMqlVelocityMonths)
	}
	if update.MqlToDealVelocityMonths != nil {
		m.MqlToDealVelocityMonths = parseVelocity(problems, "mqlToDealVelocityMonths", *update.MqlToDealVelocityMonths)
	}
	if update.AllowedGoalStages != nil {
		m.AllowedGoalStages = parseGoalStages(problems, *update.AllowedGoalStages)
	}
	return m
}

func parseVelocity(problems *errs.Problems, field string, months int) int {
	if months < 0 || months > maxVelocityMonths {
		problems.Add(field, velocityMessage)
		return 0
	}
	return months
}

// parseGoalStages keeps the funnel's own order whatever order the form sent, and refuses an empty set or an
// unknown stage with one message (047-C2).
func parseGoalStages(problems *errs.Problems, stages []string) []string {
	kept := make([]string, 0, 3)
	for _, stage := range allGoalStages() {
		if slices.Contains(stages, stage) {
			kept = append(kept, stage)
		}
	}
	unknown := slices.ContainsFunc(stages, func(s string) bool { return !slices.Contains(allGoalStages(), s) })
	if len(kept) == 0 || unknown {
		problems.Add("allowedGoalStages", goalStagesMessage)
	}
	return kept
}

func valueOrZero(months *int) *int {
	if months == nil {
		zero := 0
		return &zero
	}
	return months
}

const stageRemovedMessage = "%s set their goal at a stage this model would no longer allow. They keep it " +
	"and it keeps projecting; new goals can't choose it."

// stageRemovedWarnings warns, per goal stage the change removes, of the live plans whose goal sits there
// (FR-003, 047-C3). Existing goals are never rewritten; the warning is the whole effect.
func stageRemovedWarnings(before, after Model, refs []Reference) []ImpactWarning {
	warnings := []ImpactWarning{}
	for _, stage := range before.AllowedGoalStages {
		if slices.Contains(after.AllowedGoalStages, stage) {
			continue
		}
		n := 0
		for _, ref := range refs {
			if !ref.InBin && ref.Level == LevelPlan && ref.ModelID == before.ID && ref.GoalStage == stage {
				n++
			}
		}
		if n > 0 {
			warnings = append(warnings, ImpactWarning{Kind: ImpactGoalStageRemoved,
				Message: fmt.Sprintf(stageRemovedMessage, plural(n, "plan", "plans")), Count: n})
		}
	}
	return warnings
}

// The goal stages' words in the activity log, Shell's own (design system §8).
const (
	wordINQ     = "INQ"
	wordMQL     = "MQL"
	wordRevenue = "Revenue"
)

func goalStageWord(stage string) string {
	switch stage {
	case GoalStageINQ:
		return wordINQ
	case GoalStageMQL:
		return wordMQL
	case GoalStageRevenue:
		return wordRevenue
	}
	return ""
}

func monthsText(value any) string {
	number, ok := value.(json.Number)
	if !ok {
		return ""
	}
	if number == "1" {
		return "1 month"
	}
	return string(number) + " months"
}

func stagesText(value any) string {
	items, _ := value.([]any)
	words := make([]string, 0, len(items))
	for _, item := range items {
		stage, _ := item.(string)
		words = append(words, goalStageWord(stage))
	}
	return strings.Join(words, ", ")
}
