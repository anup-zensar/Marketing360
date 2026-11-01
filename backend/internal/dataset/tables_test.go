package dataset

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"mapp/internal/platform/date"
	"mapp/internal/platform/money"
	"mapp/internal/widgetdata"
)

func TestTables_VisibleOnlyAndTyped_FR044(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	leads := f.create(t, priya, "Leads")
	f.create(t, mei, "Mei's private data")

	visibleToSara, err := f.service.Visible(t.Context(), sara)
	if err != nil || len(visibleToSara) != 0 {
		t.Fatalf("Visible(Sara) = %v, %v; want none before the share", visibleToSara, err)
	}
	_, hidden := f.service.Table(t.Context(), sara, leads.ID)
	if _, err := f.service.Share(t.Context(), priya, leads.ID, SharingInput{Version: ptr(1), TeamIDs: []string{europe}}); err != nil {
		t.Fatalf("Share: %v", err)
	}
	visibleToSara, _ = f.service.Visible(t.Context(), sara)
	table, err := f.service.Table(t.Context(), sara, leads.ID)

	if statusOf(hidden) != 404 || err != nil {
		t.Fatalf("Table() before %v, after %v; want 404 then the table", hidden, err)
	}
	if len(visibleToSara) != 1 || visibleToSara[0].ID != leads.ID || visibleToSara[0].Owner.Name != "Priya Raman" {
		t.Errorf("Visible(Sara) = %+v", visibleToSara)
	}
	wantColumns := []widgetdata.DatasetColumn{
		{Key: "c1", Label: "Event", Kind: widgetdata.KindText}, {Key: "c2", Label: "Date", Kind: widgetdata.KindDate},
		{Key: "c3", Label: "Leads", Kind: widgetdata.KindNumber},
		{Key: "c4", Label: "Agency cost", Kind: widgetdata.KindMoney, Currency: "EUR"},
	}
	if diff := cmp.Diff(wantColumns, table.Info.Columns); diff != "" {
		t.Errorf("columns (-want +got):\n%s", diff)
	}
	first := table.Rows[0]
	wantMoney := money.Money{Amount: money.FromCents(1850000), Currency: "EUR"}
	if first[0].Text != "Expo Hamburg" || first[1].Date != date.MustParse("2026-01-22") ||
		first[2].Number != widgetdata.NumberOfUnits(1400000) || first[3].Money != wantMoney {
		t.Errorf("first row = %+v", first)
	}
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
	if diff := cmp.Diff([]string{"Fleet Day", "2026-02-12", "85", "9200.00"}, texts); diff != "" {
		t.Errorf("row (-want +got):\n%s", diff)
	}
	if page.Total != 2 || page.Offset != 1 || page.Limit != 5 || len(page.Columns) != 4 {
		t.Errorf("page = %+v", page)
	}
	if len(fieldErrors(t, badLimit)) != 2 || statusOf(hidden) != 404 {
		t.Errorf("bad page %v, hidden %v", badLimit, hidden)
	}
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
