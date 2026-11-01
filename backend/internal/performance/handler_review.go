package performance

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"mapp/internal/platform/actor"
	"mapp/internal/platform/web"
)

func (h *Handler) registerReviewRoutes(r gin.IRoutes) {
	r.GET("/performance/review-rows", h.reviewRows)
	r.GET("/performance/review-rows/:rowId", h.reviewRow)
	r.POST("/performance/review-rows/:rowId/dismiss", h.dismiss)
	r.POST("/performance/review-rows/:rowId/reopen", h.reopen)
	r.POST("/performance/review-rows/:rowId/correct", h.correct)
	r.POST("/performance/sources/:sourceId/recheck", h.recheckSource)
}

func (h *Handler) reviewRows(c *gin.Context) {
	a, _ := actor.From(c.Request.Context())
	filter, err := filterOf(c)
	if err != nil {
		web.RespondError(c, err)
		return
	}
	found, err := h.service.ReviewRows(c.Request.Context(), a, filter)
	if err != nil {
		web.RespondError(c, err)
		return
	}
	web.Respond(c, http.StatusOK, ReviewRowPage{Items: viewsOfRows(found)})
}

func filterOf(c *gin.Context) (ReviewFilter, error) {
	since, until, err := dateRangeOf(c)
	if err != nil {
		return ReviewFilter{}, err
	}
	return ReviewFilter{
		Status: c.Query("status"), SourceID: c.Query("sourceId"), ImportID: c.Query("importId"),
		Reason: c.Query("reason"), KeyText: c.Query("key"), Since: since, Until: until,
		Limit: intQuery(c, "limit"),
	}, nil
}

func (h *Handler) reviewRow(c *gin.Context) {
	a, _ := actor.From(c.Request.Context())
	row, err := h.service.ReviewRowByID(c.Request.Context(), a, c.Param("rowId"))
	if err != nil {
		web.RespondError(c, err)
		return
	}
	web.Respond(c, http.StatusOK, rowView(row))
}

func (h *Handler) correct(c *gin.Context) {
	var in CorrectInput
	if err := web.BindJSON(c, &in); err != nil {
		web.RespondError(c, err)
		return
	}
	a, _ := actor.From(c.Request.Context())
	saved, err := h.service.Correct(c.Request.Context(), a, c.Param("rowId"), in)
	h.answerReviewRow(c, a, saved, err)
}

func (h *Handler) recheckSource(c *gin.Context) {
	a, _ := actor.From(c.Request.Context())
	resolved, err := h.service.RecheckSource(c.Request.Context(), a, c.Param("sourceId"))
	answer(c, http.StatusOK, resolved, err)
}

func (h *Handler) dismiss(c *gin.Context) {
	var in DismissInput
	if err := web.BindJSON(c, &in); err != nil {
		web.RespondError(c, err)
		return
	}
	a, _ := actor.From(c.Request.Context())
	saved, err := h.service.Dismiss(c.Request.Context(), a, c.Param("rowId"), in)
	h.answerReviewRow(c, a, saved, err)
}

func (h *Handler) reopen(c *gin.Context) {
	a, _ := actor.From(c.Request.Context())
	saved, err := h.service.Reopen(c.Request.Context(), a, c.Param("rowId"))
	h.answerReviewRow(c, a, saved, err)
}

func (h *Handler) answerReviewRow(c *gin.Context, a actor.Actor, saved ReviewRow, err error) {
	if err != nil {
		web.RespondError(c, err)
		return
	}
	row, err := h.service.ReviewRowByID(c.Request.Context(), a, saved.ID)
	answer(c, http.StatusOK, rowView(row), err)
}

func intQuery(c *gin.Context, name string) int {
	value, err := strconv.Atoi(c.Query(name))
	if err != nil {
		return 0
	}
	return value
}

// rowView answers raw as the map it was stored from, so the contract's object shape holds.
type ReviewRowAnswer struct {
	ReviewRow
	Raw        map[string]string `json:"raw"`
	SourceName string            `json:"sourceName"`
	Candidates []string          `json:"candidates"`
	Faults     []Fault           `json:"faults"`
}

type ReviewRowPage struct {
	Items []ReviewRowAnswer `json:"items"`
}

func rowView(view ReviewRowView) ReviewRowAnswer {
	raw := map[string]string{}
	_ = json.Unmarshal(view.Raw, &raw)
	answerRow := ReviewRowAnswer{ReviewRow: view.ReviewRow, Raw: raw, SourceName: view.SourceName}
	answerRow.Candidates = nonNilStrings(view.Candidates)
	answerRow.Faults = view.Faults
	if answerRow.Faults == nil {
		answerRow.Faults = []Fault{}
	}
	return answerRow
}

func viewsOfRows(views []ReviewRowView) []ReviewRowAnswer {
	rows := make([]ReviewRowAnswer, 0, len(views))
	for _, view := range views {
		rows = append(rows, rowView(view))
	}
	return rows
}

func nonNilStrings(values []string) []string {
	if values == nil {
		return []string{}
	}
	return values
}
