package settings

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *Handler) listActivityCategories(c *gin.Context) {
	categories, err := h.service.ListActivityCategories(c.Request.Context(), actorOf(c))
	respond(c, http.StatusOK, ActivityCategoryPage{Items: categories, Total: len(categories)}, err)
}

func (h *Handler) getActivityCategory(c *gin.Context) {
	category, err := h.service.GetActivityCategory(c.Request.Context(), actorOf(c), c.Param("categoryId"))
	respond(c, http.StatusOK, category, err)
}

func (h *Handler) createActivityCategory(c *gin.Context) {
	var input ActivityCategoryInput
	if !readBody(c, &input) {
		return
	}
	category, err := h.service.CreateActivityCategory(c.Request.Context(), actorOf(c), input)
	respond(c, http.StatusCreated, category, err)
}

func (h *Handler) updateActivityCategory(c *gin.Context) {
	var update ActivityCategoryUpdate
	if !readBody(c, &update) {
		return
	}
	category, err := h.service.UpdateActivityCategory(c.Request.Context(), actorOf(c), c.Param("categoryId"), update)
	respond(c, http.StatusOK, category, err)
}

func (h *Handler) deleteActivityCategory(c *gin.Context) {
	entry, err := h.service.DeleteActivityCategory(c.Request.Context(), actorOf(c), c.Param("categoryId"))
	respondDeleted(c, entry, err)
}

func (h *Handler) activityCategoryRoutes(r gin.IRoutes) {
	r.GET("/settings/activity-categories", h.listActivityCategories)
	r.POST("/settings/activity-categories", h.createActivityCategory)
	r.GET("/settings/activity-categories/:categoryId", h.getActivityCategory)
	r.PATCH("/settings/activity-categories/:categoryId", h.updateActivityCategory)
	r.DELETE("/settings/activity-categories/:categoryId", h.deleteActivityCategory)
}
