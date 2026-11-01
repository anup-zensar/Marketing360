package dataset

import (
	"encoding/base64"
	"fmt"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"mapp/internal/history"
	"mapp/internal/platform/errs"
)

func TestCreate_StoresTypedRowsAndLogs_FR042_FR045(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	view := f.create(t, priya, "  Event leads 2026 ")

	eur := "EUR"
	same(t, "columns", []Column{
		{Key: "c1", Position: 0, Label: "Event", Kind: KindText}, {Key: "c2", Position: 1, Label: "Date", Kind: KindDate},
		{Key: "c3", Position: 2, Label: "Leads", Kind: KindNumber},
		{Key: "c4", Position: 3, Label: "Agency cost", Kind: KindMoney, Currency: &eur},
	}, view.Columns)
	same(t, "view", []any{"DS-1001", "Event leads 2026", 2, true, 1}, []any{view.ID, view.Name, view.RowCount, view.IsOwner, view.Version})
	rows, _ := f.store.Rows(t.Context(), view.ID)
	same(t, "rows", []Row{
		{TextCell("Expo Hamburg"), TextCell("2026-01-22"), TextCell("140"), MinorCell(1850000)},
		{TextCell("Fleet Day"), TextCell("2026-02-12"), TextCell("85"), MinorCell(920000)},
	}, rows)
	entries, _, _ := f.history.ListLog(t.Context(), priya, history.LogFilter{RecordID: view.ID, Limit: 10})
	same(t, "log", []any{1, history.ActionCreated, "Uploaded from 'leads.csv', 2 rows"},
		[]any{len(entries), entries[0].Action, entries[0].Details})
}

func TestCreate_EmptyCellsAreEmptyValues_FR042(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	in := leadsInput("Leads")
	in.Content = base64.StdEncoding.EncodeToString([]byte("Event,Date,Leads,Cost\nExpo,,,\n"))
	view, err := f.service.Create(t.Context(), priya, in)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	rows, _ := f.store.Rows(t.Context(), view.ID)
	empty := Cell{Empty: true}
	if diff := cmp.Diff([]Row{{TextCell("Expo"), empty, empty, empty}}, rows); diff != "" {
		t.Errorf("rows (-want +got):\n%s", diff)
	}
}

func TestCreate_CellProblemsAllAtOnce_FR042(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	in := leadsInput("Leads")
	csv := "Event,Date,Leads,Cost\n" + strings.Repeat("Expo,someday,n/a,12\n", 60)
	in.Content = base64.StdEncoding.EncodeToString([]byte(csv))

	_, err := f.service.Create(t.Context(), priya, in)

	fields := fieldErrors(t, err)
	want := []errs.FieldError{}
	for row := 2; row < 2+maxCellProblems; row++ {
		want = append(want, errs.FieldError{Field: "columns[1]", Message: fmt.Sprintf(msgNotADate, row, "someday")})
	}
	want = append(want, errs.FieldError{Field: "columns[1]", Message: fmt.Sprintf(msgMoreProblems, 10)},
		errs.FieldError{Field: "columns[2]", Message: fmt.Sprintf(msgMoreProblems, 60)})
	if diff := cmp.Diff(want, fields); diff != "" {
		t.Errorf("problems (-want +got):\n%s", diff)
	}
	if count, _ := f.store.CountOfOwner(t.Context(), priya.ID); count != 0 {
		t.Errorf("stored %d datasets, want none", count)
	}
}

func TestCreate_ColumnsNameAndDescription_FR042_FR043(t *testing.T) {
	t.Parallel()
	gbp, xyz := "GBP", "XYZ"
	in := leadsInput(strings.Repeat("n", maxNameRunes+1))
	in.Description = strings.Repeat("d", maxDescriptionRunes+1)
	in.Columns[0].Label = " "
	in.Columns[2].Kind = "percent"
	in.Columns[3].Currency = &gbp
	missing := leadsInput("Leads")
	missing.Columns[3].Currency = nil
	unknown := leadsInput("Leads")
	unknown.Columns[3].Currency = &xyz
	unknown.Columns[0].Label = strings.Repeat("l", maxLabelRunes+1)
	short := leadsInput("")
	short.Columns = short.Columns[:3]
	tests := []struct {
		name string
		in   CreateInput
		want []errs.FieldError
	}{
		{name: "every problem", in: in, want: []errs.FieldError{
			{Field: nameField, Message: msgNameTooLong}, {Field: descriptionField, Message: msgDescription},
			{Field: "columns[0].label", Message: msgLabelRequired}, {Field: "columns[2].kind", Message: msgKind},
			{Field: "columns[3].currency", Message: msgCurrency},
		}},
		{name: "money without a currency", in: missing, want: []errs.FieldError{{Field: "columns[3].currency", Message: msgCurrency}}},
		{name: "unknown currency, long label", in: unknown, want: []errs.FieldError{
			{Field: "columns[0].label", Message: msgLabelTooLong}, {Field: "columns[3].currency", Message: msgCurrency},
		}},
		{name: "a column missing, no name", in: short, want: []errs.FieldError{
			{Field: nameField, Message: msgNameRequired}, {Field: columnsField, Message: fmt.Sprintf(msgColumnsMismatch, 4)},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := newFixture(t).service.Create(t.Context(), priya, tt.in)
			if diff := cmp.Diff(tt.want, fieldErrors(t, err)); diff != "" {
				t.Errorf("problems (-want +got):\n%s", diff)
			}
		})
	}
}

func TestCreate_LabelsUniqueIgnoringCase_FR043(t *testing.T) {
	t.Parallel()
	in := leadsInput("Leads")
	in.Columns[2].Label = "EVENT"
	_, err := newFixture(t).service.Create(t.Context(), priya, in)
	want := []errs.FieldError{{Field: "columns[2].label", Message: fmt.Sprintf(msgLabelTaken, "EVENT")}}
	if diff := cmp.Diff(want, fieldErrors(t, err)); diff != "" {
		t.Errorf("problems (-want +got):\n%s", diff)
	}
}

func TestCreate_LimitsPerPerson_FR043(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	f.create(t, priya, "Event leads")

	_, taken := f.service.Create(t.Context(), priya, leadsInput("EVENT LEADS"))
	if _, err := f.service.Create(t.Context(), sara, leadsInput("Event leads")); err != nil {
		t.Errorf("another person's name: %v, want it free", err)
	}
	for i := 1; i < maxDatasets; i++ {
		f.create(t, priya, fmt.Sprintf("Leads %d", i))
	}
	_, full := f.service.Create(t.Context(), priya, leadsInput("One more"))

	if codeOf(taken) != errs.CodeNameTaken {
		t.Errorf("same name ignoring case = %v, want name_taken", taken)
	}
	want := []errs.FieldError{{Field: nameField, Message: msgTooManyDatasets}}
	if diff := cmp.Diff(want, fieldErrors(t, full)); diff != "" {
		t.Errorf("51st dataset (-want +got):\n%s", diff)
	}
}
