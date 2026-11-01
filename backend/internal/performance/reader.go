package performance

import (
	"bytes"
	"encoding/csv"
	"errors"
	"io"
	"strings"
	"unicode/utf8"
)

// File content is read in memory, never written to disk or logged, as feature 015's imports are.

var errFileUnreadable = errors.New("performance: the file can't be read as UTF-8 CSV")

// byteOrderUTF is the byte-order mark Excel writes in UTF-8 CSV files.
const byteOrderUTF = "\xEF\xBB\xBF"

type File struct {
	Headings []string
	Rows     []Row
}

// ReadFile turns one uploaded CSV into rows through the source's mappings (046-C23): every cell lands where
// its column's meaning says, and a column no mapping names is kept only in the row's raw copy.
func ReadFile(content []byte, mappings []Mapping) (File, error) {
	records, err := readRecords(content)
	if err != nil || len(records) == 0 {
		return File{}, errFileUnreadable
	}
	headings := records[0]
	byColumn := mappingsByColumn(mappings)
	file := File{Headings: headings}
	for _, record := range records[1:] {
		file.Rows = append(file.Rows, rowOf(headings, record, byColumn))
	}
	return file, nil
}

func readRecords(content []byte) ([][]string, error) {
	content = bytes.TrimPrefix(content, []byte(byteOrderUTF))
	if !utf8.Valid(content) {
		return nil, errFileUnreadable
	}
	reader := csv.NewReader(bytes.NewReader(content))
	firstLine, _, _ := bytes.Cut(content, []byte("\n"))
	if bytes.Count(firstLine, []byte(";")) > bytes.Count(firstLine, []byte(",")) {
		reader.Comma = ';'
	}
	reader.FieldsPerRecord = -1
	var records [][]string
	for {
		record, err := reader.Read()
		if errors.Is(err, io.EOF) {
			return records, nil
		}
		if err != nil {
			return nil, errFileUnreadable
		}
		records = append(records, record)
	}
}

func mappingsByColumn(mappings []Mapping) map[string]Mapping {
	byColumn := make(map[string]Mapping, len(mappings))
	for _, m := range mappings {
		byColumn[strings.ToLower(strings.TrimSpace(m.ColumnName))] = m
	}
	return byColumn
}

func rowOf(headings, record []string, byColumn map[string]Mapping) Row {
	row := Row{Raw: map[string]string{}}
	for i, heading := range headings {
		value := ""
		if i < len(record) {
			value = record[i]
		}
		row.Raw[heading] = value
		mapping, mapped := byColumn[strings.ToLower(strings.TrimSpace(heading))]
		if !mapped {
			continue
		}
		place(&row, mapping, Cell{Column: heading, Value: value})
	}
	return row
}

func place(row *Row, mapping Mapping, cell Cell) {
	switch mapping.Meaning {
	case MeaningKey:
		placeKey(row, mapping.KeyType, cell.Value)
	case MeaningReference:
		row.Reference = cell
	case MeaningFactDate:
		row.Date = cell
	case MeaningCurrency:
		row.Currency = cell
	case MeaningStage:
		row.Readings = append(row.Readings, Reading{Measure: Measure{Stage: mapping.Stage}, Cell: cell})
	case MeaningRevenue:
		row.Readings = append(row.Readings, Reading{Measure: Measure{Stage: stageRevenue}, Cell: cell})
	case MeaningWebMetric:
		row.Readings = append(row.Readings, Reading{Measure: Measure{WebMetric: mapping.WebMetric}, Cell: cell})
	case MeaningJourney:
	}
}

func placeKey(row *Row, on MatchedOn, value string) {
	value = strings.TrimSpace(value)
	switch on {
	case MatchedActivityID:
		row.Keys.ActivityID = value
	case MatchedMediaCode:
		row.Keys.MediaCode = value
	case MatchedIntegrationID:
		row.Keys.IntegrationID = value
	case MatchedExternalName:
		row.Keys.ExternalName = value
	case MatchedNone:
	}
}
