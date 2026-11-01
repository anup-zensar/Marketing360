package customdashboard

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"mapp/internal/history"
	"mapp/internal/platform/actor"
	"mapp/internal/platform/web"
)

// MaxBodyBytes holds a dashboard of 30 widgets with room to spare (038's body limits).
const MaxBodyBytes = 512 << 10

const dashboardIDParam = "dashboardId"

type Handler struct {
	service *Service
}

func NewHandler(s *Service) *Handler {
	return &Handler{service: s}
}

func (h *Handler) RegisterRoutes(r gin.IRoutes) {
	r.GET("/custom-dashboards", h.list)
	r.POST("/custom-dashboards", h.create)
	r.GET("/custom-dashboards/:dashboardId", h.get)
	r.PATCH("/custom-dashboards/:dashboardId", h.update)
	r.DELETE("/custom-dashboards/:dashboardId", h.delete)
	r.POST("/custom-dashboards/:dashboardId/copy", h.copy)
	r.PUT("/custom-dashboards/:dashboardId/sharing", h.share)
}

func (h *Handler) list(c *gin.Context) {
	a, _ := actor.From(c.Request.Context())
	list, err := h.service.List(c.Request.Context(), a)
	respond(c, http.StatusOK, list, err)
}

func (h *Handler) create(c *gin.Context) {
	var in CreateInput
	if !bindLimited(c, &in) {
		return
	}
	a, _ := actor.From(c.Request.Context())
	view, err := h.service.Create(c.Request.Context(), a, in)
	respond(c, http.StatusCreated, view, err)
}

func (h *Handler) get(c *gin.Context) {
	a, _ := actor.From(c.Request.Context())
	view, err := h.service.Get(c.Request.Context(), a, c.Param(dashboardIDParam))
	respond(c, http.StatusOK, view, err)
}

func (h *Handler) update(c *gin.Context) {
	var in UpdateInput
	if !bindLimited(c, &in) {
		return
	}
	a, _ := actor.From(c.Request.Context())
	view, err := h.service.Update(c.Request.Context(), a, c.Param(dashboardIDParam), in)
	respond(c, http.StatusOK, view, err)
}

func (h *Handler) delete(c *gin.Context) {
	a, _ := actor.From(c.Request.Context())
	entry, err := h.service.Delete(c.Request.Context(), a, c.Param(dashboardIDParam))
	respond(c, http.StatusOK, history.DeleteResult{RecycleBinEntry: entry.View()}, err)
}

// copy reads its body only when one is sent: the copy's name is optional.
func (h *Handler) copy(c *gin.Context) {
	var in CopyInput
	if c.Request.ContentLength != 0 {
		if err := web.BindJSON(c, &in); err != nil {
			web.RespondError(c, err)
			return
		}
	}
	a, _ := actor.From(c.Request.Context())
	view, err := h.service.Copy(c.Request.Context(), a, c.Param(dashboardIDParam), in)
	respond(c, http.StatusCreated, view, err)
}

func (h *Handler) share(c *gin.Context) {
	var in SharingInput
	if err := web.BindJSON(c, &in); err != nil {
		web.RespondError(c, err)
		return
	}
	a, _ := actor.From(c.Request.Context())
	view, err := h.service.Share(c.Request.Context(), a, c.Param(dashboardIDParam), in)
	respond(c, http.StatusOK, view, err)
}

func bindLimited(c *gin.Context, dst any) bool {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, MaxBodyBytes)
	if err := web.BindJSON(c, dst); err != nil {
		web.RespondError(c, err)
		return false
	}
	return true
}

func respond(c *gin.Context, status int, body any, err error) {
	if err != nil {
		web.RespondError(c, err)
		return
	}
	web.Respond(c, status, body)
}
