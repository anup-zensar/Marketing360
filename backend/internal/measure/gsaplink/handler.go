package gsaplink

import (
	"cmp"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"mapp/internal/measure"
	"mapp/internal/platform/actor"
	"mapp/internal/platform/errs"
	"mapp/internal/platform/web"
)

type Handler struct {
	service *Service
}

func NewHandler(s *Service) *Handler {
	return &Handler{service: s}
}

// RegisterRoutes mounts the one read of the report: nothing on it writes (FR-002).
func (h *Handler) RegisterRoutes(r gin.IRoutes) {
	r.GET("/reports/gsap-link", h.report)
}

func (h *Handler) report(c *gin.Context) {
	values := c.Request.URL.Query()
	f, err := measure.ParseFilter(values)
	if err != nil {
		web.RespondError(c, err)
		return
	}
	t, err := ParseTable(values)
	if err != nil {
		web.RespondError(c, err)
		return
	}
	a, _ := actor.From(c.Request.Context())
	r, err := h.service.Report(c.Request.Context(), a, f, t)
	if err != nil {
		web.RespondError(c, err)
		return
	}
	web.Respond(c, http.StatusOK, r)
}

// ParseTable reads tab, sort, page and pageSize from an address, each against its allowed values (FR-015).
func ParseTable(values url.Values) (Table, error) {
	var problems errs.Problems
	t := Table{Tab: values.Get("tab"), Sort: values.Get("sort")}
	if t.Tab != "" && t.Tab != TabSwbs && t.Tab != TabWaiting && t.Tab != TabOrders {
		problems.Add("tab", "Choose swbs, waiting or orders.")
	}
	if tab := cmp.Or(t.Tab, TabSwbs); t.Sort != "" && !ValidSort(tab, t.Sort) {
		problems.Add("sort", "Sort by one of: "+strings.Join(SortKeys(tab), ", ")+"; a minus sign first for descending.")
	}
	t.Page = whole(values, "page", 1, &problems)
	t.PageSize = whole(values, "pageSize", MinPageSize, &problems)
	if t.PageSize > MaxPageSize {
		problems.Add("pageSize", "Show at most 500 rows on a page.")
	}
	return t, problems.Err()
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
