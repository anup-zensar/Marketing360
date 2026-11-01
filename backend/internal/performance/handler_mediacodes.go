package performance

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"mapp/internal/platform/actor"
	"mapp/internal/platform/web"
)

func (h *Handler) registerCodeRoutes(r gin.IRoutes) {
	r.GET("/activities/:activityId/media-codes", h.mediaCodes)
	r.POST("/activities/:activityId/media-codes", h.generateCode)
	r.PATCH("/media-codes/:codeId", h.editCode)
	r.DELETE("/media-codes/:codeId", h.deleteCode)
	r.POST("/media-codes/:codeId/deactivate", h.deactivateCode)
	r.GET("/performance/configuration", h.configuration)
	r.PATCH("/performance/configuration", h.saveConfiguration)
}

func (h *Handler) mediaCodes(c *gin.Context) {
	a, _ := actor.From(c.Request.Context())
	area, err := h.service.MediaCodes(c.Request.Context(), a, c.Param("activityId"))
	answer(c, http.StatusOK, area, err)
}

func (h *Handler) generateCode(c *gin.Context) {
	var in MediaCodeInput
	if err := web.BindJSON(c, &in); err != nil {
		web.RespondError(c, err)
		return
	}
	a, _ := actor.From(c.Request.Context())
	created, err := h.service.GenerateMediaCode(c.Request.Context(), a, c.Param("activityId"), in)
	answer(c, http.StatusCreated, created, err)
}

func (h *Handler) editCode(c *gin.Context) {
	var in MediaCodeInput
	if err := web.BindJSON(c, &in); err != nil {
		web.RespondError(c, err)
		return
	}
	a, _ := actor.From(c.Request.Context())
	saved, err := h.service.EditMediaCode(c.Request.Context(), a, c.Param("codeId"), in)
	answer(c, http.StatusOK, saved, err)
}

func (h *Handler) deleteCode(c *gin.Context) {
	a, _ := actor.From(c.Request.Context())
	if err := h.service.DeleteMediaCode(c.Request.Context(), a, c.Param("codeId")); err != nil {
		web.RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) deactivateCode(c *gin.Context) {
	a, _ := actor.From(c.Request.Context())
	saved, err := h.service.DeactivateMediaCode(c.Request.Context(), a, c.Param("codeId"))
	answer(c, http.StatusOK, saved, err)
}

func (h *Handler) configuration(c *gin.Context) {
	a, _ := actor.From(c.Request.Context())
	config, err := h.service.PerformanceConfiguration(c.Request.Context(), a)
	if err != nil {
		web.RespondError(c, err)
		return
	}
	web.Respond(c, http.StatusOK, configView(config, h.now()))
}

func (h *Handler) saveConfiguration(c *gin.Context) {
	var in ConfigurationInput
	if err := web.BindJSON(c, &in); err != nil {
		web.RespondError(c, err)
		return
	}
	a, _ := actor.From(c.Request.Context())
	saved, err := h.service.SaveConfiguration(c.Request.Context(), a, in)
	if err != nil {
		web.RespondError(c, err)
		return
	}
	web.Respond(c, http.StatusOK, configView(saved, h.now()))
}

// configView answers a never-saved configuration with the reader's own moment, so the contract's required
// stamps hold before the first save.
type ConfigurationAnswer struct {
	Configuration
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func configView(config Configuration, now time.Time) ConfigurationAnswer {
	view := ConfigurationAnswer{Configuration: config, CreatedAt: config.CreatedAt, UpdatedAt: config.UpdatedAt}
	if config.CreatedAt.IsZero() {
		view.CreatedAt, view.UpdatedAt = now, now
	}
	return view
}
