// Registers the feed routes (require a valid Bearer token).
package feed

import (
	"github.com/gin-gonic/gin"

	"github.com/sebas1197-git/planetalk/internal/auth"
	"github.com/sebas1197-git/planetalk/internal/middleware"
)

func RegisterRoutes(r gin.IRouter, h *Handler, tokens *auth.TokenManager) {
	authd := r.Group("/")
	authd.Use(middleware.RequireAuth(tokens))

	authd.POST("/posts", h.Create)
	authd.DELETE("/posts/:id", h.Delete)
	authd.GET("/me/posts", h.ListMine)
	authd.GET("/users/:id/posts", h.ListByUser)
}
