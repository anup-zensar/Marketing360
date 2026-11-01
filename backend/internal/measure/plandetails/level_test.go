package plandetails

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"mapp/internal/allwork"
	"mapp/internal/measure"
	"mapp/internal/platform/actor"
	"mapp/internal/platform/money"
)

func missing(n int) func(*allwork.Node) {
	return func(node *allwork.Node) { node.Missing = n }
}

func cardsYear(t *testing.T) *year {
	t.Helper()
	return newYear(t).plan("P-1").campaign("C-1", "P-1").
		activity("A-1", "C-1", "100.00").activity("A-2", "C-1", "200.00").
		activity("A-3", "C-1", "", missing(1)).activity("A-4", "C-1", "50.00")
}

func TestCards_FR005(t *testing.T) {
	t.Parallel()
	cards := newView(cardsYear(t).matching(), LevelActivity, measure.PeriodMonth).cards()

	if cards.Records != 4 || cards.PlannedCost.Amount.String() != "350.00" || cards.PlannedCost.Currency != "USD" ||
		cards.AllFilledIn != 3 || cards.AllFilledInPercent == nil || *cards.AllFilledInPercent != 75 {
		t.Errorf("cards = %+v (%v), want 4 activities, 350.00 USD, 3 of 4 all filled in (75 %%)", cards,
			cards.AllFilledInPercent)
	}
}

func TestLevel_OnlyMatchedRecords_FR003(t *testing.T) {
	t.Parallel()
	v := newView(cardsYear(t).unmatch("A-2").matching(), LevelActivity, measure.PeriodMonth)

	cards := v.cards()
	count := v.rows([]allwork.Attribute{companyCode()})[0].Widgets[0]

	if cards.Records != 3 || cards.PlannedCost.Amount.String() != "150.00" || count.Total.Count != 3 {
		t.Errorf("cards %+v, widget total %d: want the three matching activities only", cards, count.Total.Count)
	}
}

func TestEmpty_FR016(t *testing.T) {
	t.Parallel()
	scoped := &fakeScoped{
		matching:   cardsYear(t).unmatch("P-1", "C-1", "A-1", "A-2", "A-3", "A-4").matching(),
		attributes: []allwork.Attribute{companyCode()},
	}
	s := NewService(fakeTx{}, scoped)

	r, err := s.Report(t.Context(), actor.Actor{ID: "U-1002"}, measure.Filter{})

	if err != nil {
		t.Fatalf("Report() error = %v", err)
	}
	type outline struct {
		Show            Level
		Cards           Cards
		BuiltIn, Rows   int
		AttributesSetUp bool
	}
	want := outline{Show: LevelActivity, Cards: Cards{PlannedCost: inCurrency(money.Amount{}, usdCurrency())}, Rows: 1,
		AttributesSetUp: true}
	got := outline{r.Show, r.Cards, len(r.BuiltIn), len(r.Attributes), r.AttributesSetUp}
	if diff := cmp.Diff(want, got, cmp.Comparer(func(x, y money.Amount) bool { return x.Cmp(y) == 0 })); diff != "" ||
		r.Attributes[0].Widgets[0].Total.Count != 0 {
		t.Errorf("the Activity level with nothing counted and no state of its own (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff([]string{"completeness", "inq"}, scoped.columns); diff != "" {
		t.Errorf("columns asked (-want +got):\n%s", diff)
	}
}

func TestReport_ShowAndAttributesSetUp_FR001_FR016(t *testing.T) {
	t.Parallel()
	scoped := &fakeScoped{matching: cardsYear(t).matching(), attributes: []allwork.Attribute{companyCode()}}
	s := NewService(fakeTx{}, scoped)

	r, err := s.Report(t.Context(), actor.Actor{ID: "U-1002"}, measure.Filter{Show: measure.ShowCampaign})

	if err != nil {
		t.Fatalf("Report() error = %v", err)
	}
	if r.Show != LevelCampaign || r.Cards.Records != 1 || r.AttributesSetUp || len(r.Attributes) != 0 ||
		r.Currency != "USD" {
		t.Errorf("report = %+v, want one campaign and no campaign attribute set up", r)
	}
}
