// Registers the chat routes (require a valid Bearer token).
package chat

import (
	"github.com/gin-gonic/gin"

	"github.com/sebas1197-git/planetalk/internal/auth"
	"github.com/sebas1197-git/planetalk/internal/middleware"
)

func RegisterRoutes(r gin.IRouter, h *Handler, tokens *auth.TokenManager) {
	authd := r.Group("/")
	authd.Use(middleware.RequireAuth(tokens))
	authd.POST("/matches/:id/messages", h.Send)
	authd.GET("/matches/:id/messages", h.List)
}
