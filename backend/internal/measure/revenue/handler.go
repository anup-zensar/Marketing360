package revenue

import (
	"net/http"

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

func (h *Handler) RegisterRoutes(r gin.IRoutes) {
	r.GET("/reports/revenue", h.revenue)
	r.GET("/reports/revenue/ranked", h.revenueRanked)
	r.GET("/reports/funnel", h.funnel)
	r.GET("/reports/funnel/ranked", h.funnelRanked)
}

func (h *Handler) revenue(c *gin.Context) {
	f, err := measure.ParseFilter(c.Request.URL.Query())
	if err != nil {
		web.RespondError(c, err)
		return
	}
	a, _ := actor.From(c.Request.Context())
	out, err := h.service.Report(c.Request.Context(), a, f)
	if err != nil {
		web.RespondError(c, err)
		return
	}
	web.Respond(c, http.StatusOK, out)
}

func (h *Handler) revenueRanked(c *gin.Context) {
	values := c.Request.URL.Query()
	f, err := measure.ParseFilter(values)
	if err != nil {
		web.RespondError(c, err)
		return
	}
	var problems errs.Problems
	q := parseRanked(values, []string{SortActualRevenue, SortProjectedRevenue}, SortProjectedRevenue, &problems)
	if problemErr := problems.Err(); problemErr != nil {
		web.RespondError(c, problemErr)
		return
	}
	a, _ := actor.From(c.Request.Context())
	out, err := h.service.RankedRevenue(c.Request.Context(), a, f, q)
	if err != nil {
		web.RespondError(c, err)
		return
	}
	web.Respond(c, http.StatusOK, out)
}

func (h *Handler) funnel(c *gin.Context) {
	f, err := measure.ParseFilter(c.Request.URL.Query())
	if err != nil {
		web.RespondError(c, err)
		return
	}
	a, _ := actor.From(c.Request.Context())
	out, err := h.service.FunnelReport(c.Request.Context(), a, f, c.Query(paramStage))
	if err != nil {
		web.RespondError(c, err)
		return
	}
	web.Respond(c, http.StatusOK, out)
}

func (h *Handler) funnelRanked(c *gin.Context) {
	values := c.Request.URL.Query()
	f, err := measure.ParseFilter(values)
	if err != nil {
		web.RespondError(c, err)
		return
	}
	var problems errs.Problems
	q := parseRanked(values, []string{SortActual, SortProjected}, SortProjected, &problems)
	q.Stage = values.Get(paramStage)
	if problemErr := problems.Err(); problemErr != nil {
		web.RespondError(c, problemErr)
		return
	}
	a, _ := actor.From(c.Request.Context())
	out, err := h.service.RankedFunnel(c.Request.Context(), a, f, q)
	if err != nil {
		web.RespondError(c, err)
		return
	}
	web.Respond(c, http.StatusOK, out)
}
