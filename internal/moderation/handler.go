// HTTP handlers for the moderation module.
package moderation

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

type reportBody struct {
	ReportedID string  `json:"reported_id" binding:"required"`
	Reason     string  `json:"reason" binding:"required"`
	Context    *string `json:"context"`
}

// POST /reports
func (h *Handler) Report(c *gin.Context) {
	var body reportBody
	if err := c.ShouldBindJSON(&body); err != nil {
		httputil.Error(c, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	err := h.svc.Report(c.Request.Context(), currentUser(c), body.ReportedID, body.Reason, body.Context)
	if err != nil {
		writeErr(c, err)
		return
	}
	httputil.OK(c, gin.H{"message": "report submitted"})
}

// POST /blocks/:id
func (h *Handler) Block(c *gin.Context) {
	if err := h.svc.Block(c.Request.Context(), currentUser(c), c.Param("id")); err != nil {
		writeErr(c, err)
		return
	}
	httputil.OK(c, gin.H{"message": "user blocked"})
}

// DELETE /blocks/:id
func (h *Handler) Unblock(c *gin.Context) {
	if err := h.svc.Unblock(c.Request.Context(), currentUser(c), c.Param("id")); err != nil {
		httputil.Error(c, http.StatusInternalServerError, "server_error", "could not unblock")
		return
	}
	httputil.NoContent(c)
}

// GET /blocks
func (h *Handler) ListBlocked(c *gin.Context) {
	blocked, err := h.svc.ListBlocked(c.Request.Context(), currentUser(c))
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, "server_error", "could not load blocks")
		return
	}
	httputil.OK(c, blocked)
}

func writeErr(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrSelf):
		httputil.Error(c, http.StatusBadRequest, "cannot_target_self", err.Error())
	case errors.Is(err, ErrReasonEmpty):
		httputil.Error(c, http.StatusBadRequest, "reason_required", err.Error())
	case errors.Is(err, ErrUserNotFound):
		httputil.Error(c, http.StatusNotFound, "user_not_found", "that user does not exist")
	default:
		httputil.Error(c, http.StatusInternalServerError, "server_error", "something went wrong")
	}
}
