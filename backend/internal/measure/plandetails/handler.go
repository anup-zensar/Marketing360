package plandetails

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"mapp/internal/measure"
	"mapp/internal/platform/actor"
	"mapp/internal/platform/web"
)

type Handler struct {
	service *Service
}

func NewHandler(s *Service) *Handler {
	return &Handler{service: s}
}

func (h *Handler) RegisterRoutes(r gin.IRoutes) {
	r.GET("/reports/plan-details", h.report)
}

func (h *Handler) report(c *gin.Context) {
	f, err := measure.ParseFilter(c.Request.URL.Query())
	if err != nil {
		web.RespondError(c, err)
		return
	}
	a, _ := actor.From(c.Request.Context())
	r, err := h.service.Report(c.Request.Context(), a, f)
	if err != nil {
		web.RespondError(c, err)
		return
	}
	web.Respond(c, http.StatusOK, r)
}
