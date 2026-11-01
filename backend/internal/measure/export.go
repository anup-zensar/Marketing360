package measure

import (
	"bytes"
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"mapp/internal/platform/actor"
	"mapp/internal/platform/date"
	"mapp/internal/platform/errs"
	"mapp/internal/platform/sheet"
)

const (
	FormatXLSX      = "xlsx"
	FormatCSV       = "csv"
	contentTypeXLSX = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	contentTypeCSV  = "text/csv"
	msgChooseWidget = "Choose a widget of this report."
	fieldWidget     = "widget"
	sheetAbout      = "About this export"
	sheetInView     = "In view"
)

// ExportRequest is what an export asks for: the filter, the format and, for one widget, its key (FR-022).
type ExportRequest struct {
	Filter Filter
	Format string
	Widget string
}

type File struct {
	Name        string
	ContentType string
	Body        []byte
}

// Export writes the report, or one widget of it, as the screen shows it: Excel holds the widgets' sheets, In view
// and About this export; a CSV is one widget with the notes before its rows (FR-022, research R7). Nothing is kept.
func (s *Service) Export(ctx context.Context, a actor.Actor, report string, req ExportRequest) (File, error) {
	if err := openable(ctx, report); err != nil {
		return File{}, err
	}
	name, _ := nameOf(report)
	req, err := checked(req)
	if err != nil {
		return File{}, err
	}
	var sc Scope
	var rates []string
	err = s.tx.Read(ctx, func(ctx context.Context) error {
		var readErr error
		if sc, readErr = s.Scope(ctx, a, req.Filter, nil); readErr != nil {
			return readErr
		}
		rates, readErr = s.deps.Planning.RatesUsed(ctx)
		return readErr
	})
	if err != nil {
		return File{}, err
	}
	sheets, err := s.sheetsOf(ctx, a, report, req)
	if err != nil {
		return File{}, err
	}
	own, err := s.ownNotes(ctx, a, report, req.Filter.Own)
	if err != nil {
		return File{}, err
	}
	notes := append(afterFilters(s.notes(a, name, req.Filter, sc), own), rateNotes(rates)...)
	stamp := strconv.Itoa(sc.Year) + " – " + date.Format(date.Today(s.now()))
	file := File{Name: fileName(name, sheets, req, stamp)}
	return write(file, sheets, notes, inViewRows(inViewOf(sc.Matched)))
}

// checked is the request with Excel by default; a CSV needs a widget (FR-022).
func checked(req ExportRequest) (ExportRequest, error) {
	if req.Format == "" {
		req.Format = FormatXLSX
	}
	if (req.Format != FormatXLSX && req.Format != FormatCSV) || (req.Format == FormatCSV && req.Widget == "") {
		return ExportRequest{}, errs.Validation(errs.FieldError{Field: fieldWidget, Message: msgChooseWidget})
	}
	return req, nil
}

// sheetsOf are the report's widgets as sheets, or the one asked for; a report not built yet has none.
func (s *Service) sheetsOf(ctx context.Context, a actor.Actor, report string, req ExportRequest) ([]WidgetSheet, error) {
	r, found := s.registry.Get(report)
	if !found {
		return []WidgetSheet{}, nil
	}
	sheets, err := r.Sheets(ctx, a, req.Filter)
	if err != nil || req.Widget == "" {
		return sheets, err
	}
	for _, ws := range sheets {
		if ws.Key == req.Widget {
			return []WidgetSheet{ws}, nil
		}
	}
	return nil, errs.Validation(errs.FieldError{Field: fieldWidget, Message: msgChooseWidget})
}

func (s *Service) ownNotes(ctx context.Context, a actor.Actor, report string, own url.Values) ([][2]string, error) {
	r, _ := s.registry.Get(report)
	if described, ok := r.(Described); ok {
		return described.Describe(ctx, a, own)
	}
	return nil, nil
}

func write(file File, sheets []WidgetSheet, notes [][2]string, inView [][]sheet.Cell) (File, error) {
	var body bytes.Buffer
	if strings.HasSuffix(file.Name, "."+FormatCSV) {
		file.ContentType = contentTypeCSV
		err := sheet.WriteCSV(&body, notes, sheets[0].Sheet.Rows)
		file.Body = body.Bytes()
		return file, err
	}
	wb := sheet.Workbook{}
	for _, ws := range sheets {
		wb.Sheets = append(wb.Sheets, ws.Sheet)
	}
	about := make([][]sheet.Cell, 0, len(notes))
	for _, n := range notes {
		about = append(about, []sheet.Cell{sheet.Bold(n[0]), sheet.Text(n[1])})
	}
	wb.Sheets = append(wb.Sheets, sheet.Sheet{Name: sheetInView, Rows: inView},
		sheet.Sheet{Name: sheetAbout, Rows: about})
	err := sheet.WriteXLSX(&body, wb)
	file.ContentType, file.Body = contentTypeXLSX, body.Bytes()
	return file, err
}

// fileName is "SP&A One – <report> [– <widget>] – <year> – <date>" (FR-022); stamp is "<year> – <date>".
func fileName(report string, sheets []WidgetSheet, req ExportRequest, stamp string) string {
	parts := []string{"SP&A One", report}
	if req.Widget != "" && len(sheets) == 1 {
		parts = append(parts, sheets[0].Sheet.Name)
	}
	return fmt.Sprintf("%s.%s", strings.Join(append(parts, stamp), " – "), req.Format)
}

func inViewRows(v InView) [][]sheet.Cell {
	row := func(label string, n int) []sheet.Cell { return []sheet.Cell{sheet.Text(label), sheet.Number(int64(n))} }
	return [][]sheet.Cell{{sheet.Bold("Records"), sheet.Bold("In view")}, row("Plans", v.Plans),
		row("Campaigns", v.Campaigns), row("Initiatives", v.Initiatives), row("Activities", v.Activities),
		row("Line items", v.LineItems)}
}

func nameOf(report string) (string, bool) {
	for _, e := range entries() {
		if e.key == report {
			return e.name, true
		}
	}
	return "", false
}
