// The Hub is the heart of the realtime system. It keeps track of every open
// WebSocket connection on THIS server instance and delivers events to them.
//
// To work across MULTIPLE server instances, we use Redis pub/sub:
//   - SendToUser publishes the event to a Redis channel.
//   - Every instance subscribes to that channel, and delivers the event to any
//     of that user's connections it happens to hold locally.
//
// So an event created on instance A reaches the user even if they're connected
// to instance B. This is what lets the app scale beyond one server.
package realtime

import (
	"context"
	"encoding/json"
	"log"

	goredis "github.com/redis/go-redis/v9"
)

const pubsubChannel = "planetalk:ws"

// Event is the message shape clients receive, e.g. {"type":"match_found","data":{...}}.
type Event struct {
	Type string `json:"type"`
	Data any    `json:"data,omitempty"`
}

// envelope is what travels over Redis: which user + the already-encoded event.
type envelope struct {
	UserID  string          `json:"user_id"`
	Payload json.RawMessage `json:"payload"`
}

type Hub struct {
	rdb        *goredis.Client
	clients    map[string]map[*Client]bool // userID -> set of connections
	register   chan *Client
	unregister chan *Client
	deliver    chan envelope
}

func NewHub(rdb *goredis.Client) *Hub {
	return &Hub{
		rdb:        rdb,
		clients:    make(map[string]map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		deliver:    make(chan envelope, 256),
	}
}

// Run is the hub's single event loop. Because everything that touches the
// `clients` map happens here (one goroutine), we never need locks. Call it once
// in a goroutine from main.
func (h *Hub) Run(ctx context.Context) {
	go h.subscribe(ctx)

	for {
		select {
		case c := <-h.register:
			if h.clients[c.userID] == nil {
				h.clients[c.userID] = make(map[*Client]bool)
				_ = markOnline(ctx, h.rdb, c.userID) // first connection -> online
			}
			h.clients[c.userID][c] = true

		case c := <-h.unregister:
			if conns, ok := h.clients[c.userID]; ok {
				if _, ok := conns[c]; ok {
					delete(conns, c)
					close(c.send)
				}
				if len(conns) == 0 {
					delete(h.clients, c.userID)
					_ = markOffline(ctx, h.rdb, c.userID) // last connection -> offline
				}
			}

		case env := <-h.deliver:
			for c := range h.clients[env.UserID] {
				select {
				case c.send <- env.Payload:
				default:
					// Client's buffer is full (too slow) -> drop the connection.
					close(c.send)
					delete(h.clients[env.UserID], c)
				}
			}
		}
	}
}

// SendToUser publishes an event for one user to ALL instances via Redis.
// Other modules (matchmaking, chat) call this to push live updates.
func (h *Hub) SendToUser(ctx context.Context, userID string, evt Event) error {
	payload, err := json.Marshal(evt)
	if err != nil {
		return err
	}
	data, err := json.Marshal(envelope{UserID: userID, Payload: payload})
	if err != nil {
		return err
	}
	return h.rdb.Publish(ctx, pubsubChannel, data).Err()
}

// Redis returns the underlying client (handlers use it for presence lookups).
func (h *Hub) Redis() *goredis.Client { return h.rdb }

// subscribe listens to the Redis channel and forwards events into the hub loop.
func (h *Hub) subscribe(ctx context.Context) {
	sub := h.rdb.Subscribe(ctx, pubsubChannel)
	for msg := range sub.Channel() {
		var env envelope
		if err := json.Unmarshal([]byte(msg.Payload), &env); err != nil {
			log.Printf("realtime: bad pubsub payload: %v", err)
			continue
		}
		h.deliver <- env
	}
}
