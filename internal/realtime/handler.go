// HTTP/WebSocket handlers for the realtime module.
package realtime

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"github.com/sebas1197-git/planetalk/internal/auth"
	"github.com/sebas1197-git/planetalk/internal/httputil"
	"github.com/sebas1197-git/planetalk/internal/middleware"
)

type Handler struct {
	hub    *Hub
	tokens *auth.TokenManager
}

func NewHandler(hub *Hub, tokens *auth.TokenManager) *Handler {
	return &Handler{hub: hub, tokens: tokens}
}

// upgrader turns a normal HTTP request into a WebSocket connection.
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// Dev: allow any origin. Tighten this for production (check allowed hosts).
	CheckOrigin: func(r *http.Request) bool { return true },
}

// HandleWS authenticates, upgrades to WebSocket, and registers the connection.
// The JWT can come from ?token=... (handy for browsers/Postman) or the
// Authorization: Bearer header.
func (h *Handler) HandleWS(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		token = strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer ")
	}
	claims, err := h.tokens.Parse(token)
	if err != nil {
		httputil.Error(c, http.StatusUnauthorized, "invalid_token", "token is invalid or expired")
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return // Upgrade already wrote an error response
	}

	client := &Client{hub: h.hub, conn: conn, userID: claims.UserID, send: make(chan []byte, 256)}
	h.hub.register <- client
	go client.writePump()
	go client.readPump()
}

// GET /presence -> list of online user ids (handy for testing).
func (h *Handler) ListOnline(c *gin.Context) {
	ids, err := OnlineUsers(c.Request.Context(), h.hub.Redis())
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, "server_error", "could not read presence")
		return
	}
	httputil.OK(c, gin.H{"online": ids})
}

// GET /presence/:id -> { user_id, online }
func (h *Handler) GetPresence(c *gin.Context) {
	online, err := IsOnline(c.Request.Context(), h.hub.Redis(), c.Param("id"))
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, "server_error", "could not read presence")
		return
	}
	httputil.OK(c, gin.H{"user_id": c.Param("id"), "online": online})
}

// POST /realtime/echo -> sends a test event to YOUR own connections.
// Lets you verify the WebSocket pipeline end-to-end.
type echoBody struct {
	Message string `json:"message" binding:"required"`
}

func (h *Handler) Echo(c *gin.Context) {
	var body echoBody
	if err := c.ShouldBindJSON(&body); err != nil {
		httputil.Error(c, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	userID := c.GetString(middleware.ContextUserID)
	if err := h.hub.SendToUser(c.Request.Context(), userID, Event{Type: "echo", Data: gin.H{"message": body.Message}}); err != nil {
		httputil.Error(c, http.StatusInternalServerError, "server_error", "could not send event")
		return
	}
	httputil.OK(c, gin.H{"message": "sent"})
}
