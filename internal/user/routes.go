// Registers the user module's routes. Most require a valid JWT; /interests is
// public so the app can show the catalogue during onboarding (before login).
package user

import (
	"github.com/gin-gonic/gin"

	"github.com/sebas1197-git/planetalk/internal/auth"
	"github.com/sebas1197-git/planetalk/internal/middleware"
)

func RegisterRoutes(r gin.IRouter, h *Handler, tokens *auth.TokenManager) {
	// Public reference data (still public, just versioned: /api/v1/interests).
	r.GET("/interests", h.ListInterests)

	// Everything below requires a valid Bearer token.
	authd := r.Group("/")
	authd.Use(middleware.RequireAuth(tokens))

	authd.GET("/me", h.GetMe)
	authd.PATCH("/me", h.UpdateMe)
	authd.PUT("/me/interests", h.SetMyInterests)

	authd.GET("/users/:id", h.GetUser)

	authd.GET("/friends", h.ListFriends)
	authd.GET("/friends/requests", h.ListFriendRequests)
	authd.POST("/friends/:id", h.SendFriendRequest)
	authd.POST("/friends/:id/accept", h.AcceptFriendRequest)
	authd.DELETE("/friends/:id", h.RemoveFriend)
}
