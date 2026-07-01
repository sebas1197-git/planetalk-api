// Registers the moderation routes (require a valid Bearer token).
package moderation

import (
	"github.com/gin-gonic/gin"

	"github.com/sebas1197-git/planetalk/internal/auth"
	"github.com/sebas1197-git/planetalk/internal/middleware"
)

func RegisterRoutes(r gin.IRouter, h *Handler, tokens *auth.TokenManager) {
	authd := r.Group("/")
	authd.Use(middleware.RequireAuth(tokens))

	authd.POST("/reports", h.Report)
	authd.GET("/blocks", h.ListBlocked)
	authd.POST("/blocks/:id", h.Block)
	authd.DELETE("/blocks/:id", h.Unblock)
}
