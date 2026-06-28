// Package middleware holds reusable Gin middleware.
//
// This file protects routes: it checks the "Authorization: Bearer <token>"
// header on each request, verifies the JWT, and stores the user id in the Gin
// context so handlers know who is calling. If the token is missing/invalid, it
// rejects the request with 401 before the handler ever runs.
package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/sebas1197-git/planetalk/internal/auth"
	"github.com/sebas1197-git/planetalk/internal/httputil"
)

// ContextUserID is the key under which we store the caller's user id.
const ContextUserID = "user_id"

// RequireAuth returns middleware that only lets authenticated requests through.
func RequireAuth(tokens *auth.TokenManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		authz := c.GetHeader("Authorization")
		if !strings.HasPrefix(authz, "Bearer ") {
			httputil.Abort(c, http.StatusUnauthorized, "missing_token", "missing bearer token")
			return
		}
		tokenStr := strings.TrimPrefix(authz, "Bearer ")

		claims, err := tokens.Parse(tokenStr)
		if err != nil {
			httputil.Abort(c, http.StatusUnauthorized, "invalid_token", "token is invalid or expired")
			return
		}

		c.Set(ContextUserID, claims.UserID) // hand the id to the handler
		c.Next()                            // continue to the real handler
	}
}
