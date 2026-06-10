package main

import (
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
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
	conn   *websocket.Conn
	send   chan []byte // writePump drains this; never closed explicitly
	recv   chan []byte // readPump fills this; closed when readPump exits
	alias  string
	params RegistrationParams
	hub    *Hub

	closeRecvOnce sync.Once
}

// safeSend writes data to c.send without panicking if the channel is closed,
// and drops the message silently if the buffer is full.
func (c *Client) safeSend(data []byte) {
	defer func() { recover() }() //nolint:errcheck
	select {
	case c.send <- data:
	default:
		log.Printf("send buffer full for %s, dropping message", c.alias)
	}
}

// writePump is the sole goroutine that writes to the WebSocket connection.
// It exits when the underlying connection is closed (write returns an error)
// or when a ping fails.
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
// It writes every inbound message to c.recv for the session relay to consume.
// On exit it deregisters the client, closes c.recv, and closes the connection.
func (c *Client) readPump(q *MatchingQueue) {
	defer func() {
		c.hub.Unregister(c)
		q.Remove(c)
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
				log.Printf("read error (%s): %v", c.params.AnonymousID, err)
			}
			return
		}
		// Non-blocking: if recv is full, drop the message rather than stalling the pump.
		select {
		case c.recv <- msg:
		default:
			log.Printf("recv buffer full for %s, dropping inbound message", c.params.AnonymousID)
		}
	}
}

// Hub tracks all live connections for clean-up on server shutdown.
type Hub struct {
	mu      sync.RWMutex
	clients map[*Client]struct{}
}

func NewHub() *Hub {
	return &Hub{clients: make(map[*Client]struct{})}
}

func (h *Hub) Register(c *Client) {
	h.mu.Lock()
	h.clients[c] = struct{}{}
	h.mu.Unlock()
	log.Printf("connected: %s", c.params.AnonymousID)
}

// Unregister removes the client from the hub. It does NOT close c.send —
// writePump exits naturally when the connection is closed.
func (h *Hub) Unregister(c *Client) {
	h.mu.Lock()
	delete(h.clients, c)
	h.mu.Unlock()
	log.Printf("disconnected: %s", c.params.AnonymousID)
}
