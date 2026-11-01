package performance

import (
	"encoding/base64"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"mapp/internal/platform/actor"
	"mapp/internal/platform/errs"
	"mapp/internal/platform/web"
)

// MaxUploadBytes bounds a performance file like the GSAP rate uploads: 2 MB of file, base64-encoded.
const MaxUploadBytes = 3 << 20

const fieldContent = "content"

type uploadBody struct {
	FileName string `json:"fileName"`
	Content  string `json:"content"`
}

func (h *Handler) registerImportRoutes(r gin.IRoutes) {
	r.POST("/performance/sources/:sourceId/uploads", h.upload)
	r.GET("/performance/imports", h.imports)
	r.GET("/performance/imports/:importId", h.importByID)
	r.POST("/performance/imports/:importId/undo", h.undoImport)
}

func (h *Handler) upload(c *gin.Context) {
	var body uploadBody
	if err := web.BindJSON(c, &body); err != nil {
		web.RespondError(c, err)
		return
	}
	content, err := base64.StdEncoding.DecodeString(body.Content)
	if err != nil {
		web.RespondError(c, errs.Validation(errs.FieldError{Field: fieldContent, Message: msgFileUnreadable}))
		return
	}
	a, _ := actor.From(c.Request.Context())
	sourceID := c.Param("sourceId")
	mappings, err := h.service.Mappings(c.Request.Context(), a, sourceID)
	if err != nil {
		web.RespondError(c, err)
		return
	}
	file, err := ReadFile(content, mappings)
	if err != nil {
		web.RespondError(c, errs.Validation(errs.FieldError{Field: fieldContent, Message: msgFileUnreadable}))
		return
	}
	run, err := h.service.RunImport(c.Request.Context(), a, ImportInput{
		SourceID: sourceID, FileName: body.FileName, Origin: OriginUpload,
		Headings: file.Headings, Rows: file.Rows,
	})
	if err != nil {
		web.RespondError(c, err)
		return
	}
	view, err := h.service.ImportByID(c.Request.Context(), a, run.ID)
	h.answerImport(c, http.StatusCreated, view, err)
}

func (h *Handler) imports(c *gin.Context) {
	a, _ := actor.From(c.Request.Context())
	limit, _ := strconv.Atoi(c.Query("limit"))
	found, err := h.service.Imports(c.Request.Context(), a, c.Query("sourceId"), limit)
	if err != nil {
		web.RespondError(c, err)
		return
	}
	web.Respond(c, http.StatusOK, ImportList{Items: found})
}

func (h *Handler) importByID(c *gin.Context) {
	a, _ := actor.From(c.Request.Context())
	view, err := h.service.ImportByID(c.Request.Context(), a, c.Param("importId"))
	h.answerImport(c, http.StatusOK, view, err)
}

func (h *Handler) undoImport(c *gin.Context) {
	a, _ := actor.From(c.Request.Context())
	undone, err := h.service.UndoImport(c.Request.Context(), a, c.Param("importId"))
	if err != nil {
		web.RespondError(c, err)
		return
	}
	view, err := h.service.ImportByID(c.Request.Context(), a, undone.ID)
	h.answerImport(c, http.StatusOK, view, err)
}

func (h *Handler) answerImport(c *gin.Context, status int, view ImportView, err error) {
	if err != nil {
		web.RespondError(c, err)
		return
	}
	web.Respond(c, status, view)
}

type ImportList struct {
	Items []ImportView `json:"items"`
}
