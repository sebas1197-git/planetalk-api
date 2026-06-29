// Registers the call route (requires a valid Bearer token).
package call

import (
	"github.com/gin-gonic/gin"

	"github.com/sebas1197-git/planetalk/internal/auth"
	"github.com/sebas1197-git/planetalk/internal/middleware"
)

func RegisterRoutes(r gin.IRouter, h *Handler, tokens *auth.TokenManager) {
	authd := r.Group("/")
	authd.Use(middleware.RequireAuth(tokens))
	authd.POST("/matches/:id/token", h.GetToken)
}
