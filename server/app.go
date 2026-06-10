package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:   1024,
	WriteBufferSize:  1024,
	HandshakeTimeout: 10 * time.Second,
	// Allow all origins: mobile clients connect from varied IP ranges.
	CheckOrigin: func(r *http.Request) bool { return true },
}

// App wires together the Hub and MatchingQueue and implements http.Handler.
type App struct {
	hub   *Hub
	queue *MatchingQueue
}

func NewApp() *App {
	return &App{
		hub:   NewHub(),
		queue: NewMatchingQueue(),
	}
}

// Start launches the background matching loop. Call once before serving.
func (a *App) Start() {
	go a.queue.RunMatchingLoop()
}

func (a *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/ws":
		a.handleWS(w, r)
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
	a.queue.Enqueue(c)

	go c.writePump()
	go c.readPump(a.queue)
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
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}
