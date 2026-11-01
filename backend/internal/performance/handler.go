package performance

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"mapp/internal/platform/actor"
	"mapp/internal/platform/clock"
	"mapp/internal/platform/web"
)

type Handler struct {
	service *Service
	now     clock.Clock
}

func NewHandler(s *Service, now clock.Clock) *Handler {
	return &Handler{service: s, now: now}
}

func (h *Handler) RegisterRoutes(r gin.IRoutes) {
	r.GET("/performance/sources", h.sources)
	r.POST("/performance/sources", h.createSource)
	r.GET("/performance/sources/:sourceId", h.source)
	r.PATCH("/performance/sources/:sourceId", h.updateSource)
	r.DELETE("/performance/sources/:sourceId", h.deleteSource)
	r.POST("/performance/sources/:sourceId/deactivate", h.deactivate)
	r.POST("/performance/sources/:sourceId/reactivate", h.reactivate)
	r.GET("/performance/sources/:sourceId/mappings", h.mappings)
	r.POST("/performance/sources/:sourceId/mappings", h.createMapping)
	r.GET("/performance/sources/:sourceId/template", h.template)
	r.PATCH("/performance/mappings/:mappingId", h.updateMapping)
	r.DELETE("/performance/mappings/:mappingId", h.deleteMapping)
	h.registerImportRoutes(r)
	h.registerKeyRoutes(r)
	h.registerReviewRoutes(r)
	h.registerFactRoutes(r)
	h.registerCodeRoutes(r)
}

func (h *Handler) sources(c *gin.Context) {
	a, _ := actor.From(c.Request.Context())
	found, err := h.service.Sources(c.Request.Context(), a)
	if err != nil {
		web.RespondError(c, err)
		return
	}
	web.Respond(c, http.StatusOK, SourceList{Items: viewsOf(found, h.now())})
}

func (h *Handler) source(c *gin.Context) {
	a, _ := actor.From(c.Request.Context())
	found, err := h.service.Source(c.Request.Context(), a, c.Param("sourceId"))
	if err != nil {
		web.RespondError(c, err)
		return
	}
	web.Respond(c, http.StatusOK, viewOf(found, h.now()))
}

func (h *Handler) createSource(c *gin.Context) {
	var in SourceInput
	if err := web.BindJSON(c, &in); err != nil {
		web.RespondError(c, err)
		return
	}
	a, _ := actor.From(c.Request.Context())
	created, err := h.service.CreateSource(c.Request.Context(), a, in)
	h.answerSource(c, http.StatusCreated, created, err)
}

func (h *Handler) updateSource(c *gin.Context) {
	var in SourceInput
	if err := web.BindJSON(c, &in); err != nil {
		web.RespondError(c, err)
		return
	}
	a, _ := actor.From(c.Request.Context())
	saved, err := h.service.UpdateSource(c.Request.Context(), a, c.Param("sourceId"), in)
	h.answerSource(c, http.StatusOK, saved, err)
}

func (h *Handler) deleteSource(c *gin.Context) {
	a, _ := actor.From(c.Request.Context())
	if err := h.service.DeleteSource(c.Request.Context(), a, c.Param("sourceId")); err != nil {
		web.RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) deactivate(c *gin.Context) {
	h.setStatus(c, StatusInactive)
}

func (h *Handler) reactivate(c *gin.Context) {
	h.setStatus(c, StatusActive)
}

func (h *Handler) setStatus(c *gin.Context, status Status) {
	a, _ := actor.From(c.Request.Context())
	id := c.Param("sourceId")
	if err := h.service.SetSourceStatus(c.Request.Context(), a, id, status); err != nil {
		web.RespondError(c, err)
		return
	}
	found, err := h.service.Source(c.Request.Context(), a, id)
	if err != nil {
		web.RespondError(c, err)
		return
	}
	web.Respond(c, http.StatusOK, viewOf(found, h.now()))
}

func (h *Handler) mappings(c *gin.Context) {
	a, _ := actor.From(c.Request.Context())
	found, err := h.service.Mappings(c.Request.Context(), a, c.Param("sourceId"))
	if err != nil {
		web.RespondError(c, err)
		return
	}
	web.Respond(c, http.StatusOK, MappingList{Items: found})
}

func (h *Handler) createMapping(c *gin.Context) {
	var in MappingInput
	if err := web.BindJSON(c, &in); err != nil {
		web.RespondError(c, err)
		return
	}
	a, _ := actor.From(c.Request.Context())
	created, err := h.service.CreateMapping(c.Request.Context(), a, c.Param("sourceId"), in)
	answer(c, http.StatusCreated, created, err)
}

func (h *Handler) updateMapping(c *gin.Context) {
	var in MappingInput
	if err := web.BindJSON(c, &in); err != nil {
		web.RespondError(c, err)
		return
	}
	a, _ := actor.From(c.Request.Context())
	saved, err := h.service.UpdateMapping(c.Request.Context(), a, c.Param("mappingId"), in)
	answer(c, http.StatusOK, saved, err)
}

func (h *Handler) deleteMapping(c *gin.Context) {
	a, _ := actor.From(c.Request.Context())
	if err := h.service.DeleteMapping(c.Request.Context(), a, c.Param("mappingId")); err != nil {
		web.RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) template(c *gin.Context) {
	a, _ := actor.From(c.Request.Context())
	file, err := h.service.Template(c.Request.Context(), a, c.Param("sourceId"))
	if err != nil {
		web.RespondError(c, err)
		return
	}
	c.Data(http.StatusOK, "text/csv; charset=utf-8", []byte(file))
}

func (h *Handler) answerSource(c *gin.Context, status int, src Source, err error) {
	if err != nil {
		web.RespondError(c, err)
		return
	}
	web.Respond(c, status, viewOf(SourceView{Source: src}, h.now()))
}

func answer(c *gin.Context, status int, v any, err error) {
	if err != nil {
		web.RespondError(c, err)
		return
	}
	web.Respond(c, status, v)
}
