package useraccess

import (
	"net/url"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"unicode/utf8"

	"mapp/internal/platform/errs"
)

// The report's own address parameters (contract: UserAccessSearch, UserAccessTeam, UserAccessStatus, UserAccessSort).
const (
	paramSearch = "q"
	paramTeam   = "team"
	paramStatus = "status"
	paramSort   = "sort"
)

// Paging (FR-006).
const (
	DefaultPageSize = 20
	MinPageSize     = 5
	MaxPageSize     = 500
	maxSearch       = 100
	maxTeams        = 200
	descending      = "-"
	sortName        = "name"
)

func sortKeys() []string {
	return []string{sortName, "email", "jobTitle", "roles", "teams", "status", "lastSignedIn"}
}

func teamID() *regexp.Regexp { return regexp.MustCompile(`^T-[0-9]+$`) }

// ParseQuery reads the report's parameters against their allowed values (FR-006); every problem is answered at once.
func ParseQuery(values url.Values) (Query, error) {
	var problems errs.Problems
	q := Query{Search: strings.TrimSpace(values.Get(paramSearch)), Teams: values[paramTeam],
		Status: values.Get(paramStatus), Sort: values.Get(paramSort)}
	if utf8.RuneCountInString(q.Search) > maxSearch {
		problems.Add(paramSearch, "Search with at most 100 characters.")
	}
	pattern := teamID()
	if len(q.Teams) > maxTeams || slices.ContainsFunc(q.Teams, func(id string) bool { return !pattern.MatchString(id) }) {
		problems.Add(paramTeam, "Choose teams by their IDs, such as T-1001.")
	}
	if q.Status != "" && !slices.Contains([]string{StatusActive, StatusInactive, StatusAll}, q.Status) {
		problems.Add(paramStatus, "Choose active, inactive or all.")
	}
	if q.Sort != "" && !slices.Contains(sortKeys(), strings.TrimPrefix(q.Sort, descending)) {
		problems.Add(paramSort, "Sort by one of: "+strings.Join(sortKeys(), ", ")+"; a minus sign first for descending.")
	}
	q.Page = whole(values, "page", 1, &problems)
	q.PageSize = whole(values, "pageSize", MinPageSize, &problems)
	if q.PageSize > MaxPageSize {
		problems.Add("pageSize", "Show at most 500 rows on a page.")
	}
	return withDefaults(q), problems.Err()
}

func withDefaults(q Query) Query {
	if q.Status == "" {
		q.Status = StatusActive
	}
	if q.Sort == "" {
		q.Sort = sortName
	}
	if q.Teams == nil {
		q.Teams = []string{}
	}
	q.Page, q.PageSize = max(q.Page, 1), max(q.PageSize, 0)
	if q.PageSize == 0 {
		q.PageSize = DefaultPageSize
	}
	return q
}

// whole reads a whole number of at least least; 0 when not given.
func whole(values url.Values, name string, least int, problems *errs.Problems) int {
	text := values.Get(name)
	if text == "" {
		return 0
	}
	n, err := strconv.Atoi(text)
	if err != nil || n < least {
		problems.Add(name, "Enter a whole number of at least "+strconv.Itoa(least)+".")
		return 0
	}
	return n
}
