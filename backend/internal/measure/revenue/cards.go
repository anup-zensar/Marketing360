package revenue

import (
	"mapp/internal/platform/money"
	"mapp/internal/rollup"
)

// revenueCards builds the fixed card set from the period totals (049-C9): goal, actual, projected and
// performance. A card with nothing behind it shows its reason rather than a zero (FR-009).
func revenueCards(periods []RevenuePeriod, anyGoal bool) RevenueCards {
	var goal, actual, projected money.Amount
	hasActual, hasProjected := false, false
	for _, p := range periods {
		if p.Goal != nil {
			goal = goal.Add(p.Goal.Amount)
		}
		if p.Actual != nil {
			actual = actual.Add(p.Actual.Amount)
			hasActual = true
		}
		if p.Projected != nil {
			projected = projected.Add(p.Projected.Amount)
			hasProjected = true
		}
	}
	cards := RevenueCards{}
	if anyGoal {
		shown := money.USD(goal)
		cards.Goal = MoneyCard{Value: &shown}
	} else {
		cards.Goal = MoneyCard{Reason: ReasonNoGoal}
	}
	if hasActual {
		shown := money.USD(actual)
		cards.Actual = MoneyCard{Value: &shown}
	} else {
		cards.Actual = MoneyCard{Reason: ReasonNoData}
	}
	total := actual
	switch {
	case hasProjected:
		total = total.Add(projected)
		shown := money.USD(total)
		cards.Projected = MoneyCard{Value: &shown}
	case hasActual:
		shown := money.USD(total)
		cards.Projected = MoneyCard{Value: &shown}
	default:
		cards.Projected = MoneyCard{Reason: ReasonNoGoal}
	}
	cards.Performance = performanceOf(total, goal, anyGoal)
	return cards
}

// performanceOf is (actual + projected) ÷ goal as a whole per cent (049-C15, 049-C16), computed exactly on
// cents (rollup.BudgetUsedPercent) rather than as a float division of money.
func performanceOf(actualOrProjected, goal money.Amount, anyGoal bool) NumberCard {
	if !anyGoal {
		return NumberCard{Reason: ReasonNoGoal}
	}
	a, g := money.USD(actualOrProjected), money.USD(goal)
	pct := rollup.BudgetUsedPercent(&a, &g)
	if pct == nil {
		return NumberCard{Reason: ReasonNoGoal}
	}
	value := float64(*pct)
	return NumberCard{Value: &value}
}
