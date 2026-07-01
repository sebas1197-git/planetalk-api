// Registers the match routes (all require a valid Bearer token).
//
// Note we use two prefixes to avoid a Gin static-vs-param route conflict:
//
//	/match/...    -> queue actions (enter, leave)
//	/matches/:id  -> a specific match resource (get, end)
package match

import (
	"github.com/gin-gonic/gin"

	"github.com/sebas1197-git/planetalk/internal/auth"
	"github.com/sebas1197-git/planetalk/internal/middleware"
)

func RegisterRoutes(r gin.IRouter, h *Handler, tokens *auth.TokenManager) {
	authd := r.Group("/")
	authd.Use(middleware.RequireAuth(tokens))

	authd.POST("/match/enter", h.Enter)
	authd.POST("/match/leave", h.Leave)
	authd.GET("/matches/:id", h.GetMatch)
	authd.POST("/matches/:id/end", h.EndMatch)
}
