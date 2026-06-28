// This file registers the auth routes on the Gin router.
// main.go will call auth.RegisterRoutes(...) to plug these in.
package auth

import "github.com/gin-gonic/gin"

// RegisterRoutes attaches the /auth/* endpoints under the given router group
// (e.g. /api/v1), so the full paths become /api/v1/auth/...
func RegisterRoutes(r gin.IRouter, h *Handler) {
	grp := r.Group("/auth")
	grp.POST("/otp/request", h.RequestOTP)
	grp.POST("/otp/verify", h.VerifyOTP)
	grp.POST("/refresh", h.Refresh)
}
