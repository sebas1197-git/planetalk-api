// A Client is one open WebSocket connection (one device of one user).
//
// Each client runs two goroutines:
//   - readPump:  reads from the socket (mostly to detect disconnects + pongs)
//   - writePump: writes queued messages to the socket + sends periodic pings
//
// The `send` channel is how the Hub hands messages to this connection.
package realtime

import (
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second    // max time to write one message
	pongWait       = 60 * time.Second    // if no pong in this window, drop the conn
	pingPeriod     = (pongWait * 9) / 10 // send a ping a bit before pongWait
	maxMessageSize = 4096                // limit inbound message size
)

type Client struct {
	hub    *Hub
	conn   *websocket.Conn
	userID string
	send   chan []byte
}

// readPump drains inbound messages. We don't act on client->server messages yet
// (clients mostly receive), but reading is required to detect disconnects and to
// process pong replies that keep the connection alive.
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		if _, _, err := c.conn.ReadMessage(); err != nil {
			break // client disconnected or timed out
		}
	}
}

// writePump sends queued messages and heartbeat pings.
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case msg, ok := <-c.send:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// Hub closed the channel -> tell the client we're done.
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				return
			}
		case <-ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
