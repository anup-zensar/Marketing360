package projection

import (
	"mapp/internal/platform/date"
	"mapp/internal/platform/money"
)

// Project runs the model forward from the goal stage only, never past the plan year (FR-006 - FR-009).
func Project(in Inputs) Projection {
	start, end, covered := window(in)
	out := Projection{Stamp: Stamp{ModelName: in.Model.Name, ModelChangedAt: in.Model.ModelChangedAt}}
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

// typed phases each stage's own goal: the override means the rates and velocity change nothing (FR-009).
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

func calculated(out Projection, in Inputs, start, end date.Date) Projection {
	inq := phase(*in.Goals.INQ, start, end)
	mql := shift(atRate(inq, in.Model.InqToMqlRate), in.Model.InqToMqlVelocityMonths)
	deals := shift(atRate(mql, in.Model.MqlToDealRate), in.Model.MqlToDealVelocityMonths)
	out.Stages = []Series{seriesOf(StageINQ, inq), seriesOf(StageMQL, mql), seriesOf(StageDeals, deals)}
	out.Revenue, out.Total = revenueOf(deals, in.Model.AverageDealSize)
	return out
}

// atRate rounds half-up within each month, so every month's figure is one a reader can check by hand.
func atRate(months []MonthFigure, rate money.Percent) []MonthFigure {
	converted := make([]MonthFigure, len(months))
	for i, m := range months {
		converted[i] = MonthFigure{Month: m.Month, Value: rate.OfWhole(m.Value)}
	}
	return converted
}

// shift drops what lands past the plan year: bunching it into December would invent a month (FR-008).
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

func phaseMoney(goal money.Money, start, end date.Date) ([]MoneyFigure, money.Money) {
	cents := phase(goal.Amount.Cents(), start, end)
	revenue := make([]MoneyFigure, len(cents))
	for i, m := range cents {
		revenue[i] = MoneyFigure{Month: m.Month,
			Value: money.Money{Amount: money.FromCents(m.Value), Currency: goal.Currency}}
	}
	return revenue, goal
}
