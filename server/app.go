package main

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/raaz/server/store"
)

// upgrader enforces Origin restriction for web clients; mobile (no Origin) always allowed (M-4).
var upgrader = websocket.Upgrader{
	ReadBufferSize:   1024,
	WriteBufferSize:  1024,
	HandshakeTimeout: 10 * time.Second,
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		if origin == "" {
			return true // native mobile clients never send Origin
		}
		return origin == "https://raaz.app"
	},
}

// validAgeBuckets is the accepted set of age brackets (M-10).
var validAgeBuckets = map[string]bool{
	"18-22": true, "23-28": true, "29-35": true, "36+": true,
}

// App wires together the Hub and backing stores and implements http.Handler.
type App struct {
	hub          *Hub
	queue        store.QueueStore
	sessions     store.SessionStore
	vault        store.VaultStore
	matchHistory store.MatchHistoryStore
	prompts      store.PromptStore
	freemium     store.FreemiumStore
	limiter      *ipRateLimiter
}

// ipRateLimiter enforces per-IP WebSocket connection rate limiting (H-6).
type ipRateLimiter struct {
	mu      sync.Mutex
	counts  map[string]int
	windows map[string]time.Time
}

func newIPRateLimiter() *ipRateLimiter {
	rl := &ipRateLimiter{
		counts:  make(map[string]int),
		windows: make(map[string]time.Time),
	}
	go rl.cleanup()
	return rl
}

func (l *ipRateLimiter) cleanup() {
	for range time.Tick(5 * time.Minute) {
		l.mu.Lock()
		cutoff := time.Now().Add(-time.Minute)
		for ip, w := range l.windows {
			if w.Before(cutoff) {
				delete(l.counts, ip)
				delete(l.windows, ip)
			}
		}
		l.mu.Unlock()
	}
}

func (l *ipRateLimiter) Allow(ip string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := time.Now()
	w, ok := l.windows[ip]
	if !ok || now.Sub(w) >= time.Minute {
		l.windows[ip] = now
		l.counts[ip] = 1
		return true
	}
	l.counts[ip]++
	return l.counts[ip] <= 10
}

func NewApp() *App {
	return &App{
		hub:          NewHub(),
		queue:        store.NewMemoryQueueStore(),
		sessions:     store.NewMemorySessionStore(),
		vault:        nil,
		matchHistory: store.NewMemoryMatchHistoryStore(),
		prompts:      store.NewMemoryPromptStore(),
		freemium:     store.NewMemoryFreemiumStore(),
		limiter:      newIPRateLimiter(),
	}
}

// Start launches background jobs. Call once before serving.
func (a *App) Start() {
	go a.runMatchingLoop()
	if a.vault != nil {
		go a.runVaultCleanup()
	}
}

// runVaultCleanup deletes vault messages older than 48 h every 6 hours (H-7).
func (a *App) runVaultCleanup() {
	for range time.Tick(6 * time.Hour) {
		cutoff := time.Now().Add(-48 * time.Hour)
		if err := a.vault.DeleteBefore(cutoff); err != nil {
			slog.Warn("vault cleanup error", "err", err)
		} else {
			slog.Info("vault cleanup complete", "cutoff", cutoff.Format(time.RFC3339))
		}
	}
}

func (a *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/ws":
		a.handleWS(w, r)
	case "/vault/messages":
		a.handleVaultMessages(w, r)
	case "/prompts/today":
		a.handlePromptToday(w, r)
	case "/metrics":
		a.handleMetrics(w, r)
	case "/health":
		w.WriteHeader(http.StatusOK)
	default:
		http.NotFound(w, r)
	}
}

func (a *App) handleWS(w http.ResponseWriter, r *http.Request) {
	// H-6: per-IP rate limit — max 10 WS connections per minute
	if !a.limiter.Allow(realIP(r)) {
		http.Error(w, "too many connections", http.StatusTooManyRequests)
		return
	}

	params := registrationFromQuery(r)
	if params.AnonymousID == "" || params.PromptID == "" {
		http.Error(w, "missing required params: anonymousId, promptId", http.StatusBadRequest)
		return
	}

	// M-10: input validation
	if !validAgeBuckets[params.AgeBucket] {
		params.AgeBucket = "18-22"
	}
	if len(params.City) > 100 {
		params.City = params.City[:100]
	}
	if len(params.PromptID) > 64 {
		http.Error(w, "invalid promptId", http.StatusBadRequest)
		return
	}

	// H-10: freemium daily conversation limit
	anonHash := sha256Hex(params.AnonymousID)
	if a.freemium != nil {
		allowed, remaining, err := a.freemium.CheckAndIncrementDaily(anonHash)
		if err != nil {
			slog.Warn("freemium check error", "err", err)
		} else if !allowed {
			w.Header().Set("X-Conversations-Remaining", "0")
			http.Error(w, "daily conversation limit reached", http.StatusTooManyRequests)
			return
		} else {
			w.Header().Set("X-Conversations-Remaining", fmt.Sprintf("%d", remaining))
		}
	}

	banned, err := globalStrikes.IsBanned(params.AnonymousID)
	if err != nil {
		slog.Warn("IsBanned check error", "anonymousID", params.AnonymousID, "err", err)
	} else if banned {
		http.Error(w, "temporarily banned due to repeated violations", http.StatusForbidden)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("WebSocket upgrade error", "err", err)
		return
	}

	// H-3: server-generated connToken for vault auth
	connToken, err := generateToken()
	if err != nil {
		slog.Error("connToken generation error", "err", err)
		conn.Close()
		return
	}

	c := &Client{
		conn:      conn,
		send:      make(chan []byte, sendBufferSize),
		recv:      make(chan []byte, recvBufferSize),
		params:    params,
		connToken: connToken,
		hub:       a.hub,
	}

	// M-6: start writePump before registering so REGISTERED event is deliverable
	go c.writePump()

	// Send REGISTERED immediately — client must cache connToken for vault auth.
	data, _ := marshalEnvelope(EventRegistered, RegisteredPayload{ConnToken: connToken})
	c.safeSend(data)

	a.hub.Register(c)
	if err := a.queue.Enqueue(store.WaitingEntry{
		AnonymousID: params.AnonymousID,
		PromptID:    params.PromptID,
		AgeBucket:   params.AgeBucket,
		City:        params.City,
		JoinedAt:    time.Now(),
	}); err != nil {
		slog.Error("enqueue error", "anonymousID", params.AnonymousID, "err", err)
	}

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

	// H-4: require valid connToken matching an active WebSocket connection
	anonID := r.Header.Get("X-Anon-ID")
	connToken := r.Header.Get("X-Conn-Token")
	if anonID == "" || connToken == "" {
		http.Error(w, "X-Anon-ID and X-Conn-Token headers required", http.StatusUnauthorized)
		return
	}
	if !a.hub.VerifyToken(anonID, connToken) {
		http.Error(w, "invalid or expired token", http.StatusUnauthorized)
		return
	}

	var req struct {
		MessageID   string `json:"messageId"`
		SessionID   string `json:"sessionId"`
		Prompt      string `json:"prompt"`
		Text        string `json:"text"`
		SenderAlias string `json:"senderAlias"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if req.MessageID == "" {
		http.Error(w, "messageId required", http.StatusBadRequest)
		return
	}

	// M-7: SavedAt is always server-set — client cannot influence storage timestamp.
	msg := store.VaultMessage{
		AnonIDHash:  sha256Hex(anonID),
		MessageID:   req.MessageID,
		SessionID:   req.SessionID,
		Prompt:      req.Prompt,
		Text:        req.Text,
		SenderAlias: req.SenderAlias,
		SavedAt:     time.Now(),
	}

	if err := a.vault.SaveMessage(msg); err != nil {
		// M-8: check pgx error code for unique violation (not string matching)
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			w.WriteHeader(http.StatusConflict)
			return
		}
		slog.Error("vault save error", "messageID", req.MessageID, "err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	// L-9: vault audit log (hash prefix only — never log plaintext IDs)
	slog.Info("vault save", "anonIDHash", msg.AnonIDHash[:8]+"...", "messageID", req.MessageID, "sessionID", req.SessionID)
	w.WriteHeader(http.StatusCreated)
}

// handlePromptToday serves the daily conversation prompt (M-9).
func (a *App) handlePromptToday(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	prompt, err := a.prompts.GetTodayPrompt()
	if err != nil || prompt == nil {
		http.Error(w, "no prompt available", http.StatusServiceUnavailable)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(prompt) //nolint:errcheck
}

// handleMetrics serves Prometheus metrics, gated by METRICS_SECRET if set (L-5).
func (a *App) handleMetrics(w http.ResponseWriter, r *http.Request) {
	secret := os.Getenv("METRICS_SECRET")
	if secret != "" && r.Header.Get("Authorization") != "Bearer "+secret {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	promhttp.Handler().ServeHTTP(w, r)
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

// corsMiddleware allows only https://raaz.app as web origin; mobile (no Origin) always passes (M-4).
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin == "https://raaz.app" {
			w.Header().Set("Access-Control-Allow-Origin", "https://raaz.app")
			w.Header().Set("Vary", "Origin")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Anon-ID, X-Conn-Token, Authorization")
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func realIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return strings.SplitN(xff, ",", 2)[0]
	}
	ip, _, _ := net.SplitHostPort(r.RemoteAddr)
	return ip
}

func sha256Hex(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}

// generateToken returns a 32-byte cryptographically random hex string for use as connToken.
func generateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
