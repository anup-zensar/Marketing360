package performance

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"mapp/internal/platform/actor"
	"mapp/internal/platform/money"
	"mapp/internal/platform/web"
)

func (h *Handler) registerFactRoutes(r gin.IRoutes) {
	r.GET("/activities/:activityId/performance", h.activityPanel)
	r.GET("/activities/:activityId/performance/facts", h.activityFacts)
	r.GET("/activities/:activityId/performance/facts/export", h.exportFacts)
}

func (h *Handler) activityPanel(c *gin.Context) {
	a, _ := actor.From(c.Request.Context())
	panel, err := h.service.ActivityPanel(c.Request.Context(), a, c.Param("activityId"))
	answer(c, http.StatusOK, panel, err)
}

func (h *Handler) activityFacts(c *gin.Context) {
	a, _ := actor.From(c.Request.Context())
	facts, err := h.service.ActivityFacts(c.Request.Context(), a, factQueryOf(c))
	if err != nil {
		web.RespondError(c, err)
		return
	}
	web.Respond(c, http.StatusOK, FactPage{Items: factViews(facts)})
}

func (h *Handler) exportFacts(c *gin.Context) {
	a, _ := actor.From(c.Request.Context())
	file, err := h.service.ExportFacts(c.Request.Context(), a, factQueryOf(c))
	if err != nil {
		web.RespondError(c, err)
		return
	}
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%q", file.Name))
	c.Data(http.StatusOK, file.ContentType, file.Body)
}

func factQueryOf(c *gin.Context) FactQuery {
	return FactQuery{
		ActivityID: c.Param("activityId"), Search: c.Query("search"), Limit: intQuery(c, "limit"),
	}
}

// FactAnswer shadows the domain's timestamp as the contract's plain date, and count and amount without
// omitempty: the contract answers them as null, never absent.
type FactAnswer struct {
	Fact
	SourceName string       `json:"sourceName"`
	FactDate   string       `json:"factDate"`
	Count      *int64       `json:"count"`
	Amount     *money.Money `json:"amount"`
}

type FactPage struct {
	Items []FactAnswer `json:"items"`
}

func factViews(facts []FactView) []FactAnswer {
	items := make([]FactAnswer, 0, len(facts))
	for _, view := range facts {
		items = append(items, FactAnswer{
			Fact: view.Fact, SourceName: view.SourceName, FactDate: view.FactDate.Format(time.DateOnly),
			Count: view.Count, Amount: view.Amount,
		})
	}
	return items
}
