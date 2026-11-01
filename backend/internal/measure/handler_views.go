package measure

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"mapp/internal/platform/actor"
	"mapp/internal/platform/web"
)

func (h *Handler) listViews(c *gin.Context) {
	a, _ := actor.From(c.Request.Context())
	page, err := h.service.ListViews(c.Request.Context(), a, reportOf(c))
	respond(c, http.StatusOK, page, err)
}

func (h *Handler) createView(c *gin.Context) {
	var in ViewCreate
	if err := web.BindJSON(c, &in); err != nil {
		web.RespondError(c, err)
		return
	}
	a, _ := actor.From(c.Request.Context())
	created, err := h.service.CreateView(c.Request.Context(), a, reportOf(c), in)
	respond(c, http.StatusCreated, created, err)
}

func (h *Handler) updateView(c *gin.Context) {
	var in ViewUpdate
	if err := web.BindJSON(c, &in); err != nil {
		web.RespondError(c, err)
		return
	}
	a, _ := actor.From(c.Request.Context())
	updated, err := h.service.UpdateView(c.Request.Context(), a, ViewRef{Report: reportOf(c), ID: c.Param("viewId")}, in)
	respond(c, http.StatusOK, updated, err)
}

func (h *Handler) deleteView(c *gin.Context) {
	a, _ := actor.From(c.Request.Context())
	deleted, err := h.service.DeleteView(c.Request.Context(), a, ViewRef{Report: reportOf(c), ID: c.Param("viewId")})
	respond(c, http.StatusOK, deleted, err)
}

// reportOf is the key of the path's report ("plan-budget"); "" for an unknown path, which the service answers
// with 404.
func reportOf(c *gin.Context) string {
	key, _ := KeyOfPath(c.Param("report"))
	return key
}

func respond(c *gin.Context, status int, body any, err error) {
	if err != nil {
		web.RespondError(c, err)
		return
	}
	web.Respond(c, status, body)
}
