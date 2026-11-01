package purchasereport

import (
	"slices"
	"strings"
)

// reqView is a request with the lines of it the search kept; its totals stay those of the whole request.
type reqView struct {
	r     Request
	lines []Line
}

type vendorGroup struct {
	vendor   string
	requests []reqView
}

func groupsOf(rs []Request) []vendorGroup {
	byVendor := map[string]*vendorGroup{}
	var order []string
	for _, r := range rs {
		g := byVendor[r.Vendor]
		if g == nil {
			g = &vendorGroup{vendor: r.Vendor}
			byVendor[r.Vendor] = g
			order = append(order, r.Vendor)
		}
		g.requests = append(g.requests, reqView{r: r, lines: r.Lines})
	}
	groups := make([]vendorGroup, 0, len(order))
	for _, v := range order {
		groups = append(groups, *byVendor[v])
	}
	return groups
}

// searched keeps the vendors, requests and lines whose text holds every word, with the path down to each match.
func (d data) searched(groups []vendorGroup, words []string) []vendorGroup {
	if len(words) == 0 {
		return groups
	}
	kept := []vendorGroup{}
	for _, g := range groups {
		if containsAll(vendorLabel(g.vendor), words) {
			kept = append(kept, g)
			continue
		}
		if requests := d.matchingRequests(g.requests, words); len(requests) > 0 {
			kept = append(kept, vendorGroup{vendor: g.vendor, requests: requests})
		}
	}
	return kept
}

func (d data) matchingRequests(requests []reqView, words []string) []reqView {
	var kept []reqView
	for _, rv := range requests {
		if containsAll(d.requestText(rv.r), words) {
			kept = append(kept, rv)
			continue
		}
		lines := slices.DeleteFunc(slices.Clone(rv.lines), func(l Line) bool { return !containsAll(d.lineText(l), words) })
		if len(lines) > 0 {
			kept = append(kept, reqView{r: rv.r, lines: lines})
		}
	}
	return kept
}

func (d data) requestText(r Request) string {
	parts := []string{r.ID, r.Name, string(r.Status), d.owner(r.OwnerID).Name, r.Workflow, r.GSAPNumber}
	for _, l := range r.Lines {
		parts = append(parts, l.PONumber)
	}
	return strings.Join(parts, "\n")
}

func (d data) lineText(l Line) string {
	info := d.items[l.LineItemID]
	return strings.Join([]string{l.ID, info.name, info.activity, l.PONumber, l.POLine, l.POLineStatus}, "\n")
}

func containsAll(text string, words []string) bool {
	text = strings.ToLower(text)
	return !slices.ContainsFunc(words, func(w string) bool { return !strings.Contains(text, w) })
}
