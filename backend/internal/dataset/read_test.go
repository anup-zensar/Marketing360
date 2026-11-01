package dataset

import (
	"archive/zip"
	"bytes"
	"encoding/base64"
	"fmt"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"mapp/internal/platform/date"
	"mapp/internal/platform/sheet"
)

func xlsxUpload(t *testing.T, rows [][]sheet.Cell) Upload {
	t.Helper()
	var b bytes.Buffer
	if err := sheet.WriteXLSX(&b, sheet.Workbook{Sheets: []sheet.Sheet{{Name: "Leads", Rows: rows}, {Name: "Notes"}}}); err != nil {
		t.Fatalf("WriteXLSX: %v", err)
	}
	return Upload{FileName: "leads.xlsx", Content: base64.StdEncoding.EncodeToString(b.Bytes())}
}

// formulaUpload is a workbook whose second column holds formulas with their saved results.
func formulaUpload(t *testing.T) Upload {
	t.Helper()
	parts := map[string]string{
		"xl/workbook.xml": `<workbook xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships"><sheets>` +
			`<sheet name="Data" r:id="rId1"/></sheets></workbook>`,
		"xl/_rels/workbook.xml.rels": `<Relationships><Relationship Id="rId1" Target="worksheets/sheet1.xml"/></Relationships>`,
		"xl/worksheets/sheet1.xml": `<worksheet><sheetData><row r="1"><c r="A1" t="inlineStr"><is><t>Leads</t></is></c>` +
			`<c r="B1" t="inlineStr"><is><t>Double</t></is></c></row><row r="2"><c r="A2"><v>21</v></c>` +
			`<c r="B2"><f>A2*2</f><v>42</v></c></row></sheetData></worksheet>`,
	}
	var b bytes.Buffer
	w := zip.NewWriter(&b)
	for name, content := range parts {
		part, err := w.Create(name)
		if err == nil {
			_, err = part.Write([]byte(content))
		}
		if err != nil {
			t.Fatalf("zip: %v", err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatalf("zip: %v", err)
	}
	return Upload{FileName: "formulas.xlsx", Content: base64.StdEncoding.EncodeToString(b.Bytes())}
}

func TestPreview_FormatsAndLimits_FR041_FR043(t *testing.T) {
	t.Parallel()
	wideHeader := strings.Repeat("Name,", maxColumns) + "One too many\n"
	manyRows := "Leads\n" + strings.Repeat("1\n", maxRows+1)
	tests := []struct {
		name     string
		up       Upload
		wantRows [][]string
		wantErr  string
	}{
		{name: "comma", up: upload("a.csv", "Event,Leads\nExpo,140\n"), wantRows: [][]string{{"Expo", "140"}}},
		{name: "semicolon", up: upload("a.csv", "Event;Leads\nExpo;140\n"), wantRows: [][]string{{"Expo", "140"}}},
		{name: "tab", up: upload("a.CSV", "Event\tLeads\nExpo\t140\n"), wantRows: [][]string{{"Expo", "140"}}},
		{name: "byte-order mark", up: upload("a.csv", "\xEF\xBB\xBFEvent,Leads\nExpo,140\n"), wantRows: [][]string{{"Expo", "140"}}},
		{name: "first sheet with a date cell", up: xlsxUpload(t, [][]sheet.Cell{{sheet.Text("Event"), sheet.Text("Date")},
			{sheet.Text("Expo"), sheet.Date(date.MustParse("2026-01-22"))}}), wantRows: [][]string{{"Expo", "2026-01-22"}}},
		{name: "formulas by their saved value", up: formulaUpload(t), wantRows: [][]string{{"21", "42"}}},
		{name: "header only", up: upload("a.csv", "Event,Leads\n"), wantRows: [][]string{}},
		{name: "not UTF-8", up: upload("a.csv", "Event\nCaf\xe9\n"), wantErr: msgFormat},
		{name: "neither CSV nor .xlsx", up: upload("a.txt", "Event\nExpo\n"), wantErr: msgFormat},
		{name: "not base64", up: Upload{FileName: "a.csv", Content: "%%%"}, wantErr: msgUnreadable},
		{name: "empty", up: upload("a.csv", ""), wantErr: msgEmptyFile},
		{name: "only blank lines", up: upload("a.csv", "\n,\n"), wantErr: msgEmptyFile},
		{name: "larger than 5 MB", up: upload("a.csv", "A\n"+strings.Repeat("x", maxFileBytes)), wantErr: msgTooLarge},
		{name: "more than 10,000 rows", up: upload("a.csv", manyRows), wantErr: msgTooManyRows},
		{name: "more than 50 columns", up: upload("a.csv", wideHeader), wantErr: msgTooManyColumns},
		{name: "no header row", up: upload("a.csv", "140,2026-01-22\n85,2026-02-12\n"), wantErr: msgNoHeader},
		{name: "no file name", up: upload(" ", "A\n1\n"), wantErr: msgFileName},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			p, err := newFixture(t).service.Preview(t.Context(), priya, tt.up)
			if tt.wantErr != "" {
				fields := fieldErrors(t, err)
				if len(fields) != 1 || fields[0].Message != tt.wantErr {
					t.Fatalf("Preview() problems = %v, want %q", fields, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("Preview() error = %v", err)
			}
			if diff := cmp.Diff(tt.wantRows, p.Rows); diff != "" {
				t.Errorf("Preview() rows (-want +got):\n%s", diff)
			}
		})
	}
}

func TestPreview_ColumnsSamplesAndProblems_FR041(t *testing.T) {
	t.Parallel()
	long := strings.Repeat("x", maxCellRunes+1)
	csv := "Event,Date,,Cost,event,Empty\n" +
		"Expo,2026-01-22,a,\"1,200.50\",x,\n" +
		"Expo,22/01/2026,," + long + ",y,\n"
	p, err := newFixture(t).service.Preview(t.Context(), priya, upload("leads.csv", csv))
	if err != nil {
		t.Fatalf("Preview() error = %v", err)
	}
	want := []PreviewColumn{
		{Position: 0, Label: "Event", ProposedKind: KindText, Samples: []string{"Expo"}},
		{Position: 1, Label: "Date", ProposedKind: KindDate, Samples: []string{"2026-01-22", "22/01/2026"}},
		{Position: 2, Label: "Column 3", ProposedKind: KindText, Samples: []string{"a"}, EmptyCount: 1},
		{Position: 3, Label: "Cost", ProposedKind: KindText, Samples: []string{"1,200.50", long}},
		{Position: 4, Label: "event", ProposedKind: KindText, Samples: []string{"x", "y"}},
		{Position: 5, Label: "Empty", ProposedKind: KindText, Samples: []string{}, EmptyCount: 2},
	}
	if diff := cmp.Diff(want, p.Columns); diff != "" {
		t.Errorf("columns (-want +got):\n%s", diff)
	}
	wantProblems := []string{
		fmt.Sprintf(msgNoName, 3, "Column 3"), fmt.Sprintf(msgDuplicateName, "event"),
		fmt.Sprintf(msgEmptyColumn, "Empty"), fmt.Sprintf(msgLongCell, 3, "Cost"),
	}
	if diff := cmp.Diff(wantProblems, p.Problems); diff != "" {
		t.Errorf("problems (-want +got):\n%s", diff)
	}
	if p.FileName != "leads.csv" || p.SheetName != "" || p.RowCount != 2 {
		t.Errorf("preview = %q %q %d, want leads.csv, no sheet, 2 rows", p.FileName, p.SheetName, p.RowCount)
	}
}

func TestPreview_SaysWhenNoRoomIsLeft_FR043(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	for i := range maxDatasets {
		f.create(t, priya, fmt.Sprintf("Leads %d", i))
	}
	p, err := f.service.Preview(t.Context(), priya, upload("a.csv", "Event\nExpo\n"))
	if err != nil || len(p.Problems) != 1 || p.Problems[0] != msgTooManyDatasets {
		t.Errorf("Preview() = %v, %v; want the 50-datasets problem", p.Problems, err)
	}
}
