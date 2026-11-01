package measure

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

func (h *Handler) RegisterRoutes(r gin.IRoutes) {
	r.GET("/reports", h.catalogue)
	r.GET("/reports/frame", h.frame)
}

func (h *Handler) catalogue(c *gin.Context) {
	web.Respond(c, http.StatusOK, h.service.Catalogue(c.Request.Context()))
}

func (h *Handler) frame(c *gin.Context) {
	values := c.Request.URL.Query()
	f, err := ParseFilter(values)
	if err != nil {
		web.RespondError(c, err)
		return
	}
	a, _ := actor.From(c.Request.Context())
	frame, err := h.service.Frame(c.Request.Context(), a, values.Get("report"), f)
	if err != nil {
		web.RespondError(c, err)
		return
	}
	web.Respond(c, http.StatusOK, frame)
}
