package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/raaz/server/store"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:   1024,
	WriteBufferSize:  1024,
	HandshakeTimeout: 10 * time.Second,
	// Allow all origins: mobile clients connect from varied IP ranges.
	CheckOrigin: func(r *http.Request) bool { return true },
}

// App wires together the Hub and backing stores and implements http.Handler.
type App struct {
	hub      *Hub
	queue    store.QueueStore
	sessions store.SessionStore
	vault    store.VaultStore
}

func NewApp() *App {
	return &App{
		hub:      NewHub(),
		queue:    store.NewMemoryQueueStore(),
		sessions: store.NewMemorySessionStore(),
		vault:    nil,
	}
}

// Start launches the background matching loop. Call once before serving.
func (a *App) Start() {
	go a.runMatchingLoop()
}

func (a *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/ws":
		a.handleWS(w, r)
	case "/vault/messages":
		a.handleVaultMessages(w, r)
	case "/health":
		w.WriteHeader(http.StatusOK)
	default:
		http.NotFound(w, r)
	}
}

func (a *App) handleWS(w http.ResponseWriter, r *http.Request) {
	params := registrationFromQuery(r)
	if params.AnonymousID == "" || params.PromptID == "" {
		http.Error(w, "missing required params: anonymousId, promptId", http.StatusBadRequest)
		return
	}

	banned, err := globalStrikes.IsBanned(params.AnonymousID)
	if err != nil {
		log.Printf("IsBanned error: %v", err)
		// fail open — don't block user on transient store errors
	} else if banned {
		http.Error(w, "temporarily banned due to repeated violations", http.StatusForbidden)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("upgrade error: %v", err)
		return
	}

	c := &Client{
		conn:   conn,
		send:   make(chan []byte, sendBufferSize),
		recv:   make(chan []byte, recvBufferSize),
		params: params,
		hub:    a.hub,
	}

	a.hub.Register(c)
	if err := a.queue.Enqueue(store.WaitingEntry{
		AnonymousID: params.AnonymousID,
		PromptID:    params.PromptID,
		AgeBucket:   params.AgeBucket,
		City:        params.City,
		JoinedAt:    time.Now(),
	}); err != nil {
		log.Printf("enqueue error: %v", err)
	}

	go c.writePump()
	go c.readPump(a.queue)
}

func (a *App) handleVaultMessages(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if a.vault == nil {
		http.Error(w, "vault not configured", http.StatusServiceUnavailable)
		return
	}

	var req struct {
		AnonID      string `json:"anonId"`
		MessageID   string `json:"messageId"`
		SessionID   string `json:"sessionId"`
		Prompt      string `json:"prompt"`
		Text        string `json:"text"`
		SenderAlias string `json:"senderAlias"`
		SavedAt     int64  `json:"savedAt"` // unix milliseconds
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if req.AnonID == "" || req.MessageID == "" {
		http.Error(w, "anonId and messageId required", http.StatusBadRequest)
		return
	}

	// Hash anonId before storage — plaintext IDs never touch the DB.
	h := sha256.Sum256([]byte(req.AnonID))
	anonIDHash := fmt.Sprintf("%x", h)

	msg := store.VaultMessage{
		AnonIDHash:  anonIDHash,
		MessageID:   req.MessageID,
		SessionID:   req.SessionID,
		Prompt:      req.Prompt,
		Text:        req.Text,
		SenderAlias: req.SenderAlias,
		SavedAt:     time.UnixMilli(req.SavedAt),
	}

	if err := a.vault.SaveMessage(msg); err != nil {
		if strings.Contains(err.Error(), "unique_violation") ||
			strings.Contains(err.Error(), "UNIQUE constraint") {
			w.WriteHeader(http.StatusConflict)
			return
		}
		log.Printf("vault save: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func registrationFromQuery(r *http.Request) RegistrationParams {
	q := r.URL.Query()
	return RegistrationParams{
		PromptID:    q.Get("promptId"),
		AgeBucket:   q.Get("ageBucket"),
		City:        q.Get("city"),
		AnonymousID: q.Get("anonymousId"),
	}
}

// corsMiddleware adds permissive CORS headers for mobile/web clients.
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}
