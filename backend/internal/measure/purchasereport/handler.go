package purchasereport

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
	r.GET("/reports/purchase-requests", h.get)
}

func (h *Handler) get(c *gin.Context) {
	f, err := measure.ParseFilter(c.Request.URL.Query())
	if err != nil {
		web.RespondError(c, err)
		return
	}
	q, err := ParseQuery(f.Own)
	if err != nil {
		web.RespondError(c, err)
		return
	}
	a, _ := actor.From(c.Request.Context())
	view, err := h.service.Report(c.Request.Context(), a, f, q)
	if err != nil {
		web.RespondError(c, err)
		return
	}
	web.Respond(c, http.StatusOK, view)
}
