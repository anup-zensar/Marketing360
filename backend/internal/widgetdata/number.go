package widgetdata

import (
	"errors"
	"math"
	"strconv"
	"strings"
)

const (
	numberScale      = 10000
	numberDecimals   = 4
	maxIntegerDigits = 15
)

var (
	ErrNumberFormat = errors.New("widgetdata: not a number of up to 15 integer digits and 4 decimals")
	ErrNumberRange  = errors.New("widgetdata: number too large")
)

// Number is an exact decimal of ten-thousandths: "12.5" is 125000.
type Number struct{ units int64 }

func NumberOfUnits(units int64) Number {
	return Number{units: units}
}

// ParseNumber reads "-12.5" without thousands separators; beyond ±922,337,203,685,477.5807 it is ErrNumberRange.
func ParseNumber(s string) (Number, error) {
	negative := strings.HasPrefix(s, "-")
	whole, fraction, hasPoint := strings.Cut(strings.TrimPrefix(s, "-"), ".")
	fractionOK := !hasPoint || isDigits(fraction, numberDecimals)
	if !isDigits(whole, maxIntegerDigits) || !fractionOK {
		return Number{}, ErrNumberFormat
	}
	units, err := unitsOf(whole, fraction)
	if err != nil {
		return Number{}, err
	}
	if negative {
		units = -units
	}
	return Number{units: units}, nil
}

func (n Number) Units() int64 {
	return n.units
}

func (n Number) String() string {
	whole, fraction := n.units/numberScale, n.units%numberScale
	sign := ""
	if n.units < 0 {
		sign, whole, fraction = "-", -whole, -fraction
	}
	text := sign + strconv.FormatInt(whole, 10)
	if fraction == 0 {
		return text
	}
	padded := strconv.FormatInt(fraction+numberScale, 10)[1:]
	return text + "." + strings.TrimRight(padded, "0")
}

func isDigits(s string, maxLength int) bool {
	if s == "" || len(s) > maxLength {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func unitsOf(whole, fraction string) (int64, error) {
	w, err := strconv.ParseInt(whole, 10, 64)
	if err != nil {
		return 0, ErrNumberFormat
	}
	var f int64
	if fraction != "" {
		f, err = strconv.ParseInt(fraction+strings.Repeat("0", numberDecimals-len(fraction)), 10, 64)
		if err != nil {
			return 0, ErrNumberFormat
		}
	}
	if w > (math.MaxInt64-f)/numberScale {
		return 0, ErrNumberRange
	}
	return w*numberScale + f, nil
}

func (n Number) MarshalJSON() ([]byte, error) {
	return []byte(strconv.Quote(n.String())), nil
}
