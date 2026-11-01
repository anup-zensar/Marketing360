package widgetdata

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"mapp/internal/platform/actor"
	"mapp/internal/platform/authz"
	"mapp/internal/platform/web"
)

type Handler struct {
	service *Service
}

func NewHandler(s *Service) *Handler {
	return &Handler{service: s}
}

func (h *Handler) RegisterRoutes(r gin.IRoutes) {
	r.GET("/data-sources", h.catalogue)
	r.POST("/data-sources/query", h.query)
}

type sourceList struct {
	Items []sourceView `json:"items"`
}

type sourceView struct {
	Key         string           `json:"key"`
	Label       string           `json:"label"`
	Description string           `json:"description"`
	Kind        string           `json:"kind"`
	Fields      []fieldView      `json:"fields"`
	Filters     filtersView      `json:"filters"`
	Owner       *authz.PersonRef `json:"owner"`
}

type fieldView struct {
	Key          string    `json:"key"`
	Label        string    `json:"label"`
	Kind         FieldKind `json:"kind"`
	Multi        bool      `json:"multi"`
	Groupable    bool      `json:"groupable"`
	Aggregations []string  `json:"aggregations"`
	Finance      bool      `json:"finance"`
}

type filtersView struct {
	Date *string `json:"date"`
	Plan bool    `json:"plan"`
	Team bool    `json:"team"`
}

func (h *Handler) catalogue(c *gin.Context) {
	a, _ := actor.From(c.Request.Context())
	sources, err := h.service.Catalogue(c.Request.Context(), a)
	if err != nil {
		web.RespondError(c, err)
		return
	}
	out := sourceList{Items: make([]sourceView, 0, len(sources))}
	for _, s := range sources {
		out.Items = append(out.Items, viewOf(s))
	}
	web.Respond(c, http.StatusOK, out)
}

func (h *Handler) query(c *gin.Context) {
	var r Request
	if err := web.BindJSON(c, &r); err != nil {
		web.RespondError(c, err)
		return
	}
	a, _ := actor.From(c.Request.Context())
	data, err := h.service.Query(c.Request.Context(), a, r)
	if err != nil {
		web.RespondError(c, err)
		return
	}
	web.Respond(c, http.StatusOK, data)
}

func viewOf(s Source) sourceView {
	fields := make([]fieldView, 0, len(s.Fields))
	for _, f := range s.Fields {
		fields = append(fields, fieldView{Key: f.Key, Label: f.Label, Kind: f.Kind, Multi: f.Multi, Groupable: IsGroupable(f),
			Aggregations: AggregationsOf(f), Finance: f.Finance})
	}
	return sourceView{Key: s.Key, Label: s.Label, Description: s.Description, Kind: s.Kind, Fields: fields,
		Filters: filtersView{Date: bucketOrNil(s.Filters.Date), Plan: s.Filters.Plan, Team: s.Filters.Team}, Owner: s.Owner}
}
