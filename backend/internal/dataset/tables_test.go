package dataset

import (
	"testing"

	"mapp/internal/widgetdata"
)

func TestTables_VisibleOnlyAndTyped_FR044(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	leads := f.create(t, priya, "Leads")
	f.create(t, mei, "Mei's private data")

	before, _ := f.service.Visible(t.Context(), sara)
	_, hidden := f.service.Table(t.Context(), sara, leads.ID)
	if _, err := f.service.Share(t.Context(), priya, leads.ID, SharingInput{Version: ptr(1), TeamIDs: []string{europe}}); err != nil {
		t.Fatalf("Share: %v", err)
	}
	after, _ := f.service.Visible(t.Context(), sara)
	table, err := f.service.Table(t.Context(), sara, leads.ID)
	if err != nil {
		t.Fatalf("Table() error = %v", err)
	}

	same(t, "before the share", []any{0, 404}, []any{len(before), statusOf(hidden)})
	same(t, "after the share", []any{1, leads.ID, "Priya Raman"}, []any{len(after), after[0].ID, after[0].Owner.Name})
	same(t, "columns", []widgetdata.DatasetColumn{
		{Key: "c1", Label: "Event", Kind: widgetdata.KindText}, {Key: "c2", Label: "Date", Kind: widgetdata.KindDate},
		{Key: "c3", Label: "Leads", Kind: widgetdata.KindNumber},
		{Key: "c4", Label: "Agency cost", Kind: widgetdata.KindMoney, Currency: "EUR"},
	}, table.Info.Columns)
	first := table.Rows[0]
	same(t, "first row", []string{"Expo Hamburg", "2026-01-22", "140", "18,500.00 EUR"},
		[]string{first[0].Text, first[1].Date.String(), first[2].Number.String(), first[3].Money.Format()})
}

func TestTables_EmptyCellsAndDeletedDataset_FR044(t *testing.T) {
	t.Parallel()
	got := valuesOf([]Column{{Kind: KindNumber}, {Kind: KindMoney}, {Kind: KindText}}, Row{{Empty: true}, TextCell("x")})
	if !got[0].Empty || !got[1].Empty || !got[2].Empty {
		t.Errorf("valuesOf() = %+v, want three empty values", got)
	}
	f := newFixture(t)
	leads := f.create(t, priya, "Leads")
	if _, err := f.service.Delete(t.Context(), priya, leads.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := f.service.Table(t.Context(), priya, leads.ID); statusOf(err) != 404 {
		t.Errorf("Table() of a deleted dataset = %v, want 404", err)
	}
}

func TestRows_PagesAsStored_FR040(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	leads := f.create(t, priya, "Leads")

	page, err := f.service.Rows(t.Context(), priya, leads.ID, RowsQuery{Offset: ptr(1), Limit: ptr(5)})
	_, badLimit := f.service.Rows(t.Context(), priya, leads.ID, RowsQuery{Offset: ptr(-1), Limit: ptr(201)})
	_, hidden := f.service.Rows(t.Context(), mei, leads.ID, RowsQuery{})

	if err != nil {
		t.Fatalf("Rows() error = %v", err)
	}
	texts := []string{}
	for _, cell := range page.Rows[0] {
		texts = append(texts, *cell)
	}
	same(t, "row", []string{"Fleet Day", "2026-02-12", "85", "9200.00"}, texts)
	same(t, "page", []int{2, 1, 5, 4}, []int{page.Total, page.Offset, page.Limit, len(page.Columns)})
	same(t, "refusals", []int{2, 404}, []int{len(fieldErrors(t, badLimit)), statusOf(hidden)})
}

func TestList_MineAndSharedWithMe_FR040(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	mine := f.create(t, sara, "Sara's")
	theirs := f.create(t, priya, "Priya's")
	f.create(t, mei, "Mei's")
	if _, err := f.service.Share(t.Context(), priya, theirs.ID, SharingInput{Version: ptr(1), TeamIDs: []string{europe}}); err != nil {
		t.Fatalf("Share: %v", err)
	}
	list, err := f.service.List(t.Context(), sara)
	if err != nil || len(list.Mine) != 1 || list.Mine[0].ID != mine.ID || len(list.SharedWithMe) != 1 ||
		list.SharedWithMe[0].ID != theirs.ID || list.SharedWithMe[0].Access.Delete.Available || list.Mine[0].ColumnCount != 4 {
		t.Errorf("List(Sara) = %+v, %v", list, err)
	}
}
