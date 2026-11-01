package purchasereport_test

import (
	"fmt"
	"net/url"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"mapp/internal/measure/purchasereport"
	"mapp/internal/measure/purchasereport/purchasereporttest"
)

func labels(vendors []purchasereport.VendorNode) []string {
	out := make([]string, 0, len(vendors))
	for _, v := range vendors {
		out = append(out, v.Label)
	}
	return out
}

func requestNamed(t *testing.T, grid purchasereport.Grid, name string) purchasereport.RequestNode {
	t.Helper()
	for _, v := range grid.Vendors {
		for _, r := range v.Requests {
			if r.Name == name {
				return r
			}
		}
	}
	t.Fatalf("no request named %s", name)
	return purchasereport.RequestNode{}
}

func text(s *string) string {
	if s == nil {
		return "—"
	}
	return *s
}

func TestGridHierarchy_042_FR014(t *testing.T) {
	t.Parallel()

	grid := reportOf(t, purchasereporttest.New(t), purchasereport.Query{}).Grid

	if diff := cmp.Diff([]string{"Acme", "Globex", "Initech", "Not filled in"}, labels(grid.Vendors)); diff != "" {
		t.Errorf("vendors by name, the request without a vendor under Not filled in (-want +got):\n%s", diff)
	}
	acme := grid.Vendors[0]
	if acme.PurchaseRequests != 3 || acme.Lines != 4 || amountText(acme.Committed) != "1800.00 USD" || amountText(acme.Invoiced) != "700.00 USD" {
		t.Errorf("Acme = %d requests, %d lines, committed %s, invoiced %s", acme.PurchaseRequests, acme.Lines,
			amountText(acme.Committed), amountText(acme.Invoiced))
	}
	if names := []string{acme.Requests[0].Name, acme.Requests[1].Name, acme.Requests[2].Name}; names[0] != "Brochures" || names[1] != "Event stand" || names[2] != "Plan two PR" {
		t.Errorf("requests under Acme = %v, want by name", names)
	}
}

func lineSummary(l purchasereport.LineNode) string {
	delivery := "—"
	if l.DeliveryDate != nil {
		delivery = l.DeliveryDate.String()
	}
	return strings.Join([]string{l.LineItem.Name, text(l.Activity), text(l.PONumber), text(l.POLine), text(l.POLineStatus),
		amountText(l.Invoiced), delivery}, "|")
}

func TestGridLines_042_FR019(t *testing.T) {
	t.Parallel()
	grid := reportOf(t, purchasereporttest.New(t), purchasereport.Query{}).Grid

	stand := requestNamed(t, grid, "Event stand")

	if got, want := lineSummary(stand.Lines[0]), "Line item 1|Activity 1|4500001|10|Open|400.00 USD|2026-03-10"; got != want {
		t.Errorf("first line = %s, want %s", got, want)
	}
}

func TestGridInvoiced_042_C5(t *testing.T) {
	t.Parallel()
	grid := reportOf(t, purchasereporttest.New(t), purchasereport.Query{}).Grid

	withPO := requestNamed(t, grid, "Event stand").Lines[1]
	withoutPO := requestNamed(t, grid, "Stage").Lines[0]

	if got := amountText(withPO.Invoiced); got != "0.00 USD" {
		t.Errorf("a line with a purchase order and no invoice shows %s, want 0.00 USD", got)
	}
	if got := lineSummary(withoutPO); got != "Line item 4|Activity 4|—|—|—|—|2026-04-15" {
		t.Errorf("a line without a purchase order = %s, want dashes for it and the invoiced amount", got)
	}
}

func TestGridRequest_042_FR014(t *testing.T) {
	t.Parallel()
	grid := reportOf(t, purchasereporttest.New(t), purchasereport.Query{}).Grid

	stand, stage := requestNamed(t, grid, "Event stand"), requestNamed(t, grid, "Stage")

	standGot := strings.Join([]string{amountText(stand.Amount), text(stand.PONumber), text(stand.Workflow), stand.Owner.Name}, "|")
	stageGot := fmt.Sprintf("%v|%s", stage.WaitingForGSAP, text(stage.GSAPNumber))
	if standGot != "1500.00 USD|4500001|Standard PR|Alice Ng" || stageGot != "true|GSAP-3" {
		t.Errorf("Event stand = %s, Stage = %s", standGot, stageGot)
	}
}

func TestGridHiddenPlan_042_FR020(t *testing.T) {
	t.Parallel()
	grid := reportOf(t, purchasereporttest.New(t), purchasereport.Query{}).Grid

	hidden := requestNamed(t, grid, "Plan two PR")

	if hidden.Amount != nil || hidden.Lines[0].Amount != nil || hidden.Lines[0].Invoiced != nil {
		t.Errorf("amounts of a plan without finance.view = %v, %v, want dashes", hidden.Amount, hidden.Lines[0].Amount)
	}
}

func TestGridSearch_042_FR017(t *testing.T) {
	t.Parallel()
	f := purchasereporttest.New(t)

	stage := reportOf(t, f, purchasereport.Query{Search: "STAGE"}).Grid
	lines := reportOf(t, f, purchasereport.Query{Search: "line item 2"}).Grid
	vendor := reportOf(t, f, purchasereport.Query{Search: "acme"}).Grid

	if stage.Total != 1 || labels(stage.Vendors)[0] != "Globex" || len(stage.Vendors[0].Requests) != 1 {
		t.Errorf("search stage = %v, want only Globex with the one request", labels(stage.Vendors))
	}
	if diff := cmp.Diff([]string{"Acme", "Initech"}, labels(lines.Vendors)); diff != "" {
		t.Errorf("search by line item keeps the path to each match (-want +got):\n%s", diff)
	}
	event := requestNamed(t, lines, "Event stand")
	if len(event.Lines) != 1 || event.Lines[0].LineItem.Name != "Line item 2" || amountText(event.Amount) != "1500.00 USD" {
		t.Errorf("Event stand under the search = %d lines and %s, want one line and the whole request's 1500.00", len(event.Lines), amountText(event.Amount))
	}
	if len(vendor.Vendors[0].Requests) != 3 {
		t.Errorf("search by vendor kept %d of its requests, want all 3", len(vendor.Vendors[0].Requests))
	}
}

func TestGridSort_042_FR017(t *testing.T) {
	t.Parallel()
	f := purchasereporttest.New(t)

	byAmount := reportOf(t, f, purchasereport.Query{Sort: "amount", Order: "desc"}).Grid
	byNameDesc := reportOf(t, f, purchasereport.Query{Order: "desc"}).Grid

	if diff := cmp.Diff([]string{"Globex", "Acme", "Initech", "Not filled in"}, labels(byAmount.Vendors)); diff != "" {
		t.Errorf("vendors by amount, the hidden plan counting as nothing (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff([]string{"Not filled in", "Initech", "Globex", "Acme"}, labels(byNameDesc.Vendors)); diff != "" {
		t.Errorf("vendors by name descending (-want +got):\n%s", diff)
	}
}

func TestGridPaging_042_C3(t *testing.T) {
	t.Parallel()
	f := purchasereporttest.New(t)
	f.Deps.Purchases = purchasereporttest.Purchases{Requests: purchasereporttest.Vendors(12)}

	third := reportOf(t, f, purchasereport.Query{PageSize: 5, Page: 3}).Grid
	beyond := reportOf(t, f, purchasereport.Query{PageSize: 5, Page: 9}).Grid
	defaults := reportOf(t, f, purchasereport.Query{}).Grid

	if third.Total != 12 || len(third.Vendors) != 2 || third.Page != 3 || third.PageSize != 5 {
		t.Errorf("third page = total %d, %d vendors, page %d of %d, want 12, 2, 3, 5", third.Total, len(third.Vendors), third.Page, third.PageSize)
	}
	if len(beyond.Vendors) != 0 || beyond.Total != 12 {
		t.Errorf("a page past the end = %d vendors of %d, want none of 12", len(beyond.Vendors), beyond.Total)
	}
	if defaults.PageSize != 25 || len(defaults.Vendors) != 12 {
		t.Errorf("defaults = page size %d, %d vendors, want 25 and all 12", defaults.PageSize, len(defaults.Vendors))
	}
}

func TestGridChoices_042_C6(t *testing.T) {
	t.Parallel()

	choices := reportOf(t, purchasereporttest.New(t), purchasereport.Query{Vendors: []string{"Acme"}}).Grid.Choices

	if diff := cmp.Diff([]string{"Acme", "Globex", "Initech", "none"}, choices.Vendors); diff != "" {
		t.Errorf("vendor choices come from everything in view, whatever is filtered (-want +got):\n%s", diff)
	}
	if len(choices.Owners) != 3 || choices.Owners[0].Name != "Alice Ng" || len(choices.Statuses) != 6 {
		t.Errorf("owners, statuses = %v, %v, want three people by name and the six statuses", choices.Owners, choices.Statuses)
	}
}

func TestParseQuery_042_FR016(t *testing.T) {
	t.Parallel()

	good, err := purchasereport.ParseQuery(url.Values{"sort": {"amount"}, "order": {"desc"}, "pageSize": {"25"}, "page": {"2"},
		"columns": {"id,invoiced,status"}, "vendor": {"Acme", "x"}, "requestOwner": {"p-bob"}, "status": {"approved"}, "q": {" stand "}})
	if err != nil || good.Sort != "amount" || good.PageSize != 25 || good.Page != 2 || good.Search != "stand" ||
		len(good.Vendors) != 2 || cmp.Diff([]string{"id", "invoiced", "status"}, good.Columns) != "" {
		t.Errorf("ParseQuery(good) = %+v, %v, want the values kept, columns in order and unknown filter values allowed", good, err)
	}
	bad := []url.Values{{"sort": {"nope"}}, {"order": {"up"}}, {"page": {"0"}}, {"pageSize": {"7"}}, {"columns": {"id,id"}},
		{"columns": {"nonsense"}}}
	for _, values := range bad {
		if _, err := purchasereport.ParseQuery(values); err == nil {
			t.Errorf("ParseQuery(%v) = nil error, want a problem", values)
		}
	}
}

func TestColumnsEchoed_042_FR016(t *testing.T) {
	t.Parallel()

	grid := reportOf(t, purchasereporttest.New(t), purchasereport.Query{Columns: []string{"status", "name"}}).Grid

	if diff := cmp.Diff([]string{"status", "name"}, grid.Columns); diff != "" {
		t.Errorf("Columns (-want +got):\n%s", diff)
	}
}
