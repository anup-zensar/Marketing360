package dataset

import (
	"fmt"
	"slices"
	"strings"
	"unicode"
	"unicode/utf8"

	"mapp/internal/platform/money"
)

const (
	inferRows     = 200
	previewRows   = 20
	maxSamples    = 5
	maxListed     = 10
	maxLabelRunes = 60
	moneyWords    = "cost amount spend budget price"
	currencyCodes = " USD EUR GBP CNY JPY INR CHF AUD CAD SGD HKD NZD SEK NOK DKK PLN CZK HUF ZAR BRL MXN AED SAR MYR THB IDR PHP KRW TRY "
)

// proposeKind follows data model §7: Date when every non-empty cell is a date, Number when every one is a
// number, Money when such a number column's name speaks of money, else Text.
func proposeKind(label string, cells []string) Kind {
	filled := slices.DeleteFunc(slices.Clone(cells), func(c string) bool { return c == "" })
	switch {
	case len(filled) == 0:
		return KindText
	case allFit(KindDate, filled):
		return KindDate
	case !allFit(KindNumber, filled):
		return KindText
	case isMoneyName(label) && allFit(KindMoney, filled):
		return KindMoney
	}
	return KindNumber
}

func allFit(k Kind, cells []string) bool {
	for _, c := range cells {
		if _, ok := parseCell(k, c); !ok {
			return false
		}
	}
	return true
}

func isMoneyName(label string) bool {
	lower := strings.ToLower(label)
	for _, word := range strings.Fields(moneyWords) {
		if strings.Contains(lower, word) {
			return true
		}
	}
	for _, word := range strings.FieldsFunc(label, func(r rune) bool { return !unicode.IsLetter(r) }) {
		if strings.Contains(currencyCodes, " "+word+" ") {
			return true
		}
	}
	return false
}

// labelOf names column i from the header: trimmed, at most 60 characters, "Column 3" when the cell is empty.
func labelOf(f file, i int) string {
	label := f.header.Cell(i)
	if label == "" {
		return fmt.Sprintf(msgColumnN, i+1)
	}
	if utf8.RuneCountInString(label) > maxLabelRunes {
		return string([]rune(label)[:maxLabelRunes])
	}
	return label
}

func previewOf(f file) Preview {
	p := Preview{FileName: f.name, SheetName: f.sheet, RowCount: len(f.rows), Columns: []PreviewColumn{}, Rows: [][]string{}}
	for i := range f.width {
		p.Columns = append(p.Columns, previewColumn(f, i))
	}
	for _, r := range f.rows[:min(previewRows, len(f.rows))] {
		row := make([]string, f.width)
		for i := range row {
			row[i] = r.Cell(i)
		}
		p.Rows = append(p.Rows, row)
	}
	p.Problems = previewProblems(f, p.Columns)
	return p
}

func previewColumn(f file, i int) PreviewColumn {
	c := PreviewColumn{Position: i, Label: labelOf(f, i), Samples: []string{}}
	c.ProposedKind = proposeKind(c.Label, f.column(i, inferRows))
	if c.ProposedKind == KindMoney {
		usd := money.BaseCurrency
		c.ProposedCurrency = &usd
	}
	for _, cell := range f.column(i, len(f.rows)) {
		switch {
		case cell == "":
			c.EmptyCount++
		case len(c.Samples) < maxSamples && !slices.Contains(c.Samples, cell):
			c.Samples = append(c.Samples, cell)
		}
	}
	return c
}

func previewProblems(f file, columns []PreviewColumn) []string {
	problems := []string{}
	seen := map[string]bool{}
	for _, c := range columns {
		key := strings.ToLower(c.Label)
		switch {
		case seen[key]:
			problems = append(problems, fmt.Sprintf(msgDuplicateName, c.Label))
		case f.header.Cell(c.Position) == "":
			problems = append(problems, fmt.Sprintf(msgNoName, c.Position+1, c.Label))
		case c.EmptyCount == len(f.rows) && len(f.rows) > 0:
			problems = append(problems, fmt.Sprintf(msgEmptyColumn, c.Label))
		}
		seen[key] = true
	}
	return append(problems, longCells(f, columns)...)
}

func longCells(f file, columns []PreviewColumn) []string {
	problems, more := []string{}, 0
	for _, r := range f.rows {
		for _, c := range columns {
			if utf8.RuneCountInString(r.Cell(c.Position)) <= maxCellRunes {
				continue
			}
			if len(problems) < maxListed {
				problems = append(problems, fmt.Sprintf(msgLongCell, r.Number, c.Label))
			} else {
				more++
			}
		}
	}
	if more > 0 {
		problems = append(problems, fmt.Sprintf(msgMoreLongCells, more))
	}
	return problems
}
