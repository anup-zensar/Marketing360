package performance

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"mapp/internal/platform/actor"
	"mapp/internal/platform/web"
)

func (h *Handler) registerKeyRoutes(r gin.IRoutes) {
	r.GET("/activities/:activityId/performance/sources", h.sourcesArea)
	r.GET("/activities/:activityId/external-name-aliases", h.aliases)
	r.DELETE("/activities/:activityId/external-name-aliases/:aliasId", h.deleteAlias)
	r.GET("/activities/:activityId/integration-ids", h.integrationIDs)
	r.PUT("/activities/:activityId/integration-ids/:sourceId", h.setIntegrationID)
	r.DELETE("/activities/:activityId/integration-ids/:sourceId", h.deleteIntegrationID)
}

func (h *Handler) sourcesArea(c *gin.Context) {
	a, _ := actor.From(c.Request.Context())
	area, err := h.service.SourcesArea(c.Request.Context(), a, c.Param("activityId"))
	answer(c, http.StatusOK, area, err)
}

func (h *Handler) aliases(c *gin.Context) {
	a, _ := actor.From(c.Request.Context())
	found, err := h.service.Aliases(c.Request.Context(), a, c.Param("activityId"))
	if err != nil {
		web.RespondError(c, err)
		return
	}
	web.Respond(c, http.StatusOK, AliasList{Items: found})
}

func (h *Handler) deleteAlias(c *gin.Context) {
	a, _ := actor.From(c.Request.Context())
	if err := h.service.DeleteAlias(c.Request.Context(), a, c.Param("activityId"), c.Param("aliasId")); err != nil {
		web.RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) integrationIDs(c *gin.Context) {
	a, _ := actor.From(c.Request.Context())
	found, err := h.service.IntegrationIDs(c.Request.Context(), a, c.Param("activityId"))
	if err != nil {
		web.RespondError(c, err)
		return
	}
	web.Respond(c, http.StatusOK, IntegrationIDList{Items: found})
}

func (h *Handler) setIntegrationID(c *gin.Context) {
	var in IntegrationIDInput
	if err := web.BindJSON(c, &in); err != nil {
		web.RespondError(c, err)
		return
	}
	a, _ := actor.From(c.Request.Context())
	ref := KeyRef{ActivityID: c.Param("activityId"), SourceID: c.Param("sourceId")}
	saved, err := h.service.SetIntegrationID(c.Request.Context(), a, ref, in)
	answer(c, http.StatusCreated, saved, err)
}

func (h *Handler) deleteIntegrationID(c *gin.Context) {
	a, _ := actor.From(c.Request.Context())
	ref := KeyRef{ActivityID: c.Param("activityId"), SourceID: c.Param("sourceId")}
	if err := h.service.DeleteIntegrationID(c.Request.Context(), a, ref); err != nil {
		web.RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

type AliasList struct {
	Items []Alias `json:"items"`
}

type IntegrationIDList struct {
	Items []IntegrationID `json:"items"`
}
