// HTTP handlers for the feed module.
package feed

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/sebas1197-git/planetalk/internal/httputil"
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

type createBody struct {
	MediaURL   *string `json:"media_url"`
	MediaType  *string `json:"media_type"`
	Caption    *string `json:"caption"`
	Visibility string  `json:"visibility"` // public (default) | friends | private
}

// POST /posts
func (h *Handler) Create(c *gin.Context) {
	var body createBody
	if err := c.ShouldBindJSON(&body); err != nil {
		httputil.Error(c, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	if body.MediaURL == nil && body.Caption == nil {
		httputil.Error(c, http.StatusBadRequest, "empty_post", "a post needs media_url or caption")
		return
	}
	post, err := h.svc.Create(c.Request.Context(), currentUser(c), CreateInput(body))
	if err != nil {
		if errors.Is(err, ErrInvalidVisibility) {
			httputil.Error(c, http.StatusBadRequest, "invalid_visibility", err.Error())
			return
		}
		httputil.Error(c, http.StatusInternalServerError, "server_error", "could not create post")
		return
	}
	httputil.Created(c, post)
}

// DELETE /posts/:id
func (h *Handler) Delete(c *gin.Context) {
	err := h.svc.Delete(c.Request.Context(), currentUser(c), c.Param("id"))
	if errors.Is(err, ErrNotFound) {
		httputil.Error(c, http.StatusNotFound, "not_found", "post not found")
		return
	}
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, "server_error", "could not delete post")
		return
	}
	httputil.NoContent(c)
}

// GET /me/posts — all of my own posts.
func (h *Handler) ListMine(c *gin.Context) {
	me := currentUser(c)
	h.list(c, me, me)
}

// GET /users/:id/posts — another user's posts, filtered by visibility.
func (h *Handler) ListByUser(c *gin.Context) {
	h.list(c, currentUser(c), c.Param("id"))
}

// list is the shared pagination + response logic.
func (h *Handler) list(c *gin.Context, viewerID, targetID string) {
	page := atoiDefault(c.Query("page"), 1)
	perPage := atoiDefault(c.Query("per_page"), 20)
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}
	offset := (page - 1) * perPage

	posts, total, err := h.svc.ListUserPosts(c.Request.Context(), viewerID, targetID, perPage, offset)
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, "server_error", "could not load posts")
		return
	}
	httputil.List(c, posts, httputil.Meta{Total: total, Page: page, PerPage: perPage})
}

func atoiDefault(s string, def int) int {
	if n, err := strconv.Atoi(s); err == nil {
		return n
	}
	return def
}
