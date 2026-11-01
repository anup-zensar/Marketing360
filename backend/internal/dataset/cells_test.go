package dataset

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"mapp/internal/platform/money"
	"mapp/internal/widgetdata"
)

func TestParseCells_EveryKind_FR042(t *testing.T) {
	t.Parallel()
	tests := []struct {
		kind Kind
		text string
		want Cell
		ok   bool
	}{
		{KindNumber, "", Cell{Empty: true}, true},
		{KindNumber, "140", TextCell("140"), true},
		{KindNumber, "1 234,5", TextCell("1234.5"), true},
		{KindNumber, "1 234,5", TextCell("1234.5"), true},
		{KindNumber, "1,234.50", TextCell("1234.5"), true},
		{KindNumber, "1.234,50", TextCell("1234.5"), true},
		{KindNumber, "-1.234.567", TextCell("-1234567"), true},
		{KindNumber, "1,234,567", TextCell("1234567"), true},
		{KindNumber, "1,234", TextCell("1234"), true},
		{KindNumber, "0,123", TextCell("0.123"), true},
		{KindNumber, "12,5", TextCell("12.5"), true},
		{KindNumber, "12.3456", TextCell("12.3456"), true},
		{KindNumber, "12.34567", Cell{}, false},
		{KindNumber, "1,23,456", Cell{}, false},
		{KindNumber, "12,", Cell{}, false},
		{KindNumber, "n/a", Cell{}, false},
		{KindNumber, "1234567890123456", Cell{}, false},
		{KindDate, "2026-03-05", TextCell("2026-03-05"), true},
		{KindDate, "05/03/2026", TextCell("2026-03-05"), true},
		{KindDate, "5/3/2026", TextCell("2026-03-05"), true},
		{KindDate, "5 Mar 2026", TextCell("2026-03-05"), true},
		{KindDate, "05 mar 2026", TextCell("2026-03-05"), true},
		{KindDate, "31/02/2026", Cell{}, false},
		{KindDate, "2026-13-01", Cell{}, false},
		{KindDate, "March 5", Cell{}, false},
		{KindMoney, "1,200.50", MinorCell(120050), true},
		{KindMoney, "-7", MinorCell(-700), true},
		{KindMoney, "12.345", Cell{}, false},
		{KindMoney, "€12", Cell{}, false},
		{KindText, "Expo Hamburg", TextCell("Expo Hamburg"), true},
		{KindText, strings.Repeat("é", maxCellRunes), TextCell(strings.Repeat("é", maxCellRunes)), true},
		{KindText, strings.Repeat("x", maxCellRunes+1), Cell{}, false},
	}
	for _, tt := range tests {
		got, ok := parseCell(tt.kind, tt.text)
		if ok != tt.ok {
			t.Errorf("parseCell(%s, %q) ok = %v, want %v", tt.kind, tt.text, ok, tt.ok)
			continue
		}
		if ok {
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("parseCell(%s, %q) (-want +got):\n%s", tt.kind, tt.text, diff)
			}
		}
	}
}

func FuzzParseCells_RoundTrip_FR042(f *testing.F) {
	for _, seed := range []string{"1 234,5", "-0.0001", "1,234", "05/03/2026", "1.234.567,89", "abc", ""} {
		f.Add(seed, int64(12345))
	}
	f.Fuzz(func(t *testing.T, text string, units int64) {
		for _, k := range []Kind{KindNumber, KindDate, KindMoney, KindText} {
			cell, ok := parseCell(k, text)
			if !ok || cell.Empty || k == KindMoney {
				continue
			}
			again, ok := parseCell(k, cell.Text)
			if !ok || again.Text != cell.Text {
				t.Fatalf("parseCell(%s, %q) = %q, which reads back as %q, %v", k, text, cell.Text, again.Text, ok)
			}
		}
		n := widgetdata.NumberOfUnits(units % 1_000_000_000_000_000_000)
		if cell, ok := parseCell(KindNumber, n.String()); !ok || cell.Text != n.String() {
			t.Fatalf("number %s reads back as %q, %v", n, cell.Text, ok)
		}
		cents := money.FromCents(units % 99_999_999_999_999_999)
		if cell, ok := parseCell(KindMoney, cents.String()); !ok || *cell.Minor != cents.Cents() {
			t.Fatalf("amount %s reads back as %v, %v", cents, cell.Minor, ok)
		}
	})
}

func TestInferKinds_FR041(t *testing.T) {
	t.Parallel()
	tests := []struct {
		label string
		cells []string
		want  Kind
	}{
		{"Date", []string{"2026-01-22", "", "22/01/2026", "5 Mar 2026"}, KindDate},
		{"Leads", []string{"140", "1,200", ""}, KindNumber},
		{"Agency cost", []string{"18500", "9200.50"}, KindMoney},
		{"Total spend", []string{"1"}, KindMoney},
		{"Amount (EUR)", []string{"1"}, KindMoney},
		{"Fee EUR", []string{"1"}, KindMoney},
		{"Unit price", []string{"1.2345"}, KindNumber},
		{"Leads", []string{"140", "n/a"}, KindText},
		{"Event", []string{"Expo"}, KindText},
		{"Notes", []string{"", ""}, KindText},
		{"Year", []string{"2026"}, KindNumber},
	}
	for _, tt := range tests {
		if got := proposeKind(tt.label, tt.cells); got != tt.want {
			t.Errorf("proposeKind(%q, %q) = %s, want %s", tt.label, tt.cells, got, tt.want)
		}
	}
}

func TestInferKinds_FromTheFirst200Rows_FR041(t *testing.T) {
	t.Parallel()
	csv := "Leads\n" + strings.Repeat("1\n", inferRows) + "n/a\n"
	p, err := newFixture(t).service.Preview(t.Context(), priya, upload("a.csv", csv))
	if err != nil || p.Columns[0].ProposedKind != KindNumber {
		t.Errorf("Preview() = %+v, %v; want Number from the first 200 rows", p.Columns, err)
	}
	p, err = newFixture(t).service.Preview(t.Context(), priya, upload("a.csv", "Cost\n12.50\n"))
	if err != nil || p.Columns[0].ProposedCurrency == nil || *p.Columns[0].ProposedCurrency != "USD" {
		t.Errorf("Preview() = %+v, %v; want Money in USD", p.Columns, err)
	}
}
