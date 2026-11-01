package performance_test

import (
	"testing"

	"mapp/internal/performance"
)

func eloquaMappings() []performance.Mapping {
	return []performance.Mapping{
		{ColumnName: "Campaign ID", Meaning: performance.MeaningKey, KeyType: performance.MatchedActivityID},
		{ColumnName: "Media code", Meaning: performance.MeaningKey, KeyType: performance.MatchedMediaCode},
		{ColumnName: "Row ref", Meaning: performance.MeaningReference},
		{ColumnName: "Date", Meaning: performance.MeaningFactDate},
		{ColumnName: "INQs", Meaning: performance.MeaningStage, Stage: "inq"},
		{ColumnName: "Revenue", Meaning: performance.MeaningRevenue},
		{ColumnName: "Currency", Meaning: performance.MeaningCurrency},
	}
}

func TestReadFile_PlacesEveryCellWhereItsMappingSays_046C23(t *testing.T) {
	t.Parallel()

	content := []byte("Row ref,Date,Campaign ID,Media code,INQs,Revenue,Currency,Comment\n" +
		"r1,2026-02-01,A-1,9000001,5,1234.50,USD,hello\n")

	file, err := performance.ReadFile(content, eloquaMappings())
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if len(file.Rows) != 1 {
		t.Fatalf("ReadFile gave %d rows, want 1", len(file.Rows))
	}
	row := file.Rows[0]
	if row.Reference.Value != "r1" || row.Date.Value != "2026-02-01" || row.Currency.Value != "USD" {
		t.Errorf("reference/date/currency = %q %q %q, want r1 2026-02-01 USD",
			row.Reference.Value, row.Date.Value, row.Currency.Value)
	}
	if row.Keys.ActivityID != "A-1" || row.Keys.MediaCode != "9000001" {
		t.Errorf("keys = %+v, want the activity id and media code from their own typed columns", row.Keys)
	}
	assertReadings(t, row)
}

func assertReadings(t *testing.T, row performance.Row) {
	t.Helper()
	if len(row.Readings) != 2 {
		t.Fatalf("readings = %d, want the INQ count and the revenue amount", len(row.Readings))
	}
	if row.Readings[0].Stage != "inq" || row.Readings[0].Value != "5" {
		t.Errorf("first reading = %+v, want inq counting 5", row.Readings[0])
	}
	if row.Readings[1].Stage != "revenue" || row.Readings[1].Value != "1234.50" {
		t.Errorf("second reading = %+v, want revenue carrying 1234.50", row.Readings[1])
	}
	if row.Raw["Comment"] != "hello" {
		t.Errorf("an unmapped column is kept in the raw copy, got %q", row.Raw["Comment"])
	}
}

func TestReadFile_HeadingsSurviveForThePersonalDataGate_FR012(t *testing.T) {
	t.Parallel()

	content := []byte("Row ref,Email\nr1,someone@shell.com\n")

	file, err := performance.ReadFile(content, eloquaMappings())
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if found := performance.PersonalDataIn(file.Headings); len(found) != 1 {
		t.Errorf("PersonalDataIn(headings) = %v, want the Email column named", found)
	}
}

func TestReadFile_AnUnreadableFileIsRefusedWhole_FR009(t *testing.T) {
	t.Parallel()

	if _, err := performance.ReadFile([]byte{0xff, 0xfe, 0x00}, eloquaMappings()); err == nil {
		t.Fatal("ReadFile read a non-UTF-8 file, want a refusal")
	}
}

func TestReadFile_SemicolonSeparatedFilesAreRead_FR009(t *testing.T) {
	t.Parallel()

	content := []byte("Row ref;Date;INQs\nr1;2026-02-01;5\n")

	file, err := performance.ReadFile(content, eloquaMappings())
	if err != nil || len(file.Rows) != 1 || file.Rows[0].Reference.Value != "r1" {
		t.Fatalf("ReadFile(semicolons) = %+v (%v), want the row read", file.Rows, err)
	}
}
