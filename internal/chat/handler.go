// HTTP handlers for the chat module.
package chat

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/sebas1197-git/planetalk/internal/httputil"
	"github.com/sebas1197-git/planetalk/internal/match"
	"github.com/sebas1197-git/planetalk/internal/middleware"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func currentUser(c *gin.Context) string {
	return c.GetString(middleware.ContextUserID)
}

type sendBody struct {
	Body string `json:"body" binding:"required"`
}

// POST /matches/:id/messages
func (h *Handler) Send(c *gin.Context) {
	var body sendBody
	if err := c.ShouldBindJSON(&body); err != nil {
		httputil.Error(c, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	msg, err := h.svc.Send(c.Request.Context(), currentUser(c), c.Param("id"), body.Body)
	if err != nil {
		writeErr(c, err)
		return
	}
	httputil.Created(c, msg)
}

// GET /matches/:id/messages?page=1&per_page=20
func (h *Handler) List(c *gin.Context) {
	page := atoiDefault(c.Query("page"), 1)
	perPage := atoiDefault(c.Query("per_page"), 20)
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20 // cap page size
	}
	offset := (page - 1) * perPage

	msgs, total, err := h.svc.List(c.Request.Context(), currentUser(c), c.Param("id"), perPage, offset)
	if err != nil {
		writeErr(c, err)
		return
	}
	httputil.List(c, msgs, httputil.Meta{Total: total, Page: page, PerPage: perPage})
}

func writeErr(c *gin.Context, err error) {
	switch {
	case errors.Is(err, match.ErrNotFound):
		httputil.Error(c, http.StatusNotFound, "not_found", "match not found")
	case errors.Is(err, match.ErrForbidden):
		httputil.Error(c, http.StatusForbidden, "forbidden", "you are not part of this match")
	case errors.Is(err, ErrBlocked):
		httputil.Error(c, http.StatusForbidden, "blocked", "you cannot message this user")
	default:
		httputil.Error(c, http.StatusInternalServerError, "server_error", "something went wrong")
	}
}

func atoiDefault(s string, def int) int {
	if n, err := strconv.Atoi(s); err == nil {
		return n
	}
	return def
}
