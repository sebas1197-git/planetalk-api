// This file holds the HTTP handlers: they read the JSON request, call the
// service, and write the response via httputil (consistent shape).
package auth

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/sebas1197-git/planetalk/internal/httputil"
)

// Handler wraps the service so Gin handlers can call it.
type Handler struct {
	svc     *Service
	devMode bool // when true, OTP request responses include the code (DEV ONLY)
}

func NewHandler(svc *Service, devMode bool) *Handler {
	return &Handler{svc: svc, devMode: devMode}
}

// ---- Request body shapes (DTOs) ----------------------------------------
// `binding:"required"` makes Gin reject the request if the field is missing.

type requestOTPBody struct {
	Phone string `json:"phone" binding:"required"`
}

type verifyOTPBody struct {
	Phone string `json:"phone" binding:"required"`
	Code  string `json:"code" binding:"required"`
}

type refreshBody struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// POST /auth/otp/request
func (h *Handler) RequestOTP(c *gin.Context) {
	var body requestOTPBody
	if err := c.ShouldBindJSON(&body); err != nil {
		httputil.Error(c, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	code, err := h.svc.RequestOTP(c.Request.Context(), body.Phone)
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, "send_failed", "could not send code")
		return
	}
	resp := gin.H{"message": "code sent"}
	if h.devMode {
		// DEV ONLY: expose the code so test tools can auto-fill it.
		// This branch is never taken in production (APP_ENV != development).
		resp["dev_code"] = code
	}
	httputil.OK(c, resp)
}

// POST /auth/otp/verify
func (h *Handler) VerifyOTP(c *gin.Context) {
	var body verifyOTPBody
	if err := c.ShouldBindJSON(&body); err != nil {
		httputil.Error(c, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	access, refresh, err := h.svc.VerifyOTP(c.Request.Context(), body.Phone, body.Code)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidCode):
			httputil.Error(c, http.StatusUnauthorized, "invalid_code", "the code is incorrect")
		case errors.Is(err, ErrNoValidCode):
			httputil.Error(c, http.StatusUnauthorized, "no_valid_code", "no valid code; request a new one")
		case errors.Is(err, ErrTooManyAttempts):
			httputil.Error(c, http.StatusTooManyRequests, "too_many_attempts", "too many attempts; request a new code")
		default:
			httputil.Error(c, http.StatusInternalServerError, "server_error", "something went wrong")
		}
		return
	}
	httputil.OK(c, gin.H{"access_token": access, "refresh_token": refresh})
}

// POST /auth/refresh
func (h *Handler) Refresh(c *gin.Context) {
	var body refreshBody
	if err := c.ShouldBindJSON(&body); err != nil {
		httputil.Error(c, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	access, newRefresh, err := h.svc.RefreshTokens(c.Request.Context(), body.RefreshToken)
	if err != nil {
		httputil.Error(c, http.StatusUnauthorized, "invalid_refresh_token", "refresh token is invalid or expired")
		return
	}
	httputil.OK(c, gin.H{"access_token": access, "refresh_token": newRefresh})
}

// POST /auth/logout
func (h *Handler) Logout(c *gin.Context) {
	var body refreshBody
	if err := c.ShouldBindJSON(&body); err != nil {
		httputil.Error(c, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	if err := h.svc.Logout(c.Request.Context(), body.RefreshToken); err != nil {
		httputil.Error(c, http.StatusInternalServerError, "server_error", "could not log out")
		return
	}
	httputil.OK(c, gin.H{"message": "logged out"})
}
