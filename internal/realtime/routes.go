// Registers the realtime routes.
package realtime

import (
	"github.com/gin-gonic/gin"

	"github.com/sebas1197-git/planetalk/internal/auth"
	"github.com/sebas1197-git/planetalk/internal/middleware"
)

func RegisterRoutes(r gin.IRouter, h *Handler, tokens *auth.TokenManager) {
	// WebSocket endpoint. Auth is handled inside HandleWS (token via ?token= or
	// Authorization header), so it is NOT behind the usual middleware.
	r.GET("/ws", h.HandleWS)

	// Presence + test helpers require a valid Bearer token.
	authd := r.Group("/")
	authd.Use(middleware.RequireAuth(tokens))
	authd.GET("/presence", h.ListOnline)
	authd.GET("/presence/:id", h.GetPresence)
	authd.POST("/realtime/echo", h.Echo)
}
