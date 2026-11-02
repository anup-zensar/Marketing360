package revenue

import "math"

// funnelCards builds the fixed Funnel card set (049-C9, 049-C13).
func funnelCards(periods []FunnelPeriod, anyGoal bool, previousActual *int64) FunnelCards {
	var goal, actual, projected int64
	hasActual, hasProjected, hasGoal := false, false, false
	for _, p := range periods {
		if p.Goal != nil {
			goal += *p.Goal
			hasGoal = true
		}
		if p.Actual != nil {
			actual += *p.Actual
			hasActual = true
		}
		if p.Projected != nil {
			projected += *p.Projected
			hasProjected = true
		}
	}
	cards := FunnelCards{}
	cards.Goal = intCard(goal, hasGoal && anyGoal, ReasonNoGoal)
	cards.Actual = intCard(actual, hasActual, ReasonNoData)
	total := actual
	if hasProjected {
		total += projected
	}
	cards.Projected = intCard(total, hasActual || hasProjected, ReasonNoGoal)
	cards.Performance = ratioCard(total, goal, anyGoal)
	cards.Conversion = conversionCard(actual, previousActual)
	return cards
}

func intCard(value int64, present bool, reason string) NumberCard {
	if !present {
		return NumberCard{Reason: reason}
	}
	v := float64(value)
	return NumberCard{Value: &v}
}

// ratioCard is (actual + projected) ÷ goal as a whole per cent (049-C15).
func ratioCard(value, goal int64, anyGoal bool) NumberCard {
	if !anyGoal || goal == 0 {
		return NumberCard{Reason: ReasonNoGoal}
	}
	pct := math.Round(float64(value) / float64(goal) * 100)
	return NumberCard{Value: &pct}
}

// conversionCard mirrors 046's conversionOf: one decimal, "—" on a zero denominator (FR-017).
func conversionCard(actual int64, previous *int64) NumberCard {
	if previous == nil || *previous == 0 {
		return NumberCard{Reason: ReasonNoPreviousStage}
	}
	rate := roundToOneDecimal(float64(actual) / float64(*previous) * 100)
	return NumberCard{Value: &rate}
}

func roundToOneDecimal(v float64) float64 {
	return float64(int64(v*10+0.5)) / 10
}
