package dataset

import (
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"mapp/internal/platform/date"
	"mapp/internal/platform/money"
	"mapp/internal/widgetdata"
)

const (
	maxCellRunes = 500
	groupDigits  = 3
)

// parseCell types the text of one cell for a column of kind k; ok is false when the text does not fit the kind.
func parseCell(k Kind, text string) (Cell, bool) {
	if text == "" {
		return Cell{Empty: true}, true
	}
	switch k {
	case KindNumber:
		n, ok := parseNumber(text)
		return TextCell(n.String()), ok
	case KindDate:
		d, ok := parseDate(text)
		return TextCell(d.String()), ok
	case KindMoney:
		minor, ok := parseMoney(text)
		return MinorCell(minor), ok
	case KindText:
	}
	return TextCell(text), utf8.RuneCountInString(text) <= maxCellRunes
}

func parseNumber(text string) (widgetdata.Number, bool) {
	plain, ok := plainDecimal(text)
	if !ok {
		return widgetdata.Number{}, false
	}
	n, err := widgetdata.ParseNumber(plain)
	return n, err == nil
}

func parseMoney(text string) (int64, bool) {
	plain, ok := plainDecimal(text)
	if !ok {
		return 0, false
	}
	amount, err := money.ParseAmount(plain)
	return amount.Cents(), err == nil
}

func parseDate(text string) (date.Date, bool) {
	for _, layout := range [...]string{time.DateOnly, "2/1/2006", "2 Jan 2006"} {
		if t, err := time.Parse(layout, text); err == nil {
			return date.New(t.Year(), t.Month(), t.Day()), true
		}
	}
	return date.Date{}, false
}

// plainDecimal reads "1 234,5", "1,234.50" or "-1.234.567" as "1234.5", "1234.50", "-1234567". With one
// separator only, a comma followed by three digits groups thousands, as English files write them.
func plainDecimal(text string) (string, bool) {
	s := strings.Map(dropSpace, text)
	sign := ""
	if rest, negative := strings.CutPrefix(s, "-"); negative {
		sign, s = "-", rest
	}
	mark := decimalMark(s)
	whole, fraction, hasFraction := s, "", false
	if mark != 0 {
		i := strings.LastIndexByte(s, mark)
		whole, fraction, hasFraction = s[:i], s[i+1:], true
	}
	digits, ok := ungrouped(whole, thousandsMark(mark))
	if !ok || (hasFraction && fraction == "") {
		return "", false
	}
	if hasFraction {
		return sign + digits + "." + fraction, true
	}
	return sign + digits, true
}

func dropSpace(r rune) rune {
	if unicode.IsSpace(r) || r == ' ' {
		return -1
	}
	return r
}

func decimalMark(s string) byte {
	dot, comma := strings.LastIndexByte(s, '.'), strings.LastIndexByte(s, ',')
	switch {
	case dot >= 0 && comma >= 0 && dot > comma:
		return '.'
	case dot >= 0 && comma >= 0:
		return ','
	case comma >= 0:
		return commaAlone(s, comma)
	case strings.Count(s, ".") == 1:
		return '.'
	}
	return 0
}

func commaAlone(s string, at int) byte {
	if strings.Count(s, ",") > 1 || (len(s)-at-1 == groupDigits && s[:at] != "0") {
		return 0
	}
	return ','
}

func thousandsMark(decimal byte) byte {
	if decimal == ',' {
		return '.'
	}
	return ','
}

// ungrouped removes the thousands separator from whole, which must then group its digits by three; with no
// decimal mark, either separator may group.
func ungrouped(whole string, sep byte) (string, bool) {
	mark := sep
	if strings.IndexByte(whole, sep) < 0 {
		mark = thousandsMark(sep)
	}
	groups := strings.Split(whole, string(mark))
	if len(groups) == 1 {
		return whole, true
	}
	return strings.Join(groups, ""), areGroupsOfThree(groups)
}

func areGroupsOfThree(groups []string) bool {
	if first := len(groups[0]); first == 0 || first > groupDigits {
		return false
	}
	for _, g := range groups[1:] {
		if len(g) != groupDigits {
			return false
		}
	}
	return true
}
