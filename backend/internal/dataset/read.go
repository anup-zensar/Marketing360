package dataset

import (
	"bytes"
	"encoding/base64"
	"errors"
	"strings"
	"unicode/utf8"

	"mapp/internal/platform/errs"
	"mapp/internal/platform/sheet"
)

const (
	MaxUploadBodyBytes = 7 << 20
	maxFileBytes       = 5 << 20
	maxRows            = 10000
	maxColumns         = 50
	maxFileName        = 255
	contentField       = "content"
	fileNameField      = "fileName"
	zipSignature       = "PK\x03\x04"
	csvSuffix          = ".csv"
)

// file is an upload read as text: its header (first row) and data rows, at most width columns wide.
type file struct {
	name   string
	sheet  string
	header sheet.Row
	rows   []sheet.Row
	width  int
}

func (f file) column(i, limit int) []string {
	cells := make([]string, 0, min(limit, len(f.rows)))
	for _, r := range f.rows[:min(limit, len(f.rows))] {
		cells = append(cells, r.Cell(i))
	}
	return cells
}

// readFile reads the first sheet of an .xlsx file or a CSV file; any problem refuses the whole file.
func readFile(up Upload) (file, error) {
	name := strings.TrimSpace(up.FileName)
	if name == "" || utf8.RuneCountInString(name) > maxFileName {
		return file{}, problem(fileNameField, msgFileName)
	}
	data, err := base64.StdEncoding.DecodeString(up.Content)
	switch {
	case err != nil:
		return file{}, problem(contentField, msgUnreadable)
	case len(data) == 0:
		return file{}, problem(contentField, msgEmptyFile)
	case len(data) > maxFileBytes:
		return file{}, problem(contentField, msgTooLarge)
	}
	table, err := firstTable(name, data)
	if err != nil {
		return file{}, err
	}
	return fileOf(name, table)
}

func firstTable(name string, data []byte) (sheet.Table, error) {
	if bytes.HasPrefix(data, []byte(zipSignature)) {
		tables, err := sheet.ReadXLSX(data, maxRows+1)
		if err == nil && len(tables) == 0 {
			err = sheet.ErrUnreadable
		}
		if err != nil {
			return sheet.Table{}, readProblem(err)
		}
		return tables[0], nil
	}
	if !strings.HasSuffix(strings.ToLower(name), csvSuffix) {
		return sheet.Table{}, problem(contentField, msgFormat)
	}
	table, err := sheet.ReadCSV(data, maxRows+1)
	return table, readProblem(err)
}

func fileOf(name string, t sheet.Table) (file, error) {
	if len(t.Rows) == 0 {
		return file{}, problem(contentField, msgEmptyFile)
	}
	f := file{name: name, sheet: t.Name, header: t.Rows[0], rows: t.Rows[1:]}
	for _, r := range t.Rows {
		f.width = max(f.width, filledWidth(r.Cells))
	}
	switch {
	case f.width > maxColumns:
		return file{}, problem(contentField, msgTooManyColumns)
	case !hasHeader(f.header):
		return file{}, problem(contentField, msgNoHeader)
	}
	return f, nil
}

func filledWidth(cells []string) int {
	for i := len(cells); i > 0; i-- {
		if cells[i-1] != "" {
			return i
		}
	}
	return 0
}

// hasHeader is false when every name in the first row reads as a number or a date: the file starts with data.
func hasHeader(header sheet.Row) bool {
	for _, c := range header.Cells {
		_, isNumber := parseNumber(c)
		_, isDate := parseDate(c)
		if c != "" && !isNumber && !isDate {
			return true
		}
	}
	return false
}

func readProblem(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, sheet.ErrTooManyRows):
		return problem(contentField, msgTooManyRows)
	}
	return problem(contentField, msgFormat)
}

func problem(field, message string) error {
	return errs.Validation(errs.FieldError{Field: field, Message: message})
}
