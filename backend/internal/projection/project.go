package projection

import (
	"mapp/internal/platform/date"
	"mapp/internal/platform/money"
)

// Project runs the plan's model forward from the activity's goal: the goal stage phased across the
// activity's months, each later stage that stage's volume at the model's rate placed by its velocity, and
// revenue the deals at the model's average deal size — never past the plan year (FR-006–FR-009).
func Project(in Inputs) Projection {
	start, end, covered := window(in)
	out := Projection{Stamp: Stamp{ModelName: in.Model.Name, ParametersChangedAt: in.Model.ParametersChangedAt}}
	if !hasGoal(in.Goals) {
		return out
	}
	out.Present = true
	if !covered {
		return out
	}
	if in.Goals.SetLaterStagesMyself {
		return typed(out, in, start, end)
	}
	return calculated(out, in, start, end)
}

func hasGoal(g Goals) bool {
	if g.SetLaterStagesMyself {
		return g.INQ != nil || g.MQL != nil || g.Revenue != nil
	}
	return g.INQ != nil
}

// typed projects an activity that sets its later stages itself: each stage phases its own goal and the
// model's rates and velocity change nothing (FR-009, 047-C6).
func typed(out Projection, in Inputs, start, end date.Date) Projection {
	for _, stage := range []struct {
		name string
		goal *int64
	}{{StageINQ, in.Goals.INQ}, {StageMQL, in.Goals.MQL}} {
		if stage.goal != nil {
			out.Stages = append(out.Stages, seriesOf(stage.name, phase(*stage.goal, start, end)))
		}
	}
	if in.Goals.Revenue != nil {
		out.Revenue, out.Total = phaseMoney(*in.Goals.Revenue, start, end)
	}
	return out
}

// calculated projects an activity whose later stages come from the model (domain model §4.5).
func calculated(out Projection, in Inputs, start, end date.Date) Projection {
	inq := phase(*in.Goals.INQ, start, end)
	mql := shift(atRate(inq, in.Model.InqToMqlRate), in.Model.InqToMqlVelocityMonths)
	deals := shift(atRate(mql, in.Model.MqlToDealRate), in.Model.MqlToDealVelocityMonths)
	out.Stages = []Series{seriesOf(StageINQ, inq), seriesOf(StageMQL, mql), seriesOf(StageDeals, deals)}
	out.Revenue, out.Total = revenueOf(deals, in.Model.AverageDealSize)
	return out
}

// atRate converts each month at the rate, rounded half-up in that month, so a month's figure is the one a
// reader can check by hand (FR-007).
func atRate(months []MonthFigure, rate money.Percent) []MonthFigure {
	converted := make([]MonthFigure, len(months))
	for i, m := range months {
		converted[i] = MonthFigure{Month: m.Month, Value: rate.OfWhole(m.Value)}
	}
	return converted
}

// shift moves each month's figure forward by the velocity. What lands past the plan year is dropped:
// nothing projects beyond it, and bunching it into December would invent a month that never happens
// (FR-008, 020 FR-051).
func shift(months []MonthFigure, by int) []MonthFigure {
	shifted := make([]MonthFigure, len(months))
	for i, m := range months {
		shifted[i] = MonthFigure{Month: m.Month}
	}
	for i, m := range months {
		if landing := i + by; landing < len(shifted) {
			shifted[landing].Value = m.Value
		}
	}
	return shifted
}

func seriesOf(stage string, months []MonthFigure) Series {
	total := int64(0)
	for _, m := range months {
		total += m.Value
	}
	return Series{Stage: stage, Months: months, Total: total}
}

// revenueOf prices each month's deals at the model's average deal size. A month whose revenue overflows is
// left empty rather than failing the whole projection, as the goals themselves are shown (app adapter).
func revenueOf(deals []MonthFigure, size money.Money) ([]MoneyFigure, money.Money) {
	revenue := make([]MoneyFigure, len(deals))
	total := money.Money{Currency: size.Currency}
	for i, m := range deals {
		amount, err := size.Amount.MulInt(m.Value)
		if err != nil {
			amount = money.Amount{}
		}
		revenue[i] = MoneyFigure{Month: m.Month, Value: money.Money{Amount: amount, Currency: size.Currency}}
		total.Amount = total.Amount.Add(amount)
	}
	return revenue, total
}

// phaseMoney spreads a typed revenue goal across the window in its own currency, in whole minor units, so
// the months add up to the goal exactly.
func phaseMoney(goal money.Money, start, end date.Date) ([]MoneyFigure, money.Money) {
	cents := phase(goal.Amount.Cents(), start, end)
	revenue := make([]MoneyFigure, len(cents))
	for i, m := range cents {
		revenue[i] = MoneyFigure{Month: m.Month,
			Value: money.Money{Amount: money.FromCents(m.Value), Currency: goal.Currency}}
	}
	return revenue, goal
}
