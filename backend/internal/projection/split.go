package projection

import (
	"mapp/internal/platform/date"
	"mapp/internal/platform/money"
)

// SplitMonth is one month: what happened, what is expected, and whether it is the one in progress.
type SplitMonth struct {
	Month      string `json:"month"`
	Actual     *int64 `json:"actual"`
	Projected  *int64 `json:"projected"`
	InProgress bool   `json:"inProgress"`
}

type Split struct {
	Months    []SplitMonth `json:"months"`
	Actual    int64        `json:"actual"`
	Projected int64        `json:"projected"`
	Total     int64        `json:"total"`
}

// splitAt is where today falls: 0 before the plan year, Months after it.
func splitAt(today date.Date, planYear int) int {
	switch {
	case today.Year() < planYear:
		return 0
	case today.Year() > planYear:
		return Months
	default:
		return int(today.Month()) - 1
	}
}

func SplitCounts(actual, projected []MonthFigure, today date.Date, planYear int) Split {
	at := splitAt(today, planYear)
	out := Split{Months: make([]SplitMonth, 0, Months)}
	for i := range Months {
		month := SplitMonth{Month: monthLabel(actual, projected, i), InProgress: i == at}
		if i <= at {
			value := valueAt(actual, i)
			month.Actual = &value
			out.Actual += valueBefore(value, i, at)
		}
		if i >= at {
			value := valueAt(projected, i)
			month.Projected = &value
			out.Projected += value
		}
		out.Months = append(out.Months, month)
	}
	out.Total = out.Actual + out.Projected
	return out
}

// valueBefore keeps the month in progress out of the actual total, so nothing is counted twice (FR-010).
func valueBefore(value int64, i, at int) int64 {
	if i < at {
		return value
	}
	return 0
}

func valueAt(months []MonthFigure, i int) int64 {
	if i < len(months) {
		return months[i].Value
	}
	return 0
}

func monthLabel(actual, projected []MonthFigure, i int) string {
	if i < len(projected) && projected[i].Month != "" {
		return projected[i].Month
	}
	if i < len(actual) {
		return actual[i].Month
	}
	return ""
}

// SplitMoneyMonth and SplitMoney carry the same rule for revenue, in one currency.
type SplitMoneyMonth struct {
	Month      string       `json:"month"`
	Actual     *money.Money `json:"actual"`
	Projected  *money.Money `json:"projected"`
	InProgress bool         `json:"inProgress"`
}

type SplitMoney struct {
	Months    []SplitMoneyMonth `json:"months"`
	Actual    money.Money       `json:"actual"`
	Projected money.Money       `json:"projected"`
	Total     money.Money       `json:"total"`
}

func SplitAmounts(actual, projected []MoneyFigure, today date.Date, planYear int) SplitMoney {
	at := splitAt(today, planYear)
	currency := currencyOf(actual, projected)
	out := SplitMoney{Months: make([]SplitMoneyMonth, 0, Months),
		Actual: money.Money{Currency: currency}, Projected: money.Money{Currency: currency}}
	for i := range Months {
		month := SplitMoneyMonth{Month: moneyLabel(actual, projected, i), InProgress: i == at}
		if i <= at {
			value := amountAt(actual, i, currency)
			month.Actual = &value
			if i < at {
				out.Actual.Amount = out.Actual.Amount.Add(value.Amount)
			}
		}
		if i >= at {
			value := amountAt(projected, i, currency)
			month.Projected = &value
			out.Projected.Amount = out.Projected.Amount.Add(value.Amount)
		}
		out.Months = append(out.Months, month)
	}
	out.Total = money.Money{Amount: out.Actual.Amount.Add(out.Projected.Amount), Currency: currency}
	return out
}

func amountAt(months []MoneyFigure, i int, currency string) money.Money {
	if i < len(months) {
		return money.Money{Amount: months[i].Value.Amount, Currency: currency}
	}
	return money.Money{Currency: currency}
}

func currencyOf(actual, projected []MoneyFigure) string {
	for _, months := range [][]MoneyFigure{projected, actual} {
		for _, m := range months {
			if m.Value.Currency != "" {
				return m.Value.Currency
			}
		}
	}
	return money.BaseCurrency
}

func moneyLabel(actual, projected []MoneyFigure, i int) string {
	if i < len(projected) && projected[i].Month != "" {
		return projected[i].Month
	}
	if i < len(actual) {
		return actual[i].Month
	}
	return ""
}
