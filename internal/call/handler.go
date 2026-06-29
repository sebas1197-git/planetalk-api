// HTTP handler for the call module.
package call

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/sebas1197-git/planetalk/internal/httputil"
	"github.com/sebas1197-git/planetalk/internal/match"
	"github.com/sebas1197-git/planetalk/internal/middleware"
	"github.com/sebas1197-git/planetalk/pkg/agora"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// POST /matches/:id/token
func (h *Handler) GetToken(c *gin.Context) {
	userID := c.GetString(middleware.ContextUserID)
	res, err := h.svc.Token(c.Request.Context(), userID, c.Param("id"))
	if err != nil {
		switch {
		case errors.Is(err, match.ErrNotFound):
			httputil.Error(c, http.StatusNotFound, "not_found", "match not found")
		case errors.Is(err, match.ErrForbidden):
			httputil.Error(c, http.StatusForbidden, "forbidden", "you are not part of this match")
		case errors.Is(err, agora.ErrNotConfigured):
			httputil.Error(c, http.StatusServiceUnavailable, "agora_not_configured",
				"video is not configured on this server (set AGORA_APP_ID and AGORA_APP_CERTIFICATE)")
		default:
			httputil.Error(c, http.StatusInternalServerError, "server_error", "could not create token")
		}
		return
	}
	httputil.OK(c, res)
}
