package main

import (
	"log/slog"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/raaz/server/store"
)

const (
	sendBufferSize = 256
	recvBufferSize = 256

	writeWait  = 10 * time.Second
	pongWait   = 60 * time.Second
	pingPeriod = (pongWait * 9) / 10

	maxMessageSize = 64 * 1024 // 64 KiB
)

// Client represents a single active WebSocket connection.
type Client struct {
	conn      *websocket.Conn
	send      chan []byte // writePump drains this; never closed explicitly
	recv      chan []byte // readPump fills this; closed when readPump exits
	alias     string
	connToken string // server-generated; required on /vault/messages POST
	params    RegistrationParams
	hub       *Hub

	closeRecvOnce sync.Once
}

// safeSend writes data to c.send without panicking if the channel is closed,
// and drops the message if the buffer is full.
func (c *Client) safeSend(data []byte) {
	defer func() { recover() }() //nolint:errcheck
	select {
	case c.send <- data:
	default:
		droppedMsgsTotal.Inc()
		slog.Warn("send buffer full, dropping message", "alias", c.alias)
	}
}

// writePump is the sole goroutine that writes to the WebSocket connection.
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case msg := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// readPump is the sole goroutine that reads from the WebSocket connection.
// On exit it deregisters the client from the hub and queue, closes c.recv,
// and closes the connection.
func (c *Client) readPump(q store.QueueStore) {
	defer func() {
		c.hub.Unregister(c)
		q.Remove(c.params.AnonymousID) //nolint:errcheck
		c.closeRecvOnce.Do(func() { close(c.recv) })
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, msg, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				slog.Warn("read error", "anonymousID", c.params.AnonymousID, "err", err)
			}
			return
		}
		select {
		case c.recv <- msg:
		default:
			slog.Warn("recv buffer full, dropping inbound message", "anonymousID", c.params.AnonymousID)
		}
	}
}

// Hub tracks all live connections. byID enables fast lookup by anonymousID
// so the matching loop can resolve queued IDs back to live *Client pointers.
type Hub struct {
	mu      sync.RWMutex
	clients map[*Client]struct{}
	byID    map[string]*Client
}

func NewHub() *Hub {
	return &Hub{
		clients: make(map[*Client]struct{}),
		byID:    make(map[string]*Client),
	}
}

// Register adds c to the hub. If another client is already registered with
// the same anonymousID, that old client is kicked first (M-5).
func (h *Hub) Register(c *Client) {
	h.mu.Lock()
	if old, ok := h.byID[c.params.AnonymousID]; ok && old != c {
		delete(h.clients, old)
		// Signal the old connection to close asynchronously; don't block.
		go func() {
			data, _ := marshalEnvelope(EventDisconnect, DisconnectPayload{Reason: "replaced_by_new_connection"})
			old.safeSend(data)
			old.conn.Close()
		}()
		slog.Warn("duplicate anonymousID kicked old connection", "anonymousID", c.params.AnonymousID)
	}
	h.clients[c] = struct{}{}
	h.byID[c.params.AnonymousID] = c
	h.mu.Unlock()
	wsConnectionsActive.Inc()
	slog.Info("client connected", "anonymousID", c.params.AnonymousID)
}

// Unregister removes the client from the hub only if it's still the current
// owner of that anonymousID slot (pointer-identity check prevents the new
// client from being evicted when the old connection's readPump eventually exits).
func (h *Hub) Unregister(c *Client) {
	h.mu.Lock()
	_, exists := h.clients[c]
	if exists {
		delete(h.clients, c)
		if h.byID[c.params.AnonymousID] == c {
			delete(h.byID, c.params.AnonymousID)
		}
	}
	h.mu.Unlock()
	if exists {
		wsConnectionsActive.Dec()
		slog.Info("client disconnected", "anonymousID", c.params.AnonymousID)
	}
}

// LookupByID returns the live *Client for anonymousID, or nil if not connected.
func (h *Hub) LookupByID(anonymousID string) *Client {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.byID[anonymousID]
}

// VerifyToken returns true if the given anonymousID has an active connection
// whose connToken matches token (H-3/H-4 vault auth).
func (h *Hub) VerifyToken(anonymousID, token string) bool {
	h.mu.RLock()
	c := h.byID[anonymousID]
	h.mu.RUnlock()
	return c != nil && c.connToken == token
}
