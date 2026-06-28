// HTTP handlers for the user module. They read input, call the service, and
// write responses via httputil (consistent shape). The caller's id comes from
// the auth middleware (the JWT).
package user

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/sebas1197-git/planetalk/internal/httputil"
	"github.com/sebas1197-git/planetalk/internal/middleware"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// currentUser reads the logged-in user's id that RequireAuth stored.
func currentUser(c *gin.Context) string {
	return c.GetString(middleware.ContextUserID)
}

// ---- DTOs ----

type updateProfileBody struct {
	DisplayName *string `json:"display_name"`
	Bio         *string `json:"bio"`
	AvatarURL   *string `json:"avatar_url"`
	Gender      *string `json:"gender"`
	Birthdate   *string `json:"birthdate"` // "YYYY-MM-DD"
	Country     *string `json:"country"`
	Language    *string `json:"language"`
}

type setInterestsBody struct {
	InterestIDs []int `json:"interest_ids"`
}

// ---- Profile handlers ----

// GET /me
func (h *Handler) GetMe(c *gin.Context) {
	profile, err := h.svc.GetProfile(c.Request.Context(), currentUser(c))
	if err != nil {
		writeProfileErr(c, err)
		return
	}
	httputil.OK(c, profile)
}

// GET /users/:id
func (h *Handler) GetUser(c *gin.Context) {
	profile, err := h.svc.GetProfile(c.Request.Context(), c.Param("id"))
	if err != nil {
		writeProfileErr(c, err)
		return
	}
	httputil.OK(c, profile)
}

// PATCH /me
func (h *Handler) UpdateMe(c *gin.Context) {
	var body updateProfileBody
	if err := c.ShouldBindJSON(&body); err != nil {
		httputil.Error(c, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	profile, err := h.svc.UpdateProfile(c.Request.Context(), currentUser(c), UpdateProfileInput(body))
	if err != nil {
		writeProfileErr(c, err)
		return
	}
	httputil.OK(c, profile)
}

// ---- Interest handlers ----

// GET /interests (public)
func (h *Handler) ListInterests(c *gin.Context) {
	items, err := h.svc.ListInterests(c.Request.Context())
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, "server_error", "could not load interests")
		return
	}
	httputil.OK(c, items)
}

// PUT /me/interests
func (h *Handler) SetMyInterests(c *gin.Context) {
	var body setInterestsBody
	if err := c.ShouldBindJSON(&body); err != nil {
		httputil.Error(c, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	items, err := h.svc.SetMyInterests(c.Request.Context(), currentUser(c), body.InterestIDs)
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, "server_error", "could not save interests")
		return
	}
	httputil.OK(c, items)
}

// ---- Friend handlers ----

// POST /friends/:id
func (h *Handler) SendFriendRequest(c *gin.Context) {
	err := h.svc.SendFriendRequest(c.Request.Context(), currentUser(c), c.Param("id"))
	if errors.Is(err, ErrCannotFriendSelf) {
		httputil.Error(c, http.StatusBadRequest, "cannot_friend_self", "you cannot add yourself")
		return
	}
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, "server_error", "could not send request")
		return
	}
	httputil.OK(c, gin.H{"message": "request sent"})
}

// POST /friends/:id/accept
func (h *Handler) AcceptFriendRequest(c *gin.Context) {
	err := h.svc.AcceptFriendRequest(c.Request.Context(), currentUser(c), c.Param("id"))
	if errors.Is(err, ErrNotFound) {
		httputil.Error(c, http.StatusNotFound, "no_pending_request", "no pending request from this user")
		return
	}
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, "server_error", "could not accept request")
		return
	}
	httputil.OK(c, gin.H{"message": "friend added"})
}

// GET /friends
func (h *Handler) ListFriends(c *gin.Context) {
	friends, err := h.svc.ListFriends(c.Request.Context(), currentUser(c))
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, "server_error", "could not load friends")
		return
	}
	httputil.OK(c, friends)
}

// GET /friends/requests
func (h *Handler) ListFriendRequests(c *gin.Context) {
	reqs, err := h.svc.ListFriendRequests(c.Request.Context(), currentUser(c))
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, "server_error", "could not load requests")
		return
	}
	httputil.OK(c, reqs)
}

// DELETE /friends/:id
func (h *Handler) RemoveFriend(c *gin.Context) {
	if err := h.svc.RemoveFriend(c.Request.Context(), currentUser(c), c.Param("id")); err != nil {
		httputil.Error(c, http.StatusInternalServerError, "server_error", "could not remove friend")
		return
	}
	httputil.NoContent(c)
}

// writeProfileErr turns a "not found" into 404, anything else into 500.
func writeProfileErr(c *gin.Context, err error) {
	if errors.Is(err, ErrNotFound) {
		httputil.Error(c, http.StatusNotFound, "not_found", "user not found")
		return
	}
	httputil.Error(c, http.StatusInternalServerError, "server_error", "something went wrong")
}
