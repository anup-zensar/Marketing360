package measure

import (
	"errors"
	"net/url"
	"slices"
	"strconv"
	"strings"

	"mapp/internal/allwork"
	"mapp/internal/platform/errs"
)

const (
	msgChooseListed   = "Choose one of the listed values."
	msgChooseCurrency = "Choose a currency by its code, such as EUR."
	msgChooseTables   = "Choose at most 50 widgets by their keys."
	maxTables         = 50
	maxWidgetKey      = 40
)

// allWorkParams are the parameters of a report filter that All work reads (data-model §2, research R1).
func allWorkParams() []string {
	return []string{"year", "plan", "planType", "team", "owner", "activityType", "activityStatus", "attr"}
}

// ParseFilter reads a report filter from the page address (data-model §2): All work's filters through
// allwork.ParseQuery, then period, show, currency and table; every problem is answered at once.
func ParseFilter(values url.Values) (Filter, error) {
	var problems errs.Problems
	query := url.Values{}
	for _, name := range allWorkParams() {
		if v, given := values[name]; given {
			query[name] = v
		}
	}
	q, err := allwork.ParseQuery(query)
	if e := (*errs.Error)(nil); errors.As(err, &e) {
		problems = append(problems, e.Fields...)
	} else if err != nil {
		return Filter{}, err
	}
	f := controls(values, &problems)
	f.Query, f.Own = q, ownParams(values)
	if err := problems.Err(); err != nil {
		return Filter{}, err
	}
	return f, nil
}

// controls reads the report's own controls: period (months by default), show, currency and table.
func controls(values url.Values, problems *errs.Problems) Filter {
	f := Filter{Period: values.Get("period"), Show: values.Get("show"), Currency: values.Get("currency"),
		Tables: uniqueList(values["table"])}
	if !listed(f.Period, []string{PeriodMonth, PeriodQuarter}) {
		problems.Add("period", msgChooseListed)
	}
	if !listed(f.Show, []string{ShowPlan, ShowCampaign, ShowInitiative, ShowActivity}) {
		problems.Add("show", msgChooseListed)
	}
	if f.Currency != "" && !isCurrencyCode(f.Currency) {
		problems.Add("currency", msgChooseCurrency)
	}
	if len(f.Tables) > maxTables || slices.ContainsFunc(f.Tables, func(key string) bool { return !isWidgetKey(key) }) {
		problems.Add("table", msgChooseTables)
	}
	if f.Period == "" {
		f.Period = PeriodMonth
	}
	return f
}

// Canonical is the filter as one query string: the parameters in data-model §2's order, values sorted, defaults left out.
func Canonical(f Filter) string {
	var parts []string
	add := func(name string, values ...string) {
		for _, v := range slices.Sorted(slices.Values(values)) {
			parts = append(parts, url.QueryEscape(name)+"="+url.QueryEscape(v))
		}
	}
	if f.Query.Year != 0 {
		add("year", strconv.Itoa(f.Query.Year))
	}
	add("plan", f.Query.Plans...)
	add("planType", f.Query.PlanTypes...)
	add("team", f.Query.Teams...)
	add("owner", f.Query.Owners...)
	add("activityType", f.Query.ActivityTypes...)
	add("activityStatus", f.Query.ActivityStatuses...)
	var attrs []string
	for _, a := range f.Query.Attributes {
		attrs = append(attrs, a.AttributeID+":"+strings.Join(slices.Sorted(slices.Values(a.Values)), ","))
	}
	add("attr", attrs...)
	if f.Period != PeriodMonth {
		add("period", nonEmpty(f.Period)...)
	}
	add("show", nonEmpty(f.Show)...)
	add("currency", nonEmpty(f.Currency)...)
	add("table", f.Tables...)
	return strings.Join(parts, "&")
}

// ownParams are the parameters the shared filter does not read: a report's own, such as a tab or a sort.
func ownParams(values url.Values) url.Values {
	shared := append(allWorkParams(), "period", "show", "currency", "table", "report")
	own := url.Values{}
	for name, v := range values {
		if !slices.Contains(shared, name) {
			own[name] = v
		}
	}
	return own
}

func listed(value string, allowed []string) bool {
	return value == "" || slices.Contains(allowed, value)
}

func uniqueList(values []string) []string {
	kept := []string{}
	for _, v := range values {
		if v = strings.TrimSpace(v); v != "" && !slices.Contains(kept, v) {
			kept = append(kept, v)
		}
	}
	return kept
}

func nonEmpty(value string) []string {
	if value == "" {
		return nil
	}
	return []string{value}
}

func isCurrencyCode(code string) bool {
	return len(code) == 3 && strings.Trim(code, "ABCDEFGHIJKLMNOPQRSTUVWXYZ") == ""
}

// isWidgetKey is 039's ReportWidget pattern: a lower-case letter, then at most 39 letters and digits.
func isWidgetKey(key string) bool {
	if key == "" || len(key) > maxWidgetKey || key[0] < 'a' || key[0] > 'z' {
		return false
	}
	return strings.Trim(key, "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789") == ""
}
