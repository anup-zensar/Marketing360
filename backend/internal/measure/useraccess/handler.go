package useraccess

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"mapp/internal/platform/actor"
	"mapp/internal/platform/web"
)

type Handler struct {
	service *Service
}

func NewHandler(s *Service) *Handler {
	return &Handler{service: s}
}

// RegisterRoutes mounts the one read of the report; 039's route rule asks for access.manage (FR-002, FR-004).
func (h *Handler) RegisterRoutes(r gin.IRoutes) {
	r.GET("/reports/user-access", h.report)
}

func (h *Handler) report(c *gin.Context) {
	q, err := ParseQuery(c.Request.URL.Query())
	if err != nil {
		web.RespondError(c, err)
		return
	}
	a, _ := actor.From(c.Request.Context())
	r, err := h.service.Report(c.Request.Context(), a, q)
	if err != nil {
		web.RespondError(c, err)
		return
	}
	web.Respond(c, http.StatusOK, r)
}
