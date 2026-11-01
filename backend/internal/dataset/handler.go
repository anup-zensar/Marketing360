package dataset

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"mapp/internal/history"
	"mapp/internal/platform/actor"
	"mapp/internal/platform/errs"
	"mapp/internal/platform/web"
)

const idParam = "datasetId"

type Handler struct {
	service *Service
}

func NewHandler(s *Service) *Handler {
	return &Handler{service: s}
}

func (h *Handler) RegisterRoutes(r gin.IRoutes) {
	r.GET("/datasets", h.list)
	r.POST("/datasets/preview", h.preview)
	r.POST("/datasets", h.create)
	r.GET("/datasets/:datasetId", h.get)
	r.PATCH("/datasets/:datasetId", h.update)
	r.DELETE("/datasets/:datasetId", h.delete)
	r.GET("/datasets/:datasetId/rows", h.rows)
	r.PUT("/datasets/:datasetId/sharing", h.share)
}

func (h *Handler) list(c *gin.Context) {
	a, _ := actor.From(c.Request.Context())
	body, err := h.service.List(c.Request.Context(), a)
	respond(c, http.StatusOK, body, err)
}

func (h *Handler) preview(c *gin.Context) {
	var up Upload
	if err := web.BindJSON(c, &up); err != nil {
		web.RespondError(c, err)
		return
	}
	a, _ := actor.From(c.Request.Context())
	body, err := h.service.Preview(c.Request.Context(), a, up)
	respond(c, http.StatusOK, body, err)
}

func (h *Handler) create(c *gin.Context) {
	var in CreateInput
	if err := web.BindJSON(c, &in); err != nil {
		web.RespondError(c, err)
		return
	}
	a, _ := actor.From(c.Request.Context())
	body, err := h.service.Create(c.Request.Context(), a, in)
	respond(c, http.StatusCreated, body, err)
}

func (h *Handler) get(c *gin.Context) {
	a, _ := actor.From(c.Request.Context())
	body, err := h.service.Get(c.Request.Context(), a, c.Param(idParam))
	respond(c, http.StatusOK, body, err)
}

func (h *Handler) update(c *gin.Context) {
	var in UpdateInput
	if err := web.BindJSON(c, &in); err != nil {
		web.RespondError(c, err)
		return
	}
	a, _ := actor.From(c.Request.Context())
	body, err := h.service.Update(c.Request.Context(), a, c.Param(idParam), in)
	respond(c, http.StatusOK, body, err)
}

func (h *Handler) delete(c *gin.Context) {
	a, _ := actor.From(c.Request.Context())
	entry, err := h.service.Delete(c.Request.Context(), a, c.Param(idParam))
	respond(c, http.StatusOK, history.DeleteResult{RecycleBinEntry: entry.View()}, err)
}

func (h *Handler) rows(c *gin.Context) {
	var q RowsQuery
	var problems errs.Problems
	q.Offset = intParam(c, offsetField, msgOffset, &problems)
	q.Limit = intParam(c, limitField, msgLimit, &problems)
	if err := problems.Err(); err != nil {
		web.RespondError(c, err)
		return
	}
	a, _ := actor.From(c.Request.Context())
	body, err := h.service.Rows(c.Request.Context(), a, c.Param(idParam), q)
	respond(c, http.StatusOK, body, err)
}

func intParam(c *gin.Context, name, message string, problems *errs.Problems) *int {
	text, given := c.GetQuery(name)
	if !given {
		return nil
	}
	n, err := strconv.Atoi(text)
	if err != nil {
		problems.Add(name, message)
		return nil
	}
	return &n
}

func (h *Handler) share(c *gin.Context) {
	var in SharingInput
	if err := web.BindJSON(c, &in); err != nil {
		web.RespondError(c, err)
		return
	}
	a, _ := actor.From(c.Request.Context())
	body, err := h.service.Share(c.Request.Context(), a, c.Param(idParam), in)
	respond(c, http.StatusOK, body, err)
}

func respond(c *gin.Context, status int, body any, err error) {
	if err != nil {
		web.RespondError(c, err)
		return
	}
	web.Respond(c, status, body)
}
