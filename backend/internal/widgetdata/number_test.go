package widgetdata_test

import (
	"errors"
	"math"
	"testing"

	"mapp/internal/widgetdata"
)

func TestNumber_ParseAndString(t *testing.T) {
	t.Parallel()
	valid := []struct {
		name, in, text string
		units          int64
	}{
		{"FR-019 a decimal keeps its shortest text", "12.5", "12.5", 125000},
		{"FR-019 a negative whole number", "-3", "-3", -30000},
		{"FR-019 the smallest step", "0.0001", "0.0001", 1},
		{"FR-019 a small negative fraction keeps its sign", "-0.0005", "-0.0005", -5},
		{"FR-019 leading and trailing zeros are dropped", "0012.5000", "12.5", 125000},
		{"FR-019 minus zero is zero", "-0.0", "0", 0},
		{"FR-018 the largest number that fits", "922337203685477.5807", "922337203685477.5807", math.MaxInt64},
	}
	for _, tt := range valid {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			n, err := widgetdata.ParseNumber(tt.in)
			if err != nil || n.Units() != tt.units || n.String() != tt.text {
				t.Errorf("ParseNumber(%q) = %d %q, %v; want %d %q", tt.in, n.Units(), n.String(), err, tt.units, tt.text)
			}
		})
	}

	invalid := []struct {
		name, in string
		want     error
	}{
		{"FR-019 empty", "", widgetdata.ErrNumberFormat},
		{"FR-019 a lone minus", "-", widgetdata.ErrNumberFormat},
		{"FR-019 a point without decimals", "1.", widgetdata.ErrNumberFormat},
		{"FR-019 decimals without a whole part", ".5", widgetdata.ErrNumberFormat},
		{"FR-019 thousands separators", "1,000", widgetdata.ErrNumberFormat},
		{"FR-019 five decimals", "1.23456", widgetdata.ErrNumberFormat},
		{"FR-019 sixteen integer digits", "1234567890123456", widgetdata.ErrNumberFormat},
		{"FR-019 a plus sign", "+1", widgetdata.ErrNumberFormat},
		{"FR-019 spaces", " 1", widgetdata.ErrNumberFormat},
		{"FR-019 an exponent", "1e3", widgetdata.ErrNumberFormat},
		{"FR-018 fifteen digits beyond the range", "999999999999999", widgetdata.ErrNumberRange},
	}
	for _, tt := range invalid {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if _, err := widgetdata.ParseNumber(tt.in); !errors.Is(err, tt.want) {
				t.Errorf("ParseNumber(%q) error = %v, want %v", tt.in, err, tt.want)
			}
		})
	}

	t.Run("FR-018 the smallest units print exactly", func(t *testing.T) {
		t.Parallel()
		if got := widgetdata.NumberOfUnits(math.MinInt64).String(); got != "-922337203685477.5808" {
			t.Errorf("String() = %q, want -922337203685477.5808", got)
		}
	})
}
