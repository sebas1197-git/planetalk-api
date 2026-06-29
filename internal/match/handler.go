// HTTP handlers for the match module.
package match

import (
	"errors"
	"net/http"

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

// POST /match/enter
func (h *Handler) Enter(c *gin.Context) {
	res, err := h.svc.Enter(c.Request.Context(), currentUser(c))
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, "server_error", "could not enter matchmaking")
		return
	}
	httputil.OK(c, res)
}

// POST /match/leave
func (h *Handler) Leave(c *gin.Context) {
	if err := h.svc.Leave(c.Request.Context(), currentUser(c)); err != nil {
		httputil.Error(c, http.StatusInternalServerError, "server_error", "could not leave matchmaking")
		return
	}
	httputil.OK(c, gin.H{"message": "left queue"})
}

// GET /matches/:id
func (h *Handler) GetMatch(c *gin.Context) {
	m, err := h.svc.Get(c.Request.Context(), currentUser(c), c.Param("id"))
	if err != nil {
		writeMatchErr(c, err)
		return
	}
	httputil.OK(c, m)
}

// POST /matches/:id/end
func (h *Handler) EndMatch(c *gin.Context) {
	if err := h.svc.End(c.Request.Context(), currentUser(c), c.Param("id")); err != nil {
		writeMatchErr(c, err)
		return
	}
	httputil.OK(c, gin.H{"message": "match ended"})
}

func writeMatchErr(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrNotFound):
		httputil.Error(c, http.StatusNotFound, "not_found", "match not found")
	case errors.Is(err, ErrForbidden):
		httputil.Error(c, http.StatusForbidden, "forbidden", "you are not part of this match")
	default:
		httputil.Error(c, http.StatusInternalServerError, "server_error", "something went wrong")
	}
}
